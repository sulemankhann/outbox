package outbox

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type Outbox struct {
	store        Store
	broker       MessageBroker
	pollInterval time.Duration
	stopChan     chan struct{}
	isLeader     bool
}

func NewOutbox(
	store Store,
	broker MessageBroker,
	pollInterval time.Duration,
) *Outbox {
	return &Outbox{
		store:        store,
		broker:       broker,
		pollInterval: pollInterval,
		stopChan:     make(chan struct{}),
		isLeader:     false,
	}
}

func (o *Outbox) AddMessage(
	ctx context.Context,
	tx *sql.Tx,
	subject string,
	data []byte,
) error {
	msg := Message{Subject: subject, Data: data}

	return o.store.AddMessageTx(ctx, msg, tx)
}

func (o *Outbox) Start() {
	go o.poller()
}

func (o *Outbox) Stop() {
	close(o.stopChan)
}

func (o *Outbox) SetLeader(isLeader bool) {
	o.isLeader = isLeader
}

func (o *Outbox) poller() {
	ticker := time.NewTicker(o.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if o.isLeader {
				o.processMessages()
			}

		case <-o.stopChan:
			return
		}
	}
}

func (o *Outbox) processMessages() {
	ctx := context.Background()

	messages, err := o.store.FetchUnsentMessages(ctx, 100)
	if err != nil {
		log.Print(err)
		return
	}

	if len(messages) == 0 {
		return
	}

	ids := make([]int64, len(messages))
	for i, msg := range messages {
		if err = o.broker.Publish(msg.Subject, msg.Data); err != nil {
			log.Printf("Error publishing message %d: %v", msg.ID, err)
			return

		}
		ids[i] = msg.ID
	}

	err = o.store.MarkMessagesAsSent(ctx, ids)
	if err != nil {
		log.Printf("Error updating sen_at: %v", err)
		return
	}

	log.Printf("Processed %d messages", len(messages))
}
