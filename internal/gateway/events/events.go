package events

import (
	"encoding/json"
	"errors"

	"github.com/gorilla/websocket"
)

type OpCode int

const (
	Dispatch                  OpCode = iota
	Heartbeat                        = 1
	Identify                         = 2
	Presence_Update                  = 3
	Voice_State_Update               = 4
	Resume                           = 6
	Reconnect                        = 7
	Request_Guild_Members            = 8
	Invalid_Session                  = 9
	Hello                            = 10
	Heartbeat_ACK                    = 11
	Request_Soundboard_Sounds        = 31
)

var operationName = map[OpCode]string{
	Dispatch:                  "dispatch",
	Heartbeat:                 "heartbeat",
	Identify:                  "identify",
	Presence_Update:           "presence_update",
	Voice_State_Update:        "voice_state_update",
	Resume:                    "resume",
	Reconnect:                 "reconect",
	Request_Guild_Members:     "request_guild_members",
	Invalid_Session:           "invalid_session",
	Hello:                     "hello",
	Heartbeat_ACK:             "heartbeat_ack",
	Request_Soundboard_Sounds: "request_soundboard_sounds",
}

func (op OpCode) String() string {
	return operationName[op]
}

// An event that can be sent over the websocket
type SendableEvent interface {
	// Serializes the event to a json byte slice to be sent over the websocket
	PrepareToSend() ([]byte, error)
}

// An event that can be received over the websocket
type ReceivableEvent interface {
	// Decodes the event data from a json byte slice received over the websocket
	DecodeData(gen_event Event) error
}

// An generic event that is sent or received over the websocket
type Event struct {
	Operation OpCode          `json:"op"`
	Sequence  *int            `json:"s"`
	Type      *string         `json:"t"`
	RawData   json.RawMessage `json:"d"`
}

func NewEvent(msg []byte) (*Event, error) {
	var event Event
	err := json.Unmarshal(msg, &event)
	if err != nil {
		errMsg := "error unmarshaling event: " + err.Error()
		return &Event{}, errors.New(errMsg)
	}
	return &event, nil
}

func SendEvent(conn *websocket.Conn, e SendableEvent) error {
	msg, err := e.PrepareToSend()
	if err != nil {
		return err
	}
	err = conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		errMsg := "error sending event: " + err.Error()
		return errors.New(errMsg)
	}
	return nil
}
