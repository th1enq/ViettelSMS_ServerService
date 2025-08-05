package es

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type EventType string

type Event struct {
	EventID       string
	AggregateID   string
	EventType     EventType
	AggregateType AggregateType
	Version       uint64
	Data          []byte
	Metadata      []byte
	Timestamp     time.Time
}

func NewBaseEvent(aggregate Aggregate, eventType EventType) Event {
	return Event{
		EventID:       uuid.NewV4().String(),
		AggregateType: aggregate.GetType(),
		AggregateID:   aggregate.GetID(),
		Version:       aggregate.GetVersion(),
		EventType:     eventType,
		Timestamp:     time.Now().UTC(),
	}
}

func NewEvent(aggregate Aggregate, eventType EventType, data []byte, metadata []byte) Event {
	return Event{
		EventID:       uuid.NewV4().String(),
		AggregateID:   aggregate.GetID(),
		EventType:     eventType,
		AggregateType: aggregate.GetType(),
		Version:       aggregate.GetVersion(),
		Data:          data,
		Metadata:      metadata,
		Timestamp:     time.Now().UTC(),
	}
}
