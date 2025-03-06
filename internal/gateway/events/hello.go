package events

import (
	"encoding/json"
	"errors"
)

type HelloEvent struct {
	Event
	Heartbeat_Interval float64 `json:"d"`
}

func NewHelloEvent() *HelloEvent {
	return &HelloEvent{
		Event: Event{
			Operation: Hello,
		},
	}
}

func (e *HelloEvent) DecodeData(msg []byte) (ReceivableEvent, error) {
	var interval float64
	err := json.Unmarshal(msg, &interval)
	if err != nil {
		errMsg := "error decoding Hello event: " + err.Error()
		return nil, errors.New(errMsg)
	}
	e.Heartbeat_Interval = interval
	return e, nil
}
