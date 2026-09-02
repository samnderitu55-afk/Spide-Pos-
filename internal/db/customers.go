package db

import (
	"database/sql"
	"fmt"
	"log"
)

type Customer struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Phone          string  `json:"phone"`
	Email          string  `json:"email"`
	IDNumber       sql.NullString  `json:"id_number"`
	Address        sql.NullString  `json:"address"`
	Balance        float64 `json:"balance"`
	DepositBalance float64 `json:"deposit_balance"`
	CreditLimit    float64 `json:"credit_limit"`
	Notes          string  `json:"notes"`
	CreatedBy      string  `json:"created_by"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type CreditSale struct {
	ID           int             `json:"id"`
	CustomerID   int             `json:"customer_id"`
	CustomerName string          `json:"customer_name"`
	ShopID       int             `json:"shop_id"`
	SaleID       int             `json:"sale_id"`
	TotalAmount  float64         `json:"total_amount"`
	AmountPaid   float64         `json:"amount_paid"`
	Balance      float64         `json:"balance"`
	DueDate      string          `json:"due_date"`
	Status       string          `json:"status"` // pending, partial, paid, overdue
	Notes        string          `json:"notes"`
	CreatedBy    string          `json:"created_by"`
	CreatedAt    string          `json:"created_at"`
	UpdatedAt    string          `json:"updated_at"`
	Payments     []CreditPayment `json:"payments,omitempty"`
}

type CreditPayment struct {
	ID            int     `json:"id"`
	CreditSaleID  int     `json:"credit_sale_id"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"payment_method"` // cash, mpesa, deposit
	Reference     string  `json:"reference"`
	Notes         string  `json:"notes"`
	CreatedBy     string  `json:"created_by"`
	CreatedAt     string  `json:"created_at"`
}

type CustomerTransaction struct {
    Date        string  `json:"date"`
    Description string  `json:"description"`
    Type        string  `json:"type"` // sale, payment, deposit
    Amount      float64 `json:"amount"`
    SaleID      int     `json:"sale_id,omitempty"`
    Balance     float64 `json:"balance"`
}

type CustomerSummary struct {
    TotalSales     float64 `json:"total_sales"`
    TotalPayments  float64 `json:"total_payments"`
    TotalDeposits  float64 `json:"total_deposits"`
    CurrentBalance float64 `json:"current_balance"`
}


func CreateCustomer(db *sql.DB, customer *Customer) error {
    log.Printf("Creating customer: Name='%s', Phone='%s'", customer.Name, customer.Phone)
    
    // created_by should be the user ID, not username
    // We need to get the user ID from the username
    var userID int
    err := db.QueryRow("SELECT id FROM users WHERE username = ?", customer.CreatedBy).Scan(&userID)
    if err != nil {
        log.Printf("Error getting user ID for '%s': %v", customer.CreatedBy, err)
        // Default to 1 (admin) if user not found
        userID = 1
    }
    log.Printf("Using user_id: %d for created_by", userID)
    
    query := `
        INSERT INTO customers (name, phone, email, id_number, address, 
                               credit_limit, balance, deposit_balance, notes, created_by, created_at)
        VALUES (?, ?, ?, ?, ?, ?, 0, 0, ?, ?, NOW())
    `
    
    result, err := db.Exec(query,
        customer.Name,
        customer.Phone,
        customer.Email,
        customer.IDNumber,
        customer.Address,
        customer.CreditLimit,
        customer.Notes,
        userID,  // Use user ID instead of username
    )
    if err != nil {
        log.Printf("❌ Error creating customer: %v", err)
        return fmt.Errorf("failed to create customer: %w", err)
    }

    id, err := result.LastInsertId()
    if err != nil {
        log.Printf("❌ Error getting last insert ID: %v", err)
        return fmt.Errorf("failed to get last insert ID: %w", err)
    }
    
    customer.ID = int(id)
    log.Printf("✅ Customer created with ID: %d", customer.ID)

    return nil
}

func GetCustomers(db *sql.DB) ([]Customer, error) {
    log.Println("GetCustomers called")
    
    query := `
        SELECT id, name, phone, email, id_number, address, 
               COALESCE(balance, 0) as balance,
               COALESCE(deposit_balance, 0) as deposit_balance,
               COALESCE(credit_limit, 0) as credit_limit,
               COALESCE(notes, '') as notes,
               COALESCE(created_by, '') as created_by,
               created_at, updated_at
        FROM customers
        ORDER BY name
    `
    rows, err := db.Query(query)
    if err != nil {
        log.Printf("GetCustomers query error: %v", err)
        return nil, fmt.Errorf("failed to get customers: %w", err)
    }
    defer rows.Close()

    var customers []Customer
    for rows.Next() {
        var c Customer
        err := rows.Scan(
            &c.ID,
            &c.Name,
            &c.Phone,
            &c.Email,
            &c.IDNumber,
            &c.Address,
            &c.Balance,
            &c.DepositBalance,
            &c.CreditLimit,
            &c.Notes,
            &c.CreatedBy,
            &c.CreatedAt,
            &c.UpdatedAt,
        )
        if err != nil {
            log.Printf("GetCustomers scan error: %v", err)
            return nil, err
        }
        customers = append(customers, c)
    }

    if err := rows.Err(); err != nil {
        log.Printf("GetCustomers rows error: %v", err)
        return nil, fmt.Errorf("error iterating customers: %w", err)
    }

    log.Printf("GetCustomers found %d customers", len(customers))
    return customers, nil
}

func GetCustomerByID(db *sql.DB, id int) (*Customer, error) {
	query := `
        SELECT id, name, phone, email, id_number, address, 
               COALESCE(balance, 0) as balance,
               COALESCE(deposit_balance, 0) as deposit_balance,
               COALESCE(credit_limit, 0) as credit_limit,
               notes, created_by, created_at, updated_at
        FROM customers
        WHERE id = ?
    `
	var c Customer
	err := db.QueryRow(query, id).Scan(
		&c.ID,
		&c.Name,
		&c.Phone,
		&c.Email,
		&c.IDNumber,
		&c.Address,
		&c.Balance,
		&c.DepositBalance,
		&c.CreditLimit,
		&c.Notes,
		&c.CreatedBy,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &c, nil
}

func UpdateCustomer(db *sql.DB, customer *Customer) error {
	query := `
        UPDATE customers 
        SET name = ?, phone = ?, email = ?, id_number = ?, address = ?,
            credit_limit = ?, notes = ?, updated_at = NOW()
        WHERE id = ?
    `
	_, err := db.Exec(query,
		customer.Name,
		customer.Phone,
		customer.Email,
		customer.IDNumber,
		customer.Address,
		customer.CreditLimit,
		customer.Notes,
		customer.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

func DeleteCustomer(db *sql.DB, id int) error {
	// Check if customer has any credit sales
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM credit_sales WHERE customer_id = ?", id).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check credit sales: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("cannot delete customer with credit sales")
	}

	query := `DELETE FROM customers WHERE id = ?`
	_, err = db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	return nil
}

// Search customers by name or phone
func SearchCustomers(db *sql.DB, query string) ([]Customer, error) {
	sqlQuery := `
        SELECT id, name, phone, email, id_number, address, 
               COALESCE(balance, 0) as balance,
               COALESCE(deposit_balance, 0) as deposit_balance,
               COALESCE(credit_limit, 0) as credit_limit,
               notes, created_by, created_at, updated_at
        FROM customers
        WHERE name LIKE ? OR phone LIKE ?
        ORDER BY name
        LIMIT 20
    `
	searchTerm := "%" + query + "%"
	rows, err := db.Query(sqlQuery, searchTerm, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}
	defer rows.Close()

	var customers []Customer
	for rows.Next() {
		var c Customer
		err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Phone,
			&c.Email,
			&c.IDNumber,
			&c.Address,
			&c.Balance,
			&c.DepositBalance,
			&c.CreditLimit,
			&c.Notes,
			&c.CreatedBy,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		customers = append(customers, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customers: %w", err)
	}

	return customers, nil
}

func CreateCreditSale(db *sql.DB, creditSale *CreditSale) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get customer's current balance
	var currentBalance float64
	err = tx.QueryRow("SELECT COALESCE(balance, 0) FROM customers WHERE id = ?", creditSale.CustomerID).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("failed to get customer balance: %w", err)
	}

	// Insert credit sale
	query := `
        INSERT INTO credit_sales (customer_id, shop_id, sale_id, total_amount, amount_paid, 
                                  balance, due_date, status, notes, created_by, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
    `
	newBalance := currentBalance + creditSale.Balance
	result, err := tx.Exec(query,
		creditSale.CustomerID,
		creditSale.ShopID,
		creditSale.SaleID,
		creditSale.TotalAmount,
		creditSale.AmountPaid,
		creditSale.Balance,
		creditSale.DueDate,
		creditSale.Status,
		creditSale.Notes,
		creditSale.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to create credit sale: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		creditSale.ID = int(id)
	}

	// Update customer balance
	_, err = tx.Exec(`
        UPDATE customers SET balance = ?, updated_at = NOW() WHERE id = ?
    `, newBalance, creditSale.CustomerID)
	if err != nil {
		return fmt.Errorf("failed to update customer balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit credit sale: %w", err)
	}

	return nil
}

func GetCreditSales(db *sql.DB, customerID int) ([]CreditSale, error) {
	var query string
	var args []interface{}

	if customerID > 0 {
		query = `
            SELECT cs.id, cs.customer_id, c.name as customer_name, cs.shop_id, cs.sale_id,
                   cs.total_amount, cs.amount_paid, cs.balance, cs.due_date, cs.status,
                   cs.notes, cs.created_by, cs.created_at, cs.updated_at
            FROM credit_sales cs
            JOIN customers c ON cs.customer_id = c.id
            WHERE cs.customer_id = ?
            ORDER BY cs.created_at DESC
        `
		args = append(args, customerID)
	} else {
		query = `
            SELECT cs.id, cs.customer_id, c.name as customer_name, cs.shop_id, cs.sale_id,
                   cs.total_amount, cs.amount_paid, cs.balance, cs.due_date, cs.status,
                   cs.notes, cs.created_by, cs.created_at, cs.updated_at
            FROM credit_sales cs
            JOIN customers c ON cs.customer_id = c.id
            WHERE cs.balance > 0
            ORDER BY cs.due_date ASC
        `
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get credit sales: %w", err)
	}
	defer rows.Close()

	var sales []CreditSale
	for rows.Next() {
		var s CreditSale
		err := rows.Scan(
			&s.ID,
			&s.CustomerID,
			&s.CustomerName,
			&s.ShopID,
			&s.SaleID,
			&s.TotalAmount,
			&s.AmountPaid,
			&s.Balance,
			&s.DueDate,
			&s.Status,
			&s.Notes,
			&s.CreatedBy,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		sales = append(sales, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating credit sales: %w", err)
	}

	return sales, nil
}

func GetCreditPayments(db *sql.DB, creditSaleID int) ([]CreditPayment, error) {
	query := `
        SELECT id, credit_sale_id, amount, payment_method, reference, notes, created_by, created_at
        FROM credit_payments
        WHERE credit_sale_id = ?
        ORDER BY created_at DESC
    `
	rows, err := db.Query(query, creditSaleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credit payments: %w", err)
	}
	defer rows.Close()

	var payments []CreditPayment
	for rows.Next() {
		var p CreditPayment
		err := rows.Scan(
			&p.ID,
			&p.CreditSaleID,
			&p.Amount,
			&p.PaymentMethod,
			&p.Reference,
			&p.Notes,
			&p.CreatedBy,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating credit payments: %w", err)
	}

	return payments, nil
}

func AddCreditPayment(db *sql.DB, payment *CreditPayment) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the credit sale
	var creditSale CreditSale
	err = tx.QueryRow(`
        SELECT id, customer_id, balance, total_amount, amount_paid
        FROM credit_sales WHERE id = ?
    `, payment.CreditSaleID).Scan(
		&creditSale.ID,
		&creditSale.CustomerID,
		&creditSale.Balance,
		&creditSale.TotalAmount,
		&creditSale.AmountPaid,
	)
	if err != nil {
		return fmt.Errorf("failed to get credit sale: %w", err)
	}

	if payment.Amount > creditSale.Balance {
		return fmt.Errorf("payment amount exceeds remaining balance")
	}

	// Insert payment
	query := `
        INSERT INTO credit_payments (credit_sale_id, amount, payment_method, reference, notes, created_by, created_at)
        VALUES (?, ?, ?, ?, ?, ?, NOW())
    `
	_, err = tx.Exec(query,
		payment.CreditSaleID,
		payment.Amount,
		payment.PaymentMethod,
		payment.Reference,
		payment.Notes,
		payment.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	// Update credit sale
	newBalance := creditSale.Balance - payment.Amount
	newAmountPaid := creditSale.AmountPaid + payment.Amount
	status := "partial"
	if newBalance <= 0 {
		status = "paid"
	}

	_, err = tx.Exec(`
        UPDATE credit_sales 
        SET balance = ?, amount_paid = ?, status = ?, updated_at = NOW()
        WHERE id = ?
    `, newBalance, newAmountPaid, status, payment.CreditSaleID)
	if err != nil {
		return fmt.Errorf("failed to update credit sale: %w", err)
	}

	// Update customer balance
	_, err = tx.Exec(`
        UPDATE customers 
        SET balance = balance - ?, updated_at = NOW()
        WHERE id = ?
    `, payment.Amount, creditSale.CustomerID)
	if err != nil {
		return fmt.Errorf("failed to update customer balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit payment: %w", err)
	}

	return nil
}

func GetCustomerTransactions(db *sql.DB, customerID int) ([]CustomerTransaction, error) {
    var transactions []CustomerTransaction
    var runningBalance float64

    // Get sales transactions
    salesQuery := `
        SELECT 
            DATE(created_at) as date,
            CONCAT('Sale #', id) as description,
            'sale' as type,
            total_amount as amount,
            id as sale_id
        FROM sales
        WHERE customer_id = ?
        ORDER BY created_at ASC
    `
    salesRows, err := db.Query(salesQuery, customerID)
    if err != nil {
        return nil, fmt.Errorf("failed to get sales: %w", err)
    }
    defer salesRows.Close()

    for salesRows.Next() {
        var t CustomerTransaction
        err := salesRows.Scan(&t.Date, &t.Description, &t.Type, &t.Amount, &t.SaleID)
        if err != nil {
            return nil, fmt.Errorf("failed to scan sale row: %w", err)
        }
        // Sales increase what customer owes
        runningBalance += t.Amount
        t.Balance = runningBalance
        transactions = append(transactions, t)
    }

    // ✅ Check for salesRows iteration errors
    if err = salesRows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating sales: %w", err)
    }

    // Get deposits (payments made by customer)
    depositsQuery := `
        SELECT 
            DATE(created_at) as date,
            CONCAT('Deposit - ', COALESCE(payment_method, 'cash')) as description,
            'deposit' as type,
            amount as amount,
            0 as sale_id
        FROM customer_deposits
        WHERE customer_id = ?
        ORDER BY created_at ASC
    `
    depositRows, err := db.Query(depositsQuery, customerID)
    if err != nil {
        // Table might not exist - log but don't fail
        log.Printf("⚠️ Could not fetch deposits: %v", err)
        return transactions, nil
    }
    defer depositRows.Close()

    for depositRows.Next() {
        var t CustomerTransaction
        err := depositRows.Scan(&t.Date, &t.Description, &t.Type, &t.Amount, &t.SaleID)
        if err != nil {
            log.Printf("⚠️ Error scanning deposit row: %v", err)
            continue
        }
        // Deposits reduce what customer owes
        runningBalance -= t.Amount
        t.Balance = runningBalance
        transactions = append(transactions, t)
    }

    // ✅ Check for depositRows iteration errors
    if err = depositRows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating deposits: %w", err)
    }

    return transactions, nil
}

func GetCustomerSummary(db *sql.DB, customerID int) (CustomerSummary, error) {
    var summary CustomerSummary

    // Get total sales for this customer
    salesQuery := `
        SELECT COALESCE(SUM(total_amount), 0) 
        FROM sales 
        WHERE customer_id = ?
    `
    err := db.QueryRow(salesQuery, customerID).Scan(&summary.TotalSales)
    if err != nil {
        return summary, fmt.Errorf("failed to get total sales: %w", err)
    }

    // ✅ Get total deposits from customer_deposits table
    depositsQuery := `
        SELECT COALESCE(SUM(amount), 0) 
        FROM customer_deposits 
        WHERE customer_id = ?
    `
    err = db.QueryRow(depositsQuery, customerID).Scan(&summary.TotalDeposits)
    if err != nil {
        log.Printf("⚠️ Could not get deposits: %v", err)
        summary.TotalDeposits = 0
    }

    // ✅ Get current balance from customers table (deposit_balance)
    balanceQuery := `
        SELECT COALESCE(deposit_balance, 0) 
        FROM customers 
        WHERE id = ?
    `
    err = db.QueryRow(balanceQuery, customerID).Scan(&summary.CurrentBalance)
    if err != nil {
        return summary, fmt.Errorf("failed to get current balance: %w", err)
    }

    log.Printf("📊 Customer %d summary - Sales: %.2f, Deposits: %.2f, Balance: %.2f", 
        customerID, summary.TotalSales, summary.TotalDeposits, summary.CurrentBalance)

    return summary, nil
}
