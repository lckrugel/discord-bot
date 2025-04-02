package events

import (
	"encoding/json"
	"errors"
)

type ReadyEvent struct {
	Event
	Data ReadyPayload
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

func (e *ReadyEvent) DecodeData(gen_event Event) error {
	if gen_event.Operation != Dispatch {
		errMsg := "unexpected event received: expected Dispatch, got " + gen_event.Operation.String()
		return errors.New(errMsg)
	}

	if *gen_event.Type != "READY" {
		errMsg := "unexpected event type received: expected READY, got " + *gen_event.Type
		return errors.New(errMsg)
	}

	var payload ReadyPayload
	err := json.Unmarshal(gen_event.RawData, &payload)
	if err != nil {
		errMsg := "error decoding Ready event: " + err.Error()
		return errors.New(errMsg)
	}

	e.Event = gen_event
	e.Data = payload
	return nil
}
