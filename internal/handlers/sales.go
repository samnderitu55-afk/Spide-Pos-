package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"spide-pos/internal/db"
	"spide-pos/internal/middleware"
	"strconv"
)

func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req db.SaleRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // ✅ Log the incoming request with full breakdown
    log.Printf("💰 Sale received - Total: %.2f, Customer: %d", req.TotalAmount, req.CustomerID)
    log.Printf("💰 Payment Breakdown - Cash: %.2f, Mpesa: %.2f, Deposit: %.2f, Credit: %.2f, Type: %s", 
        req.CashAmount, req.MpesaAmount, req.DepositAmount, req.CreditAmount, req.PaymentType)

    // ✅ Get shop_id from request
    shopID := req.ShopID
    if shopID == 0 {
        claims := middleware.GetUserFromContext(r)
        if claims != nil {
            shopID = claims.ShopID
        }
    }
    if shopID == 0 {
        shopID = 1
    }
    req.ShopID = shopID

    log.Printf("💰 Sale - Shop ID: %d, Total: %.2f", shopID, req.TotalAmount)

    // ✅ Calculate total paid FIRST (before any modifications)
    totalPaid := req.CashAmount + req.MpesaAmount + req.DepositAmount + req.CreditAmount
    log.Printf("💰 Total Paid before validation: %.2f", totalPaid)

    // ✅ Only modify amounts for specific payment types
    switch req.PaymentType {
    case "split":
        // Split between cash and mpesa
        if req.CashAmount > 0 && req.MpesaAmount > 0 {
            // Keep both as is
        }
        // Calculate change if cash exceeds remaining
        remainingAfterMpesa := req.TotalAmount - req.MpesaAmount - req.DepositAmount - req.CreditAmount
        if remainingAfterMpesa < 0 {
            remainingAfterMpesa = 0
        }
        if req.CashAmount > remainingAfterMpesa {
            req.ChangeGiven = req.CashAmount - remainingAfterMpesa
            req.CashAmount = remainingAfterMpesa
        }
    case "cash":
        // Only cash payment
        req.MpesaAmount = 0
        req.DepositAmount = 0
        req.CreditAmount = 0
        if req.CashAmount > req.TotalAmount {
            req.ChangeGiven = req.CashAmount - req.TotalAmount
            req.CashAmount = req.TotalAmount
        }
    case "mpesa":
        // Only Mpesa payment
        req.CashAmount = 0
        req.DepositAmount = 0
        req.CreditAmount = 0
        req.MpesaAmount = req.TotalAmount
        req.ChangeGiven = 0
    case "deposit":
        // Only deposit payment (customer has enough deposit)
        req.CashAmount = 0
        req.MpesaAmount = 0
        //req.CreditAmount = 0
        req.ChangeGiven = 0
        // Deposit amount should already be set
    case "credit":
        // ✅ Credit payment - KEEP all payment amounts!
        // Don't reset anything - just validate that total is covered
        // Cash, Mpesa, Deposit, and Credit all work together
        req.ChangeGiven = 0
        
        // If credit is the only payment, set it to total
        if req.CashAmount == 0 && req.MpesaAmount == 0 && req.DepositAmount == 0 && req.CreditAmount > 0 {
            req.CreditAmount = req.TotalAmount
        }
        // Otherwise, keep all amounts as they are (cash + deposit + credit)
    default:
        // Unknown payment type - treat as cash
        log.Printf("⚠️ Unknown payment type: %s, defaulting to cash", req.PaymentType)
        req.PaymentType = "cash"
        req.CashAmount = req.TotalAmount
        req.MpesaAmount = 0
        req.DepositAmount = 0
        req.CreditAmount = 0
        req.ChangeGiven = 0
    }

    // ✅ Recalculate total paid after modifications
    totalPaid = req.CashAmount + req.MpesaAmount + req.DepositAmount + req.CreditAmount
    log.Printf("💰 Total Paid after validation: %.2f (Cash: %.2f, Mpesa: %.2f, Deposit: %.2f, Credit: %.2f)", 
        totalPaid, req.CashAmount, req.MpesaAmount, req.DepositAmount, req.CreditAmount)

    // ✅ Validate: Cash + Mpesa + Deposit + Credit must cover total
    if totalPaid < req.TotalAmount-0.009 {
        log.Printf("❌ Insufficient payment: Paid=%.2f, Required=%.2f", totalPaid, req.TotalAmount)
        http.Error(w, "Insufficient payment", http.StatusBadRequest)
        return
    }

    saleID, err := db.CreateSale(db.GetDB(), req)
    if err != nil {
        log.Printf("❌ Error creating sale: %v", err)
        http.Error(w, "Failed to record transaction", http.StatusInternalServerError)
        return
    }

    log.Printf("✅ Sale created with ID: %d, Payment Type: %s", saleID, req.PaymentType)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":       "success",
        "sale_id":      saleID,
        "change_given": req.ChangeGiven,
    })
}

func RecentSalesHandler(w http.ResponseWriter, r *http.Request) {
    // Get shop_id from cookie
    shopID := 1
    cookie, err := r.Cookie("current_shop")
    if err == nil && cookie.Value != "" {
        if id, parseErr := strconv.Atoi(cookie.Value); parseErr == nil && id > 0 {
            shopID = id
        }
    }

    sales, err := db.GetRecentSalesForShop(db.GetDB(), shopID, 20)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(sales)
}
