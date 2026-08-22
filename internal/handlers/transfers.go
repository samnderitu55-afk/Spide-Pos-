package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "spide-pos/internal/db"
)

func CreateTransferHandler(w http.ResponseWriter, r *http.Request) {
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

    fromShopID := 1 // Default shop

    var req db.StockTransferRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
        return
    }

    if len(req.Items) == 0 {
        http.Error(w, `{"error":"At least one item is required"}`, http.StatusBadRequest)
        return
    }
    if req.ToShopID == 0 {
        http.Error(w, `{"error":"Destination shop is required"}`, http.StatusBadRequest)
        return
    }
    if req.ToShopID == fromShopID {
        http.Error(w, `{"error":"Cannot transfer to the same shop"}`, http.StatusBadRequest)
        return
    }

    transfer, err := db.CreateStockTransfer(db.GetDB(), &req, fromShopID, username)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    items, err := db.GetStockTransferItems(db.GetDB(), transfer.ID)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "transfer": transfer,
        "items":    items,
    })
}

func GetTransfersHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    username := GetUsernameFromCookie(r)
    if username == "" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    shopID := 1 // Default shop
    status := r.URL.Query().Get("status")
    transfers, err := db.GetTransfersByShop(db.GetDB(), shopID, status)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if transfers == nil {
        transfers = []db.StockTransfer{}
    }
    json.NewEncoder(w).Encode(transfers)
}

func GetTransferDetailHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    username := GetUsernameFromCookie(r)
    if username == "" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    transferIDStr := r.URL.Query().Get("id")
    if transferIDStr == "" {
        http.Error(w, `{"error":"Transfer ID is required"}`, http.StatusBadRequest)
        return
    }
    transferID, err := strconv.Atoi(transferIDStr)
    if err != nil {
        http.Error(w, `{"error":"Invalid transfer ID"}`, http.StatusBadRequest)
        return
    }
    transfer, err := db.GetStockTransfer(db.GetDB(), transferID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    items, err := db.GetStockTransferItems(db.GetDB(), transferID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(map[string]interface{}{
        "transfer": transfer,
        "items":    items,
    })
}




