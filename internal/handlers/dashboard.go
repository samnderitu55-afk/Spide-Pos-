package handlers

import (
    "encoding/json"
    "net/http"
    "spide-pos/internal/db"
)

func DashboardStatsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    stats, err := db.GetDashboardStats(db.GetDB())
    if err != nil {
        http.Error(w, `{"error":"Failed to fetch dashboard data"}`, http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(stats)
}

