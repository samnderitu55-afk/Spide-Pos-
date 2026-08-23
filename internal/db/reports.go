package db

import (
    "database/sql"
    "fmt"
    "time"
)

func GetDailyZReportWithExpenses(db *sql.DB, dateParam string, shopID int) (*ZReport, error) {
    if dateParam == "" {
        dateParam = time.Now().Format("2006-01-02")
    }

    report := &ZReport{
        ReportDate:       dateParam,
        ExpenseBreakdown: make(map[string]float64),
    }

    salesQuery := `
        SELECT 
            COUNT(*) as total_sales,
            COALESCE(SUM(total_amount), 0) as total_revenue,
            COALESCE(SUM(cash_amount), 0) as total_cash,
            COALESCE(SUM(mpesa_amount), 0) as total_mpesa,
            SUM(CASE WHEN payment_type = 'cash' THEN 1 ELSE 0 END) as cash_count,
            SUM(CASE WHEN payment_type = 'mpesa' THEN 1 ELSE 0 END) as mpesa_count,
            SUM(CASE WHEN payment_type = 'split' THEN 1 ELSE 0 END) as split_count
        FROM sales 
        WHERE DATE(created_at) = ? AND shop_id = ?
    `
    err := db.QueryRow(salesQuery, dateParam, shopID).Scan(
        &report.TotalSalesCount,
        &report.TotalRevenue,
        &report.TotalCash,
        &report.TotalMpesa,
        &report.CashSalesCount,
        &report.MpesaSalesCount,
        &report.SplitSalesCount,
    )
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get sales summary: %w", err)
    }

    cogsQuery := `
        SELECT COALESCE(SUM(si.quantity * p.cost_price), 0)
        FROM sale_items si
        JOIN products p ON si.product_id = p.id
        JOIN sales s ON si.sale_id = s.id
        WHERE DATE(s.created_at) = ? AND s.shop_id = ?
    `
    err = db.QueryRow(cogsQuery, dateParam, shopID).Scan(&report.TotalCost)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get COGS: %w", err)
    }

    report.TotalProfit = report.TotalRevenue - report.TotalCost
    if report.TotalRevenue > 0 {
        report.MarginPercent = (report.TotalProfit / report.TotalRevenue) * 100
    }

    expenseQuery := `
        SELECT category, COALESCE(SUM(amount), 0) as total
        FROM expenses
        WHERE DATE(expense_date) = ? AND shop_id = ?
        GROUP BY category
    `
    rows, err := db.Query(expenseQuery, dateParam, shopID)
    if err != nil {
        return nil, fmt.Errorf("failed to get expenses: %w", err)
    }
    defer rows.Close()

    var totalExpenses float64
    for rows.Next() {
        var category string
        var amount float64
        if err := rows.Scan(&category, &amount); err != nil {
            continue
        }
        report.ExpenseBreakdown[category] = amount
        totalExpenses += amount
        report.ExpenseCount++
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating expenses: %w", err)
    }

    report.TotalExpenses = totalExpenses
    report.NetProfit = report.TotalProfit - totalExpenses

    return report, nil
}

func GetProductSalesReport(db *sql.DB, startDate, endDate string, shopID int) ([]ProductSalesReportItem, error) {
    if startDate == "" || endDate == "" {
        return nil, fmt.Errorf("start and end dates are required")
    }

    query := `
        SELECT 
            p.name as product_name,
            p.category,
            SUM(si.quantity) as units_sold,
            SUM(si.quantity * p.cost_price) as total_cost,
            SUM(si.subtotal) as total_revenue,
            SUM(si.subtotal) - SUM(si.quantity * p.cost_price) as net_profit
        FROM sale_items si
        JOIN products p ON si.product_id = p.id
        JOIN sales s ON si.sale_id = s.id
        WHERE DATE(s.created_at) BETWEEN ? AND ? AND s.shop_id = ?
        GROUP BY p.id, p.name, p.category
        ORDER BY units_sold DESC
    `
    rows, err := db.Query(query, startDate, endDate, shopID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var items []ProductSalesReportItem
    for rows.Next() {
        var item ProductSalesReportItem
        err := rows.Scan(
            &item.ProductName, &item.Category,
            &item.UnitsSold, &item.TotalCost,
            &item.TotalRevenue, &item.NetProfit,
        )
        if err != nil {
            return nil, err
        }
        if item.TotalRevenue > 0 {
            item.MarginPct = (item.NetProfit / item.TotalRevenue) * 100
        }
        items = append(items, item)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating product sales: %w", err)
    }

    return items, nil
}

// GetInventoryValuationReport - returns inventory valuation by category
func GetInventoryValuationReport(db *sql.DB) ([]InventoryValuationItem, error) {
    query := `
        SELECT 
            category,
            COUNT(*) as total_items,
            SUM(stock_quantity) as total_quantity,
            SUM(stock_quantity * cost_price) as total_cost,
            SUM(stock_quantity * retail_price) as total_retail,
            SUM(stock_quantity * (retail_price - cost_price)) as potential_profit
        FROM products
        WHERE is_active = 1 AND stock_quantity > 0
        GROUP BY category
        ORDER BY total_cost DESC
    `
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
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
        return nil, fmt.Errorf("error iterating inventory valuation: %w", err)
    }

    return items, nil
}

// GetCategoryValuationDetails - returns detailed product list for a category
func GetCategoryValuationDetails(db *sql.DB, category string) ([]struct {
    ProductName string  `json:"product_name"`
    InStock     int     `json:"in_stock"`
    CostPrice   float64 `json:"cost_price"`
    RetailPrice float64 `json:"retail_price"`
    TotalCost   float64 `json:"total_cost"`
}, error) {
    query := `
        SELECT name, stock_quantity, cost_price, retail_price, 
               stock_quantity * cost_price as total_cost
        FROM products
        WHERE is_active = 1 AND category = ? AND stock_quantity > 0
        ORDER BY name ASC
    `
    rows, err := db.Query(query, category)
    if err != nil {
        return nil, err
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
        return nil, fmt.Errorf("error iterating category details: %w", err)
    }

    return details, nil
}
