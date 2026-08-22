package handlers

import (
    "encoding/json"
    "net/http"
    "spide-pos/internal/db"
)

func CreatePurchaseHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    if r.Method != http.MethodPost {
        http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
        return
    }

    username := GetUsernameFromCookie(r)
    if username == "" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    // For now, allow all authenticated users
    // Later we'll add role-based checks

    var req db.PurchaseRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
        return
    }

    if len(req.Items) == 0 {
        http.Error(w, `{"error":"At least one item is required"}`, http.StatusBadRequest)
        return
    }
    if req.SupplierID == 0 {
        http.Error(w, `{"error":"Supplier is required"}`, http.StatusBadRequest)
        return
    }

    purchase, err := db.CreatePurchase(db.GetDB(), &req, username)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    items, err := db.GetPurchaseItems(db.GetDB(), purchase.ID)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "purchase": purchase,
        "items":    items,
    })
}

func GetSuppliersHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    username := GetUsernameFromCookie(r)
    if username == "" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    suppliers, err := db.GetAllSuppliers(db.GetDB())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if suppliers == nil {
        suppliers = []db.Supplier{}
    }
    json.NewEncoder(w).Encode(suppliers)
}

func CreateSupplierHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    if r.Method != http.MethodPost {
        http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
        return
    }

    username := GetUsernameFromCookie(r)
    if username == "" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    var supplier db.Supplier
    if err := json.NewDecoder(r.Body).Decode(&supplier); err != nil {
        http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
        return
    }

    if supplier.Name == "" {
        http.Error(w, `{"error":"Supplier name is required"}`, http.StatusBadRequest)
        return
    }

    id, err := db.CreateSupplier(db.GetDB(), &supplier)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    supplier.ID = int(id)
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(supplier)
}




