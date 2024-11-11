package main

import (
	"fmt"

	"github.com/streadway/amqp"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
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
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(`{"Id":"0","Retailer":"Target","PurchaseDate":"2021-01-01","PurchaseTime":"15:04:05","Items":[{"Id":"0","ShortDescription":"item1","Price":"10"},{"Id":"1","ShortDescription":"item2","Price":"20"}],"Total":"30"}`),
		},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully published message to RabbitMQ")
}
