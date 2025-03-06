package events

import (
	"encoding/json"
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

func (e *ResumedEvent) DecodeData(data []byte) (ReceivableEvent, error) {
	var event ResumedEvent
	err := json.Unmarshal(data, &event)
	if err != nil {
		errMsg := "error decoding Resumed event: " + err.Error()
		return nil, errors.New(errMsg)
	}

	if event.Operation != Reconnect {
		errMsg := "unexpected event received: expected Resumed, got " + event.Operation.String()
		return nil, errors.New(errMsg)
	}

	e = &event
	return e, nil
}
