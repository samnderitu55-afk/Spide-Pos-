package handlers

import (
    "encoding/json"
    "net/http"
    "spide-pos/internal/db"
    "spide-pos/internal/middleware"
)

func DirectorDashboardHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
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
    
    dashboard, err := db.GetDirectorDashboard(db.GetDB())
    if err != nil {
        http.Error(w, `{"error":"Failed to fetch director dashboard data: `+err.Error()+`"}`, http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(dashboard)
}
