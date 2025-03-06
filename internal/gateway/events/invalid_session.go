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
	var resumable bool
	err := json.Unmarshal(data, &resumable)
	if err != nil {
		errMsg := "error decoding InvalidSession event: " + err.Error()
		return nil, errors.New(errMsg)
	}
	e.Resumable = resumable
	return e, nil
}
