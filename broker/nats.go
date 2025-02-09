package broker

import "github.com/nats-io/nats.go"

type NATSBroker struct {
	nc *nats.Conn
}

func NewNATSMessageBroker(nc *nats.Conn) *NATSBroker {
	return &NATSBroker{nc: nc}
}

func (nb *NATSBroker) Publish(subject string, data []byte) error {
	return nb.nc.Publish(subject, data)
}
