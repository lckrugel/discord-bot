package events

type ReconnectEvent struct {
	Event
}

func NewReconnectEvent() ReconnectEvent {
	return ReconnectEvent{
		Event: Event{
			Operation: Reconnect,
		},
	}
}

func (e *ReconnectEvent) DecodeData(data []byte) (ReceivableEvent, error) {
	return e, nil
}
