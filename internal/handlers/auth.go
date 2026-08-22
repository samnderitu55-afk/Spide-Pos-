package handlers

import (
    "encoding/json"
    "net/http"
    "spide-pos/internal/auth"
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

    user, err := db.GetUserByUsername(db.GetDB(), credentials.Username)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Invalid username or password",
        })
        return
    }

    // Password check enabled
    if !auth.CheckPassword(credentials.Password, user.Password) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Invalid username or password",
        })
        return
    }

    db.UpdateLastLogin(db.GetDB(), user.ID)

    token, err := auth.GenerateToken(
        user.ID,
        user.Username,
        user.Name,
        user.Role,
        user.ShopID,
        user.ShopName,
    )
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Failed to generate token",
        })
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "user":    user,
        "token":   token,
        "message": "Login successful",
    })
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "success": "true",
        "message": "Logged out successfully",
    })
}
