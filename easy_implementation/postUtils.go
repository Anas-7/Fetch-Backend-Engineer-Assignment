package main

import (
	"reflect"
	"strconv"
	"time"
)

func isValidDate(dateStr string) bool {
	// dates can have leading zeros
	layouts := []string{"2006-01-02", "2006-1-2"}
	for _, layout := range layouts {
		_, err := time.Parse(layout, dateStr)
		if err == nil {
			return true
		}
	}
	return false
}

func isValidTime(timeStr string) bool {
	// need to allow "HH:mm" and "H:m"
	layouts := []string{"15:04", "3:4"}

	for _, layout := range layouts {
		if _, err := time.Parse(layout, timeStr); err == nil {
			return true
		}
	}
	return false
}

func isValidNumber(numStr string) bool {
	_, err := strconv.ParseFloat(numStr, 64)
	return err == nil
}

func validateReceiptFields(receipt Receipt) bool {
	// Check for empty fields
	if receipt.Retailer == "" || receipt.PurchaseDate == "" || receipt.PurchaseTime == "" ||
		len(receipt.Items) == 0 || receipt.Total == "" {
		return false
	}

	// Check type
	if reflect.TypeOf(receipt.Retailer).Kind() != reflect.String ||
		reflect.TypeOf(receipt.PurchaseDate).Kind() != reflect.String ||
		reflect.TypeOf(receipt.PurchaseTime).Kind() != reflect.String ||
		reflect.TypeOf(receipt.Total).Kind() != reflect.String {
		return false
	}
	// Check if valid date and time
	if !isValidDate(receipt.PurchaseDate) || !isValidTime(receipt.PurchaseTime) || !isValidNumber(receipt.Total) {
		return false
	}

	// Validate each item and check if price is number and description is string
	for _, item := range receipt.Items {
		if item.ShortDescription == "" || item.Price == "" || !isValidNumber(item.Price) {
			return false
		}
		if reflect.TypeOf(item.ShortDescription).Kind() != reflect.String ||
			reflect.TypeOf(item.Price).Kind() != reflect.String {
			return false
		}
	}

	return true
}
