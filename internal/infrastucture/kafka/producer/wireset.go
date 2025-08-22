package producer

import "github.com/google/wire"

var WireSet = wire.NewSet(
	MessageBrokerSet,
)
