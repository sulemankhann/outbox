package outbox

import "github.com/stretchr/testify/mock"

type MockBroker struct {
	mock.Mock
}

func (m *MockBroker) Publish(subject string, data []byte) error {
	args := m.Called(subject, data)
	return args.Error(0)
}
