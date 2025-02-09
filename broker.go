package outbox

type MessageBroker interface {
	Publish(subject string, data []byte) error
}
