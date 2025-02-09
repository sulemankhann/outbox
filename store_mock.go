package outbox

import (
	"context"
	"database/sql"

	"github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) AddMessageTx(ctx context.Context,
	msg Message,
	tx *sql.Tx,
) error {
	args := m.Called(ctx, msg, tx)
	return args.Error(0)
}

func (m *MockStore) FetchUnsentMessages(
	ctx context.Context,
	limit int,
) ([]Message, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]Message), args.Error(1)
}

func (m *MockStore) MarkMessagesAsSent(
	ctx context.Context,
	messageIDs []int64,
) error {
	args := m.Called(ctx, messageIDs)
	return args.Error(0)
}
