package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"spide-pos/internal/db"
	"spide-pos/internal/middleware"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

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

    users, err := db.GetAllUsers(db.GetDB())
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
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    claims := middleware.GetUserFromContext(r)
    if claims == nil || (claims.Role != "admin" && claims.Role != "director") {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    

    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
        Email    string `json:"email"`
        FullName string `json:"full_name"`
        Role     string `json:"role"`
        ShopID   int    `json:"shop_id"`
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

    // ✅ Log the shop assignment
    log.Printf("📝 Creating user: %s, Shop: %d, Role: %s", req.Username, req.ShopID, req.Role)

    err = db.CreateUser(db.GetDB(), req.Username, string(hashedPassword), 
        req.Email, req.FullName, req.Role, req.ShopID)
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
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    claims := middleware.GetUserFromContext(r)
    if claims == nil || (claims.Role != "admin" && claims.Role != "director") {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var req struct {
        ID       int    `json:"id"`
        Username string `json:"username"`
        Email    string `json:"email"`
        FullName string `json:"full_name"`
        Role     string `json:"role"`
        ShopID   int    `json:"shop_id"`
        IsActive bool   `json:"is_active"`
        Password string `json:"password,omitempty"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    var hashedPassword string
    if req.Password != "" {
        hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
        if err != nil {
            http.Error(w, "Failed to hash password", http.StatusInternalServerError)
            return
        }
        hashedPassword = string(hash)
    }

    err := db.UpdateUser(db.GetDB(), req.ID, req.Username, hashedPassword, 
        req.Email, req.FullName, req.Role, req.ShopID, req.IsActive)
    if err != nil {
        log.Printf("❌ Error updating user: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":  "success",
        "message": "User updated successfully",
    })
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

    // Prevent deleting yourself
    if id == claims.UserID {
        http.Error(w, "Cannot delete yourself", http.StatusBadRequest)
        return
    }

    err = db.DeleteUser(db.GetDB(), id)
    if err != nil {
        log.Printf("❌ Error deleting user: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":  "success",
        "message": "User deleted successfully",
    })
}