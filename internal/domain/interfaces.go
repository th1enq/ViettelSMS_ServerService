package domain

import "context"

type ServerRepository interface {
	ExistByNameOrID(ctx context.Context, serverID string, serverName string) (bool, error)
	Create(ctx context.Context, server *Server) error
	Delete(ctx context.Context, serverID string) error
	GetByField(ctx context.Context, field string, value interface{}) (*Server, error)
	Update(ctx context.Context, server *Server) error
	GetServers(ctx context.Context, filter ServerFilterOptions, pagination ServerPaginationOptions) ([]*Server, int, error)
	BatchCreate(ctx context.Context, servers []*Server) error
}
