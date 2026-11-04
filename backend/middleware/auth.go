package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"smart-mt-backend/config"
	"smart-mt-backend/models"

	"github.com/golang-jwt/jwt/v5"
)

// jwtSecret is loaded from config in production. Here we borrow config.JWTSecret.
func GetCurrentUser(r *http.Request) (*models.AppUser, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, false
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, false
	}
	tokenStr := parts[1]

	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, false
	}

	var user models.AppUser
	if err := config.DB.First(&user, claims.UserID).Error; err != nil {
		return nil, false
	}
	return &user, true
}

// JWTClaims defines the payload stored in JWTs.
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// RequireAuth returns HTTP 401 if the request does not carry a valid JWT.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if _, ok := GetCurrentUser(r); !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Silakan masuk terlebih dahulu."})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireManager returns HTTP 403 if the authenticated user is not a Manager.
func RequireManager(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		user, ok := GetCurrentUser(r)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Silakan masuk terlebih dahulu."})
			return
		}
		if user.Role != "Manager" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Akses ditolak. Hanya Manager yang diperbolehkan."})
			return
		}
		next.ServeHTTP(w, r)
	})
}

