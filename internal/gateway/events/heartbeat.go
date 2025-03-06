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
	var lastSeq float64
	err := json.Unmarshal(msg, &lastSeq)
	if err != nil {
		errMsg := "error decoding Heartbeat event: " + err.Error()
		return nil, errors.New(errMsg)
	}
	e.LastSequence = lastSeq
	return e, nil
}

func NewHeartbeatAckEvent() *HeartbeatAckEvent {
	return &HeartbeatAckEvent{
		Event: Event{
			Operation: Heartbeat_ACK,
		},
	}
}

func (e HeartbeatAckEvent) DecodeData(msg []byte) (ReceivableEvent, error) {
	return e, nil
}
