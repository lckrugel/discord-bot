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
	"github.com/lckrugel/discord-bot/internal/config"
	"github.com/lckrugel/discord-bot/internal/gateway/events"
)

type Client struct {
	token              string
	intents            uint64
	conn               *websocket.Conn
	last_sequence      *int
	events             chan events.Event
	stop_signal        chan struct{}
	heartbeat_interval int
	session_id         string
	reconnect_url      string
}

/* Establishes connection to the Discord gateway */
func (c *Client) Connect() error {
	// Get the websocket URL
	wssURL, err := getWebsocketURL(c.token)
	if err != nil {
		errMsg := fmt.Sprint("error getting websocket url: ", err)
		return errors.New(errMsg)
	}

	// Connect to the gateway
	conn, resp, err := websocket.DefaultDialer.Dial(wssURL, nil)
	if err != nil {
		errMsg := fmt.Sprint("error establishing connection to gateway: ", err)
		return errors.New(errMsg)
	}
	c.conn = conn

	conn.SetCloseHandler(func(code int, text string) error {
		handleCloseCode(code, text, *c)
		return nil
	})

	// Check if the connection was successfully upgraded to WSS
	if resp.StatusCode != 101 {
		errMsg := fmt.Sprint("failed to switch protocols with status: ", resp.StatusCode)
		return errors.New(errMsg)
	}

	// Create a channel to stop goroutines
	c.stop_signal = make(chan struct{}, 2)
	// Create a channel to receive events
	c.events = make(chan events.Event, 1)

	go listener(c) // Start listening for events

	event := <-c.events
	if event.Operation != events.Hello {
		return errors.New("didn't receive Hello event")
	}
	hello_event := events.NewHelloEvent()
	err = hello_event.DecodeData(event)
	if err != nil {
		return err
	}
	// Set the heartbeat interval from the Hello event
	c.heartbeat_interval = int(hello_event.Heartbeat_Interval)

	// Send Identify event finishing the handshake
	identify_event := events.NewIdentifyEvent(events.IdentifyPayload{
		Token:   c.token,
		Intents: c.intents,
	})
	err = events.SendEvent(*&c.conn, identify_event)
	if err != nil {
		errMsg := fmt.Sprint("error sending Identify event: ", err)
		return errors.New(errMsg)
	}

	// Expect to receive a Ready event
	event = <-c.events
	if event.Operation != events.Dispatch || *event.Type != "READY" {
		return errors.New("didn't receive Ready event")
	}
	ready_event := events.NewReadyEvent()
	err = ready_event.DecodeData(event)

	// Pick up the session_id and reconnect_url for future reconnections
	c.session_id = ready_event.Data.Session_id
	c.reconnect_url = ready_event.Data.Resume_url

	go handleHeartbeat(c) // Start heartbeat exchange

	return nil
}

func (c *Client) Reconnect() {
	log.Println("Attempting to reconnect...")

	// Para a execução das goroutines e fecha a conexão
	c.Disconnect()

	// Reinicia a conexão com o gateway usando a url fornecida
	conn, resp, err := websocket.DefaultDialer.Dial(c.reconnect_url, nil)
	if err != nil {
		log.Print("error restablishing connection to gateway: ", err)
		log.Print("restarting connection...")
		c.Connect()
		return
	}
	c.conn = conn

	conn.SetCloseHandler(func(code int, text string) error {
		handleCloseCode(code, text, *c)
		return nil
	})

	// Espera-se que ocorra a troca de HTTP -> WSS
	if resp.StatusCode != 101 {
		log.Print("failed to switch protocols with status: ", resp.StatusCode)
		log.Print("restarting connection...")
		c.Connect()
		return
	}

	// Recria o canal para os sinais de parada
	c.stop_signal = make(chan struct{}, 2)

	go listener(c) // Recomeça a ouvir os eventos
	SendResume(*c)

	// Checa se foi resumida com sucesso a conexão
	resumedPayload := <-c.events
	if *resumedPayload.Type != "RESUMED" {
		log.Print("failed to resume connection")
		log.Print("restarting connection...")
		c.Connect()
		return
	}

	go handleHeartbeat(c)
}

func (c *Client) Disconnect() {
	log.Println("Disconnecting...")

	c.sendStopSignal()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	close(c.events)
}

func (client *Client) sendStopSignal() {
	if client.stop_signal != nil {
		close(client.stop_signal)
		client.stop_signal = nil
	}
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		token:   cfg.GetSecretKey(),
		intents: cfg.GetIntents(),
	}
}

/* Get the websocket URL from the Discord API */
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

/* Lida com o envio periódico dos "heartbeats" para manter a conexão */
func handleHeartbeat(client *Client) error {
	// Envia o primeiro heartbeat
	log.Println("[heartbeat] start sending heartbeats...")
	jitter := rand.Float64() // Intervalo aleatorio antes de começar a enviar heartbeat
	intervalDuration := time.Duration(time.Millisecond * time.Duration(client.heartbeat_interval))

	time.Sleep(time.Duration(intervalDuration.Milliseconds() * int64(jitter)))
	client.last_sequence = nil
	lastHeartbeartSentAt, err := SendHeartbeat(*client)
	if err != nil {
		errMsg := fmt.Sprint("[heartbeat] failed to send heartbeat: ", err)
		return errors.New(errMsg)
	}

	// Loop de envio de heartbeats
	for lastEvent := range client.events {
		select {
		case <-client.stop_signal:
			return nil

		default:
			if time.Since(lastHeartbeartSentAt) > intervalDuration {
				client.Reconnect()
			}
			client.last_sequence = lastEvent.Sequence

			switch lastEvent.Operation {
			case Heartbeat_ACK:
				log.Print("[heartbeat] received heartbeat ack")
				time.Sleep(intervalDuration)
				lastHeartbeartSentAt, err = SendHeartbeat(*client)
				if err != nil {
					errMsg := fmt.Sprint("[heartbeat] failed to send heartbeat: ", err)
					return errors.New(errMsg)
				}

			case Heartbeat:
				log.Print("[heartbeat] received heartbeat")
				lastHeartbeartSentAt, err = SendHeartbeat(*client)
				if err != nil {
					errMsg := fmt.Sprint("[heartbeat] failed to send heartbeat: ", err)
					return errors.New(errMsg)
				}
			}
		}
	}
	return nil
}

func handleCloseCode(code int, text string, c Client) {
	log.Printf("webSocket closed with code: %d, reason: %s", code, text)
	// Se for possível, tenta a reconexão, se não cria uma nova conexão
	if code > 4010 {
		c.Connect()
	} else {
		c.Reconnect()
	}
}
