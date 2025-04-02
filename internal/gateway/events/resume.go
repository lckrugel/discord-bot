package events

import (
	"encoding/json"
	"errors"
)

type ResumeEvent struct {
	Event
	Data ResumePayload `json:"d"`
}

type ResumePayload struct {
	Token     string `json:"token"`
	SessionId string `json:"session_id"`
	Sequence  int    `json:"seq"`
}

func NewResumeEvent(payload ResumePayload) ResumeEvent {
	return ResumeEvent{
		Event: Event{
			Operation: Resume,
		},
		Data: payload,
	}
}

func (e ResumeEvent) prepareToSend() ([]byte, error) {
	json, err := json.Marshal(e)
	if err != nil {
		errMsg := "error preparing Resume event: " + err.Error()
		return nil, errors.New(errMsg)
	}
	return json, nil
}
