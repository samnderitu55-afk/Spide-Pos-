package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"spide-pos/internal/auth"
	"strings"
)

type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🔐 MW: %s %s", r.Method, r.URL.Path)
		publicRoutes := []string{"/login", "/api/login", "/api/health", "/favicon.ico", "/static/"}
		for _, route := range publicRoutes {
			if r.URL.Path == route || strings.HasPrefix(r.URL.Path, route) {
				next(w, r)
				return
			}
		}
		authHeader := r.Header.Get("Authorization")
		tokenString := ""

		if authHeader != "" {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			tokenString = strings.TrimSpace(tokenString)
		}

		if tokenString == "" {
			cookie, err := r.Cookie("spide_token")
			if err == nil && cookie.Value != "" {
				tokenString = cookie.Value
			}
		}

		if tokenString == "" {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Unauthorized",
				})
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// ✅ Validate token - returns Claims with CompanyID
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Invalid token",
				})
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// ✅ Store claims with CompanyID in context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

func GetUserFromContext(r *http.Request) *auth.Claims {
	if claims, ok := r.Context().Value(UserContextKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}
