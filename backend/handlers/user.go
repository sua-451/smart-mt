package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"

	"smart-mt-backend/config"
	"smart-mt-backend/middleware"
	"smart-mt-backend/models"

	"golang.org/x/crypto/bcrypt"
)

// HandleUsers routes GET/POST /api/users.
func HandleUsers(w http.ResponseWriter, r *http.Request) {
	SetJSONHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	user, ok := middleware.GetCurrentUser(r)
	if !ok || user.Role != "Manager" {
		WriteError(w, http.StatusForbidden, "Hanya Manager yang dapat mengelola pengguna")
		return
	}

	switch r.Method {
	case http.MethodGet:
		listUsers(w, r)
	case http.MethodPost:
		createUser(w, r)
	default:
		WriteError(w, http.StatusMethodNotAllowed, "Method tidak diizinkan")
	}
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.UserWithCount
	config.DB.Raw(`
		SELECT u.user_id, u.full_name, u.role, u.username, u.created_at,
		       COUNT(a.wo_number) AS assigned_count
		FROM app_users u
		LEFT JOIN app_work_order a ON a.technician_id = u.user_id
		  AND a.workflow_status IN ('Ditugaskan','Dikerjakan','Perlu Revisi')
		GROUP BY u.user_id
		ORDER BY u.full_name
	`).Scan(&users)
	WriteJSON(w, http.StatusOK, users)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FullName string `json:"full_name"`
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Format permintaan tidak valid")
		return
	}
	req.FullName = strings.TrimSpace(req.FullName)
	req.Username = strings.TrimSpace(req.Username)

	if req.FullName == "" || req.Username == "" || req.Password == "" {
		WriteError(w, http.StatusBadRequest, "Nama, username, dan kata sandi wajib diisi.")
		return
	}
	if utf8.RuneCountInString(req.Password) < 8 {
		WriteError(w, http.StatusBadRequest, "Kata sandi minimal 8 karakter.")
		return
	}
	if req.Role != "Manager" && req.Role != "Teknisi" {
		WriteError(w, http.StatusBadRequest, "Peran harus Manager atau Teknisi.")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Gagal mengenkripsi kata sandi")
		return
	}

	newUser := models.AppUser{
		FullName:     req.FullName,
		Username:     req.Username,
		PasswordHash: string(hash),
		Role:         req.Role,
	}
	if err := config.DB.Create(&newUser).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			WriteError(w, http.StatusConflict, "Username sudah digunakan.")
		} else {
			WriteError(w, http.StatusInternalServerError, "Gagal membuat akun")
		}
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]string{"message": "Akun berhasil dibuat."})
}

// HandleTechnicians returns GET /api/technicians – just a slim list for dropdowns.
func HandleTechnicians(w http.ResponseWriter, r *http.Request) {
	SetJSONHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	var technicians []models.AppUser
	config.DB.Where("role = 'Teknisi'").Order("full_name").Find(&technicians)
	type slim struct {
		UserID   uint   `json:"user_id"`
		FullName string `json:"full_name"`
	}
	result := make([]slim, len(technicians))
	for i, t := range technicians {
		result[i] = slim{UserID: t.UserID, FullName: t.FullName}
	}
	WriteJSON(w, http.StatusOK, result)
}
