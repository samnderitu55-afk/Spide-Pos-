package handlers

import (
    "encoding/json"
    "net/http"
    "spide-pos/internal/db"
    "spide-pos/internal/middleware"
)

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
    claims := middleware.GetUserFromContext(r)
    if claims == nil || claims.Role != "director" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    users, err := db.GetAllUsers(db.GetDB())
    if err != nil {
        http.Error(w, `{"error":"Failed to load users"}`, http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    claims := middleware.GetUserFromContext(r)
    if claims == nil || claims.Role != "director" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    if r.Method != http.MethodPost {
        http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
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
        http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
        return
    }

    if req.Username == "" || req.Password == "" || req.Name == "" {
        http.Error(w, `{"error":"Username, password, and name are required"}`, http.StatusBadRequest)
        return
    }

    user, err := db.CreateUser(db.GetDB(), req.Username, req.Password, req.Name, req.Role, req.ShopID)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
    claims := middleware.GetUserFromContext(r)
    if claims == nil || claims.Role != "director" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    if r.Method != http.MethodPut {
        http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
        return
    }

    var req struct {
        ID     int    `json:"id"`
        Name   string `json:"name"`
        Role   string `json:"role"`
        ShopID int    `json:"shop_id"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
        return
    }

    if req.ID == 0 {
        http.Error(w, `{"error":"User ID is required"}`, http.StatusBadRequest)
        return
    }

    err := db.UpdateUser(db.GetDB(), req.ID, req.Name, req.Role, req.ShopID)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "User updated successfully",
    })
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
    claims := middleware.GetUserFromContext(r)
    if claims == nil || claims.Role != "director" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    if r.Method != http.MethodDelete && r.Method != http.MethodPost {
        http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
        return
    }

    var req struct {
        ID int `json:"id"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
        return
    }

    if req.ID == 0 {
        http.Error(w, `{"error":"User ID is required"}`, http.StatusBadRequest)
        return
    }

    err := db.DeleteUser(db.GetDB(), req.ID)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "User deleted successfully",
    })
}
