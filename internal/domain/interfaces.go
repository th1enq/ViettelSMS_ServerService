package domain

import "context"

type ServerRepository interface {
	ExistByNameOrID(ctx context.Context, serverID string, serverName string) (bool, error)
	Create(ctx context.Context, server *Server) error
}
