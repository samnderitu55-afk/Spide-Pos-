package handlers

import (
    "encoding/json"
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

    req.CreatedBy = claims.Username // Store username, DB will convert to ID

    err := db.CreateCustomer(db.GetDB(), &req)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Failed to create customer: " + err.Error(),
        })
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Customer created successfully",
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
        "success": true,
        "message": "Customer updated successfully",
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
    w.Header().Set("Content-Type", "application/json")
    
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Unauthorized",
        })
        return
    }

    query := r.URL.Query().Get("q")
    if query == "" {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Search query is required",
        })
        return
    }

    customers, err := db.SearchCustomers(db.GetDB(), query)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Failed to search customers: " + err.Error(),
        })
        return
    }

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
    req.CreatedBy = claims.Username // Store username, DB will convert to ID
    req.Status = "pending"

    err := db.CreateCreditSale(db.GetDB(), &req)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Failed to create credit sale: " + err.Error(),
        })
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Credit sale created successfully",
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

    req.CreatedBy = claims.Username // Store username, DB will convert to ID

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


