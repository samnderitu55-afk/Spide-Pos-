package db

import (
	"database/sql"
	"fmt"
	"log"
)

func CreateSale(db *sql.DB, req SaleRequest) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Use shop_id from request or default to 1
	shopID := req.ShopID
	if shopID == 0 {
		shopID = 1
	}

	// Get company_id from the shop
	var companyID int
	err = tx.QueryRow("SELECT company_id FROM shops WHERE id = ?", shopID).Scan(&companyID)
	if err != nil {
		companyID = 1 // Default if not found
	}

	log.Printf("📝 Creating sale for shop: %d, company: %d, total: %.2f", shopID, companyID, req.TotalAmount)

	// ✅ Calculate total from items
	var calculatedTotal float64
	for _, item := range req.Items {
		calculatedTotal += item.Subtotal
	}

	// ✅ Use calculated total if mismatch
	if req.TotalAmount != calculatedTotal {
		log.Printf("⚠️ Total mismatch - Request: %.2f, Calculated: %.2f", req.TotalAmount, calculatedTotal)
		req.TotalAmount = calculatedTotal
	}

	// ✅ Insert sale
	saleQuery := `
        INSERT INTO sales (total_amount, cash_amount, mpesa_amount, credit_amount, deposit_amount, 
                           mpesa_code, payment_type, change_given, shop_id, customer_id, company_id, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
    `
	result, err := tx.Exec(saleQuery,
		req.TotalAmount,
		req.CashAmount,
		req.MpesaAmount,
		req.CreditAmount,
		req.DepositAmount, // ← NEW
		req.MpesaCode,
		req.PaymentType,
		req.ChangeGiven,
		shopID,
		req.CustomerID,
		companyID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert sale: %w", err)
	}

	saleID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get sale ID: %w", err)
	}

	log.Printf("✅ Sale inserted with ID: %d", saleID)

	// ✅ Insert sale items and deduct stock
	for _, item := range req.Items {
		// Insert sale item
		itemQuery := `
            INSERT INTO sale_items (sale_id, product_id, quantity, unit_price, subtotal, company_id)
            VALUES (?, ?, ?, ?, ?, ?)
        `
		_, err := tx.Exec(itemQuery, saleID, item.ProductID, item.Quantity, item.UnitPrice, item.Subtotal, companyID)
		if err != nil {
			return 0, fmt.Errorf("failed to insert sale item: %w", err)
		}

		// Deduct stock
		stockQuery := `
            UPDATE shop_stock 
            SET quantity = quantity - ?, updated_at = NOW()
            WHERE shop_id = ? AND product_id = ? AND company_id = ? AND quantity >= ?
        `
		result, err := tx.Exec(stockQuery, item.Quantity, shopID, item.ProductID, companyID, item.Quantity)
		if err != nil {
			return 0, fmt.Errorf("failed to update shop stock: %w", err)
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			log.Printf("⚠️ Warning: No stock deducted for product %d - insufficient stock", item.ProductID)
		}
	}

	// ✅ Update customer balance if credit was used
	if req.CreditAmount > 0 && req.CustomerID > 0 {
		_, err = tx.Exec(`
            UPDATE customers 
            SET balance = balance + ? 
            WHERE id = ?
        `, req.CreditAmount, req.CustomerID)
		if err != nil {
			return 0, fmt.Errorf("failed to update customer credit balance: %w", err)
		}

		// Record credit sale
		_, err = tx.Exec(`
            INSERT INTO credit_sales (customer_id, shop_id, sale_id, total_amount, balance, status, created_at, company_id)
            VALUES (?, ?, ?, ?, ?, 'pending', NOW(), ?)
        `, req.CustomerID, shopID, saleID, req.CreditAmount, req.CreditAmount, companyID)
		if err != nil {
			log.Printf("⚠️ Warning: Failed to record credit sale: %v", err)
		}
	}

	// ✅ Update deposit balance
	if req.DepositAmount > 0 && req.CustomerID > 0 {
		_, err = tx.Exec(`
            UPDATE customers 
            SET deposit_balance = deposit_balance - ? 
            WHERE id = ?
        `, req.DepositAmount, req.CustomerID)
		if err != nil {
			return 0, fmt.Errorf("failed to update customer deposit balance: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("✅ Sale %d completed successfully", saleID)
	return saleID, nil
}

func GetRecentSales(db *sql.DB, limit int) ([]Sale, error) {
	query := `
        SELECT id, total_amount, cash_amount, mpesa_amount, mpesa_code, 
               payment_type, shop_id, created_at
        FROM sales
        ORDER BY id DESC
        LIMIT ?
    `
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sales := []Sale{}
	for rows.Next() {
		var s Sale
		err := rows.Scan(
			&s.ID, &s.TotalAmount, &s.CashAmount, &s.MpesaAmount,
			&s.MpesaCode, &s.PaymentType, &s.ShopID, &s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		sales = append(sales, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sales: %w", err)
	}

	// Get items for each sale
	for i := range sales {
		items, err := getSaleItems(db, sales[i].ID)
		if err != nil {
			continue
		}
		sales[i].Items = items
	}

	return sales, nil
}

func GetRecentSalesForShop(db *sql.DB, shopID int, limit int) ([]Sale, error) {
	query := `
        SELECT id, total_amount, cash_amount, mpesa_amount, mpesa_code, 
               payment_type, shop_id, created_at
        FROM sales
        WHERE shop_id = ?
        ORDER BY id DESC
        LIMIT ?
    `
	rows, err := db.Query(query, shopID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sales := []Sale{}
	for rows.Next() {
		var s Sale
		err := rows.Scan(
			&s.ID, &s.TotalAmount, &s.CashAmount, &s.MpesaAmount,
			&s.MpesaCode, &s.PaymentType, &s.ShopID, &s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		sales = append(sales, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sales: %w", err)
	}

	for i := range sales {
		items, err := getSaleItems(db, sales[i].ID)
		if err != nil {
			continue
		}
		sales[i].Items = items
	}

	return sales, nil
}

func getSaleItems(db *sql.DB, saleID int64) ([]SaleItem, error) {
	query := `
        SELECT si.id, si.sale_id, si.product_id, p.name, si.quantity, si.unit_price, si.subtotal
        FROM sale_items si
        JOIN products p ON si.product_id = p.id
        WHERE si.sale_id = ?
    `
	rows, err := db.Query(query, saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []SaleItem{}
	for rows.Next() {
		var item SaleItem
		err := rows.Scan(
			&item.ID, &item.SaleID, &item.ProductID,
			&item.ProductName, &item.Quantity, &item.UnitPrice, &item.Subtotal,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sale items: %w", err)
	}

	return items, nil
}

func GetRecentSalesForShopAndCompany(db *sql.DB, shopID, companyID int, limit int) ([]Sale, error) {
	query := `
        SELECT id, total_amount, cash_amount, mpesa_amount, credit_amount, deposit_amount,
               mpesa_code, payment_type, shop_id, customer_id, created_at
        FROM sales
        WHERE shop_id = ? AND company_id = ?
        ORDER BY created_at DESC
        LIMIT ?
    `
	rows, err := db.Query(query, shopID, companyID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sales := []Sale{}
	for rows.Next() {
		var s Sale
		err := rows.Scan(
			&s.ID,
			&s.TotalAmount,
			&s.CashAmount,
			&s.MpesaAmount,
			&s.CreditAmount,  // ← new
			&s.DepositAmount, // ← new
			&s.MpesaCode,
			&s.PaymentType,
			&s.ShopID,
			&s.CustomerID,
			&s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// ✅ Attach the sale's items so reprints and detail views work
		items, err := getSaleItems(db, s.ID)
		if err == nil {
			s.Items = items
		}

		sales = append(sales, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return sales, nil
}

func GetSalesByCompany(db *sql.DB, companyID int, limit int) ([]Sale, error) {
	query := `
        SELECT id, total_amount, cash_amount, mpesa_amount, mpesa_code, 
               payment_type, shop_id, customer_id, created_at
        FROM sales
        WHERE company_id = ?
        ORDER BY created_at DESC
        LIMIT ?
    `
	rows, err := db.Query(query, companyID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sales := []Sale{}
	for rows.Next() {
		var s Sale
		err := rows.Scan(
			&s.ID, &s.TotalAmount, &s.CashAmount, &s.MpesaAmount,
			&s.MpesaCode, &s.PaymentType, &s.ShopID, &s.CustomerID, &s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		sales = append(sales, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return sales, nil
}
