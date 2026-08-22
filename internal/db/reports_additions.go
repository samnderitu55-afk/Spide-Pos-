package db

import (
    "database/sql"
)

// GetInventoryValuationReportForShop - with fallback for missing shop_stock
func GetInventoryValuationReportForShop(db *sql.DB, filterShopID int) ([]InventoryValuationItem, error) {
    // Check if shop_stock table exists
    var tableExists bool
    err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'shop_stock'").Scan(&tableExists)
    if err != nil {
        tableExists = false
    }

    var query string
    var args []interface{}

    if filterShopID > 0 && tableExists {
        query = `
            SELECT 
                p.category,
                COUNT(*) as total_items,
                SUM(ss.quantity) as total_quantity,
                SUM(ss.quantity * p.cost_price) as total_cost,
                SUM(ss.quantity * p.retail_price) as total_retail,
                SUM(ss.quantity * (p.retail_price - p.cost_price)) as potential_profit
            FROM products p
            JOIN shop_stock ss ON p.id = ss.product_id
            WHERE ss.shop_id = ? AND ss.quantity > 0
            GROUP BY p.category
            ORDER BY total_cost DESC
        `
        args = []interface{}{filterShopID}
    } else {
        // Fallback: use products table directly
        query = `
            SELECT 
                category,
                COUNT(*) as total_items,
                SUM(stock_quantity) as total_quantity,
                SUM(stock_quantity * cost_price) as total_cost,
                SUM(stock_quantity * retail_price) as total_retail,
                SUM(stock_quantity * (retail_price - cost_price)) as potential_profit
            FROM products
            WHERE stock_quantity > 0
            GROUP BY category
            ORDER BY total_cost DESC
        `
    }

    rows, err := db.Query(query, args...)
    if err != nil {
        return []InventoryValuationItem{}, nil
    }
    defer rows.Close()

    var items []InventoryValuationItem
    for rows.Next() {
        var item InventoryValuationItem
        err := rows.Scan(
            &item.Category, &item.TotalItems, &item.TotalQuantity,
            &item.TotalCost, &item.TotalRetail, &item.PotentialProfit,
        )
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

// GetLowStockReportForShop - with fallback for missing shop_stock
func GetLowStockReportForShop(db *sql.DB, threshold int, filterShopID int) ([]LowStockReportItem, error) {
    // Check if shop_stock table exists
    var tableExists bool
    err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'shop_stock'").Scan(&tableExists)
    if err != nil {
        tableExists = false
    }

    var query string
    var args []interface{}

    if filterShopID > 0 && tableExists {
        query = `
            SELECT 
                p.name as product_name,
                p.category,
                ss.quantity as stock_quantity,
                p.reorder_level,
                p.cost_price,
                (p.reorder_level - ss.quantity) * p.cost_price as restock_cost
            FROM products p
            JOIN shop_stock ss ON p.id = ss.product_id
            WHERE ss.shop_id = ? AND ss.quantity <= ?
            ORDER BY ss.quantity ASC
        `
        args = []interface{}{filterShopID, threshold}
    } else if filterShopID > 0 && !tableExists {
        return []LowStockReportItem{}, nil
    } else {
        query = `
            SELECT 
                name as product_name,
                category,
                stock_quantity,
                reorder_level,
                cost_price,
                (reorder_level - stock_quantity) * cost_price as restock_cost
            FROM products
            WHERE stock_quantity <= ?
            ORDER BY stock_quantity ASC
        `
        args = []interface{}{threshold}
    }

    rows, err := db.Query(query, args...)
    if err != nil {
        return []LowStockReportItem{}, nil
    }
    defer rows.Close()

    var items []LowStockReportItem
    for rows.Next() {
        var item LowStockReportItem
        err := rows.Scan(
            &item.ProductName, &item.Category,
            &item.StockQuantity, &item.ReorderLevel,
            &item.CostPrice, &item.RestockCost,
        )
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

// GetCategoryValuationDetailsForShop - with fallback for missing shop_stock
func GetCategoryValuationDetailsForShop(db *sql.DB, category string, filterShopID int) ([]struct {
    ProductName string  `json:"product_name"`
    InStock     int     `json:"in_stock"`
    CostPrice   float64 `json:"cost_price"`
    RetailPrice float64 `json:"retail_price"`
    TotalCost   float64 `json:"total_cost"`
}, error) {
    // Check if shop_stock table exists
    var tableExists bool
    err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'shop_stock'").Scan(&tableExists)
    if err != nil {
        tableExists = false
    }

    var query string
    var args []interface{}

    if filterShopID > 0 && tableExists {
        query = `
            SELECT p.name, ss.quantity as in_stock, p.cost_price, p.retail_price, 
                   ss.quantity * p.cost_price as total_cost
            FROM products p
            JOIN shop_stock ss ON p.id = ss.product_id
            WHERE p.category = ? AND ss.shop_id = ? AND ss.quantity > 0
            ORDER BY p.name ASC
        `
        args = []interface{}{category, filterShopID}
    } else if filterShopID > 0 && !tableExists {
        return []struct {
            ProductName string  `json:"product_name"`
            InStock     int     `json:"in_stock"`
            CostPrice   float64 `json:"cost_price"`
            RetailPrice float64 `json:"retail_price"`
            TotalCost   float64 `json:"total_cost"`
        }{}, nil
    } else {
        query = `
            SELECT name, stock_quantity as in_stock, cost_price, retail_price, 
                   stock_quantity * cost_price as total_cost
            FROM products
            WHERE category = ? AND stock_quantity > 0
            ORDER BY name ASC
        `
        args = []interface{}{category}
    }

    rows, err := db.Query(query, args...)
    if err != nil {
        return []struct {
            ProductName string  `json:"product_name"`
            InStock     int     `json:"in_stock"`
            CostPrice   float64 `json:"cost_price"`
            RetailPrice float64 `json:"retail_price"`
            TotalCost   float64 `json:"total_cost"`
        }{}, nil
    }
    defer rows.Close()

    var details []struct {
        ProductName string  `json:"product_name"`
        InStock     int     `json:"in_stock"`
        CostPrice   float64 `json:"cost_price"`
        RetailPrice float64 `json:"retail_price"`
        TotalCost   float64 `json:"total_cost"`
    }
    for rows.Next() {
        var d struct {
            ProductName string  `json:"product_name"`
            InStock     int     `json:"in_stock"`
            CostPrice   float64 `json:"cost_price"`
            RetailPrice float64 `json:"retail_price"`
            TotalCost   float64 `json:"total_cost"`
        }
        err := rows.Scan(&d.ProductName, &d.InStock, &d.CostPrice, &d.RetailPrice, &d.TotalCost)
        if err != nil {
            return nil, err
        }
        details = append(details, d)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return details, nil
}


