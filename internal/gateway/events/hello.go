package events

import (
	"encoding/json"
	"errors"
)

type HelloEvent struct {
	Event
	Heartbeat_Interval float64
}

func NewHelloEvent() *HelloEvent {
	return &HelloEvent{
		Event: Event{
			Operation: Hello,
		},
	}
}

func (e *HelloEvent) DecodeData(gen_event Event) error {
	if gen_event.Operation != Hello {
		errMsg := "unexpected event received: expected Hello, got " + gen_event.Operation.String()
		return errors.New(errMsg)
	}
	var payload struct {
		Heartbeat_Interval float64 `json:"d"`
	}
	err := json.Unmarshal(gen_event.RawData, &payload)
	if err != nil {
		errMsg := "error decoding Hello event: " + err.Error()
		return errors.New(errMsg)
	}

	e.Event = gen_event
	e.Heartbeat_Interval = payload.Heartbeat_Interval
	return nil
}
