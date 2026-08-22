package db

import (
    "database/sql"
    "fmt"
)

func CreateSale(db *sql.DB, req SaleRequest) (int64, error) {
    tx, err := db.Begin()
    if err != nil {
        return 0, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Set default shop_id if not provided
    if req.ShopID == 0 {
        req.ShopID = 1
    }

    saleQuery := `
        INSERT INTO sales (total_amount, cash_amount, mpesa_amount, mpesa_code, 
                           payment_type, shop_id, created_at)
        VALUES (?, ?, ?, ?, ?, ?, NOW())
    `
    result, err := tx.Exec(saleQuery,
        req.TotalAmount, req.CashAmount, req.MpesaAmount,
        req.MpesaCode, req.PaymentType, req.ShopID,
    )
    if err != nil {
        return 0, fmt.Errorf("failed to insert sale: %w", err)
    }

    saleID, err := result.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("failed to get sale ID: %w", err)
    }

    for _, item := range req.Items {
        // Insert sale item
        itemQuery := `
            INSERT INTO sale_items (sale_id, product_id, quantity, unit_price, subtotal)
            VALUES (?, ?, ?, ?, ?)
        `
        _, err := tx.Exec(itemQuery, saleID, item.ProductID, item.Quantity, item.UnitPrice, item.Subtotal)
        if err != nil {
            return 0, fmt.Errorf("failed to insert sale item: %w", err)
        }

        // Update shop_stock (not products!)
        stockQuery := `
            UPDATE shop_stock 
            SET quantity = quantity - ?, updated_at = NOW()
            WHERE shop_id = ? AND product_id = ? AND quantity >= ?
        `
        result, err := tx.Exec(stockQuery, item.Quantity, req.ShopID, item.ProductID, item.Quantity)
        if err != nil {
            return 0, fmt.Errorf("failed to update shop stock: %w", err)
        }

        // Check if stock was updated
        rowsAffected, _ := result.RowsAffected()
        if rowsAffected == 0 {
            return 0, fmt.Errorf("insufficient stock for product %d in shop %d", item.ProductID, req.ShopID)
        }
    }

    if err := tx.Commit(); err != nil {
        return 0, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return saleID, nil
}

// GetRecentSalesForShop - filter by shop_id
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

    var sales []Sale
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

    var sales []Sale
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

    var items []SaleItem
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
