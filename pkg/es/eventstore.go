package es

import "context"

type AggregateStore interface {
	Save(ctx context.Context, aggregate Aggregate) error

	Load(ctx context.Context, aggregateID Aggregate) error

	Exists(ctx context.Context, aggregateID string) (bool, error)

	EventStore
	SnapshotStore
}

type EventStore interface {
	SaveEvents(ctx context.Context, events []Event) error
	LoadEvents(ctx context.Context, aggregateID string) ([]Event, error)
}

type SnapshotStore interface {
	SaveSnapshot(ctx context.Context, aggregate Aggregate) error

	GetSnapshot(ctx context.Context, id string) (*Snapshot, error)
}
