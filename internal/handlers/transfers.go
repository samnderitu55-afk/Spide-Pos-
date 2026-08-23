package handlers

import (
    "encoding/json"
    "net/http"
    "spide-pos/internal/db"
    "spide-pos/internal/middleware"
    "strconv"
    "time"
)

type TransferRequest struct {
    ToShopID     int    `json:"to_shop_id"`
    TransferDate string `json:"transfer_date"`
    Notes        string `json:"notes"`
    Items        []struct {
        ProductID int     `json:"product_id"`
        Quantity  int     `json:"quantity"`
        CostPrice float64 `json:"cost_price"`
    } `json:"items"`
}

func CreateTransferHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    var req TransferRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
        return
    }

    // Validate
    if req.ToShopID == 0 {
        http.Error(w, `{"error":"Destination shop is required"}`, http.StatusBadRequest)
        return
    }
    if len(req.Items) == 0 {
        http.Error(w, `{"error":"At least one item is required"}`, http.StatusBadRequest)
        return
    }

    // Use current shop as source
    fromShopID := claims.ShopID
    if fromShopID == 0 {
        fromShopID = 1
    }

    // Can't transfer to the same shop
    if fromShopID == req.ToShopID {
        http.Error(w, `{"error":"Cannot transfer to the same shop"}`, http.StatusBadRequest)
        return
    }

    // Parse transfer date or use today
    transferDate := time.Now().Format("2006-01-02")
    if req.TransferDate != "" {
        if parsed, err := time.Parse("2006-01-02", req.TransferDate); err == nil {
            transferDate = parsed.Format("2006-01-02")
        }
    }

    transfer := &db.StockTransfer{
        FromShopID:   fromShopID,
        ToShopID:     req.ToShopID,
        TransferDate: transferDate,
        Notes:        req.Notes,
        CreatedBy:    claims.Username,
        Status:       "completed",
        Items:        req.Items,
    }

    err := db.CreateTransfer(db.GetDB(), transfer)
    if err != nil {
        http.Error(w, `{"error":"Failed to create transfer: `+err.Error()+`"}`, http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Transfer created successfully",
        "transfer": transfer,
    })
}

func GetTransfersHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    // Get transfers for the user's shop (or all if director)
    shopID := claims.ShopID
    if shopID == 0 {
        shopID = 1
    }

    transfers, err := db.GetTransfers(db.GetDB(), shopID)
    if err != nil {
        http.Error(w, `{"error":"Failed to get transfers: `+err.Error()+`"}`, http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(transfers)
}

func GetTransferDetailHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    idStr := r.URL.Query().Get("id")
    if idStr == "" {
        http.Error(w, `{"error":"Transfer ID required"}`, http.StatusBadRequest)
        return
    }

    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, `{"error":"Invalid transfer ID"}`, http.StatusBadRequest)
        return
    }

    transfer, err := db.GetTransferDetail(db.GetDB(), id)
    if err != nil {
        http.Error(w, `{"error":"Failed to get transfer detail: `+err.Error()+`"}`, http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(transfer)
}
