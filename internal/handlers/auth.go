package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"spide-pos/internal/auth"
	"spide-pos/internal/db"
	"strings"

	"golang.org/x/crypto/bcrypt"
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
        log.Printf("❌ Failed to decode login request: %v", err)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Invalid request format",
        })
        return
    }

    log.Printf("🔐 Login attempt - Username: %s", credentials.Username)

    // ✅ Get user from database
    user, err := db.GetUserByUsername(db.GetDB(), credentials.Username)
    if err != nil {
        log.Printf("❌ Database error for user %s: %v", credentials.Username, err)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Invalid username or password",
        })
        return
    }

    // ✅ Check if user exists
    if user == nil {
        log.Printf("❌ User not found: %s", credentials.Username)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Invalid username or password",
        })
        return
    }

    log.Printf("✅ User found: %s, Role: %s, Active: %v", user.Username, user.Role, user.IsActive)

    // ✅ Check if user is active
    if !user.IsActive {
        log.Printf("❌ User account disabled: %s", credentials.Username)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Account is disabled. Please contact administrator.",
        })
        return
    }

    // ✅ VERIFY PASSWORD
    log.Printf("🔍 Verifying password for user: %s", credentials.Username)
    log.Printf("   Stored hash length: %d", len(user.Password))
    
    err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
    if err != nil {
        log.Printf("❌ Password mismatch for user: %s - Error: %v", credentials.Username, err)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Invalid username or password",
        })
        return
    }

    log.Printf("✅ Password verified for user: %s", credentials.Username)

    // ✅ Update last login
    db.UpdateLastLogin(db.GetDB(), user.ID)

    // ✅ Generate token
    token, err := auth.GenerateToken(
        user.ID,
        user.Username,
        user.FullName,
        user.Role,
        user.ShopID,
        user.ShopName,
    )
    if err != nil {
        log.Printf("❌ Failed to generate token for user %s: %v", credentials.Username, err)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Failed to generate token",
        })
        return
    }

    log.Printf("✅ Login successful for user: %s", credentials.Username)

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
        "shop_id": user.ShopID,
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

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Hash the new password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Failed to hash password", http.StatusInternalServerError)
        return
    }

    // Update the password
    _, err = db.GetDB().Exec(
        "UPDATE users SET password = ? WHERE username = ?",
        string(hashedPassword), req.Username,
    )
    if err != nil {
        http.Error(w, "Failed to update password", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Password reset successfully",
    })
}