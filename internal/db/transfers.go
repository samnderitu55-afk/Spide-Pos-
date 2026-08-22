package db

import (
    "database/sql"
    "fmt"
    "time"
)

func CreateStockTransfer(db *sql.DB, req *StockTransferRequest, fromShopID int, createdBy string) (*StockTransfer, error) {
    tx, err := db.Begin()
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    transferNumber := fmt.Sprintf("TRF-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)

    var totalItems int
    var totalCost float64
    for _, item := range req.Items {
        totalItems += item.Quantity
        totalCost += float64(item.Quantity) * item.CostPrice
    }

    query := `
        INSERT INTO stock_transfers 
        (transfer_number, from_shop_id, to_shop_id, total_items, total_cost, 
         transfer_date, status, notes, created_by, created_at)
        VALUES (?, ?, ?, ?, ?, ?, 'pending', ?, ?, NOW())
    `
    result, err := tx.Exec(query,
        transferNumber, fromShopID, req.ToShopID,
        totalItems, totalCost, req.TransferDate,
        req.Notes, createdBy,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to insert transfer: %w", err)
    }

    transferID, err := result.LastInsertId()
    if err != nil {
        return nil, fmt.Errorf("failed to get transfer ID: %w", err)
    }

    for _, item := range req.Items {
        itemQuery := `
            INSERT INTO transfer_items (transfer_id, product_id, quantity, cost_price, subtotal)
            VALUES (?, ?, ?, ?, ?)
        `
        subtotal := float64(item.Quantity) * item.CostPrice
        _, err := tx.Exec(itemQuery, transferID, item.ProductID, item.Quantity, item.CostPrice, subtotal)
        if err != nil {
            return nil, fmt.Errorf("failed to insert transfer item: %w", err)
        }

        stockQuery := `
            UPDATE shop_stock 
            SET quantity = quantity - ?, updated_at = NOW()
            WHERE shop_id = ? AND product_id = ? AND quantity >= ?
        `
        _, err = tx.Exec(stockQuery, item.Quantity, fromShopID, item.ProductID, item.Quantity)
        if err != nil {
            return nil, fmt.Errorf("failed to deduct stock: %w", err)
        }

        addQuery := `
            INSERT INTO shop_stock (shop_id, product_id, quantity, created_at, updated_at)
            VALUES (?, ?, ?, NOW(), NOW())
            ON DUPLICATE KEY UPDATE quantity = quantity + ?, updated_at = NOW()
        `
        _, err = tx.Exec(addQuery, req.ToShopID, item.ProductID, item.Quantity, item.Quantity)
        if err != nil {
            return nil, fmt.Errorf("failed to add stock to destination: %w", err)
        }
    }

    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    transfer, err := GetStockTransfer(db, int(transferID))
    return transfer, err
}

func GetStockTransfer(db *sql.DB, id int) (*StockTransfer, error) {
    query := `
        SELECT id, transfer_number, from_shop_id, to_shop_id, total_items, total_cost,
               transfer_date, status, notes, created_by, created_at
        FROM stock_transfers
        WHERE id = ?
    `
    var t StockTransfer
    err := db.QueryRow(query, id).Scan(
        &t.ID, &t.TransferNumber, &t.FromShopID, &t.ToShopID,
        &t.TotalItems, &t.TotalCost, &t.TransferDate,
        &t.Status, &t.Notes, &t.CreatedBy, &t.CreatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &t, nil
}

func GetTransfersByShop(db *sql.DB, shopID int, status string) ([]StockTransfer, error) {
    query := `
        SELECT id, transfer_number, from_shop_id, to_shop_id, total_items, total_cost,
               transfer_date, status, notes, created_by, created_at
        FROM stock_transfers
        WHERE from_shop_id = ? OR to_shop_id = ?
    `
    args := []interface{}{shopID, shopID}

    if status != "" {
        query += " AND status = ?"
        args = append(args, status)
    }
    query += " ORDER BY id DESC"

    rows, err := db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var transfers []StockTransfer
    for rows.Next() {
        var t StockTransfer
        err := rows.Scan(
            &t.ID, &t.TransferNumber, &t.FromShopID, &t.ToShopID,
            &t.TotalItems, &t.TotalCost, &t.TransferDate,
            &t.Status, &t.Notes, &t.CreatedBy, &t.CreatedAt,
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

func GetStockTransferItems(db *sql.DB, transferID int) ([]struct {
    ProductName string
    Quantity    int
    CostPrice   float64
    Subtotal    float64
}, error) {
    query := `
        SELECT p.name, ti.quantity, ti.cost_price, ti.subtotal
        FROM transfer_items ti
        JOIN products p ON ti.product_id = p.id
        WHERE ti.transfer_id = ?
    `
    rows, err := db.Query(query, transferID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var items []struct {
        ProductName string
        Quantity    int
        CostPrice   float64
        Subtotal    float64
    }
    for rows.Next() {
        var item struct {
            ProductName string
            Quantity    int
            CostPrice   float64
            Subtotal    float64
        }
        err := rows.Scan(&item.ProductName, &item.Quantity, &item.CostPrice, &item.Subtotal)
        if err != nil {
            return nil, err
        }
        items = append(items, item)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return items, nil
}

// GetAllTransfers retrieves all transfers (for admin/director)
func GetAllTransfers(db *sql.DB, status string) ([]StockTransfer, error) {
    query := `
        SELECT id, transfer_number, from_shop_id, to_shop_id, total_items, total_cost,
               transfer_date, status, notes, created_by, created_at
        FROM stock_transfers
    `
    args := []interface{}{}
    if status != "" {
        query += " WHERE status = ?"
        args = append(args, status)
    }
    query += " ORDER BY id DESC"

    rows, err := db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var transfers []StockTransfer
    for rows.Next() {
        var t StockTransfer
        err := rows.Scan(
            &t.ID, &t.TransferNumber, &t.FromShopID, &t.ToShopID,
            &t.TotalItems, &t.TotalCost, &t.TransferDate,
            &t.Status, &t.Notes, &t.CreatedBy, &t.CreatedAt,
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


