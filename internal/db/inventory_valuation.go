package db

import (
    "database/sql"
    "fmt"
)

type InventoryValuation struct {
    BranchID         int                  `json:"branch_id"`
    BranchName       string               `json:"branch_name"`
    TotalItems       int                  `json:"total_items"`
    TotalQuantity    int                  `json:"total_quantity"`
    TotalCostValue   float64              `json:"total_cost_value"`
    TotalRetailValue float64              `json:"total_retail_value"`
    PotentialProfit  float64              `json:"potential_profit"`
    Categories       []CategoryValuation  `json:"categories"`
}

type CategoryValuation struct {
    Category         string  `json:"category"`
    ItemCount        int     `json:"item_count"`
    TotalQuantity    int     `json:"total_quantity"`
    TotalCostValue   float64 `json:"total_cost_value"`
    TotalRetailValue float64 `json:"total_retail_value"`
    PotentialProfit  float64 `json:"potential_profit"`
}

type ProductValuation struct {
    ProductID    int     `json:"product_id"`
    ProductName  string  `json:"product_name"`
    Barcode      string  `json:"barcode"`
    Category     string  `json:"category"`
    Quantity     int     `json:"quantity"`
    CostPrice    float64 `json:"cost_price"`
    RetailPrice  float64 `json:"retail_price"`
    CostValue    float64 `json:"cost_value"`
    RetailValue  float64 `json:"retail_value"`
    Profit       float64 `json:"profit"`
}

func GetInventoryValuation(db *sql.DB, branchID int) (*InventoryValuation, error) {
    valuation := &InventoryValuation{}
    
    // Get branch info
    if branchID > 0 {
        err := db.QueryRow(`
            SELECT id, name FROM branches WHERE id = ? AND is_active = 1
        `, branchID).Scan(&valuation.BranchID, &valuation.BranchName)
        if err != nil {
            return nil, fmt.Errorf("branch not found: %w", err)
        }
    } else {
        valuation.BranchID = 0
        valuation.BranchName = "All Branches"
    }

    // Build query based on branch filter
    var branchCondition string
    var args []interface{}
    if branchID > 0 {
        branchCondition = "AND ss.shop_id = ?"
        args = append(args, branchID)
    }

    // Get totals
    query := fmt.Sprintf(`
        SELECT 
            COUNT(DISTINCT p.id),
            COALESCE(SUM(ss.quantity), 0),
            COALESCE(SUM(ss.quantity * p.cost_price), 0),
            COALESCE(SUM(ss.quantity * p.retail_price), 0)
        FROM products p
        JOIN shop_stock ss ON p.id = ss.product_id
        WHERE p.is_active = 1 %s
    `, branchCondition)
    
    err := db.QueryRow(query, args...).Scan(
        &valuation.TotalItems,
        &valuation.TotalQuantity,
        &valuation.TotalCostValue,
        &valuation.TotalRetailValue,
    )
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get valuation totals: %w", err)
    }
    
    valuation.PotentialProfit = valuation.TotalRetailValue - valuation.TotalCostValue

    // Get category breakdown
    categoryQuery := fmt.Sprintf(`
        SELECT 
            COALESCE(p.category, 'Uncategorized') as category,
            COUNT(DISTINCT p.id),
            COALESCE(SUM(ss.quantity), 0),
            COALESCE(SUM(ss.quantity * p.cost_price), 0),
            COALESCE(SUM(ss.quantity * p.retail_price), 0)
        FROM products p
        JOIN shop_stock ss ON p.id = ss.product_id
        WHERE p.is_active = 1 %s
        GROUP BY p.category
        ORDER BY SUM(ss.quantity * p.cost_price) DESC
    `, branchCondition)
    
    rows, err := db.Query(categoryQuery, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to get category breakdown: %w", err)
    }
    defer rows.Close()
    
    var categories []CategoryValuation
    for rows.Next() {
        var cat CategoryValuation
        err := rows.Scan(
            &cat.Category,
            &cat.ItemCount,
            &cat.TotalQuantity,
            &cat.TotalCostValue,
            &cat.TotalRetailValue,
        )
        if err != nil {
            return nil, err
        }
        cat.PotentialProfit = cat.TotalRetailValue - cat.TotalCostValue
        categories = append(categories, cat)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating categories: %w", err)
    }
    valuation.Categories = categories
    
    return valuation, nil
}

func GetProductValuation(db *sql.DB, branchID int) ([]ProductValuation, error) {
    var branchCondition string
    var args []interface{}
    if branchID > 0 {
        branchCondition = "AND ss.shop_id = ?"
        args = append(args, branchID)
    }
    
    query := fmt.Sprintf(`
        SELECT 
            p.id,
            p.name,
            p.barcode,
            COALESCE(p.category, 'Uncategorized') as category,
            COALESCE(ss.quantity, 0) as quantity,
            p.cost_price,
            p.retail_price,
            COALESCE(ss.quantity * p.cost_price, 0) as cost_value,
            COALESCE(ss.quantity * p.retail_price, 0) as retail_value
        FROM products p
        JOIN shop_stock ss ON p.id = ss.product_id
        WHERE p.is_active = 1 AND COALESCE(ss.quantity, 0) > 0 %s
        ORDER BY cost_value DESC
    `, branchCondition)
    
    rows, err := db.Query(query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to get product valuation: %w", err)
    }
    defer rows.Close()
    
    var products []ProductValuation
    for rows.Next() {
        var p ProductValuation
        err := rows.Scan(
            &p.ProductID,
            &p.ProductName,
            &p.Barcode,
            &p.Category,
            &p.Quantity,
            &p.CostPrice,
            &p.RetailPrice,
            &p.CostValue,
            &p.RetailValue,
        )
        if err != nil {
            return nil, err
        }
        p.Profit = p.RetailValue - p.CostValue
        products = append(products, p)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating products: %w", err)
    }
    
    return products, nil
}
