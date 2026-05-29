package main

import (
	"net/http"

	"lab10-medium-gin/internal/contracts"

	"github.com/gin-gonic/gin"
)

var catalog = map[string]contracts.CatalogItem{
	"BOOK-001": {
		SKU:         "BOOK-001",
		Name:        "Go для начинающих",
		Available:   12,
		WeightGrams: 420,
		Tags:        []string{"book", "programming", "go"},
	},
	"MUG-404": {
		SKU:         "MUG-404",
		Name:        "Кружка Debug Mode",
		Available:   5,
		WeightGrams: 310,
		Tags:        []string{"mug", "gift"},
	},
	"USB-128": {
		SKU:         "USB-128",
		Name:        "USB Flash 128 GB",
		Available:   20,
		WeightGrams: 40,
		Tags:        []string{"electronics", "storage"},
	},
}

func main() {
	router := gin.Default()

	router.GET("/health", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "catalog-service",
		})
	})

	router.POST("/catalog/check", checkCatalog)

	router.Run(":8081")
}

func checkCatalog(context *gin.Context) {
	var request contracts.CatalogCheckRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := contracts.ValidateCatalogCheckRequest(request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items := make([]contracts.CatalogItem, 0, len(request.Items))
	missing := make([]string, 0)

	for _, requestedItem := range request.Items {
		catalogItem, exists := catalog[requestedItem.SKU]
		if !exists || catalogItem.Available < requestedItem.Quantity {
			missing = append(missing, requestedItem.SKU)
			continue
		}
		items = append(items, catalogItem)
	}

	status := "available"
	message := "all requested items are available"
	if len(missing) > 0 {
		status = "partial"
		message = "some requested items are missing or not enough in stock"
	}

	context.JSON(http.StatusOK, contracts.CatalogCheckResponse{
		RequestID: request.RequestID,
		Status:    status,
		Message:   message,
		Items:     items,
		Missing:   missing,
	})
}
