package events

import (
	"encoding/json"
	"errors"
)

type InvalidSessionEvent struct {
	Event
	Resumable bool
}

func NewInvalidSessionEvent(resumable bool) InvalidSessionEvent {
	return InvalidSessionEvent{
		Event: Event{
			Operation: Invalid_Session,
		},
	}
}

func (e *InvalidSessionEvent) DecodeData(gen_event Event) error {
	if gen_event.Operation != Invalid_Session {
		errMsg := "unexpected event received: expected Invalid_Session, got " + gen_event.Operation.String()
		return errors.New(errMsg)
	}

	var payload struct {
		Resumable bool `json:"d"`
	}
	err := json.Unmarshal(gen_event.RawData, &payload)
	if err != nil {
		errMsg := "error decoding InvalidSession event: " + err.Error()
		return errors.New(errMsg)
	}

	e.Event = gen_event
	e.Resumable = payload.Resumable
	return nil
}
