package handlers

import (
    "encoding/json"
    "net/http"
    "time"
    "spide-pos/internal/db"
)

func CreateExpenseHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    if r.Method != http.MethodPost {
        http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
        return
    }

    username := GetUsernameFromCookie(r)
    if username == "" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }

    var expense db.Expense
    if err := json.NewDecoder(r.Body).Decode(&expense); err != nil {
        http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
        return
    }

    if expense.ExpenseDate == "" {
        expense.ExpenseDate = time.Now().Format("2006-01-02")
    }
    if expense.CreatedBy == "" {
        expense.CreatedBy = username
    }
    // Default ShopID to 1 for now (will come from user settings later)
    if expense.ShopID == 0 {
        expense.ShopID = 1
    }

    id, err := db.CreateExpense(db.GetDB(), &expense)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    expense.ID = int(id)
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(expense)
}

func ExpenseReportHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    username := GetUsernameFromCookie(r)
    if username == "" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    startDate := r.URL.Query().Get("start")
    endDate := r.URL.Query().Get("end")
    
    // Default ShopID to 1
    shopID := 1
    
    report, err := db.GetExpenseReport(db.GetDB(), startDate, endDate, shopID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(report)
}

func ExpenseCategoriesHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    username := GetUsernameFromCookie(r)
    if username == "" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    // Default ShopID to 1
    shopID := 1
    
    categories, err := db.GetExpenseCategories(db.GetDB(), shopID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(categories)
}

func DeleteExpenseHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    username := GetUsernameFromCookie(r)
    if username == "" {
        http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    var request struct {
        ID int `json:"id"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
        return
    }
    
    if request.ID == 0 {
        http.Error(w, `{"error":"Expense ID is required"}`, http.StatusBadRequest)
        return
    }
    
    if err := db.DeleteExpense(db.GetDB(), request.ID); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Expense deleted successfully",
        "id":      request.ID,
    })
}





