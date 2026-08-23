package handlers

import (
    "encoding/json"
    "net/http"
    "spide-pos/internal/db"
    "spide-pos/internal/middleware"
)

func BranchesHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    branches, err := db.GetBranches(db.GetDB())
    if err != nil {
        http.Error(w, `{"error":"Failed to fetch branches"}`, http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(branches)
}
