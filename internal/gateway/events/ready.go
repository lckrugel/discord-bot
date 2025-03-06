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
	var payload ReadyPayload
	err := json.Unmarshal(msg, &payload)
	if err != nil {
		errMsg := "error decoding Ready payload: " + err.Error()
		return nil, errors.New(errMsg)
	}
	e.Data = payload
	return e, nil
}
