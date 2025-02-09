package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
)

func main() {
	natsURL := flag.String(
		"natsdsn",
		nats.DefaultURL,
		"Postgres data source name",
	)

	flag.Parse()

	nc, err := nats.Connect(*natsURL)
	if err != nil {
		log.Fatal("❌ Failed to connect to NATS:", err)
	}
	defer nc.Close()

	_, err = nc.Subscribe("orders.created", func(msg *nats.Msg) {
		log.Printf("✅ Received message: %s", string(msg.Data))
	})
	if err != nil {
		log.Fatal("❌ Failed to subscribe:", err)
	}

	log.Println("👂 Listening for messages on 'orders.created'")

	// Keep the program running until interrupted
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	log.Println("🛑 Shutting down subscriber...")
}
