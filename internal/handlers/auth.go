package handlers

import (
    "encoding/json"
    "net/http"
    "spide-pos/internal/db"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var credentials struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Invalid request format",
        })
        return
    }

    // Get user from database
    user, err := db.GetUserByUsername(db.GetDB(), credentials.Username)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Invalid username or password",
        })
        return
    }

    // Check password
    if user.Password != credentials.Password {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Invalid username or password",
        })
        return
    }

    // Update last login
    db.UpdateLastLogin(db.GetDB(), user.ID)

    // Set session cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "spide_session",
        Value:    user.Username,
        Path:     "/",
        HttpOnly: true,
        Secure:   false,
        MaxAge:   86400,
        SameSite: http.SameSiteLaxMode,
    })

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "user":    user,
        "message": "Login successful",
    })
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
    // Clear session cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "spide_session",
        Value:    "",
        Path:     "/",
        HttpOnly: true,
        MaxAge:   -1,
    })

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "success": "true",
        "message": "Logged out successfully",
    })
}




