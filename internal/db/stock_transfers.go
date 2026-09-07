package db

import (
	"database/sql"
)

type TransferItem struct {
	ID          int     `json:"id"`
	TransferID  int     `json:"transfer_id"`
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Barcode     string  `json:"barcode"`
	Quantity    int     `json:"quantity"`
	CostPrice   float64 `json:"cost_price"`
	Subtotal    float64 `json:"subtotal"`
}

func CreateStockTransfer(db *sql.DB, transfer StockTransfer, items []TransferItem) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Insert stock transfer
	query := `
        INSERT INTO stock_transfers (company_id, transfer_number, from_shop_id, to_shop_id, 
            total_items, total_cost, status, notes, created_by, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
    `
	result, err := tx.Exec(query,
		transfer.CompanyID, transfer.TransferNumber, transfer.FromShopID,
		transfer.ToShopID, transfer.TotalItems, transfer.TotalCost,
		transfer.Status, transfer.Notes, transfer.CreatedBy,
	)
	if err != nil {
		return 0, err
	}

	transferID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert transfer items and update stock
	for _, item := range items {
		// Insert transfer item
		itemQuery := `
            INSERT INTO transfer_items (transfer_id, product_id, quantity, cost_price, subtotal)
            VALUES (?, ?, ?, ?, ?)
        `
		_, err := tx.Exec(itemQuery, transferID, item.ProductID, item.Quantity, item.CostPrice, item.Subtotal)
		if err != nil {
			return 0, err
		}

		// Deduct from source shop stock
		stockQuery := `
            UPDATE shop_stock 
            SET quantity = quantity - ?, updated_at = NOW()
            WHERE shop_id = ? AND product_id = ?
        `
		_, err = tx.Exec(stockQuery, item.Quantity, transfer.FromShopID, item.ProductID)
		if err != nil {
			return 0, err
		}

		// Add to destination shop stock (or create if doesn't exist)
		addStockQuery := `
            INSERT INTO shop_stock (company_id, shop_id, product_id, quantity, updated_at)
            VALUES (?, ?, ?, ?, NOW())
            ON DUPLICATE KEY UPDATE quantity = quantity + ?, updated_at = NOW()
        `
		_, err = tx.Exec(addStockQuery, transfer.CompanyID, transfer.ToShopID, item.ProductID, item.Quantity, item.Quantity)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return transferID, nil
}

func GetStockTransfersByCompany(db *sql.DB, companyID int) ([]StockTransfer, error) {
	query := `
        SELECT id, company_id, transfer_number, from_shop_id, to_shop_id, 
               total_items, total_cost, status, notes, created_by, created_at, updated_at
        FROM stock_transfers
        WHERE company_id = ?
        ORDER BY created_at DESC
    `
	rows, err := db.Query(query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transfers []StockTransfer
	for rows.Next() {
		var t StockTransfer
		err := rows.Scan(
			&t.ID, &t.CompanyID, &t.TransferNumber, &t.FromShopID,
			&t.ToShopID, &t.TotalItems, &t.TotalCost, &t.Status,
			&t.Notes, &t.CreatedBy, &t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transfers, nil
}

func GetStockTransferByID(db *sql.DB, transferID int) (*StockTransfer, []TransferItem, error) {
	var transfer StockTransfer
	query := `
        SELECT id, company_id, transfer_number, from_shop_id, to_shop_id, 
              total_items, total_cost, status, notes, created_by, created_at
        FROM stock_transfers
        WHERE id = ?
    `
	err := db.QueryRow(query, transferID).Scan(
		&transfer.ID, &transfer.CompanyID, &transfer.TransferNumber,
		&transfer.FromShopID, &transfer.ToShopID, &transfer.TotalItems,
		&transfer.TotalCost, &transfer.Status, &transfer.Notes,
		&transfer.CreatedBy, &transfer.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	// Get transfer items
	itemsQuery := `
        SELECT ti.id, ti.transfer_id, ti.product_id, p.name as product_name, 
               p.barcode, ti.quantity, ti.cost_price, ti.subtotal
        FROM transfer_items ti
        JOIN products p ON ti.product_id = p.id
        WHERE ti.transfer_id = ?
    `
	rows, err := db.Query(itemsQuery, transferID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []TransferItem
	for rows.Next() {
		var item TransferItem
		err := rows.Scan(
			&item.ID, &item.TransferID, &item.ProductID, &item.ProductName,
			&item.Barcode, &item.Quantity, &item.CostPrice, &item.Subtotal,
		)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return &transfer, items, nil
}

func UpdateStockTransferStatus(db *sql.DB, transferID int, status string) error {
	query := `UPDATE stock_transfers SET status = ?, updated_at = NOW() WHERE id = ?`
	_, err := db.Exec(query, status, transferID)
	return err
}
