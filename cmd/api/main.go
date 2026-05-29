package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"lab10-medium-gin/internal/contracts"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	catalogURL := getEnv("CATALOG_URL", "http://localhost:8081")
	httpClient := &http.Client{Timeout: 3 * time.Second}

	router.GET("/health", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"status":          "ok",
			"service":         "order-api",
			"catalog_service": catalogURL,
		})
	})

	router.GET("/examples/quote", func(context *gin.Context) {
		context.JSON(http.StatusOK, exampleQuoteRequest())
	})

	router.POST("/orders/quote", func(context *gin.Context) {
		createQuote(context, httpClient, catalogURL)
	})

	router.Run(":8080")
}

func createQuote(context *gin.Context, httpClient *http.Client, catalogURL string) {
	var request contracts.QuoteRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := contracts.ValidateQuoteRequest(request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	catalogResponse, err := requestCatalogCheck(httpClient, catalogURL, request)
	if err != nil {
		context.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	deliveryPrice := calculateDeliveryPrice(request)
	total := deliveryPrice.Amount
	for _, item := range request.Items {
		total += item.Price.Amount * float64(item.Quantity)
	}

	status := "accepted"
	message := "quote was calculated successfully"
	if catalogResponse.Status != "available" {
		status = "needs_review"
		message = catalogResponse.Message
	}

	context.JSON(http.StatusOK, contracts.QuoteResponse{
		RequestID: request.RequestID,
		Status:    status,
		Message:   message,
		Customer:  request.Customer,
		Delivery: contracts.DeliveryCalculation{
			Method:        request.Preferences.Method,
			City:          request.DeliveryTo.City,
			Date:          request.Preferences.Date,
			TimeWindow:    request.Preferences.TimeWindow,
			AllowedZones:  request.Preferences.AllowedZones,
			DeliveryPrice: deliveryPrice,
		},
		Items: catalogResponse.Items,
		Total: contracts.Money{
			Amount:   total,
			Currency: "RUB",
		},
	})
}

func requestCatalogCheck(httpClient *http.Client, catalogURL string, request contracts.QuoteRequest) (contracts.CatalogCheckResponse, error) {
	payload := contracts.CatalogCheckRequest{
		RequestID: request.RequestID,
		City:      request.DeliveryTo.City,
		Items:     request.Items,
		Zones:     request.Preferences.AllowedZones,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return contracts.CatalogCheckResponse{}, fmt.Errorf("failed to encode catalog request: %w", err)
	}

	response, err := httpClient.Post(catalogURL+"/catalog/check", "application/json", bytes.NewReader(body))
	if err != nil {
		return contracts.CatalogCheckResponse{}, fmt.Errorf("catalog service is unavailable: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return contracts.CatalogCheckResponse{}, fmt.Errorf("catalog service returned status %d", response.StatusCode)
	}

	var catalogResponse contracts.CatalogCheckResponse
	if err := json.NewDecoder(response.Body).Decode(&catalogResponse); err != nil {
		return contracts.CatalogCheckResponse{}, fmt.Errorf("failed to decode catalog response: %w", err)
	}

	return catalogResponse, nil
}

func calculateDeliveryPrice(request contracts.QuoteRequest) contracts.Money {
	price := 300.0
	if request.Preferences.Method == "pickup" {
		price = 0
	}
	if request.Preferences.Method == "post" {
		price = 250
	}
	if request.Preferences.Fragile {
		price += 150
	}

	return contracts.Money{
		Amount:   price,
		Currency: "RUB",
	}
}

func exampleQuoteRequest() contracts.QuoteRequest {
	return contracts.QuoteRequest{
		RequestID: "REQ-2026-0001",
		Customer: contracts.Customer{
			Name:  "Артем Иванов",
			Email: "artem@example.com",
			Phone: "+79990001122",
		},
		DeliveryTo: contracts.Address{
			City:    "Москва",
			Street:  "Тверская",
			House:   "10",
			Comment: "Позвонить за 10 минут",
		},
		Items: []contracts.OrderItem{
			{
				SKU:      "BOOK-001",
				Name:     "Go для начинающих",
				Quantity: 2,
				Price: contracts.Money{
					Amount:   1200,
					Currency: "RUB",
				},
			},
			{
				SKU:      "MUG-404",
				Name:     "Кружка Debug Mode",
				Quantity: 1,
				Price: contracts.Money{
					Amount:   650,
					Currency: "RUB",
				},
			},
		},
		Preferences: contracts.DeliveryPreferences{
			Method:       "courier",
			Date:         time.Now().AddDate(0, 0, 2).Format("2006-01-02"),
			TimeWindow:   "day",
			Fragile:      true,
			AllowedZones: []string{"center", "north"},
		},
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
