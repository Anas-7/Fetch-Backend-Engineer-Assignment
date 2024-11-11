package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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
	// receipts["0"] = Receipt{
	// 	Id:           "0",
	// 	Retailer:     "Target",
	// 	PurchaseDate: "2022-01-01",
	// 	PurchaseTime: "13:01",
	// 	Items: []Item{
	// 		{
	// 			ShortDescription: "Mountain Dew 12PK",
	// 			Price:            "6.49",
	// 		},
	// 		{
	// 			ShortDescription: "Emils Cheese Pizza",
	// 			Price:            "12.25",
	// 		},
	// 		{
	// 			ShortDescription: "Knorr Creamy Chicken",
	// 			Price:            "1.26",
	// 		},
	// 		{
	// 			ShortDescription: "Doritos Nacho Cheese",
	// 			Price:            "3.35",
	// 		},
	// 		{
	// 			ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ",
	// 			Price:            "12.00",
	// 		},
	// 	},
	// 	Total: "35.35",
	// }
	// receipts["1"] = Receipt{
	// 	Id:           "1",
	// 	Retailer:     "M&M Corner Market",
	// 	PurchaseDate: "2022-03-20",
	// 	PurchaseTime: "14:33",
	// 	Items: []Item{
	// 		{
	// 			ShortDescription: "Gatorade",
	// 			Price:            "2.25",
	// 		},
	// 		{
	// 			ShortDescription: "Gatorade",
	// 			Price:            "2.25",
	// 		},
	// 		{
	// 			ShortDescription: "Gatorade",
	// 			Price:            "2.25",
	// 		},
	// 		{
	// 			ShortDescription: "Gatorade",
	// 			Price:            "2.25",
	// 		},
	// 	},
	// 	Total: "9.00",
	// }

	router := gin.Default()
	router.GET("/receipts/:Id/points", getPoints)
	router.POST("/receipts/process", addReceipt)

	router.Run("localhost:9090")

}
