package events

import (
	"encoding/json"
	"errors"
)

type HeartbeatEvent struct {
	Event
	LastSequence float64
}

type HeartbeatAckEvent struct {
	Event
}

func NewHeartbeatEvent() *HeartbeatEvent {
	return &HeartbeatEvent{
		Event: Event{
			Operation: Heartbeat,
		},
	}
}

func (e HeartbeatEvent) PrepareToSend() ([]byte, error) {
	msg, err := json.Marshal(e)
	if err != nil {
		errMsg := "error preparing Heartbeat event: " + err.Error()
		return []byte{}, errors.New(errMsg)
	}
	return msg, nil
}

func (e *HeartbeatEvent) DecodeData(gen_event Event) error {
	if gen_event.Operation != Heartbeat {
		errMsg := "unexpected event received: expected Heartbeat, got " + gen_event.Operation.String()
		return errors.New(errMsg)
	}

	var payload struct {
		LastSequence float64 `json:"d"`
	}
	err := json.Unmarshal(gen_event.RawData, &payload)
	if err != nil {
		errMsg := "error decoding HeartbeatEvent: " + err.Error()
		return errors.New(errMsg)
	}

	e.Event = gen_event
	e.LastSequence = payload.LastSequence
	return nil
}

func NewHeartbeatAckEvent() *HeartbeatAckEvent {
	return &HeartbeatAckEvent{
		Event: Event{
			Operation: Heartbeat_ACK,
		},
	}
}

func (e *HeartbeatAckEvent) DecodeData(gen_event Event) error {
	if gen_event.Operation != Heartbeat_ACK {
		errMsg := "unexpected event received: expected Heartbeat_ACK, got " + gen_event.Operation.String()
		return errors.New(errMsg)
	}

	e.Event = gen_event
	return nil
}
