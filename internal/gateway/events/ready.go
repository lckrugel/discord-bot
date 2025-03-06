package events

import (
	"encoding/json"
	"errors"
)

type ReadyEvent struct {
	Event
	Data ReadyPayload `json:"d"`
}

type ReadyPayload struct {
	Api_version int    `json:"api_version"`
	Session_id  string `json:"session_id"`
	Resume_url  string `json:"resume_gateway_url"`
	// TODO: User
	// TODO: UnavailableGuilds
	// TODO: Shards
}

func NewReadyEvent() *ReadyEvent {
	typeReady := "Ready"
	return &ReadyEvent{
		Event: Event{
			Operation: Dispatch,
			Type:      &typeReady,
		},
	}
}

func (e *ReadyEvent) DecodeData(msg []byte) (ReceivableEvent, error) {
	var event ReadyEvent
	err := json.Unmarshal(msg, &event)
	if err != nil {
		errMsg := "error decoding Ready event: " + err.Error()
		return nil, errors.New(errMsg)
	}

	if event.Operation != Dispatch {
		errMsg := "unexpected event received: expected Dispatch, got " + event.Operation.String()
		return nil, errors.New(errMsg)
	}

	if *event.Type != "Ready" {
		errMsg := "unexpected event type received: expected Ready, got " + *event.Type
		return nil, errors.New(errMsg)
	}

	e = &event
	return e, nil
}
