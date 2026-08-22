package handlers

import (
    "encoding/json"
    "net/http"
    "spide-pos/internal/db"
)

func DirectorDashboardHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    dashboard, err := db.GetDirectorDashboard(db.GetDB())
    if err != nil {
        http.Error(w, `{"error":"Failed to fetch director dashboard data"}`, http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(dashboard)
}




