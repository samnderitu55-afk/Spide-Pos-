package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"spide-pos/internal/db"
	"spide-pos/internal/middleware"
	"strconv"
)

func DashboardStatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		log.Println("❌ DashboardStatsHandler: No claims found")
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	log.Printf("📊 DashboardStatsHandler: User: %s, Role: %s, ShopID: %d", claims.Username, claims.Role, claims.ShopID)

	shopID := claims.ShopID
	if shopID == 0 {
		shopID = 1
	}

	// Check if shop_id is in URL (for directors viewing other shops)
	if shopIDParam := r.URL.Query().Get("shop_id"); shopIDParam != "" {
		if id, err := strconv.Atoi(shopIDParam); err == nil && id > 0 {
			shopID = id
			log.Printf("📊 DashboardStatsHandler: Using shop_id from URL: %d", shopID)
		}
	}

	if claims.Role != "director" && claims.Role != "admin" {
		shopID = claims.ShopID
		log.Printf("📊 DashboardStatsHandler: Non-director, forcing shop_id: %d", shopID)
	}

	log.Printf("📊 DashboardStatsHandler: Final shopID: %d", shopID)

	stats, err := db.GetDashboardStats(db.GetDB(), shopID)
	if err != nil {
		log.Printf("❌ DashboardStatsHandler: db.GetDashboardStats error: %v", err)
		http.Error(w, `{"error":"Failed to fetch dashboard stats: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("✅ DashboardStatsHandler: Successfully fetched stats")
	json.NewEncoder(w).Encode(stats)
}

// GetShopDashboardHandler - For cashiers and managers to view their shop stats
func GetShopDashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get shop_id from URL or use user's shop
	shopIDParam := r.URL.Query().Get("shop_id")
	var shopID int

	if shopIDParam != "" {
		id, err := strconv.Atoi(shopIDParam)
		if err == nil && id > 0 {
			shopID = id
		}
	}

	// If no shop_id provided, use user's assigned shop
	if shopID == 0 {
		shopID = claims.ShopID
	}

	// For cashiers and managers, only allow their own shop
	if claims.Role == "cashier" || claims.Role == "manager" {
		if shopID != claims.ShopID {
			http.Error(w, "Forbidden: You can only view your assigned shop", http.StatusForbidden)
			return
		}
	}

	// Get shop stats
	stats, err := db.GetShopStats(db.GetDB(), shopID)
	if err != nil {
		log.Printf("❌ Error getting shop stats: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get shop name
	var shopName string
	err = db.GetDB().QueryRow("SELECT name FROM shops WHERE id = ?", shopID).Scan(&shopName)
	if err != nil {
		shopName = "Store #" + strconv.Itoa(shopID)
	}

	// Get recent sales for this shop
	recentSales, err := db.GetRecentSalesForShop(db.GetDB(), shopID, 5)
	if err != nil {
		recentSales = []db.Sale{}
	}

	response := map[string]interface{}{
		"stats":        stats,
		"shop_id":      shopID,
		"shop_name":    shopName,
		"role":         claims.Role,
		"recent_sales": recentSales,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
