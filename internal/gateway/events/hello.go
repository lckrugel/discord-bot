package events

import (
	"encoding/json"
	"errors"
)

type HelloEvent struct {
	Event
	Data HelloPayload `json:"d"`
}

type HelloPayload struct {
	Interval float64 `json:"heartbeat_interval"`
}

func NewHelloEvent() *HelloEvent {
	return &HelloEvent{
		Event: Event{
			Operation: Hello,
			Sequence:  nil,
			Type:      nil,
		},
		Data: HelloPayload{},
	}
}

func (e *HelloEvent) DecodeData(msg []byte) (ReceivableEvent, error) {
	var payload HelloPayload
	err := json.Unmarshal(msg, &payload)
	if err != nil {
		errMsg := "error decoding Hello event: " + err.Error()
		return nil, errors.New(errMsg)
	}
	e.Data = payload
	return e, nil
}
