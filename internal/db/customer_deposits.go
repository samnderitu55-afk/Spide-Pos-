package db

import (
    "database/sql"
    "fmt"
)

// Add deposit to customer account
func AddCustomerDeposit(db *sql.DB, customerID int, amount float64, paymentMethod, reference, notes, createdBy string) error {
    tx, err := db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Add deposit record
    query := `
        INSERT INTO customer_deposits (customer_id, amount, payment_method, reference, notes, created_by, created_at)
        VALUES (?, ?, ?, ?, ?, ?, NOW())
    `
    _, err = tx.Exec(query, customerID, amount, paymentMethod, reference, notes, createdBy)
    if err != nil {
        return fmt.Errorf("failed to create deposit: %w", err)
    }

    // Update customer deposit balance
    _, err = tx.Exec(`
        UPDATE customers 
        SET deposit_balance = deposit_balance + ?, updated_at = NOW()
        WHERE id = ?
    `, amount, customerID)
    if err != nil {
        return fmt.Errorf("failed to update customer balance: %w", err)
    }

    // Record in credit_sales as a deposit transaction
    _, err = tx.Exec(`
        INSERT INTO credit_sales (customer_id, total_amount, amount_paid, balance, 
                                  due_date, status, notes, created_by, transaction_type, reference, created_at)
        VALUES (?, ?, ?, ?, NULL, 'completed', ?, ?, 'deposit', ?, NOW())
    `, customerID, amount, amount, 0, notes, createdBy, reference)

    if err != nil {
        return fmt.Errorf("failed to record deposit transaction: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit deposit: %w", err)
    }

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

    // Record in credit_sales as a purchase transaction
    _, err = tx.Exec(`
        INSERT INTO credit_sales (customer_id, sale_id, total_amount, amount_paid, balance, 
                                  due_date, status, notes, created_by, transaction_type, created_at)
        VALUES (?, ?, ?, ?, ?, NULL, 'completed', ?, ?, 'purchase', NOW())
    `, customerID, saleID, amount, amount, 0, notes, createdBy)

    if err != nil {
        return fmt.Errorf("failed to record purchase transaction: %w", err)
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
               due_date, status, notes, created_by, created_at, updated_at, 
               COALESCE(transaction_type, 'purchase') as transaction_type,
               COALESCE(reference, '') as reference
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
            &t.TransactionType,
            &t.Reference,
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
