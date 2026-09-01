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