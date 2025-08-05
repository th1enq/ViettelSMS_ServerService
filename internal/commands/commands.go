package commands

type ServerCommands struct {
	CreateServer
}

func NewServerCommands(
	createServer CreateServer,
) *ServerCommands {
	return &ServerCommands{
		CreateServer: createServer,
	}
}
