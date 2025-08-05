package es

import "fmt"

const (
	changesEventsCap = 10
	startVersion     = 0
)

type AggregateType string

type when func(event any) error

type AggregateBase struct {
	ID      string
	Version uint64
	Changes []any
	Type    AggregateType
	when    when
}

type When interface {
	When(event any) error
}

type AggregateRoot interface {
	GetID() string
	SetID(id string) *AggregateBase
	GetType() AggregateType
	SetType(aggregateType AggregateType)
	GetChanges() []any
	ClearChanges()
	GetVersion() uint64
	ToSnapshot()
	String() string
	Load
	Apply
	RaiseEvent
}

type RaiseEvent interface {
	RaiseEvent(event any) error
}

type Apply interface {
	Apply(event any) error
}

type Load interface {
	Load(events []any) error
}

type Aggregate interface {
	When
	AggregateRoot
	RaiseEvent
}

func NewAggregateBase(when when) *AggregateBase {
	if when == nil {
		return nil
	}

	return &AggregateBase{
		Version: startVersion,
		Changes: make([]any, 0, changesEventsCap),
		when:    when,
	}
}

func (a *AggregateBase) SetID(id string) *AggregateBase {
	a.ID = id
	return a
}

func (a *AggregateBase) GetID() string {
	return a.ID
}

func (a *AggregateBase) SetType(aggregateType AggregateType) {
	a.Type = aggregateType
}

func (a *AggregateBase) GetType() AggregateType {
	return a.Type
}

func (a *AggregateBase) GetVersion() uint64 {
	return a.Version
}

func (a *AggregateBase) ClearChanges() {
	a.Changes = make([]any, 0, changesEventsCap)
}

func (a *AggregateBase) GetChanges() []any {
	return a.Changes
}

func (a *AggregateBase) Load(events []any) error {

	for _, evt := range events {
		if err := a.when(evt); err != nil {
			return err
		}

		a.Version++
	}

	return nil
}

func (a *AggregateBase) Apply(event any) error {
	if err := a.when(event); err != nil {
		return err
	}
	a.Version++
	a.Changes = append(a.Changes, event)
	return nil
}

func (a *AggregateBase) RaiseEvent(event any) error {
	if err := a.when(event); err != nil {
		return err
	}
	a.Version++
	return nil
}

// ToSnapshot prepare AggregateBase for saving Snapshot.
func (a *AggregateBase) ToSnapshot() {
	a.ClearChanges()
}

func (a *AggregateBase) String() string {
	return fmt.Sprintf("(Aggregate) AggregateID: %s, Type: %s, Version: %v, Changes: %d",
		a.GetID(),
		string(a.GetType()),
		a.GetVersion(),
		len(a.GetChanges()),
	)
}
