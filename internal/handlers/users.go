package handlers

import (
    "encoding/json"
    "net/http"
    "spide-pos/internal/db"
    "spide-pos/internal/middleware"
    "strconv"
)

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    // Only directors and admins can view users
    if claims.Role != "director" && claims.Role != "admin" {
        http.Error(w, `{"error":"Access denied"}`, http.StatusForbidden)
        return
    }

    users, err := db.GetAllUsers(db.GetDB())
    if err != nil {
        http.Error(w, `{"error":"Failed to get users"}`, http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(users)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    // Only directors and admins can create users
    if claims.Role != "director" && claims.Role != "admin" {
        http.Error(w, `{"error":"Access denied"}`, http.StatusForbidden)
        return
    }

    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
        Name     string `json:"name"`
        Role     string `json:"role"`
        ShopID   int    `json:"shop_id"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
        return
    }

    // Create user object
    user := &db.User{
        Username: req.Username,
        Password: req.Password,
        Name:     req.Name,
        Role:     req.Role,
        ShopID:   req.ShopID,
        IsActive: true,
    }

    err := db.CreateUser(db.GetDB(), user)
    if err != nil {
        http.Error(w, `{"error":"Failed to create user: `+err.Error()+`"}`, http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(user)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    // Only directors and admins can update users
    if claims.Role != "director" && claims.Role != "admin" {
        http.Error(w, `{"error":"Access denied"}`, http.StatusForbidden)
        return
    }

    var req struct {
        ID       int    `json:"id"`
        Username string `json:"username"`
        Name     string `json:"name"`
        Role     string `json:"role"`
        ShopID   int    `json:"shop_id"`
        IsActive bool   `json:"is_active"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
        return
    }

    // Create user object
    user := &db.User{
        ID:       req.ID,
        Username: req.Username,
        Name:     req.Name,
        Role:     req.Role,
        ShopID:   req.ShopID,
        IsActive: req.IsActive,
    }

    err := db.UpdateUser(db.GetDB(), user)
    if err != nil {
        http.Error(w, `{"error":"Failed to update user: `+err.Error()+`"}`, http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(user)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    // Only directors and admins can delete users
    if claims.Role != "director" && claims.Role != "admin" {
        http.Error(w, `{"error":"Access denied"}`, http.StatusForbidden)
        return
    }

    // Get user ID from URL
    idStr := r.URL.Path[len("/api/users/delete/"):]
    if idStr == "" {
        http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
        return
    }

    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, `{"error":"Invalid user ID"}`, http.StatusBadRequest)
        return
    }

    err = db.DeleteUser(db.GetDB(), id)
    if err != nil {
        http.Error(w, `{"error":"Failed to delete user: `+err.Error()+`"}`, http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "User deleted successfully",
    })
}
