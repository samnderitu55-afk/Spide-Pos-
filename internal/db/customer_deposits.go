package db

import (
    "log"
    "database/sql"
    "fmt"
)

// Add deposit to customer account
func AddCustomerDeposit(db *sql.DB, customerID int, amount float64, paymentMethod, reference, notes, createdBy string) error {
    log.Printf("💰 Adding deposit: customerID=%d, amount=%.2f, paymentMethod=%s", customerID, amount, paymentMethod)
    
    // Get user ID from username
    var userID int
    err := db.QueryRow("SELECT id FROM users WHERE username = ?", createdBy).Scan(&userID)
    if err != nil {
        log.Printf("⚠️ User not found: %s, defaulting to 1", createdBy)
        userID = 1
    }
    log.Printf("Using user_id: %d for created_by", userID)
    
    tx, err := db.Begin()
    if err != nil {
        log.Printf("❌ Failed to begin transaction: %v", err)
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Add deposit record with user ID
    query := `
        INSERT INTO customer_deposits (customer_id, amount, payment_method, reference, notes, created_by, created_at)
        VALUES (?, ?, ?, ?, ?, ?, NOW())
    `
    log.Printf("Inserting deposit: customer_id=%d, amount=%.2f", customerID, amount)
    
    _, err = tx.Exec(query, customerID, amount, paymentMethod, reference, notes, userID)
    if err != nil {
        log.Printf("❌ Failed to create deposit: %v", err)
        return fmt.Errorf("failed to create deposit: %w", err)
    }
    log.Printf("✅ Deposit record inserted")

    // Update customer deposit balance
    _, err = tx.Exec(`
        UPDATE customers 
        SET deposit_balance = deposit_balance + ?, updated_at = NOW()
        WHERE id = ?
    `, amount, customerID)
    if err != nil {
        log.Printf("❌ Failed to update customer balance: %v", err)
        return fmt.Errorf("failed to update customer balance: %w", err)
    }
    log.Printf("✅ Customer balance updated")

    if err := tx.Commit(); err != nil {
        log.Printf("❌ Failed to commit deposit: %v", err)
        return fmt.Errorf("failed to commit deposit: %w", err)
    }
    
    log.Printf("✅ Deposit completed successfully for customer %d", customerID)
    return nil
}

// Deduct from customer balance for purchases
func DeductCustomerBalance(db *sql.DB, customerID int, amount float64, saleID int, notes, createdBy string) error {
    tx, err := db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Check if customer has enough balance
    var balance float64
    err = tx.QueryRow("SELECT deposit_balance FROM customers WHERE id = ?", customerID).Scan(&balance)
    if err != nil {
        return fmt.Errorf("failed to get customer balance: %w", err)
    }

    if balance < amount {
        return fmt.Errorf("insufficient balance: available KES %.2f, required KES %.2f", balance, amount)
    }

    // Deduct from balance
    _, err = tx.Exec(`
        UPDATE customers 
        SET deposit_balance = deposit_balance - ?, updated_at = NOW()
        WHERE id = ?
    `, amount, customerID)
    if err != nil {
        return fmt.Errorf("failed to update customer balance: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit deduction: %w", err)
    }

    return nil
}

// Get customer deposit balance
func GetCustomerBalance(db *sql.DB, customerID int) (float64, error) {
    var balance float64
    err := db.QueryRow("SELECT deposit_balance FROM customers WHERE id = ?", customerID).Scan(&balance)
    if err != nil {
        if err == sql.ErrNoRows {
            return 0, nil
        }
        return 0, fmt.Errorf("failed to get customer balance: %w", err)
    }
    return balance, nil
}

// Get customer transaction history
func GetCustomerTransactions(db *sql.DB, customerID int) ([]CreditSale, error) {
    query := `
        SELECT id, customer_id, shop_id, sale_id, total_amount, amount_paid, balance,
               due_date, status, notes, created_by, created_at, updated_at
        FROM credit_sales
        WHERE customer_id = ?
        ORDER BY created_at DESC
    `
    rows, err := db.Query(query, customerID)
    if err != nil {
        return nil, fmt.Errorf("failed to get transactions: %w", err)
    }
    defer rows.Close()

    var transactions []CreditSale
    for rows.Next() {
        var t CreditSale
        err := rows.Scan(
            &t.ID,
            &t.CustomerID,
            &t.ShopID,
            &t.SaleID,
            &t.TotalAmount,
            &t.AmountPaid,
            &t.Balance,
            &t.DueDate,
            &t.Status,
            &t.Notes,
            &t.CreatedBy,
            &t.CreatedAt,
            &t.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        transactions = append(transactions, t)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating transactions: %w", err)
    }

    return transactions, nil
}



