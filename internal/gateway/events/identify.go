package events

import (
	"encoding/json"
	"errors"
)

type IdentifyEvent struct {
	Event
	Data IdentifyPayload `json:"d"`
}

type IdentifyPayload struct {
	Token      string             `json:"token"`
	Properties IdentifyProperties `json:"properties"`
	Intents    uint64             `json:"intents"`
}

type IdentifyProperties struct {
	Os      string `json:"os"`
	Browser string `json:"browser"`
	Device  string `json:"device"`
}

func NewIdentifyEvent(payload IdentifyPayload) *IdentifyEvent {
	return &IdentifyEvent{
		Event: Event{
			Operation: Identify,
			Sequence:  nil,
			Type:      nil,
		},
		Data: payload,
	}
}

func (e IdentifyEvent) PrepareToSend() ([]byte, error) {
	b, err := json.Marshal(e)
	if err != nil {
		errMsg := "error preparing Identify event: " + err.Error()
		return nil, errors.New(errMsg)
	}
	return b, nil
}
