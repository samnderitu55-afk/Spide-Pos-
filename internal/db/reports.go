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

// ... rest of functions remain the same ...


