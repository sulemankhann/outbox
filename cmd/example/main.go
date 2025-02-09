package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sulemankhann/outbox"
	"sulemankhann/outbox/broker"
	"sulemankhann/outbox/store"
	"time"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

type application struct {
	logger *slog.Logger
	db     *sql.DB
	outbox *outbox.Outbox
}

func main() {
	leader := flag.Bool("leader", false, "Run as leader instance")
	port := flag.String("port", "8080", "Port to listen on")
	dsn := flag.String(
		"dsn",
		"postgres://postgres:postgres@localhost:5432/outbox?sslmode=disable",
		"Postgres data source name",
	)
	natsDSN := flag.String(
		"natsdsn",
		nats.DefaultURL,
		"Postgres data source name",
	)

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := connectDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	err = createTables(db)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	nc, err := connectNATS(*natsDSN)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer nc.Close()

	store := store.NewPostgresStore(db)
	broker := broker.NewNATSMessageBroker(nc)
	pollInterval := 10 * time.Second

	ob := outbox.NewOutbox(store, broker, pollInterval)
	ob.SetLeader(*leader)
	ob.Start()
	defer ob.Stop()

	app := &application{
		logger: logger,
		db:     db,
		outbox: ob,
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", *port),
		Handler: app.routes(),
	}

	logger.Info("starting server", "port", srv.Addr, "leader", *leader)

	err = srv.ListenAndServe()

	logger.Error(err.Error())

	os.Exit(1)
}

func connectNATS(url string) (*nats.Conn, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return nc, nil
}

func connectDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			total INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS outbox (
			id BIGSERIAL PRIMARY KEY,
			subject TEXT NOT NULL,
			data BYTEA NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			sent_at TIMESTAMP
		)`)
	if err != nil {
		return err
	}

	return nil
}
