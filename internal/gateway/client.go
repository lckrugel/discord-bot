package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

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
	heartbeat_interval int
	session_id         string
	reconnect_url      string
	wg                 sync.WaitGroup
	reconnect_signal   chan struct{}
	shutdown_signal    chan struct{}
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
		handleCloseCode(code, text, c)
		return nil
	})

	// Check if the connection was successfully upgraded to WSS
	if resp.StatusCode != 101 {
		errMsg := fmt.Sprint("failed to switch protocols with status: ", resp.StatusCode)
		return errors.New(errMsg)
	}

	// Create a channel to receive events
	c.events = make(chan events.Event, 1)

	// Start goroutine to handle reconnections
	c.reconnect_signal = make(chan struct{}, 1)
	c.shutdown_signal = make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-c.reconnect_signal:
				log.Println("Received reconnect signal")
				err := c.Reconnect()
				if err != nil {
					log.Print("error reconnecting: ", err)
					c.RestartConnection()
				}
			case <-c.shutdown_signal:
				log.Println("Received shutdown signal")
				return
			}
		}
	}()

	c.wg.Add(1)
	// Start listening for events
	go func() {
		defer c.wg.Done()
		listener(c)
	}()

	event := <-c.events
	if event.Operation != events.Hello {
		return errors.New("didn't receive Hello event")
	}
	hello_event := events.NewHelloEvent()
	err = hello_event.DecodeData(event)
	if err != nil {
		errMsg := fmt.Sprint("error decoding Hello event: ", err)
		return errors.New(errMsg)
	}
	// Set the heartbeat interval from the Hello event
	c.heartbeat_interval = int(hello_event.Heartbeat_Interval)

	// Send Identify event finishing the handshake
	identify_event := events.NewIdentifyEvent(events.IdentifyPayload{
		Token:   c.token,
		Intents: c.intents,
		Properties: events.IdentifyProperties{
			Os:      "linux",
			Browser: "discord-bot",
			Device:  "discord-bot",
		},
	})
	err = events.SendEvent(c.conn, identify_event)
	if err != nil {
		errMsg := fmt.Sprint("error sending Identify event: ", err)
		return errors.New(errMsg)
	}

	// Expect to receive a Ready event
	event = <-c.events
	if event.Operation == events.Invalid_Session {
		return errors.New("invalid discord session")
	} else if event.Operation != events.Dispatch || *event.Type != "READY" {
		return errors.New("didn't receive Ready event")
	}

	ready_event := events.NewReadyEvent()
	err = ready_event.DecodeData(event)
	if err != nil {
		errMsg := fmt.Sprint("error decoding Ready event: ", err)
		return errors.New(errMsg)
	}

	// Pick up the session_id and reconnect_url for future reconnections
	c.session_id = ready_event.Data.Session_id
	c.reconnect_url = ready_event.Data.Resume_url

	// Start heartbeat exchange
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		handleHeartbeat(c)
	}()

	return nil
}

func (c *Client) Reconnect() error {
	log.Println("Attempting to reconnect...")

	// Forces the goroutines to stop and closes the connection
	c.Disconnect()

	// Restablishes the connection with the reconnect_url
	conn, resp, err := websocket.DefaultDialer.Dial(c.reconnect_url, nil)
	if err != nil {
		errMsg := fmt.Sprint("error restablishing connection to gateway: ", err)
		return errors.New(errMsg)
	}
	c.conn = conn

	conn.SetCloseHandler(func(code int, text string) error {
		handleCloseCode(code, text, c)
		return nil
	})

	// Check if the connection was successfully upgraded to WSS
	if resp.StatusCode != 101 {
		errMsg := fmt.Sprint("error while switching protocols with status: ", resp.StatusCode)
		return errors.New(errMsg)
	}

	// Start listening for events
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		listener(c)
	}()

	resumeEvent := events.NewResumeEvent(events.ResumePayload{
		Token:     c.token,
		SessionId: c.session_id,
		Sequence:  *c.last_sequence,
	})
	err = events.SendEvent(c.conn, resumeEvent)
	if err != nil {
		errMsg := fmt.Sprint("failed to send resume event: ", err)
		return errors.New(errMsg)
	}

	// Check if the connection was successfully resumed
	resumedPayload := <-c.events
	if *resumedPayload.Type != "RESUMED" {
		errMsg := fmt.Sprint("connection not resumed: ", err)
		return errors.New(errMsg)
	}

	// Start heartbeat exchange
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		handleHeartbeat(c)
	}()

	return nil
}

func (c *Client) Disconnect() {
	log.Println("Disconnecting...")

	if c.conn != nil {
		c.conn.Close()
	}
	close(c.events)
	c.wg.Wait()
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		token:   cfg.GetSecretKey(),
		intents: cfg.GetIntents(),
	}
}

func (c *Client) RestartConnection() {
	log.Println("Restarting connection...")
	close(c.shutdown_signal)
	c.Disconnect()
	err := c.Connect()
	if err != nil {
		log.Fatal("error restarting connection: ", err)
	}
}

func (c *Client) Shutdown() {
	log.Println("Shutting down...")
	c.Disconnect()
	close(c.shutdown_signal)
	log.Println("Shutdown complete.")
}

/* Get the websocket URL from the Discord API */
func getWebsocketURL(api_key string) (string, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/v9/gateway/bot", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bot "+api_key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New(string(bodyBytes))
	}

	var bodyMap map[string]any
	err = json.Unmarshal(bodyBytes, &bodyMap)
	if err != nil {
		return "", err
	}

	url, ok := bodyMap["url"].(string)
	if !ok {
		return "", errors.New("invalid url")
	}
	return url, nil
}

func handleCloseCode(code int, text string, c *Client) {
	log.Printf("webSocket closed with code: %d, reason: %s", code, text)
	// If possible, try to resume the connection
	if code > 4010 {
		c.RestartConnection()
	} else {
		c.reconnect_signal <- struct{}{}
	}
}
