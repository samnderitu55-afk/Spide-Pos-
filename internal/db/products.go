package db

import (
	"database/sql"
	"fmt"
	"log"
)

func CreateProduct(db *sql.DB, p *Product) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// ✅ Handle NULL barcode - check if nil or empty string
	var barcode interface{}
	if p.Barcode == nil || *p.Barcode == "" {
		barcode = nil // This will insert NULL in the database
	} else {
		barcode = *p.Barcode
	}

	// ✅ CategoryID: NULL if 0 (so the FK doesn't reject it)
	var categoryID interface{}
	if p.CategoryID > 0 {
		categoryID = p.CategoryID
	} else {
		categoryID = nil
	}

	// ✅ Insert with barcode as NULL if empty
	query := `
        INSERT INTO products 
        (company_id, barcode, name, category, category_id, cost_price, retail_price, 
         wholesale_price, wholesale_min_qty, reorder_level, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
    `
	result, err := tx.Exec(query,
		p.CompanyID,
		barcode, // ✅ Can be NULL
		p.Name,
		p.Category,
		categoryID,
		p.CostPrice,
		p.RetailPrice,
		p.WholesalePrice,
		p.WholesaleMinQty,
		p.ReorderLevel,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create product: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get product ID: %w", err)
	}

	// Add stock if specified
	if p.StockQuantity > 0 {
		shopQuery := `
            INSERT INTO shop_stock (shop_id, product_id, quantity, company_id, created_at, updated_at)
            SELECT id, ?, ?, ?, NOW(), NOW() 
            FROM shops 
            WHERE company_id = ?
        `
		_, err = tx.Exec(shopQuery, id, p.StockQuantity, p.CompanyID, p.CompanyID)
		if err != nil {
			log.Printf("⚠️ Warning: Failed to create shop stock: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("✅ Product created: ID=%d, Name=%s, Barcode=%v", id, p.Name, barcode)
	return id, nil
}

func UpdateProduct(db *sql.DB, p *Product) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// ✅ Handle barcode - if empty string, use NULL
	var barcode interface{}
	if p.Barcode == nil || *p.Barcode == "" {
		barcode = nil
	} else {
		barcode = *p.Barcode
	}

	var categoryID interface{}
	if p.CategoryID > 0 {
		categoryID = p.CategoryID
	} else {
		categoryID = nil
	}

	// Update products table
	query := `
        UPDATE products 
        SET barcode = ?, name = ?, category = ?, category_id = ?, cost_price = ?, retail_price = ?,
            wholesale_price = ?, wholesale_min_qty = ?, reorder_level = ?, 
            updated_at = NOW()
        WHERE id = ? AND company_id = ?
    `
	_, err = tx.Exec(query,
		barcode, p.Name, p.Category, categoryID,
		p.CostPrice, p.RetailPrice, p.WholesalePrice,
		p.WholesaleMinQty, p.ReorderLevel,
		p.ID, p.CompanyID,
	)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	// NOTE: Stock is intentionally NOT updated here.
	// Stock lives in shop_stock and is managed via /api/products/update-stock
	// (and via sales / transfers / purchases). Product edit only touches metadata.

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

	products := []Product{}
	for rows.Next() {
		var p Product
		var barcodePtr *string

		err := rows.Scan(
			&p.ID,
			&barcodePtr,
			&p.Name,
			&p.Category,
			&p.CostPrice,
			&p.RetailPrice,
			&p.WholesalePrice,
			&p.WholesaleMinQty,
			&p.ReorderLevel,
			&p.StockQuantity,
		)
		if err != nil {
			return nil, err
		}

		p.Barcode = barcodePtr
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

	products := []Product{}
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
        SELECT 
            p.id, 
            p.company_id, 
            p.barcode, 
            p.name, 
            p.category, 
            p.category_id,
            p.cost_price, 
            p.retail_price, 
            p.wholesale_price, 
            p.wholesale_min_qty,
            COALESCE(ss.quantity, 0) as stock_quantity,
            p.reorder_level, 
            p.is_active,
            p.created_at, 
            p.updated_at,
			p.unit_type,
            p.unit_label
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = ? AND ss.company_id = ?
        WHERE p.barcode = ? AND p.company_id = ?
        LIMIT 1
    `

	var p Product
	var barcodePtr *string

	err := db.QueryRow(query, shopID, companyID, barcode, companyID).Scan(
		&p.ID,
		&p.CompanyID,
		&barcodePtr,
		&p.Name,
		&p.Category,
		&p.CategoryID,
		&p.CostPrice,
		&p.RetailPrice,
		&p.WholesalePrice,
		&p.WholesaleMinQty,
		&p.StockQuantity,
		&p.ReorderLevel,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.UnitType,
		&p.UnitLabel,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	p.Barcode = barcodePtr

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

	products := []Product{}
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

	products := []Product{}
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

// SearchProductsByCompany - Search products by company only (master catalog)
// Optionally includes stock quantity for a specific shop
func SearchProductsByCompany(db *sql.DB, query string, companyID int, shopID int) ([]Product, error) {
	if query == "" {
		return []Product{}, nil
	}

	searchTerm := "%" + query + "%"

	sqlQuery := `
        SELECT 
            p.id, 
            p.company_id,
            p.barcode, 
            p.name, 
            p.category, 
            p.category_id,
            p.cost_price, 
            p.retail_price, 
            p.wholesale_price, 
            p.wholesale_min_qty,
            COALESCE(ss.quantity, 0) as stock_quantity,
            p.reorder_level, 
            p.is_active,
            p.created_at,
            p.updated_at
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = ? AND ss.company_id = ?
        WHERE p.is_active = 1 
            AND p.company_id = ?
            AND (p.name LIKE ? OR p.barcode LIKE ?)
        ORDER BY 
            CASE 
                WHEN p.barcode = ? THEN 1
                WHEN p.name LIKE ? THEN 2
                ELSE 3
            END,
            p.name
        LIMIT 20
    `

	rows, err := db.Query(sqlQuery,
		shopID,
		companyID,
		companyID,
		searchTerm, searchTerm,
		query,
		searchTerm,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		var barcodePtr *string

		err := rows.Scan(
			&p.ID,
			&p.CompanyID,
			&barcodePtr,
			&p.Name,
			&p.Category,
			&p.CategoryID,
			&p.CostPrice,
			&p.RetailPrice,
			&p.WholesalePrice,
			&p.WholesaleMinQty,
			&p.StockQuantity,
			&p.ReorderLevel,
			&p.IsActive,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// ✅ Set barcode from pointer
		p.Barcode = barcodePtr

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	if len(products) == 0 {
		log.Printf("🔍 No products found for company %d with query: %s", companyID, query)
	}

	return products, nil
}

// GetProductsByCompany - Get all products for a company (master catalog)
func GetProductsByCompany(db *sql.DB, companyID int, shopID int) ([]Product, error) {
	query := `
        SELECT 
            p.id, 
            p.company_id, 
            p.barcode, 
            p.name, 
            p.category, 
            p.category_id,
            p.cost_price, 
            p.retail_price, 
            p.wholesale_price, 
            p.wholesale_min_qty,
            p.reorder_level, 
            p.is_active,
            p.created_at, 
            p.updated_at,
            COALESCE(ss.quantity, 0) as stock_quantity
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = ? AND ss.company_id = ?
        WHERE p.company_id = ? AND p.is_active = 1
        ORDER BY p.name ASC
    `

	rows, err := db.Query(query, shopID, companyID, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get products with stock: %w", err)
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		err := rows.Scan(
			&p.ID,
			&p.CompanyID,
			&p.Barcode,
			&p.Name,
			&p.Category,
			&p.CategoryID,
			&p.CostPrice,
			&p.RetailPrice,
			&p.WholesalePrice,
			&p.WholesaleMinQty,
			&p.ReorderLevel,
			&p.IsActive,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.StockQuantity,
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

// GetProductByNameExact - Get product by exact name match
func GetProductByNameExact(db *sql.DB, name string, companyID int) (*Product, error) {
	query := `
        SELECT id, company_id, barcode, name, category, category_id,
               cost_price, retail_price, wholesale_price, wholesale_min_qty,
               reorder_level, created_at, updated_at
        FROM products
        WHERE name = ? AND company_id = ?
        LIMIT 1
    `
	var p Product
	var barcodePtr *string

	err := db.QueryRow(query, name, companyID).Scan(
		&p.ID,
		&p.CompanyID,
		&barcodePtr,
		&p.Name,
		&p.Category,
		&p.CategoryID,
		&p.CostPrice,
		&p.RetailPrice,
		&p.WholesalePrice,
		&p.WholesaleMinQty,
		&p.ReorderLevel,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	p.Barcode = barcodePtr
	return &p, nil
}

// GetProductsByNameFuzzy - Get products by fuzzy name match (contains)
func GetProductsByNameFuzzy(db *sql.DB, name string, companyID int) ([]Product, error) {
	searchTerm := "%" + name + "%"
	query := `
        SELECT id, company_id, barcode, name, category, category_id,
               cost_price, retail_price, wholesale_price, wholesale_min_qty,
               reorder_level, created_at, updated_at
        FROM products
        WHERE name LIKE ? AND company_id = ?
        LIMIT 10
    `
	rows, err := db.Query(query, searchTerm, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		var barcodePtr *string
		err := rows.Scan(
			&p.ID,
			&p.CompanyID,
			&barcodePtr,
			&p.Name,
			&p.Category,
			&p.CategoryID,
			&p.CostPrice,
			&p.RetailPrice,
			&p.WholesalePrice,
			&p.WholesaleMinQty,
			&p.ReorderLevel,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		p.Barcode = barcodePtr
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}

// UpdateProductBarcode - Update a product's barcode
func UpdateProductBarcode(db *sql.DB, productID int64, barcode string) error {
	// Convert empty string to NULL
	var barcodePtr interface{}
	if barcode == "" {
		barcodePtr = nil
	} else {
		barcodePtr = barcode
	}

	query := `
        UPDATE products 
        SET barcode = ?, updated_at = NOW()
        WHERE id = ?
    `
	_, err := db.Exec(query, barcodePtr, productID)
	return err
}

// GetProductsByCompanyWithStock - Get all products for a company with stock for a specific shop
func GetProductsByCompanyWithStock(db *sql.DB, companyID int, shopID int) ([]Product, error) {
	if shopID == 0 {
		return getAllProductsWithTotalStock(db, companyID)
	}

	query := `
        SELECT 
            p.id, 
            p.company_id, 
            p.barcode, 
            p.name, 
            p.category, 
            p.category_id,
            p.cost_price, 
            p.retail_price, 
            p.wholesale_price, 
            p.wholesale_min_qty,
            COALESCE(ss.quantity, 0) as stock_quantity,
            COALESCE(ss.stock_cap, 0) as stock_cap,
            COALESCE(ss.reorder_level, p.reorder_level) as shop_reorder_level,
            p.reorder_level, 
            p.is_active,
            p.created_at, 
            p.updated_at
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = ? AND ss.company_id = ?
        WHERE p.company_id = ? AND p.is_active = 1
        ORDER BY p.name ASC
    `

	rows, err := db.Query(query, shopID, companyID, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get products with stock: %w", err)
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		var barcodePtr *string

		err := rows.Scan(
			&p.ID,
			&p.CompanyID,
			&barcodePtr,
			&p.Name,
			&p.Category,
			&p.CategoryID,
			&p.CostPrice,
			&p.RetailPrice,
			&p.WholesalePrice,
			&p.WholesaleMinQty,
			&p.StockQuantity,
			&p.StockCap,         // ← new
			&p.ShopReorderLevel, // ← new
			&p.ReorderLevel,
			&p.IsActive,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		p.Barcode = barcodePtr
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	log.Printf("📦 Retrieved %d products with stock for company %d, shop %d", len(products), companyID, shopID)
	return products, nil
}

func getAllProductsWithTotalStock(db *sql.DB, companyID int) ([]Product, error) {
	query := `
        SELECT 
            p.id, p.company_id, p.barcode, p.name, p.category, p.category_id,
            p.cost_price, p.retail_price, p.wholesale_price, p.wholesale_min_qty,
            COALESCE(SUM(ss.quantity), 0) AS stock_quantity,
            COALESCE(SUM(ss.stock_cap), 0) AS stock_cap,
            COALESCE(MAX(ss.reorder_level), p.reorder_level) AS shop_reorder_level,
            p.reorder_level, p.is_active, p.created_at, p.updated_at
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.company_id = p.company_id
        WHERE p.company_id = ? AND p.is_active = 1
        GROUP BY p.id, p.company_id, p.barcode, p.name, p.category, p.category_id,
                 p.cost_price, p.retail_price, p.wholesale_price, p.wholesale_min_qty,
                 p.reorder_level, p.is_active, p.created_at, p.updated_at
        ORDER BY p.name ASC
    `
	rows, err := db.Query(query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get products with total stock: %w", err)
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		var barcodePtr *string
		err := rows.Scan(
			&p.ID, &p.CompanyID, &barcodePtr, &p.Name, &p.Category, &p.CategoryID,
			&p.CostPrice, &p.RetailPrice, &p.WholesalePrice, &p.WholesaleMinQty,
			&p.StockQuantity, &p.StockCap, &p.ShopReorderLevel,
			&p.ReorderLevel, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		p.Barcode = barcodePtr
		products = append(products, p)
	}
	return products, rows.Err()
}
