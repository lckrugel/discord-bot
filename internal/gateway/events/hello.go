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
	var event HelloEvent
	err := json.Unmarshal(msg, &event)
	if err != nil {
		errMsg := "error decoding Hello event: " + err.Error()
		return nil, errors.New(errMsg)
	}

	if event.Operation != Hello {
		errMsg := "unexpected event received: expected Hello, got " + event.Operation.String()
		return nil, errors.New(errMsg)
	}
	e = &event
	return e, nil
}
