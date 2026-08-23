package db

import (
    "database/sql"
    "fmt"
    "time"
)

func GetDashboardStats(db *sql.DB) (*DashboardStats, error) {
    stats := &DashboardStats{}
    today := time.Now().Format("2006-01-02")
    yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

    if err := getTodaySales(db, today, stats); err != nil {
        return nil, fmt.Errorf("failed to get today's sales: %w", err)
    }
    if err := getYesterdaySales(db, yesterday, stats); err != nil {
        return nil, fmt.Errorf("failed to get yesterday's sales: %w", err)
    }
    if err := getQuickStats(db, stats); err != nil {
        return nil, fmt.Errorf("failed to get quick stats: %w", err)
    }
    if err := getSalesTrend(db, today, stats); err != nil {
        return nil, fmt.Errorf("failed to get sales trend: %w", err)
    }
    if err := getTopProducts(db, today, stats); err != nil {
        return nil, fmt.Errorf("failed to get top products: %w", err)
    }
    if err := getRecentSalesForDashboard(db, stats); err != nil {
        return nil, fmt.Errorf("failed to get recent sales: %w", err)
    }
    if err := getLowStockForDashboard(db, stats); err != nil {
        return nil, fmt.Errorf("failed to get low stock items: %w", err)
    }

    return stats, nil
}

func getTodaySales(db *sql.DB, today string, stats *DashboardStats) error {
    query := `
        SELECT 
            COALESCE(SUM(total_amount), 0),
            COUNT(*),
            COALESCE(AVG(total_amount), 0),
            COALESCE(SUM(si.quantity), 0)
        FROM sales s
        LEFT JOIN sale_items si ON s.id = si.sale_id
        WHERE DATE(s.created_at) = ?
    `
    err := db.QueryRow(query, today).Scan(
        &stats.TodaySales.TotalRevenue,
        &stats.TodaySales.TransactionCount,
        &stats.TodaySales.AverageTicket,
        &stats.TodaySales.TotalItems,
    )
    if err != nil && err != sql.ErrNoRows {
        return err
    }
    return nil
}

func getYesterdaySales(db *sql.DB, yesterday string, stats *DashboardStats) error {
    query := `
        SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
        FROM sales
        WHERE DATE(created_at) = ?
    `
    err := db.QueryRow(query, yesterday).Scan(
        &stats.YesterdaySales.TotalRevenue,
        &stats.YesterdaySales.TransactionCount,
    )
    if err != nil && err != sql.ErrNoRows {
        return err
    }
    return nil
}

func getQuickStats(db *sql.DB, stats *DashboardStats) error {
    query := `
        SELECT 
            COUNT(*),
            COALESCE(SUM(stock_quantity * cost_price), 0)
        FROM products
        WHERE is_active = 1
    `
    err := db.QueryRow(query).Scan(
        &stats.QuickStats.TotalProducts,
        &stats.QuickStats.TotalStockValue,
    )
    if err != nil && err != sql.ErrNoRows {
        return err
    }
    return nil
}

func getSalesTrend(db *sql.DB, today string, stats *DashboardStats) error {
    query := `
        SELECT HOUR(created_at) as hour, 
               COALESCE(SUM(total_amount), 0),
               COUNT(*)
        FROM sales
        WHERE DATE(created_at) = ?
        GROUP BY HOUR(created_at)
        ORDER BY hour ASC
    `
    rows, err := db.Query(query, today)
    if err != nil {
        return err
    }
    defer rows.Close()

    var trend []struct {
        Hour   int     `json:"hour"`
        Amount float64 `json:"amount"`
        Count  int     `json:"count"`
    }

    for rows.Next() {
        var hour int
        var amount float64
        var count int
        if err := rows.Scan(&hour, &amount, &count); err != nil {
            continue
        }
        trend = append(trend, struct {
            Hour   int     `json:"hour"`
            Amount float64 `json:"amount"`
            Count  int     `json:"count"`
        }{hour, amount, count})
    }

    if err := rows.Err(); err != nil {
        return fmt.Errorf("error iterating sales trend: %w", err)
    }

    stats.SalesTrend = trend
    return nil
}

func getTopProducts(db *sql.DB, today string, stats *DashboardStats) error {
    query := `
        SELECT p.name, SUM(si.quantity) as units, SUM(si.subtotal) as revenue
        FROM sale_items si
        JOIN products p ON si.product_id = p.id
        JOIN sales s ON si.sale_id = s.id
        WHERE DATE(s.created_at) = ?
        GROUP BY p.id, p.name
        ORDER BY units DESC
        LIMIT 5
    `
    rows, err := db.Query(query, today)
    if err != nil {
        return err
    }
    defer rows.Close()

    var topProducts []struct {
        ProductName string  `json:"product_name"`
        UnitsSold   int     `json:"units_sold"`
        Revenue     float64 `json:"revenue"`
    }

    for rows.Next() {
        var name string
        var units int
        var revenue float64
        if err := rows.Scan(&name, &units, &revenue); err != nil {
            continue
        }
        topProducts = append(topProducts, struct {
            ProductName string  `json:"product_name"`
            UnitsSold   int     `json:"units_sold"`
            Revenue     float64 `json:"revenue"`
        }{name, units, revenue})
    }

    if err := rows.Err(); err != nil {
        return fmt.Errorf("error iterating top products: %w", err)
    }

    stats.TopProducts = topProducts
    return nil
}

func getRecentSalesForDashboard(db *sql.DB, stats *DashboardStats) error {
    sales, err := GetRecentSales(db, 5)
    if err != nil {
        return err
    }
    stats.RecentSales = sales
    return nil
}

func getLowStockForDashboard(db *sql.DB, stats *DashboardStats) error {
    // Count low stock items
    countQuery := `
        SELECT COUNT(*)
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = 1
        WHERE p.is_active = 1 
          AND COALESCE(ss.quantity, 0) <= p.reorder_level
          AND COALESCE(ss.quantity, 0) >= 0
    `
    var count int
    err := db.QueryRow(countQuery).Scan(&count)
    if err != nil {
        stats.LowStockItems = []LowStockReportItem{}
        return nil
    }

    // Get the actual items
    itemsQuery := `
        SELECT 
            p.name as product_name,
            p.category,
            COALESCE(ss.quantity, 0) as stock_quantity,
            p.reorder_level,
            p.cost_price,
            (p.reorder_level - COALESCE(ss.quantity, 0)) * p.cost_price as restock_cost
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = 1
        WHERE p.is_active = 1 AND COALESCE(ss.quantity, 0) <= p.reorder_level
        ORDER BY stock_quantity ASC
        LIMIT 10
    `
    rows, err := db.Query(itemsQuery)
    if err != nil {
        stats.LowStockItems = []LowStockReportItem{}
        return nil
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
            continue
        }
        items = append(items, item)
    }

    if err := rows.Err(); err != nil {
        stats.LowStockItems = []LowStockReportItem{}
        return nil
    }

    stats.LowStockItems = items
    return nil
}
