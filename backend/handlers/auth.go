package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"smart-mt-backend/config"
	"smart-mt-backend/middleware"
	"smart-mt-backend/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest is the JSON body for POST /api/auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is returned on successful login.
type LoginResponse struct {
	Token    string          `json:"token"`
	User     publicUser      `json:"user"`
}

type publicUser struct {
	UserID   uint   `json:"user_id"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	Username string `json:"username"`
}

// HandleLogin handles POST /api/auth/login.
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	SetJSONHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "Method tidak diizinkan")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Format permintaan tidak valid")
		return
	}

	var user models.AppUser
	if err := config.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		WriteError(w, http.StatusUnauthorized, "Username atau kata sandi tidak sesuai.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		WriteError(w, http.StatusUnauthorized, "Username atau kata sandi tidak sesuai.")
		return
	}

	claims := &middleware.JWTClaims{
		UserID:   user.UserID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Gagal membuat token autentikasi")
		return
	}

	WriteJSON(w, http.StatusOK, LoginResponse{
		Token: tokenStr,
		User: publicUser{
			UserID:   user.UserID,
			FullName: user.FullName,
			Role:     user.Role,
			Username: user.Username,
		},
	})
}

// HandleMe handles GET /api/auth/me – returns the currently authenticated user.
func HandleMe(w http.ResponseWriter, r *http.Request) {
	SetJSONHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	user, ok := middleware.GetCurrentUser(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "Tidak terautentikasi")
		return
	}
	WriteJSON(w, http.StatusOK, publicUser{
		UserID:   user.UserID,
		FullName: user.FullName,
		Role:     user.Role,
		Username: user.Username,
	})
}
