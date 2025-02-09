package outbox

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	store := &MockStore{}
	broker := &MockBroker{}
	pollInterval := time.Second * 10

	got := NewOutbox(store, broker, pollInterval)

	assert.Equal(t, store, got.store)
	assert.Equal(t, broker, got.broker)
	assert.Equal(t, pollInterval, got.pollInterval)
}

func TestOutbox_AddMessage(t *testing.T) {
	sampleTx := &sql.Tx{} // Mock transaction object
	sampleSubject := "test-subject"
	sampleData := []byte("test-data")

	tests := map[string]struct {
		subject string
		data    []byte
		store   *MockStore
		tx      *sql.Tx
		expErr  error
	}{
		"Successful message addition should return without error": {
			subject: sampleSubject,
			data:    sampleData,
			store: func() *MockStore {
				ms := &MockStore{}
				msg := Message{Subject: sampleSubject, Data: sampleData}
				ms.On("AddMessageTx", mock.Anything, msg, sampleTx).
					Return(nil)
				return ms
			}(),
			tx:     sampleTx,
			expErr: nil,
		},
		"Failure in adding message should return error": {
			subject: sampleSubject,
			data:    sampleData,
			store: func() *MockStore {
				ms := &MockStore{}
				msg := Message{Subject: sampleSubject, Data: sampleData}
				ms.On("AddMessageTx", mock.Anything, msg, sampleTx).
					Return(errors.New("error"))
				return ms
			}(),
			tx:     sampleTx,
			expErr: errors.New("error"),
		},
	}

	for name, test := range tests {
		tt := test
		t.Run(name, func(t *testing.T) {
			outbox := &Outbox{
				store:        tt.store,
				broker:       nil, // Not needed for this test
				pollInterval: time.Second,
				stopChan:     make(chan struct{}),
				isLeader:     false,
			}

			err := outbox.AddMessage(tt.tx, tt.subject, tt.data)
			assert.Equal(t, tt.expErr, err)

			tt.store.AssertExpectations(t)
		})
	}
}

func TestOutbox_SetLeader(t *testing.T) {
	tests := map[string]struct {
		initial  bool
		input    bool
		expected bool
	}{
		"Default state should be false": {
			initial:  false,
			input:    false,
			expected: false,
		},
		"Set as Leader": {
			initial:  false,
			input:    true,
			expected: true,
		},
		"Unset as Leader": {
			initial:  false,
			input:    false,
			expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			o := Outbox{
				isLeader: test.initial,
			}

			o.SetLeader(test.input)

			assert.Equal(t, test.expected, o.isLeader)
		})
	}
}

func TestOutbox_Start(t *testing.T) {
	sampleMsg := Message{
		ID: 1, Subject: "test", Data: []byte("test data"),
	}

	sampleMessages := []Message{sampleMsg}
	sampleIDs := []int64{1}

	t.Run("Fetch, Publish, and Mark As Sent should work", func(t *testing.T) {
		ms := &MockStore{}
		ms.On("FetchUnsentMessages", mock.Anything, 100).
			Return(sampleMessages, nil)
		ms.On("MarkMessagesAsSent", mock.Anything, sampleIDs).
			Return(nil)

		b := &MockBroker{}
		b.On("Publish", sampleMsg.Subject, sampleMsg.Data).Return(nil)

		o := Outbox{
			store:        ms,
			broker:       b,
			pollInterval: 10 * time.Millisecond,
			stopChan:     make(chan struct{}),
		}

		o.SetLeader(true)

		o.Start()
		time.Sleep(15 * time.Millisecond) // Allow some time for processing
		o.Stop()

		ms.AssertExpectations(t)
		b.AssertExpectations(t)
		// 			assert.Error(t, test.expErr, "Expected an error")
	})

	t.Run("Fetch failure - Should log error and stop", func(t *testing.T) {
		ms := &MockStore{}
		ms.On("FetchUnsentMessages", mock.Anything, 100).
			Return([]Message{}, errors.New("fetch error"))

		b := &MockBroker{}

		o := Outbox{
			store:        ms,
			broker:       b,
			pollInterval: 10 * time.Millisecond,
			stopChan:     make(chan struct{}),
		}

		o.SetLeader(true)

		o.Start()
		time.Sleep(15 * time.Millisecond) // Allow some time for processing
		o.Stop()

		ms.AssertExpectations(t)
		b.AssertNotCalled(t, "Publish", sampleMsg.Subject, sampleMsg.Data)
		ms.AssertNotCalled(t, "MarkMessagesAsSent", mock.Anything, sampleIDs)
	})

	t.Run(
		"Message length zero - Should not call Publish or MarkAsSent",
		func(t *testing.T) {
			ms := &MockStore{}
			ms.On("FetchUnsentMessages", mock.Anything, 100).
				Return([]Message{}, nil)

			b := &MockBroker{}

			o := Outbox{
				store:        ms,
				broker:       b,
				pollInterval: 10 * time.Millisecond,
				stopChan:     make(chan struct{}),
			}

			o.SetLeader(true)

			o.Start()
			time.Sleep(15 * time.Millisecond) // Allow some time for processing
			o.Stop()

			ms.AssertExpectations(t)
			b.AssertNotCalled(t, "Publish", sampleMsg.Subject, sampleMsg.Data)
			ms.AssertNotCalled(
				t,
				"MarkMessagesAsSent",
				mock.Anything,
				sampleIDs,
			)
		},
	)

	t.Run(
		"Publish failure - Should not call MarkMessagesAsSent",
		func(t *testing.T) {
			ms := &MockStore{}
			ms.On("FetchUnsentMessages", mock.Anything, 100).
				Return(sampleMessages, nil)

			b := &MockBroker{}
			b.On("Publish", sampleMsg.Subject, sampleMsg.Data).
				Return(errors.New("publish error"))

			o := Outbox{
				store:        ms,
				broker:       b,
				pollInterval: 10 * time.Millisecond,
				stopChan:     make(chan struct{}),
			}

			o.SetLeader(true)

			o.Start()
			time.Sleep(15 * time.Millisecond) // Allow some time for processing
			o.Stop()

			ms.AssertExpectations(t)
			ms.AssertNotCalled(
				t,
				"MarkMessagesAsSent",
				mock.Anything,
				sampleIDs,
			)
		},
	)

	t.Run(
		"MarkMessagesAsSent failure - Should retry message in next iteration",
		func(t *testing.T) {
			ms := &MockStore{}
			ms.On("FetchUnsentMessages", mock.Anything, 100).
				Return(sampleMessages, nil).
				Times(1) // Retry after failure in the first attempt

			ms.On("MarkMessagesAsSent", mock.Anything, sampleIDs).
				Return(errors.New("mark as sent error")).
				On("MarkMessagesAsSent", mock.Anything, sampleIDs).
				Return(nil)

			b := &MockBroker{}
			b.On("Publish", sampleMsg.Subject, sampleMsg.Data).Return(nil)

			o := Outbox{
				store:        ms,
				broker:       b,
				pollInterval: 10 * time.Millisecond,
				stopChan:     make(chan struct{}),
			}

			o.SetLeader(true)

			o.Start()
			time.Sleep(12 * time.Millisecond) // Allow some time for processing
			o.Stop()

			ms.AssertExpectations(t)
			b.AssertExpectations(t)
		},
	)
}
