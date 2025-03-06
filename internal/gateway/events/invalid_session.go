package events

import (
	"encoding/json"
	"errors"
)

type InvalidSessionEvent struct {
	Event
	Resumable bool `json:"d"`
}

func NewInvalidSessionEvent(resumable bool) InvalidSessionEvent {
	return InvalidSessionEvent{
		Event: Event{
			Operation: Invalid_Session,
		},
	}
}

func (e *InvalidSessionEvent) DecodeData(data []byte) (ReceivableEvent, error) {
	var event InvalidSessionEvent
	err := json.Unmarshal(data, &event)
	if err != nil {
		errMsg := "error decoding InvalidSession event: " + err.Error()
		return nil, errors.New(errMsg)
	}

	if event.Operation != Invalid_Session {
		errMsg := "unexpected event received: expected Invalid_Session, got " + event.Operation.String()
		return nil, errors.New(errMsg)
	}

	e = &event
	return e, nil
}
