package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"smart-mt-backend/config"
	"smart-mt-backend/models"
)

// Helper untuk set response JSON & CORS
func setJSONHeader(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
}

// Handler utama /api/assets
func HandleAssets(w http.ResponseWriter, r *http.Request) {
	setJSONHeader(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parsing ID dari URL (Contoh: /api/assets/5)
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var id int
	if len(pathParts) > 2 {
		id, _ = strconv.Atoi(pathParts[2])
	}

	switch r.Method {
	case http.MethodGet:
		if id > 0 {
			getAssetByID(w, r, uint(id))
		} else {
			getAllAssets(w, r)
		}
	case http.MethodPost:
		createAsset(w, r)
	case http.MethodPut:
		if id > 0 {
			updateAsset(w, r, uint(id))
		} else {
			http.Error(w, `{"error":"ID Asset diperlukan"}`, http.StatusBadRequest)
		}
	case http.MethodDelete:
		if id > 0 {
			deleteAsset(w, r, uint(id))
		} else {
			http.Error(w, `{"error":"ID Asset diperlukan"}`, http.StatusBadRequest)
		}
	default:
		http.Error(w, `{"error":"Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
	}
}

func getAllAssets(w http.ResponseWriter, r *http.Request) {
	var assets []models.MasterAsset
	config.DB.Order("id_asset desc").Find(&assets)
	json.NewEncoder(w).Encode(assets)
}

func getAssetByID(w http.ResponseWriter, r *http.Request, id uint) {
	var asset models.MasterAsset
	if err := config.DB.First(&asset, id).Error; err != nil {
		http.Error(w, `{"error":"Data aset tidak ditemukan"}`, http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(asset)
}

func createAsset(w http.ResponseWriter, r *http.Request) {
	var asset models.MasterAsset
	if err := json.NewDecoder(r.Body).Decode(&asset); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}
	config.DB.Create(&asset)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(asset)
}

func updateAsset(w http.ResponseWriter, r *http.Request, id uint) {
	var asset models.MasterAsset
	if err := config.DB.First(&asset, id).Error; err != nil {
		http.Error(w, `{"error":"Data aset tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	var updatedData models.MasterAsset
	json.NewDecoder(r.Body).Decode(&updatedData)

	config.DB.Model(&asset).Updates(updatedData)
	json.NewEncoder(w).Encode(asset)
}

func deleteAsset(w http.ResponseWriter, r *http.Request, id uint) {
	if err := config.DB.Delete(&models.MasterAsset{}, id).Error; err != nil {
		http.Error(w, `{"error":"Gagal menghapus aset"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Aset berhasil dihapus"})
}