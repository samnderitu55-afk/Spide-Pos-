package db

import (
	"database/sql"
	"fmt"
	"log"
)

func CreateProduct(db *sql.DB, p *Product) (int64, error) {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert into products table - Updated to include company_id and category_id
	query := `
        INSERT INTO products 
        (company_id, barcode, name, category, category_id, cost_price, retail_price, 
         wholesale_price, wholesale_min_qty, reorder_level, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
    `
	result, err := tx.Exec(query,
		p.CompanyID, p.Barcode, p.Name, p.Category, p.CategoryID,
		p.CostPrice, p.RetailPrice, p.WholesalePrice,
		p.WholesaleMinQty, p.ReorderLevel,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create product: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get product ID: %w", err)
	}

	// ✅ Insert into shop_stock for all active shops (not branches)
	// Only if StockQuantity > 0
	if p.StockQuantity > 0 {
		shopQuery := `
            INSERT INTO shop_stock (shop_id, product_id, quantity, company_id, created_at, updated_at)
            SELECT id, ?, ?, ?, NOW(), NOW() 
            FROM shops 
            WHERE company_id = ? AND is_active = 1
        `
		_, err = tx.Exec(shopQuery, id, p.StockQuantity, p.CompanyID, p.CompanyID)
		if err != nil {
			// Log error but don't fail the transaction - stock can be added later
			log.Printf("⚠️ Warning: Failed to create shop stock: %v", err)
			// Continue without stock
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("✅ Product created: %s (ID: %d) with stock: %d", p.Name, id, p.StockQuantity)

	return id, nil
}

func UpdateProduct(db *sql.DB, p *Product) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update products table
	query := `
        UPDATE products 
        SET barcode = ?, name = ?, category = ?, cost_price = ?, retail_price = ?,
            wholesale_price = ?, wholesale_min_qty = ?, reorder_level = ?, 
            updated_at = NOW()
        WHERE id = ?
    `
	_, err = tx.Exec(query,
		p.Barcode, p.Name, p.Category,
		p.CostPrice, p.RetailPrice, p.WholesalePrice,
		p.WholesaleMinQty, p.ReorderLevel,
		p.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	// Update stock in shop_stock for the current shop
	stockQuery := `
        UPDATE shop_stock 
        SET quantity = ?, updated_at = NOW()
        WHERE shop_id = 1 AND product_id = ?
    `
	_, err = tx.Exec(stockQuery, p.StockQuantity, p.ID)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func GetAllProducts(db *sql.DB) ([]Product, error) {
	query := `
        SELECT p.id, p.barcode, p.name, p.category, p.cost_price, p.retail_price, 
               p.wholesale_price, p.wholesale_min_qty, p.reorder_level,
               COALESCE(ss.quantity, 0) as stock_quantity
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = 1
        WHERE p.is_active = 1
        ORDER BY p.name ASC
    `
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(
			&p.ID, &p.Barcode, &p.Name, &p.Category,
			&p.CostPrice, &p.RetailPrice, &p.WholesalePrice,
			&p.WholesaleMinQty, &p.ReorderLevel, &p.StockQuantity,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

func SearchProducts(db *sql.DB, query string) ([]Product, error) {
	searchTerm := "%" + query + "%"
	sqlQuery := `
        SELECT p.id, p.barcode, p.name, p.category, p.cost_price, p.retail_price, 
               p.wholesale_price, p.wholesale_min_qty, p.reorder_level,
               COALESCE(ss.quantity, 0) as stock_quantity
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = 1
        WHERE p.is_active = 1 AND (p.name LIKE ? OR p.barcode LIKE ?)
        LIMIT 10
    `
	rows, err := db.Query(sqlQuery, searchTerm, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(
			&p.ID, &p.Barcode, &p.Name, &p.Category,
			&p.CostPrice, &p.RetailPrice, &p.WholesalePrice,
			&p.WholesaleMinQty, &p.ReorderLevel, &p.StockQuantity,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating search results: %w", err)
	}

	return products, nil
}

func ArchiveProduct(db *sql.DB, id int64) error {
	_, err := db.Exec("UPDATE products SET is_active = 0, updated_at = NOW() WHERE id = ?", id)
	return err
}

func RestoreProduct(db *sql.DB, id int64) error {
	_, err := db.Exec("UPDATE products SET is_active = 1, updated_at = NOW() WHERE id = ?", id)
	return err
}

func GetProductByBarcode(db *sql.DB, barcode string, companyID int) (*Product, error) {
	if barcode == "" {
		return nil, nil
	}

	query := `
        SELECT id, company_id, barcode, name, category, category_id,
               cost_price, retail_price, wholesale_price, wholesale_min_qty,
               reorder_level, created_at, updated_at
        FROM products
        WHERE barcode = ? AND company_id = ?
    `
	var p Product
	err := db.QueryRow(query, barcode, companyID).Scan(
		&p.ID, &p.CompanyID, &p.Barcode, &p.Name, &p.Category, &p.CategoryID,
		&p.CostPrice, &p.RetailPrice, &p.WholesalePrice,
		&p.WholesaleMinQty, &p.ReorderLevel,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetProductByBarcodeAndShop - Check by barcode, company, and shop (4 parameters)
func GetProductByBarcodeAndShop(db *sql.DB, barcode string, companyID int, shopID int) (*Product, error) {
	if barcode == "" {
		return nil, nil
	}

	query := `
        SELECT p.id, p.company_id, p.barcode, p.name, p.category, p.category_id,
               p.cost_price, p.retail_price, p.wholesale_price, p.wholesale_min_qty,
               p.reorder_level, p.created_at, p.updated_at,
               COALESCE(ss.quantity, 0) as stock_quantity
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = ? AND ss.company_id = ?
        WHERE p.barcode = ? AND p.company_id = ?
        LIMIT 1
    `
	var p Product
	err := db.QueryRow(query, shopID, companyID, barcode, companyID).Scan(
		&p.ID, &p.CompanyID, &p.Barcode, &p.Name, &p.Category, &p.CategoryID,
		&p.CostPrice, &p.RetailPrice, &p.WholesalePrice,
		&p.WholesaleMinQty, &p.ReorderLevel,
		&p.CreatedAt, &p.UpdatedAt,
		&p.StockQuantity,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func SearchProductsByShop(db *sql.DB, query string, shopID int) ([]Product, error) {
	searchTerm := "%" + query + "%"
	sqlQuery := `
        SELECT p.id, p.barcode, p.name, p.category, p.cost_price, p.retail_price, 
               p.wholesale_price, p.wholesale_min_qty, 
               COALESCE(ss.quantity, 0) as stock_quantity,
               p.reorder_level, p.is_active
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = ?
        WHERE p.is_active = 1 AND (p.name LIKE ? OR p.barcode LIKE ?)
        ORDER BY p.name
        LIMIT 20
    `
	rows, err := db.Query(sqlQuery, shopID, searchTerm, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(
			&p.ID,
			&p.Barcode,
			&p.Name,
			&p.Category,
			&p.CostPrice,
			&p.RetailPrice,
			&p.WholesalePrice,
			&p.WholesaleMinQty,
			&p.StockQuantity,
			&p.ReorderLevel,
			&p.IsActive,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

func GetProductsByShop(db *sql.DB, shopID int) ([]Product, error) {
	query := `
        SELECT 
            p.id, p.barcode, p.name, p.category, p.cost_price, 
            p.retail_price, p.wholesale_price, p.wholesale_min_qty,
            p.reorder_level, p.created_at, p.updated_at,
            COALESCE(ss.quantity, 0) as stock_quantity
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = ?
        ORDER BY p.name
    `
	rows, err := db.Query(query, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(
			&p.ID, &p.Barcode, &p.Name, &p.Category,
			&p.CostPrice, &p.RetailPrice, &p.WholesalePrice,
			&p.WholesaleMinQty, &p.ReorderLevel,
			&p.CreatedAt, &p.UpdatedAt, &p.StockQuantity,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}
	return products, nil
}

func GetProductsByShopAndCompany(db *sql.DB, shopID, companyID int) ([]Product, error) {
	query := `
        SELECT 
            p.id, p.barcode, p.name, p.category, p.cost_price, 
            p.retail_price, p.wholesale_price, p.wholesale_min_qty,
            p.reorder_level, p.created_at, p.updated_at,
            COALESCE(ss.quantity, 0) as stock_quantity
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = ? AND ss.company_id = ?
        WHERE p.company_id = ?
        ORDER BY p.name
    `
	rows, err := db.Query(query, shopID, companyID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(
			&p.ID, &p.Barcode, &p.Name, &p.Category,
			&p.CostPrice, &p.RetailPrice, &p.WholesalePrice,
			&p.WholesaleMinQty, &p.ReorderLevel,
			&p.CreatedAt, &p.UpdatedAt, &p.StockQuantity,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}
	return products, nil
}

func AddStockToShop(db *sql.DB, shopID, productID, quantity, companyID int) error {
	query := `
        INSERT INTO shop_stock (shop_id, product_id, quantity, company_id, created_at, updated_at)
        VALUES (?, ?, ?, ?, NOW(), NOW())
        ON DUPLICATE KEY UPDATE quantity = quantity + ?, updated_at = NOW()
    `
	_, err := db.Exec(query, shopID, productID, quantity, companyID, quantity)
	return err
}
