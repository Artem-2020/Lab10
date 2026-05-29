package contracts

import (
	"fmt"
	"time"
)

type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type Customer struct {
	Name  string `json:"name" binding:"required,min=2,max=80"`
	Email string `json:"email" binding:"required,email"`
	Phone string `json:"phone" binding:"required,min=7,max=20"`
}

type Address struct {
	City    string `json:"city" binding:"required,min=2,max=60"`
	Street  string `json:"street" binding:"required,min=2,max=120"`
	House   string `json:"house" binding:"required,min=1,max=20"`
	Comment string `json:"comment" binding:"omitempty,max=200"`
}

type OrderItem struct {
	SKU      string `json:"sku" binding:"required,min=3,max=30"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Quantity int    `json:"quantity" binding:"required,min=1,max=20"`
	Price    Money  `json:"price" binding:"required"`
}

type DeliveryPreferences struct {
	Method       string   `json:"method" binding:"required,oneof=courier pickup post"`
	Date         string   `json:"date" binding:"required"`
	TimeWindow   string   `json:"time_window" binding:"required,oneof=morning day evening"`
	Fragile      bool     `json:"fragile"`
	AllowedZones []string `json:"allowed_zones" binding:"omitempty,dive,min=2,max=30"`
}

type QuoteRequest struct {
	RequestID   string              `json:"request_id" binding:"required,min=6,max=40"`
	Customer    Customer            `json:"customer" binding:"required"`
	DeliveryTo  Address             `json:"delivery_to" binding:"required"`
	Items       []OrderItem         `json:"items" binding:"required,min=1,dive"`
	Preferences DeliveryPreferences `json:"preferences" binding:"required"`
}

type CatalogItem struct {
	SKU         string   `json:"sku"`
	Name        string   `json:"name"`
	Available   int      `json:"available"`
	WeightGrams int      `json:"weight_grams"`
	Tags        []string `json:"tags"`
}

type CatalogCheckRequest struct {
	RequestID string      `json:"request_id" binding:"required"`
	City      string      `json:"city" binding:"required"`
	Items     []OrderItem `json:"items" binding:"required,min=1,dive"`
	Zones     []string    `json:"zones"`
}

type CatalogCheckResponse struct {
	RequestID string        `json:"request_id"`
	Status    string        `json:"status"`
	Message   string        `json:"message"`
	Items     []CatalogItem `json:"items"`
	Missing   []string      `json:"missing"`
}

type QuoteResponse struct {
	RequestID string              `json:"request_id"`
	Status    string              `json:"status"`
	Message   string              `json:"message"`
	Customer  Customer            `json:"customer"`
	Delivery  DeliveryCalculation `json:"delivery"`
	Items     []CatalogItem       `json:"items"`
	Total     Money               `json:"total"`
}

type DeliveryCalculation struct {
	Method        string   `json:"method"`
	City          string   `json:"city"`
	Date          string   `json:"date"`
	TimeWindow    string   `json:"time_window"`
	AllowedZones  []string `json:"allowed_zones"`
	DeliveryPrice Money    `json:"delivery_price"`
}

func ValidateQuoteRequest(request QuoteRequest) error {
	if err := validateMoneyList(request.Items); err != nil {
		return err
	}

	deliveryDate, err := time.Parse("2006-01-02", request.Preferences.Date)
	if err != nil {
		return fmt.Errorf("preferences.date must use YYYY-MM-DD format")
	}

	today := time.Now().Truncate(24 * time.Hour)
	if deliveryDate.Before(today) {
		return fmt.Errorf("preferences.date must not be in the past")
	}

	return nil
}

func ValidateCatalogCheckRequest(request CatalogCheckRequest) error {
	return validateMoneyList(request.Items)
}

func validateMoneyList(items []OrderItem) error {
	for index, item := range items {
		if item.Price.Amount <= 0 {
			return fmt.Errorf("items[%d].price.amount must be greater than 0", index)
		}
		if item.Price.Currency == "" {
			return fmt.Errorf("items[%d].price.currency is required", index)
		}
		if item.Price.Currency != "RUB" {
			return fmt.Errorf("items[%d].price.currency must be RUB", index)
		}
	}

	return nil
}
