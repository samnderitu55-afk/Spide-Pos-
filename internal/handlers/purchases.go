package handlers

import (
    "encoding/json"
    "net/http"
    "spide-pos/internal/db"
    "spide-pos/internal/middleware"
    "strconv"
    "time"
)

type PurchaseRequest struct {
    SupplierID   int    `json:"supplier_id"`
    PurchaseDate string `json:"purchase_date"`
    Notes        string `json:"notes"`
    Items        []struct {
        ProductID int     `json:"product_id"`
        Quantity  int     `json:"quantity"`
        CostPrice float64 `json:"cost_price"`
    } `json:"items"`
}

func CreatePurchaseHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    var req PurchaseRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
        return
    }

    if req.SupplierID == 0 {
        http.Error(w, `{"error":"Supplier is required"}`, http.StatusBadRequest)
        return
    }
    if len(req.Items) == 0 {
        http.Error(w, `{"error":"At least one item is required"}`, http.StatusBadRequest)
        return
    }

    shopID := claims.ShopID
    if shopID == 0 {
        shopID = 1
    }

    purchaseDate := time.Now().Format("2006-01-02")
    if req.PurchaseDate != "" {
        if parsed, err := time.Parse("2006-01-02", req.PurchaseDate); err == nil {
            purchaseDate = parsed.Format("2006-01-02")
        }
    }

    purchase := &db.Purchase{
        SupplierID:   req.SupplierID,
        PurchaseDate: purchaseDate,
        Notes:        req.Notes,
        CreatedBy:    claims.Username,
        ShopID:       shopID,
        Items:        req.Items,
    }

    err := db.CreatePurchase(db.GetDB(), purchase)
    if err != nil {
        http.Error(w, `{"error":"Failed to create purchase: `+err.Error()+`"}`, http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success":  true,
        "message":  "Purchase created successfully",
        "purchase": purchase,
    })
}

func GetSuppliersHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    suppliers, err := db.GetSuppliers(db.GetDB())
    if err != nil {
        http.Error(w, `{"error":"Failed to get suppliers: `+err.Error()+`"}`, http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(suppliers)
}

func CreateSupplierHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    var req db.Supplier
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
        return
    }

    if req.Name == "" {
        http.Error(w, `{"error":"Supplier name is required"}`, http.StatusBadRequest)
        return
    }

    err := db.CreateSupplier(db.GetDB(), &req)
    if err != nil {
        http.Error(w, `{"error":"Failed to create supplier: `+err.Error()+`"}`, http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success":  true,
        "message":  "Supplier created successfully",
        "supplier": req,
    })
}

func PurchaseReportHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    startDate := r.URL.Query().Get("start_date")
    endDate := r.URL.Query().Get("end_date")
    supplierID := 0
    
    if supplierIDParam := r.URL.Query().Get("supplier_id"); supplierIDParam != "" {
        if id, err := strconv.Atoi(supplierIDParam); err == nil {
            supplierID = id
        }
    }

    if startDate == "" {
        now := time.Now()
        startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
    }
    if endDate == "" {
        endDate = time.Now().Format("2006-01-02")
    }

    purchases, total, err := db.GetPurchaseReport(db.GetDB(), startDate, endDate, supplierID)
    if err != nil {
        http.Error(w, `{"error":"Failed to get purchase report: `+err.Error()+`"}`, http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "purchases":  purchases,
        "total":      total,
        "count":      len(purchases),
        "start_date": startDate,
        "end_date":   endDate,
    })
}

func GetPurchaseItemsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    idStr := r.URL.Query().Get("id")
    if idStr == "" {
        http.Error(w, `{"error":"Purchase ID required"}`, http.StatusBadRequest)
        return
    }

    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, `{"error":"Invalid purchase ID"}`, http.StatusBadRequest)
        return
    }

    items, err := db.GetPurchaseItems(db.GetDB(), id)
    if err != nil {
        http.Error(w, `{"error":"Failed to get purchase items: `+err.Error()+`"}`, http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(items)
}
