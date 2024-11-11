package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dgraph-io/badger"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

// Declaring the structure of the receipt and item
type Item struct {
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

// Map to store the receipts in memory
var receipts = map[string]Receipt{}

func receiptExists(Id string) error {
	_, ok := receipts[Id]
	if ok {
		return nil
	}
	return errors.New("no receipt found for that id")
}

// GET request handler to calculate the points of a receipt
func getPoints(context *gin.Context) {
	Id := context.Param("Id")
	err := receiptExists(Id)
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	points, err := calculatePoints(receipts[Id])
	if err != nil {
		context.IndentedJSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	context.IndentedJSON(http.StatusOK, gin.H{"The number of points awarded are": points})
}

// POST request handler to publish a receipt to the RabbitMQ or add it to the database if the RabbitMQ is down
func addReceipt(ch *amqp.Channel, queueName string) gin.HandlerFunc {
	fn := func(context *gin.Context) {
		var newReceipt Receipt
		Id := generateId(16)
		newReceipt.Id = Id

		decoder := json.NewDecoder(context.Request.Body)
		decoder.DisallowUnknownFields() // Disallow unknown fields

		err := decoder.Decode(&newReceipt)
		if err != nil {
			context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "The receipt is invalid"})
			return
		}
		// Validate the fields of the receipt
		if !validateReceiptFields(newReceipt) {
			context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "The receipt is invalid. Check missing fields and ensure all values are strings"})
			return
		}
		receipts[Id] = newReceipt // Add the receipt to the map so that it can be retrieved even if the database is down
		context.IndentedJSON(http.StatusOK, gin.H{"Id": Id})

		receiptJSON, _ := json.Marshal(newReceipt)
		// Publish the receipt to the RabbitMQ
		err = ch.Publish(
			"",
			queueName,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        receiptJSON,
			},
		)
		if err != nil {
			fmt.Println("Failed to publish the receipt to the RabbitMQ. Saving it to the database")
			// Save the receipt to the database if the RabbitMQ is down
		}
	}
	return gin.HandlerFunc(fn)
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/") // Connect to RabbitMQ. Changed it for docker compose. It was "amqp://guest:guest@localhost:5672/"
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

	q, err := ch.QueueDeclare( // Declare a queue to store the receipts
		"POST_receipts",
		false,
		false,
		false,
		false,
		nil,
	)
	if q.Messages > 1000 {
		// Send an alert to the monitoring system that the queue is getting full
		fmt.Println("POST_requests queue is getting full")
	}
	if err != nil {
		fmt.Println("Failed to declare a queue")
		panic(err)
	}
	/**
		I understand it is not the best practice, but I am doing this to avoid using a trickier concurrency database
		The idea is to have a map of receipts in memory and then write to the database as a different process, thus ensuring the fetch here remains read-only
	**/
	db, err := badger.Open(badger.DefaultOptions("../badger/data").WithBypassLockGuard(true))
	if err != nil {
		fmt.Println("Failed to open the database")
		panic(err)
	}
	defer db.Close()
	// Get all the receipts from the database
	err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				var receipt Receipt
				json.Unmarshal(v, &receipt) // Convert the byte slice to a receipt
				receipts[string(k)] = receipt
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("Failed to retrieve receipts from the database")
		panic(err)
	}
	fmt.Println("Successfully retrieved receipts from the database")
	fmt.Println(receipts)
	router := gin.Default()
	router.GET("/receipts/:Id/points", getPoints)
	router.POST("/receipts/process", addReceipt(ch, q.Name))

	router.Run("0.0.0.0:9090")

}
