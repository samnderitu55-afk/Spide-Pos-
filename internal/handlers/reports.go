package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"spide-pos/internal/db"
	"spide-pos/internal/middleware"
	"strconv"
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

	// Get query parameters
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	// Debug: log to terminal
	log.Println("ProductSalesReportHandler called")
	log.Printf("startDate: '%s', endDate: '%s'", startDate, endDate)

	if startDate == "" {
		http.Error(w, `{"error":"start_date parameter is required"}`, http.StatusBadRequest)
		return
	}
	if endDate == "" {
		http.Error(w, `{"error":"end_date parameter is required"}`, http.StatusBadRequest)
		return
	}

	// Get shop_id - default to user's shop
	shopID := claims.ShopID
	if claims.Role == "director" || claims.Role == "admin" {
		if shopIDParam := r.URL.Query().Get("shop_id"); shopIDParam != "" {
			if id, err := strconv.Atoi(shopIDParam); err == nil && id > 0 {
				shopID = id
			}
		}
	}

	products, err := db.GetProductSalesReport(db.GetDB(), startDate, endDate, shopID)
	if err != nil {
		http.Error(w, `{"error":"Failed to generate product sales report: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(products)
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

	// If user is director/admin, they can view all shops
	if claims.Role == "director" || claims.Role == "admin" {
		if shopIDParam := r.URL.Query().Get("shop_id"); shopIDParam != "" {
			if id, err := strconv.Atoi(shopIDParam); err == nil && id > 0 {
				shopID = id
			}
		}
	}

	query := `
        SELECT 
            p.name,
            COALESCE(p.category, 'Uncategorized') as category,
            COALESCE(ss.quantity, 0) as stock_quantity,
            p.reorder_level,
            p.cost_price,
            COALESCE((p.reorder_level - COALESCE(ss.quantity, 0)) * p.cost_price, 0) as restock_cost
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = ?
        WHERE p.is_active = 1 
          AND COALESCE(ss.quantity, 0) <= ?
        ORDER BY stock_quantity ASC
        LIMIT 50
    `

	rows, err := db.GetDB().Query(query, shopID, threshold)
	if err != nil {
		http.Error(w, `{"error":"Database error: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []db.LowStockReportItem // ← Use db.LowStockReportItem
	for rows.Next() {
		var item db.LowStockReportItem // ← Use db.LowStockReportItem
		err := rows.Scan(
			&item.ProductName,
			&item.Category,
			&item.StockQuantity,
			&item.ReorderLevel,
			&item.CostPrice,
			&item.RestockCost,
		)
		if err != nil {
			http.Error(w, `{"error":"Scan error: `+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}
    
    if err := rows.Err(); err != nil {
        http.Error(w, `{"error":"Rows error: `+err.Error()+`"}`, http.StatusInternalServerError)
        return
    }

	if items == nil {
		items = []db.LowStockReportItem{}
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

	branchID := 0
	if branchIDParam := r.URL.Query().Get("branch_id"); branchIDParam != "" {
		if id, err := strconv.Atoi(branchIDParam); err == nil && id > 0 {
			branchID = id
		}
	}

	// If user is not director, force their branch
	if claims.Role != "director" && claims.Role != "admin" {
		branchID = claims.ShopID
	}

	valuation, err := db.GetInventoryValuation(db.GetDB(), branchID)
	if err != nil {
		http.Error(w, `{"error":"Failed to get inventory valuation: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	products, err := db.GetProductValuation(db.GetDB(), branchID)
	if err != nil {
		http.Error(w, `{"error":"Failed to get product details: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"valuation": valuation,
		"products":  products,
	}

	json.NewEncoder(w).Encode(response)
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
