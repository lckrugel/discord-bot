package events

import (
	"encoding/json"
	"errors"
)

type ReconnectEvent struct {
	Event
}

func NewReconnectEvent() ReconnectEvent {
	return ReconnectEvent{
		Event: Event{
			Operation: Reconnect,
		},
	}
}

func (e *ReconnectEvent) DecodeData(data []byte) (ReceivableEvent, error) {
	var event ReconnectEvent
	err := json.Unmarshal(data, &event)
	if err != nil {
		errMsg := "error decoding Reconnect event: " + err.Error()
		return nil, errors.New(errMsg)
	}

	if event.Operation != Reconnect {
		errMsg := "unexpected event received: expected Reconnect, got " + event.Operation.String()
		return nil, errors.New(errMsg)
	}

	e = &event
	return e, nil
}
