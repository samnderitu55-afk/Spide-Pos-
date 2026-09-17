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

type TransferRequest struct {
	FromShopID   int                     `json:"from_shop_id"`
	ToShopID     int                     `json:"to_shop_id"`
	TransferDate string                  `json:"transfer_date"`
	Notes        string                  `json:"notes"`
	Items        []db.TransferItemDetail `json:"items"`
}

func CreateTransferHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if claims.CompanyID == 0 {
		writeJSONError(w, http.StatusBadRequest, "No company assigned")
		return
	}

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ToShopID == 0 {
		writeJSONError(w, http.StatusBadRequest, "Destination shop is required")
		return
	}
	if len(req.Items) == 0 {
		writeJSONError(w, http.StatusBadRequest, "At least one item is required")
		return
	}

	// Default source shop = user's own shop
	fromShopID := claims.ShopID

	// Directors/admins can override with any shop in their company
	if claims.Role == "director" || claims.Role == "admin" {
		if req.FromShopID > 0 {
			fromShopID = req.FromShopID
		}
	}

	if fromShopID == 0 {
		writeJSONError(w, http.StatusBadRequest, "No source shop assigned")
		return
	}

	if fromShopID == req.ToShopID {
		writeJSONError(w, http.StatusBadRequest, "Source and destination shops must differ")
		return
	}

	// Verify both shops belong to the user's company
	var shopCount int
	err := db.GetDB().QueryRow(
		"SELECT COUNT(*) FROM shops WHERE id IN (?, ?) AND company_id = ?",
		fromShopID, req.ToShopID, claims.CompanyID,
	).Scan(&shopCount)
	if err != nil || shopCount != 2 {
		writeJSONError(w, http.StatusBadRequest, "One or both shops don't belong to your company")
		return
	}

	transferDate := time.Now().Format("2006-01-02")
	if req.TransferDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.TransferDate); err == nil {
			transferDate = parsed.Format("2006-01-02")
		}
	}

	transfer := &db.StockTransfer{
		CompanyID:    claims.CompanyID, // ← fixed
		FromShopID:   fromShopID,
		ToShopID:     req.ToShopID,
		TransferDate: transferDate,
		Notes:        req.Notes,
		CreatedBy:    claims.Username,
		Status:       "completed",
		Items:        req.Items,
	}

	if err := db.CreateTransfer(db.GetDB(), transfer); err != nil {
		log.Printf("❌ CreateTransfer error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"message":  "Transfer created successfully",
		"transfer": transfer,
	})
}

func GetTransfersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if claims.CompanyID == 0 {
		writeJSONError(w, http.StatusBadRequest, "No company assigned")
		return
	}

	transfers, err := db.GetTransfers(db.GetDB(), claims.CompanyID) // ← company, not shop
	if err != nil {
		log.Printf("❌ GetTransfers error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, transfers)
}

func GetTransferDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeJSONError(w, http.StatusBadRequest, "Transfer ID required")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "Invalid transfer ID")
		return
	}

	transfer, err := db.GetTransferDetail(db.GetDB(), id, claims.CompanyID) // ← companyID
	if err != nil {
		log.Printf("❌ GetTransferDetail error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, transfer)
}
