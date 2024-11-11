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

var receipts = map[string]Receipt{}

func receiptExists(Id string) error {
	_, ok := receipts[Id]
	if ok {
		return nil
	}
	return errors.New("no receipt found for that id")
}

func getPoints(context *gin.Context) {
	Id := context.Param("Id")
	err := receiptExists(Id)
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	points, err := calculatePoints(Id)
	if err != nil {
		context.IndentedJSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	context.IndentedJSON(http.StatusOK, gin.H{"The number of points awarded are": points})
}

func addReceipt(ch *amqp.Channel, queueName string) gin.HandlerFunc {
	fn := func(context *gin.Context) {
		var newReceipt Receipt
		Id := generateId(16)
		newReceipt.Id = Id

		decoder := json.NewDecoder(context.Request.Body)
		decoder.DisallowUnknownFields()

		err := decoder.Decode(&newReceipt)
		if err != nil {
			context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "The receipt is invalid"})
			return
		}
		if !validateReceiptFields(newReceipt) {
			context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "The receipt is invalid. Check missing fields and ensure all values are strings"})
			return
		}
		receipts[Id] = newReceipt // Add the receipt to the map so that it can be retrieved even if the database is down
		context.IndentedJSON(http.StatusOK, gin.H{"Id": Id})

		receiptJSON, _ := json.Marshal(newReceipt)
		// Publish the receipt to the RabbitMQ
		ch.Publish(
			"",
			queueName,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        receiptJSON,
			},
		)
	}
	return gin.HandlerFunc(fn)
}

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
	if q.Messages > 1000 {
		// Send an alert to the monitoring system that the queue is getting full
		fmt.Println("POST_requests queue is getting full")
	}
	if err != nil {
		panic(err)
	}
	/**
		not the best practice, but I am doing this to avoid using a tricker concurrency database
		The idea is to have a map of receipts in memory and then write to the database as a different process, thus ensuring the fetch here remains read-only
	**/
	db, err := badger.Open(badger.DefaultOptions("../badger/data").WithBypassLockGuard(true))
	if err != nil {
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
				json.Unmarshal(v, &receipt)
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

	router.Run("localhost:9090")

}
