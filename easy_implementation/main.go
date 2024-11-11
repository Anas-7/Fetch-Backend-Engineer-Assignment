package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
	points, err := calculatePoints(receipts[Id])
	if err != nil {
		context.IndentedJSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	context.IndentedJSON(http.StatusOK, gin.H{"The number of points awarded are": points})
}

func addReceipt(context *gin.Context) {
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
	receipts[Id] = newReceipt
	context.IndentedJSON(http.StatusOK, gin.H{"Id": Id})
	fmt.Println(receipts)
}

func main() {
	router := gin.Default()
	router.GET("/receipts/:Id/points", getPoints)
	router.POST("/receipts/process", addReceipt)
	router.Run("localhost:9090")
}
