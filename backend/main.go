package main

import (
	"fmt"
	"net/http"

	"smart-mt-backend/config"
	"smart-mt-backend/handlers"
)

func main() {
	// 1. Inisialisasi DB MySQL
	config.ConnectDatabase()

	// 2. Register Router API
	http.HandleFunc("/api/assets", handlers.HandleAssets)
	http.HandleFunc("/api/assets/", handlers.HandleAssets) // Handle /api/assets/{id}

	fmt.Println("🚀 Backend Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
