package middleware

import (
    "net/http"
)

// SimpleAuth checks if user is logged in via cookie/session
func SimpleAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Public routes
        if r.URL.Path == "/login" || r.URL.Path == "/api/login" || r.URL.Path == "/api/health" || r.URL.Path == "/favicon.ico" {
            next(w, r)
            return
        }

        // Check for session cookie
        cookie, err := r.Cookie("spide_session")
        if err != nil || cookie.Value == "" {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }

        // For demo, just check if cookie exists
        // In production, validate session from database
        next(w, r)
    }
}

// GetUserFromContext retrieves user from context (placeholder for compatibility)
func GetUserFromContext(r *http.Request) interface{} {
    // Since we're using cookies, we don't store user in context
    // This is just for compatibility with existing handlers
    return nil
}

