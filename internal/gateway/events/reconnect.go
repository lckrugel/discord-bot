package events

import (
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

func (e *ReconnectEvent) DecodeData(gen_event Event) error {
	if gen_event.Operation != Reconnect {
		errMsg := "unexpected event received: expected Reconnect, got " + gen_event.Operation.String()
		return errors.New(errMsg)
	}

	e.Event = gen_event
	return nil
}
