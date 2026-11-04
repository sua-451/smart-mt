package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"smart-mt-backend/config"
	"smart-mt-backend/handlers"
	"smart-mt-backend/models"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 1. Connect to MySQL
	config.ConnectDatabase()

	// 2. Auto-migrate tables that don't have GENERATED ALWAYS columns
	//    (kpi_before_after is managed by the SQL schema file due to generated cols)
	if err := config.DB.AutoMigrate(
		&models.AppUser{},
		&models.MasterAsset{},
		&models.AssetPhoto{},
		&models.KondisiAwal{},
		&models.WorkOrder{},
		&models.AppWorkOrder{},
		&models.AppReport{},
		&models.AppReportDetail{},
		&models.PmSchedule{},
		&models.HistoriKerusakan{},
		&models.Downtime{},
	); err != nil {
		log.Printf("⚠️ AutoMigrate error: %v", err)
	}

	// 3. Seed default users if the table is empty
	seedDefaultUsers()

	// Ensure uploads directory exists
	os.MkdirAll("./uploads", 0755)

	// 4. Build the HTTP mux
	mux := http.NewServeMux()

	// Handle all /api/ endpoints with CORS pre-flight
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		handlers.SetJSONHeaders(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		path := r.URL.Path

		switch {
		// Auth
		case path == "/api/auth/login":
			handlers.HandleLogin(w, r)
		case path == "/api/auth/me":
			handlers.HandleMe(w, r)

		// Dashboard
		case path == "/api/dashboard":
			handlers.HandleDashboard(w, r)

		// Assets
		case path == "/api/assets" || strings.HasPrefix(path, "/api/assets/"):
			handlers.HandleAssets(w, r)

		// Work Orders
		case path == "/api/work-orders" || strings.HasPrefix(path, "/api/work-orders/"):
			handlers.HandleWorkOrders(w, r)

		// Users
		case path == "/api/users":
			handlers.HandleUsers(w, r)

		// Technicians dropdown
		case path == "/api/technicians":
			handlers.HandleTechnicians(w, r)

		// Uploaded file serving
		case strings.HasPrefix(path, "/api/uploads/"):
			// Strip /api prefix and delegate
			r.URL.Path = strings.TrimPrefix(path, "/api")
			handlers.HandleUpload(w, r)

		default:
			handlers.WriteError(w, http.StatusNotFound, "Endpoint tidak ditemukan")
		}
	})

	// Serve uploaded files directly (non-/api path for convenience)
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("🚀 Backend Server berjalan di http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func seedDefaultUsers() {
	var count int64
	config.DB.Model(&models.AppUser{}).Count(&count)
	if count > 0 {
		return
	}

	defaults := []struct {
		name, role, username, password string
	}{
		{"Manager Demo", "Manager", "manager", "manager123"},
		{"Teknisi Demo", "Teknisi", "teknisi", "teknisi123"},
	}

	for _, d := range defaults {
		hash, err := bcrypt.GenerateFromPassword([]byte(d.password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Gagal hash password untuk %s: %v", d.username, err)
			continue
		}
		user := models.AppUser{
			FullName:     d.name,
			Role:         d.role,
			Username:     d.username,
			PasswordHash: string(hash),
		}
		if err := config.DB.Create(&user).Error; err != nil {
			log.Printf("Gagal membuat user %s: %v", d.username, err)
		} else {
			fmt.Printf("✅ User demo dibuat: %s / %s\n", d.username, d.password)
		}
	}
}
