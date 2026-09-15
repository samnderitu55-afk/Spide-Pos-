package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"spide-pos/internal/auth"
	"spide-pos/internal/db"

	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔵 LoginHandler hit: %s %s content-type=%s",
		r.Method, r.URL.Path, r.Header.Get("Content-Type"))

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

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		log.Printf("❌ Password mismatch for user: %s", credentials.Username)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid username or password",
		})
		return
	}

	// ---------- Look up shop name ----------
	// Use this value in the response below; if the shop is missing we just log a warning.
	shopName := ""
	if user.ShopID > 0 {
		if err := db.GetDB().
			QueryRow("SELECT name FROM shops WHERE id = ?", user.ShopID).
			Scan(&shopName); err != nil && err != sql.ErrNoRows {
			log.Printf("⚠️  Could not load shop name for shop_id=%d: %v", user.ShopID, err)
		}
	}

	// ---------- Look up company name ----------
	companyName := ""
	if user.CompanyID > 0 {
		if err := db.GetDB().
			QueryRow("SELECT company_name FROM companies WHERE id = ?", user.CompanyID).
			Scan(&companyName); err != nil && err != sql.ErrNoRows {
			log.Printf("⚠️  Could not load company name for company_id=%d: %v", user.CompanyID, err)
		}
	}

	db.UpdateLastLogin(db.GetDB(), user.ID)

	// ---------- Generate token ----------
	// Signature: GenerateToken(userID int, username, fullName, role string, shopID, companyID int)
	token, err := auth.GenerateToken(
		user.ID,
		user.Username,
		user.FullName,
		user.Role,
		user.ShopID,
		user.CompanyID,
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

	// ---------- Build payload once, reuse for cookie + JSON response ----------
	userPayload := map[string]interface{}{
		"id":           user.ID,
		"username":     user.Username,
		"full_name":    user.FullName,
		"role":         user.Role,
		"shop_id":      user.ShopID,
		"shop_name":    shopName,
		"company_id":   user.CompanyID,
		"company_name": companyName,
		"is_active":    user.IsActive,
	}

	// ---------- Token cookie ----------
	http.SetCookie(w, &http.Cookie{
		Name:     "spide_token",
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		Secure:   false,
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
	})

	// ---------- User cookie (URL-encoded JSON) ----------
	userJSON, err := json.Marshal(userPayload)
	if err != nil {
		log.Printf("⚠️  Could not marshal user payload: %v", err)
	} else {
		http.SetCookie(w, &http.Cookie{
			Name:     "spide_user",
			Value:    url.QueryEscape(string(userJSON)),
			Path:     "/",
			HttpOnly: false,
			Secure:   false,
			MaxAge:   86400,
			SameSite: http.SameSiteLaxMode,
		})
	}

	// ---------- JSON response ----------
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user":    userPayload,
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
