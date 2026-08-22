package middleware

import (
    "context"
    "encoding/json"
    "net/http"
    "strings"
    "spide-pos/internal/auth"
)

type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Public routes (no auth required)
        publicRoutes := []string{"/login", "/api/login", "/api/health", "/favicon.ico"}
        for _, route := range publicRoutes {
            if r.URL.Path == route {
                next(w, r)
                return
            }
        }

        // Try to get token from Authorization header first
        authHeader := r.Header.Get("Authorization")
        tokenString := ""

        if authHeader != "" {
            tokenString = strings.TrimPrefix(authHeader, "Bearer ")
            tokenString = strings.TrimSpace(tokenString)
        }

        // If no header, try to get from cookie
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
                    "error": "Unauthorized - No token provided",
                })
                return
            }
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }

        claims, err := auth.ValidateToken(tokenString)
        if err != nil {
            if strings.HasPrefix(r.URL.Path, "/api/") {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusUnauthorized)
                json.NewEncoder(w).Encode(map[string]string{
                    "error": "Invalid or expired token",
                })
                return
            }
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }

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

func RequireRole(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            claims := GetUserFromContext(r)
            if claims == nil {
                http.Redirect(w, r, "/login", http.StatusFound)
                return
            }

            for _, role := range roles {
                if claims.Role == role {
                    next(w, r)
                    return
                }
            }

            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusForbidden)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Insufficient permissions",
            })
        }
    }
}
