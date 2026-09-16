package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"spide-pos/internal/auth"
	"spide-pos/internal/db"
	"spide-pos/internal/middleware"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

// internal/handlers/users.go

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admin and director can see users
	if claims.Role != "admin" && claims.Role != "director" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// ✅ Get users for this company only
	users, err := db.GetUsersByCompany(db.GetDB(), claims.CompanyID)
	if err != nil {
		log.Printf("❌ Error getting users: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil || (claims.Role != "admin" && claims.Role != "director") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		Email     string `json:"email"`
		FullName  string `json:"full_name"`
		Role      string `json:"role"`
		ShopID    int    `json:"shop_id"`
		CompanyID int    `json:"company_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Always use the authenticated user's company
	companyID := claims.CompanyID
	if companyID == 0 {
		http.Error(w, `{"error":"No company assigned"}`, http.StatusBadRequest)
		return
	}

	// ✅ Pass companyID to CreateUser
	err = db.CreateUser(db.GetDB(), req.Username, string(hashedPassword),
		req.Email, req.FullName, req.Role, req.ShopID, companyID)
	if err != nil {
		log.Printf("❌ Error creating user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "User created successfully",
	})
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil || (claims.Role != "admin" && claims.Role != "director") {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		ID       int     `json:"id"`
		Username *string `json:"username"`
		Password *string `json:"password"`
		Email    *string `json:"email"`
		FullName *string `json:"full_name"`
		Role     *string `json:"role"`
		ShopID   *int    `json:"shop_id"`
		IsActive *bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.ID == 0 {
		writeJSONError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Load current values
	current, err := db.GetUserByID(db.GetDB(), req.ID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	if current.CompanyID != claims.CompanyID {
		writeJSONError(w, http.StatusForbidden, "Cannot update users from another company")
		return
	}

	// Merge — apply only what was sent
	username := current.Username
	fullName := current.FullName
	email := current.Email
	role := current.Role
	shopID := current.ShopID
	isActive := current.IsActive
	password := ""

	if req.Username != nil {
		username = *req.Username
	}
	if req.FullName != nil {
		fullName = *req.FullName
	}
	if req.Email != nil {
		email = *req.Email
	}
	if req.Role != nil {
		role = *req.Role
	}
	if req.ShopID != nil {
		shopID = *req.ShopID
	}
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	if req.Password != nil && *req.Password != "" {
		hashed, err := auth.HashPassword(*req.Password)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Password hashing failed")
			return
		}
		password = hashed
	}

	if err := db.UpdateUser(db.GetDB(), req.ID, username, password, email, fullName, role, shopID, claims.CompanyID, isActive); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"id":      req.ID,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil || (claims.Role != "admin" && claims.Role != "director") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// ✅ Prevent deleting yourself
	if id == claims.UserID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Cannot delete your own account"})
		return
	}

	// ✅ Prevent deleting the last admin/director
	var adminCount int
	_ = db.GetDB().QueryRow(
		`SELECT COUNT(*) FROM users 
         WHERE company_id = ? AND role IN ('admin','director') 
         AND is_active = 1 AND id != ?`,
		claims.CompanyID, id,
	).Scan(&adminCount)
	if adminCount == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Cannot delete the last admin"})
		return
	}

	if err := db.DeleteUser(db.GetDB(), id, claims.CompanyID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"success": "true"})
}
