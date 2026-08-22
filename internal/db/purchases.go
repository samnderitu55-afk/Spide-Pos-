package db

import (
	"database/sql"
	"fmt"
	"time"
)

func CreatePurchase(db *sql.DB, req *PurchaseRequest, createdBy string) (*Purchase, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	purchaseNumber := fmt.Sprintf("PO-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)

	var totalItems int
	var totalCost float64
	for _, item := range req.Items {
		totalItems += item.Quantity
		totalCost += float64(item.Quantity) * item.CostPrice
	}

	query := `
        INSERT INTO purchases 
        (purchase_number, supplier_id, shop_id, total_items, total_cost, purchase_date, notes, created_by, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
    `
	result, err := tx.Exec(query,
		purchaseNumber, req.SupplierID, totalItems,
		totalCost, req.PurchaseDate, req.Notes, createdBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert purchase: %w", err)
	}

	purchaseID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase ID: %w", err)
	}

	for _, item := range req.Items {
		itemQuery := `
            INSERT INTO purchase_items (purchase_id, product_id, quantity, cost_price, subtotal)
            VALUES (?, ?, ?, ?, ?)
        `
		subtotal := float64(item.Quantity) * item.CostPrice
		_, err := tx.Exec(itemQuery, purchaseID, item.ProductID, item.Quantity, item.CostPrice, subtotal)
		if err != nil {
			return nil, fmt.Errorf("failed to insert purchase item: %w", err)
		}

		costQuery := `
            UPDATE products 
            SET cost_price = ((cost_price * stock_quantity) + (? * ?)) / (stock_quantity + ?),
                stock_quantity = stock_quantity + ?,
                updated_at = NOW()
            WHERE id = ?
        `
		_, err = tx.Exec(costQuery,
			item.CostPrice, item.Quantity, item.Quantity,
			item.Quantity, item.ProductID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update product cost: %w", err)
		}

		stockQuery := `
            INSERT INTO shop_stock (shop_id, product_id, quantity, created_at, updated_at)
            VALUES (1, ?, ?, NOW(), NOW())
            ON DUPLICATE KEY UPDATE quantity = quantity + ?, updated_at = NOW()
        `
		_, err = tx.Exec(stockQuery, item.ProductID, item.Quantity, item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to update shop stock: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	purchase, err := GetPurchase(db, int(purchaseID))
	return purchase, err
}

func GetPurchase(db *sql.DB, id int) (*Purchase, error) {
	query := `
        SELECT p.id, p.purchase_number, p.supplier_id, s.name as supplier_name,
               p.total_items, p.total_cost, p.purchase_date, p.notes, p.created_by, p.created_at
        FROM purchases p
        JOIN suppliers s ON p.supplier_id = s.id
        WHERE p.id = ?
    `
	var purchase Purchase
	err := db.QueryRow(query, id).Scan(
		&purchase.ID, &purchase.PurchaseNumber, &purchase.SupplierID,
		&purchase.SupplierName, &purchase.TotalItems, &purchase.TotalCost,
		&purchase.PurchaseDate, &purchase.Notes, &purchase.CreatedBy,
		&purchase.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &purchase, nil
}

func GetPurchaseItems(db *sql.DB, purchaseID int) ([]struct {
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	CostPrice   float64 `json:"cost_price"`
	Subtotal    float64 `json:"subtotal"`
}, error) {
	query := `
        SELECT p.name, pi.quantity, pi.cost_price, pi.subtotal
        FROM purchase_items pi
        JOIN products p ON pi.product_id = p.id
        WHERE pi.purchase_id = ?
    `
	rows, err := db.Query(query, purchaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []struct {
		ProductName string  `json:"product_name"`
		Quantity    int     `json:"quantity"`
		CostPrice   float64 `json:"cost_price"`
		Subtotal    float64 `json:"subtotal"`
	}
	for rows.Next() {
		var item struct {
			ProductName string  `json:"product_name"`
			Quantity    int     `json:"quantity"`
			CostPrice   float64 `json:"cost_price"`
			Subtotal    float64 `json:"subtotal"`
		}
		err := rows.Scan(&item.ProductName, &item.Quantity, &item.CostPrice, &item.Subtotal)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	// Check for errors after the loop
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating purchase items: %w", err)
	}

	return items, nil
}

func GetAllSuppliers(db *sql.DB) ([]Supplier, error) {
	query := `
        SELECT id, name, contact_person, phone, email, address, notes, is_active
        FROM suppliers
        WHERE is_active = 1
        ORDER BY name ASC
    `
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []Supplier
	for rows.Next() {
		var s Supplier
		err := rows.Scan(
			&s.ID, &s.Name, &s.ContactPerson, &s.Phone,
			&s.Email, &s.Address, &s.Notes, &s.IsActive,
		)
		if err != nil {
			return nil, err
		}
		suppliers = append(suppliers, s)
	}

	// Check for errors after the loop
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating suppliers: %w", err)
	}

	return suppliers, nil
}

func CreateSupplier(db *sql.DB, s *Supplier) (int64, error) {
	query := `
        INSERT INTO suppliers 
        (name, contact_person, phone, email, address, notes, is_active)
        VALUES (?, ?, ?, ?, ?, ?, 1)
    `
	result, err := db.Exec(query,
		s.Name, s.ContactPerson, s.Phone,
		s.Email, s.Address, s.Notes,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create supplier: %w", err)
	}
	return result.LastInsertId()
}


