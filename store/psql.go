package store

import (
	"context"
	"database/sql"
	"fmt"
	"sulemankhann/outbox"

	"github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (p *PostgresStore) AddMessageTx(
	ctx context.Context,
	msg outbox.Message,
	tx *sql.Tx,
) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO outbox(subject, data) VALUES ($1, $2)`,
		msg.Subject,
		msg.Data,
	)

	return err
}

func (p *PostgresStore) FetchUnsentMessages(
	ctx context.Context,
	limit int,
) ([]outbox.Message, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, subject, data FROM outbox 
		WHERE sent_at IS NULL 
		ORDER BY id ASC 
		LIMIT $1 
		FOR UPDATE SKIP LOCKED`, limit)
	if err != nil {
		return nil, fmt.Errorf("Error querying outbox: %v", err)
	}
	defer rows.Close()

	var messages []outbox.Message

	for rows.Next() {
		var msg outbox.Message
		if err := rows.Scan(&msg.ID, &msg.Subject, &msg.Data); err != nil {
			return nil, fmt.Errorf("Error scanning message: %v", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error after scanning rows: %v", err)
	}

	return messages, nil
}

func (p *PostgresStore) MarkMessagesAsSent(
	ctx context.Context,
	messageIDs []int64,
) error {
	_, err := p.db.ExecContext(ctx,
		`UPDATE outbox SET sent_at = NOW() WHERE id = ANY($1)`,
		pq.Array(messageIDs),
	)
	return err
}
