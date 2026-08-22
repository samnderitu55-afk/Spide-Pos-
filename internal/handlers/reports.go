package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"
    "spide-pos/internal/db"
)

func ZReportHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    username := GetUsernameFromCookie(r)
    if username == "" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    dateParam := r.URL.Query().Get("date")
    shopID := 1 // Default shop
    report, err := db.GetDailyZReportWithExpenses(db.GetDB(), dateParam, shopID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(report)
}

func ProductSalesReportHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    username := GetUsernameFromCookie(r)
    if username == "" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    startDate := r.URL.Query().Get("start")
    endDate := r.URL.Query().Get("end")
    shopID := 1 // Default shop
    report, err := db.GetProductSalesReport(db.GetDB(), startDate, endDate, shopID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if report == nil {
        report = []db.ProductSalesReportItem{}
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(report)
}

func LowStockReportHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    username := GetUsernameFromCookie(r)
    if username == "" {
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
    
    report, err := getLowStockReport(db.GetDB(), threshold)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if report == nil {
        report = []LowStockItem{}
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(report)
}

type LowStockItem struct {
    ProductName   string  `json:"product_name"`
    Category      string  `json:"category"`
    StockQuantity int     `json:"stock_quantity"`
    ReorderLevel  int     `json:"reorder_level"`
    CostPrice     float64 `json:"cost_price"`
    RestockCost   float64 `json:"restock_cost"`
}

func getLowStockReport(db *sql.DB, threshold int) ([]LowStockItem, error) {
    query := `
        SELECT 
            name as product_name,
            category,
            stock_quantity,
            reorder_level,
            cost_price,
            (reorder_level - stock_quantity) * cost_price as restock_cost
        FROM products
        WHERE stock_quantity <= ?
        ORDER BY stock_quantity ASC
    `
    rows, err := db.Query(query, threshold)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var items []LowStockItem
    for rows.Next() {
        var item LowStockItem
        err := rows.Scan(
            &item.ProductName, &item.Category,
            &item.StockQuantity, &item.ReorderLevel,
            &item.CostPrice, &item.RestockCost,
        )
        if err != nil {
            return nil, err
        }
        items = append(items, item)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return items, nil
}

func InventoryValuationHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    username := GetUsernameFromCookie(r)
    if username == "" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    report, err := getInventoryValuationReport(db.GetDB())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if report == nil {
        report = []InventoryValuationItem{}
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(report)
}

type InventoryValuationItem struct {
    Category        string  `json:"category"`
    TotalItems      int     `json:"total_items"`
    TotalQuantity   int     `json:"total_quantity"`
    TotalCost       float64 `json:"total_cost"`
    TotalRetail     float64 `json:"total_retail"`
    PotentialProfit float64 `json:"potential_profit"`
}

func getInventoryValuationReport(db *sql.DB) ([]InventoryValuationItem, error) {
    query := `
        SELECT 
            category,
            COUNT(*) as total_items,
            SUM(stock_quantity) as total_quantity,
            SUM(stock_quantity * cost_price) as total_cost,
            SUM(stock_quantity * retail_price) as total_retail,
            SUM(stock_quantity * (retail_price - cost_price)) as potential_profit
        FROM products
        WHERE stock_quantity > 0
        GROUP BY category
        ORDER BY total_cost DESC
    `
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var items []InventoryValuationItem
    for rows.Next() {
        var item InventoryValuationItem
        err := rows.Scan(
            &item.Category, &item.TotalItems, &item.TotalQuantity,
            &item.TotalCost, &item.TotalRetail, &item.PotentialProfit,
        )
        if err != nil {
            return nil, err
        }
        items = append(items, item)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return items, nil
}

func CategoryDrilldownHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    username := GetUsernameFromCookie(r)
    if username == "" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    category := r.URL.Query().Get("category")
    if category == "" {
        http.Error(w, "Category parameter is required", http.StatusBadRequest)
        return
    }
    details, err := getCategoryValuationDetails(db.GetDB(), category)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if details == nil {
        details = []CategoryProductDetail{}
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(details)
}

type CategoryProductDetail struct {
    ProductName string  `json:"product_name"`
    InStock     int     `json:"in_stock"`
    CostPrice   float64 `json:"cost_price"`
    RetailPrice float64 `json:"retail_price"`
    TotalCost   float64 `json:"total_cost"`
}

func getCategoryValuationDetails(db *sql.DB, category string) ([]CategoryProductDetail, error) {
    query := `
        SELECT name, stock_quantity, cost_price, retail_price, 
               stock_quantity * cost_price as total_cost
        FROM products
        WHERE category = ? AND stock_quantity > 0
        ORDER BY name ASC
    `
    rows, err := db.Query(query, category)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var details []CategoryProductDetail
    for rows.Next() {
        var d CategoryProductDetail
        err := rows.Scan(&d.ProductName, &d.InStock, &d.CostPrice, &d.RetailPrice, &d.TotalCost)
        if err != nil {
            return nil, err
        }
        details = append(details, d)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return details, nil
}




