package events

import (
	"encoding/json"
	"errors"
)

type HeartbeatEvent struct {
	Event
	Data float64 `json:"d"`
}

type HeartbeatAckEvent struct {
	Event
}

func NewHeartbeatEvent() *HeartbeatEvent {
	return &HeartbeatEvent{
		Event: Event{
			Operation: Heartbeat,
			Sequence:  nil,
			Type:      nil,
		},
		Data: 0,
	}
}

func (e HeartbeatEvent) Prepare() ([]byte, error) {
	msg, err := json.Marshal(e)
	if err != nil {
		errMsg := "error preparing Heartbeat event: " + err.Error()
		return []byte{}, errors.New(errMsg)
	}
	return msg, nil
}

func (e *HeartbeatEvent) DecodeData(msg []byte) (ReceivableEvent, error) {
	var payload float64
	err := json.Unmarshal(msg, &payload)
	if err != nil {
		errMsg := "error decoding Heartbeat event: " + err.Error()
		return nil, errors.New(errMsg)
	}
	e.Data = payload
	return e, nil
}

func NewHeartbeatAckEvent() *HeartbeatAckEvent {
	return &HeartbeatAckEvent{
		Event: Event{
			Operation: Heartbeat_ACK,
			Sequence:  nil,
			Type:      nil,
		},
	}
}

func (e HeartbeatAckEvent) DecodeData(msg []byte) (ReceivableEvent, error) {
	return e, nil
}
