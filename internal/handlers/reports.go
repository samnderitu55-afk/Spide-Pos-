package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "spide-pos/internal/db"
    "spide-pos/internal/middleware"
)

type LowStockItem struct {
    ProductName   string  `json:"product_name"`
    Category      string  `json:"category"`
    StockQuantity int     `json:"stock_quantity"`
    ReorderLevel  int     `json:"reorder_level"`
    CostPrice     float64 `json:"cost_price"`
    RestockCost   float64 `json:"restock_cost"`
}

// ZReportHandler - Daily Z-Report
func ZReportHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    dateParam := r.URL.Query().Get("date")
    shopID := claims.ShopID
    if shopID == 0 {
        shopID = 1
    }
    
    report, err := db.GetDailyZReportWithExpenses(db.GetDB(), dateParam, shopID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(report)
}

// ProductSalesReportHandler - Product Sales Report
func ProductSalesReportHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    startDate := r.URL.Query().Get("start")
    endDate := r.URL.Query().Get("end")
    shopID := claims.ShopID
    if shopID == 0 {
        shopID = 1
    }
    
    report, err := db.GetProductSalesReport(db.GetDB(), startDate, endDate, shopID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if report == nil {
        report = []db.ProductSalesReportItem{}
    }
    json.NewEncoder(w).Encode(report)
}

// LowStockReportHandler - Low Stock Alerts
func LowStockReportHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    thresholdStr := r.URL.Query().Get("threshold")
    threshold := 5
    if thresholdStr != "" {
        if val, err := strconv.Atoi(thresholdStr); err == nil && val >= 0 {
            threshold = val
        }
    }
    
    shopID := claims.ShopID
    if shopID == 0 {
        shopID = 1
    }
    
    query := `
        SELECT 
            p.name,
            p.category,
            COALESCE(ss.quantity, 0) as stock_qty,
            p.reorder_level,
            p.cost_price,
            (p.reorder_level - COALESCE(ss.quantity, 0)) * p.cost_price as restock_cost
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = ?
        WHERE p.is_active = 1 
          AND COALESCE(ss.quantity, 0) <= ?
        ORDER BY stock_qty ASC
        LIMIT 50
    `
    
    rows, err := db.GetDB().Query(query, shopID, threshold)
    if err != nil {
        http.Error(w, `{"error":"Database error: `+err.Error()+`"}`, http.StatusInternalServerError)
        return
    }
    defer rows.Close()
    
    var items []LowStockItem
    for rows.Next() {
        var item LowStockItem
        err := rows.Scan(
            &item.ProductName,
            &item.Category,
            &item.StockQuantity,
            &item.ReorderLevel,
            &item.CostPrice,
            &item.RestockCost,
        )
        if err != nil {
            continue
        }
        items = append(items, item)
    }
    
    if items == nil {
        items = []LowStockItem{}
    }
    
    json.NewEncoder(w).Encode(items)
}

// InventoryValuationHandler - Inventory Valuation Report
func InventoryValuationHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    report, err := db.GetInventoryValuationReport(db.GetDB())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if report == nil {
        report = []db.InventoryValuationItem{}
    }
    json.NewEncoder(w).Encode(report)
}

// CategoryDrilldownHandler - Category Drilldown for Valuation
func CategoryDrilldownHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    category := r.URL.Query().Get("category")
    if category == "" {
        http.Error(w, "Category parameter is required", http.StatusBadRequest)
        return
    }
    
    details, err := db.GetCategoryValuationDetails(db.GetDB(), category)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if details == nil {
        details = []struct {
            ProductName string  `json:"product_name"`
            InStock     int     `json:"in_stock"`
            CostPrice   float64 `json:"cost_price"`
            RetailPrice float64 `json:"retail_price"`
            TotalCost   float64 `json:"total_cost"`
        }{}
    }
    json.NewEncoder(w).Encode(details)
}
