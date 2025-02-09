package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type orderCreateForm struct {
	Name  string `json:"name"`
	Total int    `json:"total"`
}

func (app *application) createOrderHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	var order orderCreateForm

	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := app.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	ctx := context.Background()
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO  orders (name, total) VALUES ($1, $2)`,
		order.Name,
		order.Total,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	event := map[string]interface{}{
		"event_type": "order_created",
		"order":      order,
	}
	eventData, _ := json.Marshal(event)

	if err := app.outbox.AddMessage(ctx, tx, "orders.created", eventData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	fmt.Fprintln(w, "Order Created")
}
