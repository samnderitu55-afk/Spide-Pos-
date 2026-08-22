package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "spide-pos/internal/db"
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

    // Get shop_id from cookie or use default
    shopID := 1
    cookie, err := r.Cookie("current_shop")
    if err == nil && cookie.Value != "" {
        if id, parseErr := strconv.Atoi(cookie.Value); parseErr == nil && id > 0 {
            shopID = id
        }
    }
    req.ShopID = shopID

    // Calculate change
    switch req.PaymentType {
    case "split":
        remainingBalance := req.TotalAmount - req.MpesaAmount
        if remainingBalance < 0 {
            remainingBalance = 0
        }
        if req.CashAmount > remainingBalance {
            req.ChangeGiven = req.CashAmount - remainingBalance
            req.CashAmount = remainingBalance
        }
    case "cash":
        req.MpesaAmount = 0
        if req.CashAmount > req.TotalAmount {
            req.ChangeGiven = req.CashAmount - req.TotalAmount
            req.CashAmount = req.TotalAmount
        }
    case "mpesa":
        req.MpesaAmount = req.TotalAmount
        req.CashAmount = 0
        req.ChangeGiven = 0
    }

    saleID, err := db.CreateSale(db.GetDB(), req)
    if err != nil {
        http.Error(w, "Failed to record transaction", http.StatusInternalServerError)
        return
    }

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
