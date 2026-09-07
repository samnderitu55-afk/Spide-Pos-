package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"spide-pos/internal/db"
	"spide-pos/internal/middleware"
)

func DirectorDashboardHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Only directors and admins can access
	if claims.Role != "director" && claims.Role != "admin" {
		http.Error(w, `{"error":"Director access required"}`, http.StatusForbidden)
		return
	}

	// ✅ Get company ID from claims
	companyID := claims.CompanyID
	if companyID == 0 {
		companyID = 1
	}

	// Get director dashboard data for this company
	data, err := db.GetDirectorDashboardData(db.GetDB(), companyID)
	if err != nil {
		log.Printf("❌ Error getting director dashboard data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
