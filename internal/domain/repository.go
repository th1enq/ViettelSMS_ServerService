package domain

import "context"

type PostgresRepository interface {
	CreateServer(ctx context.Context, server *Server) error
	DeleteServer(ctx context.Context, server_id string) error
	UpdateServer(ctx context.Context, server *Server) error
	UpdateStatus(ctx context.Context, server_id string, status string) error
}
