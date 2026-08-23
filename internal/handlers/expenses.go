package handlers

import (
	"encoding/json"
	"net/http"
	"spide-pos/internal/db"
	"spide-pos/internal/middleware"
	"strconv"
	"time"
)

type ExpenseRequest struct {
	Category      string  `json:"category"`
	Description   string  `json:"description"`
	Amount        float64 `json:"amount"`
	ExpenseDate   string  `json:"expense_date"`
	PaymentMethod string  `json:"payment_method"`
	Reference     string  `json:"reference"`
	Notes         string  `json:"notes"`
	ShopID        int     `json:"shop_id"`
}

func CreateExpenseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req ExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Category == "" || req.Description == "" || req.Amount <= 0 {
		http.Error(w, `{"error":"Category, description, and amount are required"}`, http.StatusBadRequest)
		return
	}

	// Use user's shop
	shopID := claims.ShopID
	if shopID == 0 {
		shopID = 1
	}

	// Parse expense date or use today
	expenseDate := time.Now().Format("2006-01-02")
	if req.ExpenseDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.ExpenseDate); err == nil {
			expenseDate = parsed.Format("2006-01-02")
		}
	}

	expense := &db.Expense{
		Category:      req.Category,
		Description:   req.Description,
		Amount:        req.Amount,
		ExpenseDate:   expenseDate,
		PaymentMethod: req.PaymentMethod,
		Reference:     req.Reference,
		Notes:         req.Notes,
		ShopID:        shopID,
		CreatedBy:     claims.Username,
	}

	err := db.CreateExpense(db.GetDB(), expense)
	if err != nil {
		http.Error(w, `{"error":"Failed to create expense: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Expense created successfully",
		"expense": expense,
	})
}

func ExpenseReportHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	//category := r.URL.Query().Get("category")

	// Default to current month
	if startDate == "" {
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	// Use user's shop
	shopID := claims.ShopID
	if shopID == 0 {
		shopID = 1
	}

	expenses, total, err := db.GetExpenseReport(db.GetDB(), startDate, endDate, shopID)
	if err != nil {
		http.Error(w, `{"error":"Failed to get expense report: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"expenses":   expenses,
		"total":      total,
		"count":      len(expenses),
		"start_date": startDate,
		"end_date":   endDate,
	})
}

func ExpenseCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	categories, err := db.GetExpenseCategories(db.GetDB(), claims.ShopID)
	if err != nil {
		http.Error(w, `{"error":"Failed to get expense categories: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(categories)
}

func DeleteExpenseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, `{"error":"Expense ID required"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid expense ID"}`, http.StatusBadRequest)
		return
	}

	err = db.DeleteExpense(db.GetDB(), id)
	if err != nil {
		http.Error(w, `{"error":"Failed to delete expense: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Expense deleted successfully",
	})
}
