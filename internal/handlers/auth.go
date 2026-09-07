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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request format",
		})
		return
	}

	user, err := db.GetUserByUsername(db.GetDB(), credentials.Username)
	if err != nil {
		log.Printf("❌ Login error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid username or password",
		})
		return
	}

	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid username or password",
		})
		return
	}

	if !user.IsActive {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Account is disabled. Please contact administrator.",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		log.Printf("❌ Password mismatch for user: %s", credentials.Username)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid username or password",
		})
		return
	}

	// ✅ Update last login
	db.UpdateLastLogin(db.GetDB(), user.ID)

	// ✅ Generate token with CompanyID
	token, err := auth.GenerateToken(
		user.ID,
		user.Username,
		user.FullName,
		user.Role,
		user.ShopID,
		user.CompanyID, // ✅ Pass CompanyID
	)
	if err != nil {
		log.Printf("❌ Failed to generate token: %v", err)
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

	// Set user cookie
	userJSON, err := json.Marshal(map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"full_name":  user.FullName,
		"role":       user.Role,
		"shop_id":    user.ShopID,
		"company_id": user.CompanyID, // ✅ Include CompanyID
		"is_active":  user.IsActive,
	})
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
	})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "spide_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "spide_user",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"success": "true",
		"message": "Logged out successfully",
	})
}
