package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"spide-pos/internal/db"
	"spide-pos/internal/middleware"
	"strconv"
	"time"
)

type LowStockItem struct {
	ProductName   string  `json:"product_name"`
	Category      string  `json:"category"`
	StockQuantity float64 `json:"stock_quantity"`
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
	if dateParam == "" {
		dateParam = time.Now().Format("2006-01-02")
	}

	// ✅ Role-based shop access
	var shopID int
	shopIDParam := r.URL.Query().Get("shop_id")

	// ✅ Add debug logging
	log.Printf("🔍 Z-Report Request - Date: %s, shop_id_param: %s, Role: %s",
		dateParam, shopIDParam, claims.Role)

	switch claims.Role {
	case "director", "admin":
		// Director/Admin can view any shop
		if shopIDParam != "" {
			id, err := strconv.Atoi(shopIDParam)
			if err == nil && id > 0 {
				shopID = id
			} else {
				shopID = claims.ShopID
			}
		} else {
			// If no shop_id specified, show all shops (shopID = 0)
			shopID = 0
		}
	case "manager":
		// Manager can view their own shop only
		shopID = claims.ShopID
		if shopID == 0 {
			shopID = 1
		}
	case "cashier":
		// Cashier can view their own shop only
		shopID = claims.ShopID
		if shopID == 0 {
			shopID = 1
		}
	default:
		// Default to their shop
		shopID = claims.ShopID
		if shopID == 0 {
			shopID = 1
		}
	}

	log.Printf("📊 Generating Z-Report - Date: %s, Shop: %d, User: %s, Role: %s",
		dateParam, shopID, claims.Username, claims.Role)

	report, err := db.GetZReport(db.GetDB(), dateParam, shopID)
	if err != nil {
		log.Printf("❌ Error generating Z-Report: %v", err)
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

	// Query parameters
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	category := r.URL.Query().Get("category")
	productID := r.URL.Query().Get("product_id")

	log.Printf("ProductSalesReportHandler: start=%q end=%q category=%q product_id=%q",
		startDate, endDate, category, productID)

	if startDate == "" {
		http.Error(w, `{"error":"start_date parameter is required"}`, http.StatusBadRequest)
		return
	}
	if endDate == "" {
		http.Error(w, `{"error":"end_date parameter is required"}`, http.StatusBadRequest)
		return
	}

	// Company ID from JWT
	companyID := claims.CompanyID
	if companyID == 0 {
		companyID = 1
	}

	// Resolve shop_id
	shopID := 0
	if claims.Role == "director" || claims.Role == "admin" {
		// Directors: default to all shops, allow override
		if shopIDParam := r.URL.Query().Get("shop_id"); shopIDParam != "" {
			if id, err := strconv.Atoi(shopIDParam); err == nil && id >= 0 {
				shopID = id // 0 = all shops
			}
		}
	} else {
		// Cashiers/managers: locked to their own shop
		shopID = claims.ShopID
		if shopID == 0 {
			shopID = 1
		}
	}

	// Parse product_id if provided
	productIDInt := 0
	if productID != "" {
		if id, err := strconv.Atoi(productID); err == nil {
			productIDInt = id
		}
	}

	products, err := db.GetProductSalesReport(db.GetDB(), companyID, startDate, endDate, shopID, category, productIDInt)
	if err != nil {
		log.Printf("❌ GetProductSalesReport error: %v", err)
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

	//thresholdStr := r.URL.Query().Get("threshold")
	//threshold := 5
	//if thresholdStr != "" {
	//	if val, err := strconv.Atoi(thresholdStr); err == nil && val >= 0 {
	//		threshold = val
	//	}
	//}

	// ✅ Get company ID from claims
	companyID := claims.CompanyID
	if companyID == 0 {
		companyID = 1
	}

	// Get shop_id from query (optional)
	shopIDParam := r.URL.Query().Get("shop_id")
	var shopID int
	if shopIDParam != "" {
		id, err := strconv.Atoi(shopIDParam)
		if err == nil && id > 0 {
			shopID = id
		}
	}
	// If user is director/admin, they can view all shops
	if claims.Role == "director" || claims.Role == "admin" {
		if shopIDParam := r.URL.Query().Get("shop_id"); shopIDParam != "" {
			if id, err := strconv.Atoi(shopIDParam); err == nil && id > 0 {
				shopID = id
			}
		}
	}
	// ✅ Get low stock items for this company
	items, err := db.GetLowStockItems(db.GetDB(), companyID, shopID)
	if err != nil {
		log.Printf("❌ Error getting low stock items: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// InventoryValuationHandler - Inventory Valuation Report
func InventoryValuationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// ✅ Get company ID from claims
	companyID := claims.CompanyID
	if companyID == 0 {
		companyID = 1
	}

	// Get branch/shop ID from query (optional)
	shopIDParam := r.URL.Query().Get("branch_id")
	var shopID int
	if shopIDParam != "" {
		id, err := strconv.Atoi(shopIDParam)
		if err == nil && id > 0 {
			shopID = id
		}
	}

	// ✅ Get inventory valuation for this company
	valuation, err := db.GetInventoryValuation(db.GetDB(), companyID, shopID)
	if err != nil {
		log.Printf("❌ Error getting inventory valuation: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(valuation)
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

func ProductSalesCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	companyID := claims.CompanyID
	if companyID == 0 {
		companyID = 1
	}

	shopID := 0
	if claims.Role == "director" || claims.Role == "admin" {
		if p := r.URL.Query().Get("shop_id"); p != "" {
			if id, err := strconv.Atoi(p); err == nil {
				shopID = id
			}
		}
	} else {
		shopID = claims.ShopID
	}

	cats, err := db.GetProductSalesCategories(db.GetDB(), companyID, shopID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(cats)
}

func ProductSalesProductsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	companyID := claims.CompanyID
	if companyID == 0 {
		companyID = 1
	}

	shopID := 0
	if claims.Role == "director" || claims.Role == "admin" {
		if p := r.URL.Query().Get("shop_id"); p != "" {
			if id, err := strconv.Atoi(p); err == nil {
				shopID = id
			}
		}
	} else {
		shopID = claims.ShopID
	}

	category := r.URL.Query().Get("category")

	products, err := db.GetProductSalesProducts(db.GetDB(), companyID, shopID, category)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(products)
}
