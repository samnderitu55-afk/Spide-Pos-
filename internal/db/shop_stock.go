package db

import (
	"database/sql"
	"fmt"
)

// Get stock for a specific product across all shops in a company
func GetProductStockByCompany(db *sql.DB, productID, companyID int) ([]ShopStock, error) {
	query := `
        SELECT ss.id, ss.company_id, ss.shop_id, ss.product_id, ss.quantity, ss.updated_at,
               s.name as shop_name
        FROM shop_stock ss
        JOIN shops s ON ss.shop_id = s.id
        WHERE ss.product_id = ? AND ss.company_id = ?
        ORDER BY s.name
    `
	rows, err := db.Query(query, productID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []ShopStock
	for rows.Next() {
		var s ShopStock
		err := rows.Scan(
			&s.ID, &s.CompanyID, &s.ShopID, &s.ProductID,
			&s.Quantity, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stocks, nil
}

// CreateShopStock - Add stock to a specific shop
func CreateShopStock(db *sql.DB, shopID, productID, quantity, companyID int) error {
	query := `
        INSERT INTO shop_stock (company_id, shop_id, product_id, quantity, created_at, updated_at)
        VALUES (?, ?, ?, ?, NOW(), NOW())
        ON DUPLICATE KEY UPDATE quantity = quantity + ?, updated_at = NOW()
    `
	_, err := db.Exec(query, companyID, shopID, productID, quantity, quantity)
	return err
}

// GetShopStock - Get stock for a specific shop
func GetShopStock(db *sql.DB, shopID int) ([]ShopStock, error) {
	query := `
        SELECT id, company_id, shop_id, product_id, quantity, created_at, updated_at
        FROM shop_stock
        WHERE shop_id = ?
        ORDER BY product_id
    `
	rows, err := db.Query(query, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []ShopStock
	for rows.Next() {
		var s ShopStock
		err := rows.Scan(
			&s.ID, &s.CompanyID, &s.ShopID, &s.ProductID,
			&s.Quantity, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stocks, nil
}

// UpdateShopStock - Update stock quantity for a specific shop and product
func UpdateShopStock(db *sql.DB, shopID, productID, quantity int) error {
	query := `
        UPDATE shop_stock 
        SET quantity = ?, updated_at = NOW()
        WHERE shop_id = ? AND product_id = ?
    `
	_, err := db.Exec(query, quantity, shopID, productID)
	return err
}

// DeductShopStock - Deduct stock from a specific shop
func DeductShopStock(db *sql.DB, shopID, productID, quantity int) error {
	query := `
        UPDATE shop_stock 
        SET quantity = quantity - ?, updated_at = NOW()
        WHERE shop_id = ? AND product_id = ? AND quantity >= ?
    `
	result, err := db.Exec(query, quantity, shopID, productID, quantity)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("insufficient stock for product %d in shop %d", productID, shopID)
	}
	return nil
}
