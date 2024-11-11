package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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

// POST request handler to add a receipt
func addReceipt(context *gin.Context) {
	var newReceipt Receipt
	Id := generateId(16)
	newReceipt.Id = Id

	decoder := json.NewDecoder(context.Request.Body)
	decoder.DisallowUnknownFields() // Used to return an error if the JSON contains unknown fields

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
	receipts[Id] = newReceipt
	context.IndentedJSON(http.StatusOK, gin.H{"Id": Id})
	fmt.Println(receipts)
}

func main() {
	router := gin.Default()
	// Routes
	router.GET("/receipts/:Id/points", getPoints)
	router.POST("/receipts/process", addReceipt)
	// Run the server
	router.Run("0.0.0.0:9090") // Changed from localhost to 0.0.0.0 for docker
}
