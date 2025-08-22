package kafka

import (
	"github.com/google/wire"
	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/kafka/consumer"
	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/kafka/producer"
)

var WireSet = wire.NewSet(
	consumer.WireSet,
	producer.WireSet,
)
