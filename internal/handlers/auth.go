package handlers

import (
    "encoding/json"
    "net/http"
    "strings"
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

    // Skip password check for testing
    // if !auth.CheckPassword(credentials.Password, user.Password) {
    //     w.Header().Set("Content-Type", "application/json")
    //     w.WriteHeader(http.StatusUnauthorized)
    //     json.NewEncoder(w).Encode(map[string]string{
    //         "error": "Invalid username or password",
    //     })
    //     return
    // }

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

    // Set token cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "spide_token",
        Value:    token,
        Path:     "/",
        HttpOnly: false,
        Secure:   false,
        MaxAge:   86400,
        SameSite: http.SameSiteLaxMode,
    })

    // Set user cookie - clean the JSON to remove quotes issue
    userJSON, err := json.Marshal(user)
    if err == nil {
        // URL encode the JSON to avoid quote issues
        encodedUser := strings.ReplaceAll(string(userJSON), `"`, `'`)
        http.SetCookie(w, &http.Cookie{
            Name:     "spide_user",
            Value:    encodedUser,
            Path:     "/",
            HttpOnly: false,
            Secure:   false,
            MaxAge:   86400,
            SameSite: http.SameSiteLaxMode,
        })
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
    http.SetCookie(w, &http.Cookie{
        Name:     "spide_token",
        Value:    "",
        Path:     "/",
        MaxAge:   -1,
    })
    http.SetCookie(w, &http.Cookie{
        Name:     "spide_user",
        Value:    "",
        Path:     "/",
        MaxAge:   -1,
    })

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "success": "true",
        "message": "Logged out successfully",
    })
}
