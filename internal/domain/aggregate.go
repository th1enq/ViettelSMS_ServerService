package domain

import (
	"ViettelSMS_ServerService/internal/events"
	"ViettelSMS_ServerService/pkg/es"
	"context"
)

const (
	ServerAggregateType es.AggregateType = "Server"
)

type ServerAggregate struct {
	*es.AggregateBase
	Server *Server
}

func NewServerAggregate(aggregateID string) *ServerAggregate {
	if aggregateID == "" {
		return nil
	}

	serverAggregate := &ServerAggregate{Server: NewServer(aggregateID)}
	aggregateBase := es.NewAggregateBase(serverAggregate.When)
	aggregateBase.Type = ServerAggregateType
	aggregateBase.ID = aggregateID
	serverAggregate.AggregateBase = aggregateBase

	return serverAggregate
}

func (s *ServerAggregate) CreateNewServer(
	ctx context.Context,
	serverID, serverName, ipV4, location, os string,
	intervalTime uint32,
) error {
	event := &events.ServerCreated{
		ServerID:     serverID,
		ServerName:   serverName,
		IPV4:         ipV4,
		Location:     location,
		OS:           os,
		IntervalTime: intervalTime,
	}
	return s.Apply(event)
}

func (s *ServerAggregate) When(event any) error {
	switch evt := event.(type) {
	case *events.ServerCreated:
		s.Server.ServerID = evt.ServerID
		s.Server.ServerName = evt.ServerName
		s.Server.IPv4 = evt.IPV4
		s.Server.Location = evt.Location
		s.Server.OS = evt.OS
		s.Server.IntervalTime = evt.IntervalTime
		return nil
	}

	return nil
}
