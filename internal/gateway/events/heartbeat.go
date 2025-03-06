package events

import (
	"encoding/json"
	"errors"
)

type HeartbeatEvent struct {
	Event
	LastSequence float64 `json:"d"`
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

func (e *HeartbeatEvent) DecodeData(msg []byte) (ReceivableEvent, error) {
	var event HeartbeatEvent
	err := json.Unmarshal(msg, &event)
	if err != nil {
		errMsg := "error decoding HeartbeatEvent: " + err.Error()
		return nil, errors.New(errMsg)
	}

	if event.Operation != Heartbeat {
		errMsg := "unexpected event received: expected Heartbeat, got " + event.Operation.String()
		return nil, errors.New(errMsg)
	}

	e = &event
	return e, nil
}

func NewHeartbeatAckEvent() *HeartbeatAckEvent {
	return &HeartbeatAckEvent{
		Event: Event{
			Operation: Heartbeat_ACK,
		},
	}
}

func (e *HeartbeatAckEvent) DecodeData(msg []byte) (ReceivableEvent, error) {
	var event HeartbeatAckEvent
	err := json.Unmarshal(msg, &event)
	if err != nil {
		errMsg := "error decoding HeartbeatACK event: " + err.Error()
		return nil, errors.New(errMsg)
	}

	if event.Operation != Heartbeat_ACK {
		errMsg := "unexpected event received: expected Heartbeat_ACK, got " + event.Operation.String()
		return nil, errors.New(errMsg)
	}

	e = &event
	return e, nil
}
