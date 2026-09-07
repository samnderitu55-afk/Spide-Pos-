package db

import (
	"database/sql"
	"time"
)

type ShopStock struct {
	ID        int       `json:"id"`
	CompanyID int       `json:"company_id"`
	ShopID    int       `json:"shop_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Get stock for a specific shop
func GetShopStock(db *sql.DB, shopID int) ([]ShopStock, error) {
	query := `
        SELECT ss.id, ss.company_id, ss.shop_id, ss.product_id, ss.quantity, ss.updated_at,
               p.name as product_name, p.barcode, p.retail_price, p.cost_price
        FROM shop_stock ss
        JOIN products p ON ss.product_id = p.id
        WHERE ss.shop_id = ?
        ORDER BY p.name
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

// Update stock quantity
func UpdateShopStock(db *sql.DB, shopID, productID, quantity int) error {
	query := `
        UPDATE shop_stock 
        SET quantity = ?, updated_at = NOW()
        WHERE shop_id = ? AND product_id = ?
    `
	_, err := db.Exec(query, quantity, shopID, productID)
	return err
}
