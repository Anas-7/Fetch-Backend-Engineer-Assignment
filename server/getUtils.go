package main

import (
	"errors"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func calculateRetailerNamePoints(retailerName string) int {
	count := 0
	for _, char := range retailerName {
		if unicode.IsLetter(char) || unicode.IsNumber(char) {
			count += 1
		}
	}
	return count
}

func calculateTotalPricePoints(totalPrice string) (int, error) {
	total, err := strconv.ParseFloat(totalPrice, 64)
	if err != nil {
		return -1, errors.New("error in converting total price to float")
	}
	points := 0
	if math.Mod(total, 1) == 0 {
		points += 50
	}
	if math.Mod(total, 0.25) == 0 {
		points += 25
	}
	return points, nil
}

func calculateItemCountPoints(items []Item) int {
	itemsCount := len(items)
	return (itemsCount / 2) * 5
}

func calculateItemDescriptionPoints(items []Item) (int, error) {
	total := 0
	for _, item := range items {
		shortDescription := item.ShortDescription
		trimmedString := strings.TrimSpace(shortDescription)
		if len(trimmedString)%3 == 0 {
			price, err := strconv.ParseFloat(item.Price, 64)
			if err != nil {
				return -1, errors.New("error in converting price of an item to float")
			}
			total += int(math.Ceil(price * 0.2))
		}
	}
	return total, nil
}

func calculatePurchaseDayPoints(purchaseDay int) int {
	if purchaseDay%2 == 1 {
		return 6
	}
	return 0
}

func calculatePurchaseDatePoints(purchaseDate string) (int, error) {
	parseDate := strings.Split(purchaseDate, "-")
	purchaseDay, err := strconv.Atoi(parseDate[2])
	if err != nil {
		return -1, errors.New("error in converting purchase day to integer")
	}
	total := 0
	total += calculatePurchaseDayPoints(purchaseDay)

	return total, nil
}

func calculatePurchaseTimePoints(purchaseTime string) (int, error) {
	parseTime := strings.Split(purchaseTime, ":")
	hour, err := strconv.Atoi(parseTime[0])
	if err != nil {
		return -1, errors.New("error in converting purchase hour to integer")
	}
	if hour >= 14 && hour < 16 {
		return 10, nil
	}
	return 0, nil
}

func generateId(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func calculatePoints(Id string) (int, error) {
	receipt := receipts[Id]
	totalPoints := 0

	retailerNamePoints := calculateRetailerNamePoints(receipt.Retailer)

	totalPricePoints, err := calculateTotalPricePoints(receipt.Total)
	if err != nil {
		return -1, err
	}

	itemCountPoints := calculateItemCountPoints(receipt.Items)

	itemDescriptionPoints, err := calculateItemDescriptionPoints(receipt.Items)
	if err != nil {
		return -1, err
	}

	itemPurchaseDatePoints, err := calculatePurchaseDatePoints(receipt.PurchaseDate)
	if err != nil {
		return -1, err
	}

	itemPurhcaseTimePoints, err := calculatePurchaseTimePoints(receipt.PurchaseTime)
	if err != nil {
		return -1, err
	}
	// Print the points for each category for debugging
	// fmt.Println("Retailer Name Points: ", retailerNamePoints)
	// fmt.Println("Total Price Points: ", totalPricePoints)
	// fmt.Println("Item Count Points: ", itemCountPoints)
	// fmt.Println("Item Description Points: ", itemDescriptionPoints)
	// fmt.Println("Item Purchase Date Points: ", itemPurchaseDatePoints)
	// fmt.Println("Item Purchase Time Points: ", itemPurhcaseTimePoints)

	totalPoints = retailerNamePoints + totalPricePoints + itemCountPoints +
		itemDescriptionPoints + itemPurchaseDatePoints + itemPurhcaseTimePoints
	return totalPoints, nil
}
