package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lckrugel/discord-bot/bot"
)

type reconnectParams struct {
	reconnect_url string
	session_id    string
	bot           bot.Bot
	last_sequence int
}

/* Conecta o bot ao gateway do Discord e ouve por eventos */
func ConnectToGateway(bot bot.Bot) error {
	// Descobre a URL do websocket
	wssURL, err := getWebsocketURL(bot.GetSecretKey())
	if err != nil {
		errMsg := fmt.Sprint("error getting websocket url: ", err)
		return errors.New(errMsg)
	}

	conn, heartbeat_interval, reconnectParams, err := handshake(wssURL, bot)
	if err != nil {
		errMsg := fmt.Sprint("failed the initial handshake: ", err)
		return errors.New(errMsg)
	}

	gatewayEvents := make(chan GatewayEventPayload, 5)

	go eventListener(conn, gatewayEvents) // Começa a ouvir por eventos

	for {
		last_seq, err := handleHeartbeat(conn, heartbeat_interval, gatewayEvents)
		if err != nil {
			if err.Error() == "requires reconnect" {
				reconnectParams.last_sequence = *last_seq
				conn, err = resumeConnection(reconnectParams)
				if err != nil {
					errMsg := fmt.Sprint("failed to reconnect: ", err)
					return errors.New(errMsg)
				}
				continue
			}
			return err
		}
	}
}

/* Recebe a URL de websockets que deve ser usada na conexão */
func getWebsocketURL(api_key string) (string, error) {
	// Forma o request
	req, err := http.NewRequest("GET", "https://discord.com/api/v9/gateway/bot", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bot "+api_key)

	// Envia o request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	// Lê a resposta
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New(string(bodyBytes))
	}

	// Transforma o corpo em um map
	var bodyMap map[string]any
	err = json.Unmarshal(bodyBytes, &bodyMap)
	if err != nil {
		return "", err
	}

	// Busca no corpo a URL
	url, ok := bodyMap["url"].(string)
	if !ok {
		return "", errors.New("invalid url")
	}
	return url, nil
}

func handshake(wssURL string, bot bot.Bot) (conn *websocket.Conn, interval float64, reconnectInfo *reconnectParams, err error) {
	// Estabelece a conexão com o Gateway
	conn, resp, err := websocket.DefaultDialer.Dial(wssURL, nil)
	if err != nil {
		errMsg := fmt.Sprint("error establishing connection to gateway: ", err)
		return nil, 0, nil, errors.New(errMsg)
	}
	defer conn.Close()

	// Espera-se que ocorra a troca de HTTP -> WSS
	if resp.StatusCode != 101 {
		errMsg := fmt.Sprint("failed to switch protocols with status: ", resp.StatusCode)
		return nil, 0, nil, errors.New(errMsg)
	}

	helloPayload := <-gatewayEvents
	if helloPayload.Operation != Hello {
		return nil, 0, nil, errors.New("didn't receive Hello event")
	}
	helloData, err := helloPayload.GetPayloadData()
	if err != nil {
		errMsg := fmt.Sprint("error getting Hello payload: ", err)
		return nil, 0, nil, errors.New(errMsg)
	}

	// Envia Identify terminando o 'handshake'
	err = sendIdentify(conn, bot)
	if err != nil {
		errMsg := fmt.Sprint("error sending Identify event: ", err)
		return nil, 0, nil, errors.New(errMsg)
	}

	// Recebe o Ready com informações sobre como reconectar
	readyPayload := <-gatewayEvents
	if readyPayload.Operation != Dispatch || *readyPayload.Type != "Ready" {
		return nil, 0, nil, errors.New("didn't receive Ready event")
	}
	session_id, reconnect_url, err := extractReadyData(readyPayload)
	if err != nil {
		errMsg := fmt.Sprint("could not extract data from Ready event: ", err)
		return nil, 0, nil, errors.New(errMsg)
	}

	reconnectInfo = new(reconnectParams)
	reconnectInfo.bot = bot
	reconnectInfo.session_id = session_id
	reconnectInfo.reconnect_url = reconnect_url

	// Usando o heartbeat interval recebido no hello inicia a troca de heartbeats
	interval = helloData["heartbeat_interval"].(float64)

	return conn, interval, reconnectInfo, nil
}

/* Escuta por eventos na conexão e os envia no canal */
func eventListener(conn *websocket.Conn, ch chan<- GatewayEventPayload) error {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		responsePayload, err := unmarshalPayload(msg)
		if err != nil {
			return err
		}

		ch <- responsePayload
	}
}

func extractReadyData(readyPayload GatewayEventPayload) (session_id string, reconnect_url string, err error) {
	readyData, err := readyPayload.GetPayloadData()
	if err != nil {
		errMsg := fmt.Sprint("error getting Ready payload: ", err)
		return "", "", errors.New(errMsg)
	}
	session_id, exists := readyData["session_id"].(string)
	if !exists {
		return "", "", errors.New("session_id not set on Ready payload")
	}
	reconnect_url, exists = readyData["reconnect_url"].(string)
	if !exists {
		return "", "", errors.New("reconnect_url not set on Ready payload")
	}
	return session_id, reconnect_url, nil
}

/* Lida com o envio periódico dos "heartbeats" para manter a conexão */
func handleHeartbeat(conn *websocket.Conn, interval float64, ch <-chan GatewayEventPayload) (last_seq *int, err error) {
	// Envia o primeiro heartbeat
	log.Println("start sending heartbeats...")
	jitter := rand.Float64() // Intervalo aleatorio antes de começar a enviar heartbeat
	intervalDuration := time.Duration(time.Millisecond * time.Duration(interval))
	time.Sleep(time.Duration(intervalDuration.Milliseconds() * int64(jitter)))
	last_seq = nil
	lastHeartbeartSentAt, err := sendHeartbeat(conn, last_seq)
	if err != nil {
		errMsg := fmt.Sprint("failed to send heartbeat: ", err)
		return last_seq, errors.New(errMsg)
	}

	// Loop de envio de heartbeats
	for lastEvent := range ch {
		if time.Since(lastHeartbeartSentAt) > intervalDuration || lastEvent.Operation == Reconect {
			return last_seq, errors.New("requires reconnect")
		}
		last_seq = lastEvent.Sequence

		switch lastEvent.Operation {
		case Heartbeat_ACK:
			log.Print("received heartbeat ack")
			time.Sleep(intervalDuration)
			lastHeartbeartSentAt, err = sendHeartbeat(conn, last_seq)
			if err != nil {
				errMsg := fmt.Sprint("failed to send heartbeat: ", err)
				return last_seq, errors.New(errMsg)
			}

		case Heartbeat:
			log.Print("received heartbeat")
			lastHeartbeartSentAt, err = sendHeartbeat(conn, last_seq)
			if err != nil {
				errMsg := fmt.Sprint("failed to send heartbeat: ", err)
				return last_seq, errors.New(errMsg)
			}
		}
	}
	return last_seq, nil
}

/* Envia um evento do tipo Heartbeat */
func sendHeartbeat(conn *websocket.Conn, seq *int) (time.Time, error) {
	heartbeatPayload, err := CreateHeartbeatPayload(seq)
	if err != nil {
		return time.Now(), err
	}
	log.Println("sending heartbeat")
	conn.WriteMessage(websocket.TextMessage, heartbeatPayload)
	return time.Now(), nil
}

/* Tenta fazer a reconexão */
func resumeConnection(params reconnectParams) (*websocket.Conn, error) {
	// Tenta a reconectar usando a URL de reconexão
	conn, resp, err := websocket.DefaultDialer.Dial(params.reconnect_url, nil)
	if err != nil {
		errMsg := fmt.Sprint("error restablishing connection to gateway: ", err)
		return nil, errors.New(errMsg)
	}

	// Espera-se que ocorra a troca de HTTP -> WSS
	if resp.StatusCode != 101 {
		errMsg := fmt.Sprint("failed to switch protocols with status: ", resp.StatusCode)
		return nil, errors.New(errMsg)
	}

	sendResume(conn, params)
	return conn, nil
}

/* Envia um evento do tipo Identify */
func sendIdentify(conn *websocket.Conn, bot bot.Bot) error {
	identifyPayload, err := CreateIdentifyPayload(bot)
	if err != nil {
		return err
	}
	conn.WriteMessage(websocket.TextMessage, identifyPayload)
	return nil
}

/* Envia um evento do tipo Resume */
func sendResume(conn *websocket.Conn, params reconnectParams) error {
	resumePayload, err := CreateResumePayload(params.bot, params.session_id, params.last_sequence)
	if err != nil {
		return err
	}
	conn.WriteMessage(websocket.TextMessage, resumePayload)
	return nil
}
