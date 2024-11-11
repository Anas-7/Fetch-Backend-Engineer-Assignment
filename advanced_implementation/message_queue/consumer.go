package main

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/streadway/amqp"
)

type Item struct {
	Id               string `json:"Id"`
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type Receipt struct {
	Id           string `json:"Id"`
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/") // Changed it for docker compose. It was "amqp://guest:guest@localhost:5672/"
	if err != nil {
		panic(err)
	}

	fmt.Println("Consumer successfully connected to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()
	q, err := ch.QueueDeclare( // QueueDeclare is used to create a new queue or initialize an existing one
		"POST_receipts",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}
	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}
	deadQueue, err := ch.QueueDeclare(
		"failed_receipts",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		panic(err)
	}
	db, err := badger.Open(badger.DefaultOptions("../badger/data").WithBypassLockGuard(true))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	fetchMsg := make(chan bool)
	go func() {
		for d := range msgs {
			txn := db.NewTransaction(true)
			var receipt Receipt
			json.Unmarshal(d.Body, &receipt)
			err := txn.Set([]byte(receipt.Id), d.Body)
			if err != nil {
				ch.Publish( // We can send this message to a dead queue so that we don't lose it
					"",
					deadQueue.Name,
					false,
					false,
					amqp.Publishing{
						ContentType: "application/json",
						Body:        d.Body,
					},
				)
				// Send an alert to the admin/monitoring system as well
				fmt.Println("CRITICAL: Error while saving receipt to database")
			}
			err = txn.Commit()
			if err != nil {
				ch.Publish(
					"",
					deadQueue.Name,
					false,
					false,
					amqp.Publishing{
						ContentType: "application/json",
						Body:        d.Body,
					},
				)
				fmt.Println("CRITICAL: Error while saving receipt to database")
			}
			txn.Discard()
		}
		fetchMsg <- true
	}()
	<-fetchMsg
}
