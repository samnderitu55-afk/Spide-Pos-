package db

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
)

// GetZReport generates a Z-report for a given date with optional shop filter
// internal/db/reports.go

func GetZReport(db *sql.DB, date string, shopID int) (*ZReport, error) {
    report := &ZReport{
        ReportDate:       date,
        ExpenseBreakdown: make(map[string]float64),
        ShopID:           shopID,
    }

    // Build shop filter
    shopFilter := ""
    if shopID > 0 {
        shopFilter = " AND shop_id = " + strconv.Itoa(shopID)
    }

    // ✅ Updated query - REMOVED SPLIT
    query := `
        SELECT 
            COALESCE(SUM(total_amount), 0) as total_revenue,
            COUNT(*) as total_sales,
            COALESCE(SUM(CASE WHEN payment_type = 'cash' THEN total_amount ELSE 0 END), 0) as total_cash,
            COALESCE(SUM(CASE WHEN payment_type = 'mpesa' THEN total_amount ELSE 0 END), 0) as total_mpesa,
            COALESCE(SUM(CASE WHEN payment_type = 'deposit' THEN total_amount ELSE 0 END), 0) as total_deposit,
            COALESCE(SUM(CASE WHEN payment_type = 'credit' THEN total_amount ELSE 0 END), 0) as total_credit,
            COUNT(CASE WHEN payment_type = 'cash' THEN 1 END) as cash_count,
            COUNT(CASE WHEN payment_type = 'mpesa' THEN 1 END) as mpesa_count,
            COUNT(CASE WHEN payment_type = 'deposit' THEN 1 END) as deposit_count,
            COUNT(CASE WHEN payment_type = 'credit' THEN 1 END) as credit_count
            -- ✅ REMOVED split_count
        FROM sales
        WHERE DATE(created_at) = ?` + shopFilter

    // ✅ Updated Scan - REMOVED SplitSalesCount
    err := db.QueryRow(query, date).Scan(
        &report.TotalRevenue,
        &report.TotalSalesCount,
        &report.TotalCash,
        &report.TotalMpesa,
        &report.TotalDeposit,
        &report.TotalCredit,
        &report.CashSalesCount,
        &report.MpesaSalesCount,
        &report.DepositSalesCount,
        &report.CreditSalesCount,
        // ✅ REMOVED &report.SplitSalesCount
    )
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get sales data: %w", err)
    }

    // ✅ Calculate Total Cost (COGS) from sale items
    costQuery := `
        SELECT COALESCE(SUM(si.quantity * p.cost_price), 0) as total_cost
        FROM sale_items si
        JOIN products p ON si.product_id = p.id
        JOIN sales s ON si.sale_id = s.id
        WHERE DATE(s.created_at) = ?` + shopFilter

    err = db.QueryRow(costQuery, date).Scan(&report.TotalCost)
    if err != nil && err != sql.ErrNoRows {
        report.TotalCost = 0
    }

    // Calculate Total Profit
    report.TotalProfit = report.TotalRevenue - report.TotalCost

    // Get expenses
    expenseQuery := `
        SELECT category, COALESCE(SUM(amount), 0) as total
        FROM expenses
        WHERE DATE(created_at) = ?` + shopFilter + `
        GROUP BY category
    `
    rows, err := db.Query(expenseQuery, date)
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
    report.NetProfit = report.TotalRevenue - report.TotalCost - report.TotalExpenses

    // Calculate margin
    if report.TotalRevenue > 0 {
        report.MarginPercent = (report.TotalProfit / report.TotalRevenue) * 100
    }

    // Get shop name if shopID is specified
    if shopID > 0 {
        var shopName string
        err = db.QueryRow("SELECT name FROM shops WHERE id = ?", shopID).Scan(&shopName)
        if err == nil {
            report.ShopName = shopName
        }
    }

    log.Printf("📊 Z-Report - Revenue: %.2f, Cost: %.2f, Profit: %.2f, Margin: %.2f%%",
        report.TotalRevenue, report.TotalCost, report.TotalProfit, report.MarginPercent)

    return report, nil
}