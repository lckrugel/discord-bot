package events

import (
	"errors"
)

type ResumedEvent struct {
	Event
}

func NewResumedEvent() ResumedEvent {
	resumedType := "Resumed"
	return ResumedEvent{
		Event: Event{
			Operation: Dispatch,
			Type:      &resumedType,
		},
	}
}

func (e *ResumedEvent) DecodeData(gen_event Event) error {
	if gen_event.Operation != Reconnect {
		errMsg := "unexpected event received: expected Resumed, got " + gen_event.Operation.String()
		return errors.New(errMsg)
	}

	e.Event = gen_event
	return nil
}
