package db

import (
	"database/sql"
	"fmt"
	"time"
)

func CreateTransfer(db *sql.DB, transfer *StockTransfer) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Generate transfer number
	transferNumber := "TRF-" + time.Now().Format("20060102") + "-" + fmt.Sprintf("%04d", time.Now().UnixNano()%10000)

	// Calculate total
	var totalItems int
	var totalCost float64
	for _, item := range transfer.Items {
		totalItems += item.Quantity
		totalCost += float64(item.Quantity) * item.CostPrice
	}

	// Insert transfer
	query := `
        INSERT INTO stock_transfers (transfer_number, from_shop_id, to_shop_id, 
                                     total_items, total_cost, transfer_date, 
                                     status, notes, created_by, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
    `
	result, err := tx.Exec(query,
		transferNumber,
		transfer.FromShopID,
		transfer.ToShopID,
		totalItems,
		totalCost,
		transfer.TransferDate,
		transfer.Status,
		transfer.Notes,
		transfer.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to insert transfer: %w", err)
	}

	transferID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get transfer ID: %w", err)
	}
	transfer.ID = int(transferID)

	// Insert transfer items and update stock
	for _, item := range transfer.Items {
		// Insert transfer item
		itemQuery := `
            INSERT INTO transfer_items (transfer_id, product_id, quantity, cost_price, subtotal)
            VALUES (?, ?, ?, ?, ?)
        `
		subtotal := float64(item.Quantity) * item.CostPrice
		_, err := tx.Exec(itemQuery, transferID, item.ProductID, item.Quantity, item.CostPrice, subtotal)
		if err != nil {
			return fmt.Errorf("failed to insert transfer item: %w", err)
		}

		// Deduct from source shop
		_, err = tx.Exec(`
            UPDATE shop_stock 
            SET quantity = quantity - ?, updated_at = NOW()
            WHERE shop_id = ? AND product_id = ? AND quantity >= ?
        `, item.Quantity, transfer.FromShopID, item.ProductID, item.Quantity)
		if err != nil {
			return fmt.Errorf("failed to deduct from source shop: %w", err)
		}

		// Add to destination shop
		_, err = tx.Exec(`
            INSERT INTO shop_stock (shop_id, product_id, quantity, created_at, updated_at)
            VALUES (?, ?, ?, NOW(), NOW())
            ON DUPLICATE KEY UPDATE quantity = quantity + ?, updated_at = NOW()
        `, transfer.ToShopID, item.ProductID, item.Quantity, item.Quantity)
		if err != nil {
			return fmt.Errorf("failed to add to destination shop: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transfer: %w", err)
	}

	return nil
}

func GetTransfers(db *sql.DB, shopID int) ([]StockTransfer, error) {
	query := `
        SELECT id, transfer_number, from_shop_id, to_shop_id, 
               total_items, total_cost, transfer_date, status, 
               notes, created_by, created_at
        FROM stock_transfers
        WHERE from_shop_id = ? OR to_shop_id = ?
        ORDER BY created_at DESC
    `
	rows, err := db.Query(query, shopID, shopID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfers: %w", err)
	}
	defer rows.Close()

	transfers := []StockTransfer{}
	for rows.Next() {
		var t StockTransfer
		err := rows.Scan(
			&t.ID,
			&t.TransferNumber,
			&t.FromShopID,
			&t.ToShopID,
			&t.TotalItems,
			&t.TotalCost,
			&t.TransferDate,
			&t.Status,
			&t.Notes,
			&t.CreatedBy,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transfers: %w", err)
	}

	return transfers, nil
}

func GetTransferDetail(db *sql.DB, transferID int) (*StockTransfer, error) {
	// Get transfer header
	query := `
        SELECT id, transfer_number, from_shop_id, to_shop_id, 
               total_items, total_cost, transfer_date, status, 
               notes, created_by, created_at
        FROM stock_transfers
        WHERE id = ?
    `
	var t StockTransfer
	err := db.QueryRow(query, transferID).Scan(
		&t.ID,
		&t.TransferNumber,
		&t.FromShopID,
		&t.ToShopID,
		&t.TotalItems,
		&t.TotalCost,
		&t.TransferDate,
		&t.Status,
		&t.Notes,
		&t.CreatedBy,
		&t.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transfer not found")
		}
		return nil, fmt.Errorf("failed to get transfer: %w", err)
	}

	// Get transfer items
	itemQuery := `
        SELECT ti.id, ti.product_id, p.name, ti.quantity, ti.cost_price, ti.subtotal
        FROM transfer_items ti
        JOIN products p ON ti.product_id = p.id
        WHERE ti.transfer_id = ?
    `
	rows, err := db.Query(itemQuery, transferID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfer items: %w", err)
	}
	defer rows.Close()

	items := []struct {
		ID          int     `json:"id"`
		ProductID   int     `json:"product_id"`
		ProductName string  `json:"product_name"`
		Quantity    int     `json:"quantity"`
		CostPrice   float64 `json:"cost_price"`
		Subtotal    float64 `json:"subtotal"`
	}{}

	for rows.Next() {
		var item struct {
			ID          int     `json:"id"`
			ProductID   int     `json:"product_id"`
			ProductName string  `json:"product_name"`
			Quantity    int     `json:"quantity"`
			CostPrice   float64 `json:"cost_price"`
			Subtotal    float64 `json:"subtotal"`
		}
		err := rows.Scan(&item.ID, &item.ProductID, &item.ProductName, &item.Quantity, &item.CostPrice, &item.Subtotal)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transfer items: %w", err)
	}

	// Add items to transfer
	// We'll store them as a JSON field or in a separate struct
	// For simplicity, we'll just return the header and items separately

	return &t, nil
}

func GetTransferItems(db *sql.DB, transferID int) ([]map[string]interface{}, error) {
	query := `
        SELECT ti.id, ti.product_id, p.name as product_name, 
               COALESCE(p.barcode, '') as barcode,
               ti.quantity, ti.cost_price, ti.subtotal
        FROM transfer_items ti
        JOIN products p ON ti.product_id = p.id
        WHERE ti.transfer_id = ?
    `
	rows, err := db.Query(query, transferID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfer items: %w", err)
	}
	defer rows.Close()

	items := []map[string]interface{}{}
	for rows.Next() {
		var id, productID int
		var productName, barcode string
		var quantity int
		var costPrice, subtotal float64

		err := rows.Scan(&id, &productID, &productName, &barcode, &quantity, &costPrice, &subtotal)
		if err != nil {
			return nil, err
		}

		item := map[string]interface{}{
			"id":           id,
			"product_id":   productID,
			"product_name": productName,
			"barcode":      barcode,
			"quantity":     quantity,
			"cost_price":   costPrice,
			"subtotal":     subtotal,
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transfer items: %w", err)
	}

	return items, nil
}
