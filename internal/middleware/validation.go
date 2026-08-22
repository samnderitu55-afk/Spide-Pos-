package middleware

import (
    "encoding/json"
    "net/http"
)

type ErrorResponse struct {
    Error string `json:"error"`
}

func ValidateJSON(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Only validate POST/PUT requests
        if r.Method != http.MethodPost && r.Method != http.MethodPut {
            next(w, r)
            return
        }

        // Check content-type
        if r.Header.Get("Content-Type") != "application/json" {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(ErrorResponse{
                Error: "Content-Type must be application/json",
            })
            return
        }

        // Check if body is empty
        if r.ContentLength == 0 {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(ErrorResponse{
                Error: "Request body cannot be empty",
            })
            return
        }

        next(w, r)
    }
}

