package outbox

import (
	"context"
	"database/sql"
)

type Message struct {
	ID      int64
	Subject string
	Data    []byte
}

type Store interface {
	AddMessageTx(ctx context.Context, message Message, tx *sql.Tx) error
	FetchUnsentMessages(ctx context.Context, limit int) ([]Message, error)
	MarkMessagesAsSent(ctx context.Context, messageIDs []int64) error
}
