package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"spide-pos/internal/db"
	"spide-pos/internal/middleware"
	"strconv"
)

func GetCustomersHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Unauthorized",
        })
        return
    }

    customers, err := db.GetCustomers(db.GetDB())
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Failed to get customers: " + err.Error(),
        })
        return
    }

    json.NewEncoder(w).Encode(customers)
}

func GetCustomerHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Unauthorized",
        })
        return
    }

    idStr := r.URL.Query().Get("id")
    if idStr == "" {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Customer ID required",
        })
        return
    }

    id, err := strconv.Atoi(idStr)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Invalid customer ID",
        })
        return
    }

    customer, err := db.GetCustomerByID(db.GetDB(), id)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Failed to get customer: " + err.Error(),
        })
        return
    }

    if customer == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Customer not found",
        })
        return
    }

    json.NewEncoder(w).Encode(customer)
}

func CreateCustomerHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Unauthorized",
        })
        return
    }

    var req db.Customer
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Invalid request body: " + err.Error(),
        })
        return
    }

    if req.Name == "" {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Customer name is required",
        })
        return
    }

    req.CreatedBy = claims.Username

    err := db.CreateCustomer(db.GetDB(), &req)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Failed to create customer: " + err.Error(),
        })
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success":  true,
        "message":  "Customer created successfully",
        "customer": req,
    })
}

func UpdateCustomerHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Unauthorized",
        })
        return
    }

    var req db.Customer
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Invalid request body: " + err.Error(),
        })
        return
    }

    if req.ID == 0 {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Customer ID is required",
        })
        return
    }

    if req.Name == "" {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Customer name is required",
        })
        return
    }

    err := db.UpdateCustomer(db.GetDB(), &req)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Failed to update customer: " + err.Error(),
        })
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success":  true,
        "message":  "Customer updated successfully",
        "customer": req,
    })
}

func DeleteCustomerHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Unauthorized",
        })
        return
    }

    idStr := r.URL.Query().Get("id")
    if idStr == "" {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Customer ID required",
        })
        return
    }

    id, err := strconv.Atoi(idStr)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Invalid customer ID",
        })
        return
    }

    err = db.DeleteCustomer(db.GetDB(), id)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Failed to delete customer: " + err.Error(),
        })
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Customer deleted successfully",
    })
}

func SearchCustomersHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    query := r.URL.Query().Get("q")
    if query == "" {
        http.Error(w, "Search query required", http.StatusBadRequest)
        return
    }

    // Single company - just search all customers
    customers, err := db.SearchCustomers(db.GetDB(), query)
    if err != nil {
        log.Printf("❌ Error searching customers: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(customers)
}

func CreateCreditSaleHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Unauthorized",
        })
        return
    }

    var req db.CreditSale
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Invalid request body: " + err.Error(),
        })
        return
    }

    if req.CustomerID == 0 {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Customer is required",
        })
        return
    }
    if req.TotalAmount <= 0 {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Total amount must be greater than 0",
        })
        return
    }
    if req.Balance <= 0 {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Balance must be greater than 0",
        })
        return
    }

    req.ShopID = claims.ShopID
    if req.ShopID == 0 {
        req.ShopID = 1
    }
    req.CreatedBy = claims.Username
    req.Status = "pending"

    err := db.CreateCreditSale(db.GetDB(), &req)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Failed to create credit sale: " + err.Error(),
        })
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success":     true,
        "message":     "Credit sale created successfully",
        "credit_sale": req,
    })
}

func GetCreditSalesHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Unauthorized",
        })
        return
    }

    customerID := 0
    if customerIDParam := r.URL.Query().Get("customer_id"); customerIDParam != "" {
        if id, err := strconv.Atoi(customerIDParam); err == nil {
            customerID = id
        }
    }

    sales, err := db.GetCreditSales(db.GetDB(), customerID)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Failed to get credit sales: " + err.Error(),
        })
        return
    }

    json.NewEncoder(w).Encode(sales)
}

func GetCreditPaymentsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Unauthorized",
        })
        return
    }

    creditSaleIDStr := r.URL.Query().Get("credit_sale_id")
    if creditSaleIDStr == "" {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Credit sale ID required",
        })
        return
    }

    creditSaleID, err := strconv.Atoi(creditSaleIDStr)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Invalid credit sale ID",
        })
        return
    }

    payments, err := db.GetCreditPayments(db.GetDB(), creditSaleID)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Failed to get credit payments: " + err.Error(),
        })
        return
    }

    json.NewEncoder(w).Encode(payments)
}

func AddCreditPaymentHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Unauthorized",
        })
        return
    }

    var req db.CreditPayment
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Invalid request body: " + err.Error(),
        })
        return
    }

    if req.CreditSaleID == 0 {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Credit sale ID is required",
        })
        return
    }
    if req.Amount <= 0 {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Amount must be greater than 0",
        })
        return
    }

    req.CreatedBy = claims.Username

    err := db.AddCreditPayment(db.GetDB(), &req)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Failed to add credit payment: " + err.Error(),
        })
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Payment added successfully",
        "payment": req,
    })
}

// ============================================
// CUSTOMER DEPOSIT HANDLERS
// ============================================

// Add deposit to customer account
func AddCustomerDepositHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Unauthorized",
        })
        return
    }

    var req struct {
        CustomerID    int     `json:"customer_id"`
        Amount        float64 `json:"amount"`
        PaymentMethod string  `json:"payment_method"`
        Reference     string  `json:"reference"`
        Notes         string  `json:"notes"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Invalid request body: " + err.Error(),
        })
        return
    }

    if req.CustomerID == 0 {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Customer is required",
        })
        return
    }

    if req.Amount <= 0 {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Amount must be greater than 0",
        })
        return
    }

    err := db.AddCustomerDeposit(db.GetDB(), req.CustomerID, req.Amount, req.PaymentMethod, req.Reference, req.Notes, claims.Username)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Failed to add deposit: " + err.Error(),
        })
        return
    }

    balance, _ := db.GetCustomerBalance(db.GetDB(), req.CustomerID)

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success":     true,
        "message":     "Deposit added successfully",
        "new_balance": balance,
    })
}

// Get customer balance
func GetCustomerBalanceHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Unauthorized",
        })
        return
    }

    customerIDStr := r.URL.Query().Get("customer_id")
    if customerIDStr == "" {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Customer ID required",
        })
        return
    }

    customerID, err := strconv.Atoi(customerIDStr)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Invalid customer ID",
        })
        return
    }

    balance, err := db.GetCustomerBalance(db.GetDB(), customerID)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Failed to get balance: " + err.Error(),
        })
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "customer_id": customerID,
        "balance":     balance,
    })
}

// Get customer transactions
func GetCustomerTransactionsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Unauthorized",
        })
        return
    }

    customerIDStr := r.URL.Query().Get("customer_id")
    if customerIDStr == "" {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Customer ID required",
        })
        return
    }

    customerID, err := strconv.Atoi(customerIDStr)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Invalid customer ID",
        })
        return
    }

    transactions, err := db.GetCustomerTransactions(db.GetDB(), customerID)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Failed to get transactions: " + err.Error(),
        })
        return
    }

    json.NewEncoder(w).Encode(transactions)
}

// internal/handlers/customers.go

func GetCustomerStatementHandler(w http.ResponseWriter, r *http.Request) {
    // Get user from context (set by AuthMiddleware)
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    customerID := r.URL.Query().Get("customer_id")
    if customerID == "" {
        http.Error(w, "customer_id required", http.StatusBadRequest)
        return
    }

    id, err := strconv.Atoi(customerID)
    if err != nil {
        http.Error(w, "Invalid customer_id", http.StatusBadRequest)
        return
    }

    log.Printf("📊 Generating statement for customer: %d (user: %s)", id, claims.Username)

    // Get customer transactions
    transactions, err := db.GetCustomerTransactions(db.GetDB(), id)
    if err != nil {
        log.Printf("❌ Error getting transactions: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Get customer summary
    summary, err := db.GetCustomerSummary(db.GetDB(), id)
    if err != nil {
        log.Printf("❌ Error getting summary: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    response := map[string]interface{}{
        "transactions":     transactions,
        "total_sales":      summary.TotalSales,
        "total_payments":   summary.TotalPayments,
        "total_deposits":   summary.TotalDeposits,
        "current_balance":  summary.CurrentBalance,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
