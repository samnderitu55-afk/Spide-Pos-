package db

import (
    "database/sql"
    "fmt"
)

// GetZReport generates a Z-report for a given date
func GetZReport(db *sql.DB, date string) (*ZReport, error) {
    report := &ZReport{
        ReportDate:       date,
        ExpenseBreakdown: make(map[string]float64),
    }

    // Get total revenue and sales count
    err := db.QueryRow(`
        SELECT 
            COALESCE(SUM(total_amount), 0),
            COUNT(*),
            COALESCE(SUM(CASE WHEN payment_type = 'cash' THEN total_amount ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN payment_type = 'mpesa' THEN total_amount ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN payment_type = 'split' THEN total_amount ELSE 0 END), 0),
            COUNT(CASE WHEN payment_type = 'cash' THEN 1 END),
            COUNT(CASE WHEN payment_type = 'mpesa' THEN 1 END),
            COUNT(CASE WHEN payment_type = 'split' THEN 1 END)
        FROM sales
        WHERE DATE(created_at) = ?
    `, date).Scan(
        &report.TotalRevenue,
        &report.TotalSalesCount,
        &report.TotalCash,
        &report.TotalMpesa,
        &report.SplitSalesCount,
        &report.CashSalesCount,
        &report.MpesaSalesCount,
        &report.SplitSalesCount,
    )
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get sales data: %w", err)
    }

    // Get expenses for the day
    rows, err := db.Query(`
        SELECT category, COALESCE(SUM(amount), 0) as total
        FROM expenses
        WHERE DATE(created_at) = ?
        GROUP BY category
    `, date)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get expenses: %w", err)
    }
    defer rows.Close()

    var totalExpenses float64
    var expenseCount int
    for rows.Next() {
        var category string
        var total float64
        err := rows.Scan(&category, &total)
        if err != nil {
            return nil, err
        }
        report.ExpenseBreakdown[category] = total
        totalExpenses += total
        expenseCount++
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating expenses: %w", err)
    }

    report.TotalExpenses = totalExpenses
    report.ExpenseCount = expenseCount
    report.NetProfit = report.TotalRevenue - report.TotalExpenses

    // Calculate margin
    if report.TotalRevenue > 0 {
        report.MarginPercent = (report.NetProfit / report.TotalRevenue) * 100
    }

    return report, nil
}
