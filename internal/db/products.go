package db

import (
    "database/sql"
    "fmt"
)

func GetProductByBarcode(db *sql.DB, barcode string, qty int) (*Product, error) {
    var p Product
    query := `
        SELECT p.id, p.barcode, p.name, p.category, p.cost_price, p.retail_price, 
               p.wholesale_price, p.wholesale_min_qty, p.reorder_level,
               COALESCE(ss.quantity, 0) as stock_quantity
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = 1
        WHERE p.barcode = ? AND p.is_active = 1
    `
    err := db.QueryRow(query, barcode).Scan(
        &p.ID, &p.Barcode, &p.Name, &p.Category,
        &p.CostPrice, &p.RetailPrice, &p.WholesalePrice,
        &p.WholesaleMinQty, &p.ReorderLevel, &p.StockQuantity,
    )
    if err != nil {
        return nil, fmt.Errorf("product not found: %w", err)
    }
    return &p, nil
}

func CreateProduct(db *sql.DB, p *Product) (int64, error) {
    query := `
        INSERT INTO products 
        (barcode, name, category, cost_price, retail_price, wholesale_price, 
         wholesale_min_qty, reorder_level, is_active, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, NOW(), NOW())
    `
    result, err := db.Exec(query,
        p.Barcode, p.Name, p.Category,
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

    // Also create shop_stock entries for all shops
    shopQuery := `
        INSERT INTO shop_stock (shop_id, product_id, quantity) 
        SELECT id, ?, 0 FROM branches WHERE is_active = 1
    `
    _, err = db.Exec(shopQuery, id)
    if err != nil {
        return 0, fmt.Errorf("failed to create shop stock: %w", err)
    }

    return id, nil
}

func UpdateProduct(db *sql.DB, p *Product) error {
    query := `
        UPDATE products 
        SET barcode = ?, name = ?, category = ?, cost_price = ?, retail_price = ?,
            wholesale_price = ?, wholesale_min_qty = ?, reorder_level = ?, 
            updated_at = NOW()
        WHERE id = ?
    `
    _, err := db.Exec(query,
        p.Barcode, p.Name, p.Category,
        p.CostPrice, p.RetailPrice, p.WholesalePrice,
        p.WholesaleMinQty, p.ReorderLevel,
        p.ID,
    )
    return err
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
