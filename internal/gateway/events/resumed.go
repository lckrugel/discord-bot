package events

type ResumedEvent struct {
	Event
}

func NewResumedEvent() ResumedEvent {
	resumedType := "Resumed"
	return ResumedEvent{
		Event: Event{
			Operation: Dispatch,
			Type:      &resumedType,
		},
	}
}

func (e *ResumedEvent) DecodeData(data []byte) (ReceivableEvent, error) {
	return e, nil
}
