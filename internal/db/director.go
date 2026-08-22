package db

import (
    "database/sql"
    "fmt"
    "time"
)

// ============================================
// DIRECTOR DASHBOARD STRUCTS
// ============================================

type OutletStats struct {
    ShopID            int     `json:"shop_id"`
    ShopName          string  `json:"shop_name"`
    Location          string  `json:"location"`
    TodayRevenue      float64 `json:"today_revenue"`
    TodayTransactions int     `json:"today_transactions"`
    TodayItems        int     `json:"today_items"`
    MonthRevenue      float64 `json:"month_revenue"`
    Status            string  `json:"status"`
}

type DirectorDashboard struct {
    TotalRevenueToday   float64          `json:"total_revenue_today"`
    TotalRevenueMonth   float64          `json:"total_revenue_month"`
    TotalStores         int              `json:"total_stores"`
    ActiveStores        int              `json:"active_stores"`
    LowStockItems       int              `json:"low_stock_items"`
    TodayTransactions   int              `json:"today_transactions"`
    OutletStats         []OutletStats    `json:"outlet_stats"`
    SalesTrend          []struct {
        Date  string  `json:"date"`
        Total float64 `json:"total"`
    } `json:"sales_trend"`
    TopProducts []struct {
        ProductName string  `json:"product_name"`
        Category    string  `json:"category"`
        UnitsSold   int     `json:"units_sold"`
        Revenue     float64 `json:"revenue"`
    } `json:"top_products"`
    RecentTransactions []struct {
        SaleID      int     `json:"sale_id"`
        ShopName    string  `json:"shop_name"`
        Amount      float64 `json:"amount"`
        PaymentType string  `json:"payment_type"`
        CreatedAt   string  `json:"created_at"`
    } `json:"recent_transactions"`
    Alerts []struct {
        ShopName     string `json:"shop_name"`
        ProductName  string `json:"product_name"`
        StockLevel   int    `json:"stock_level"`
        ReorderLevel int    `json:"reorder_level"`
        Severity     string `json:"severity"`
    } `json:"alerts"`
}

// GetDirectorDashboard fetches all data for the director dashboard
func GetDirectorDashboard(db *sql.DB) (*DirectorDashboard, error) {
    dashboard := &DirectorDashboard{}
    today := time.Now().Format("2006-01-02")
    monthStart := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

    // Get total stores
    err := db.QueryRow("SELECT COUNT(*) FROM shops").Scan(&dashboard.TotalStores)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get total stores: %w", err)
    }

    // Get active stores (with sales today)
    err = db.QueryRow(`
        SELECT COUNT(DISTINCT s.id) 
        FROM shops s
        INNER JOIN sales sl ON sl.shop_id = s.id
        WHERE DATE(sl.created_at) = ?
    `, today).Scan(&dashboard.ActiveStores)
    if err != nil && err != sql.ErrNoRows {
        dashboard.ActiveStores = dashboard.TotalStores
    }

    // Get today's total revenue
    err = db.QueryRow(`
        SELECT COALESCE(SUM(total_amount), 0)
        FROM sales
        WHERE DATE(created_at) = ?
    `, today).Scan(&dashboard.TotalRevenueToday)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get today's revenue: %w", err)
    }

    // Get month revenue
    err = db.QueryRow(`
        SELECT COALESCE(SUM(total_amount), 0)
        FROM sales
        WHERE DATE(created_at) >= ?
    `, monthStart).Scan(&dashboard.TotalRevenueMonth)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get month revenue: %w", err)
    }

    // Get today's transaction count
    err = db.QueryRow(`
        SELECT COUNT(*)
        FROM sales
        WHERE DATE(created_at) = ?
    `, today).Scan(&dashboard.TodayTransactions)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get today's transactions: %w", err)
    }

    // Get low stock items count
    err = db.QueryRow(`
        SELECT COUNT(*)
        FROM products
        WHERE stock_quantity <= reorder_level AND stock_quantity > 0
    `).Scan(&dashboard.LowStockItems)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get low stock items: %w", err)
    }

    // Get outlet stats
    outletStats, err := getDirectorOutletStats(db, today, monthStart)
    if err != nil {
        return nil, fmt.Errorf("failed to get outlet stats: %w", err)
    }
    dashboard.OutletStats = outletStats

    // Get sales trend (last 30 days)
    salesTrend, err := getDirectorSalesTrend(db, monthStart, today)
    if err != nil {
        return nil, fmt.Errorf("failed to get sales trend: %w", err)
    }
    dashboard.SalesTrend = salesTrend

    // Get top products (all outlets)
    topProducts, err := getDirectorTopProducts(db, monthStart, today)
    if err != nil {
        return nil, fmt.Errorf("failed to get top products: %w", err)
    }
    dashboard.TopProducts = topProducts

    // Get recent transactions
    recentTransactions, err := getDirectorRecentTransactions(db, 10)
    if err != nil {
        return nil, fmt.Errorf("failed to get recent transactions: %w", err)
    }
    dashboard.RecentTransactions = recentTransactions

    // Get alerts
    alerts, err := getDirectorAlerts(db)
    if err != nil {
        // If alerts fail, just return empty alerts
        dashboard.Alerts = []struct {
            ShopName     string `json:"shop_name"`
            ProductName  string `json:"product_name"`
            StockLevel   int    `json:"stock_level"`
            ReorderLevel int    `json:"reorder_level"`
            Severity     string `json:"severity"`
        }{}
    }
    dashboard.Alerts = alerts

    return dashboard, nil
}

func getDirectorOutletStats(db *sql.DB, today, monthStart string) ([]OutletStats, error) {
    query := `
        SELECT 
            s.id,
            s.name,
            s.location,
            COALESCE((
                SELECT SUM(total_amount) 
                FROM sales 
                WHERE shop_id = s.id AND DATE(created_at) = ?
            ), 0) as today_revenue,
            COALESCE((
                SELECT COUNT(*) 
                FROM sales 
                WHERE shop_id = s.id AND DATE(created_at) = ?
            ), 0) as today_transactions,
            COALESCE((
                SELECT SUM(si.quantity) 
                FROM sales sl
                JOIN sale_items si ON sl.id = si.sale_id
                WHERE sl.shop_id = s.id AND DATE(sl.created_at) = ?
            ), 0) as today_items,
            COALESCE((
                SELECT SUM(total_amount) 
                FROM sales 
                WHERE shop_id = s.id AND DATE(created_at) >= ?
            ), 0) as month_revenue,
            CASE 
                WHEN EXISTS (
                    SELECT 1 FROM sales 
                    WHERE shop_id = s.id AND DATE(created_at) = ?
                ) THEN 'active'
                ELSE 'inactive'
            END as status
        FROM shops s
        ORDER BY today_revenue DESC
    `
    rows, err := db.Query(query, today, today, today, monthStart, today)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var stats []OutletStats
    for rows.Next() {
        var stat OutletStats
        err := rows.Scan(
            &stat.ShopID, &stat.ShopName, &stat.Location,
            &stat.TodayRevenue, &stat.TodayTransactions, &stat.TodayItems,
            &stat.MonthRevenue, &stat.Status,
        )
        if err != nil {
            return nil, err
        }
        stats = append(stats, stat)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return stats, nil
}

func getDirectorSalesTrend(db *sql.DB, startDate, endDate string) ([]struct {
    Date  string  `json:"date"`
    Total float64 `json:"total"`
}, error) {
    query := `
        SELECT DATE(created_at) as date, COALESCE(SUM(total_amount), 0) as total
        FROM sales
        WHERE DATE(created_at) BETWEEN ? AND ?
        GROUP BY DATE(created_at)
        ORDER BY date ASC
    `
    rows, err := db.Query(query, startDate, endDate)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var trend []struct {
        Date  string  `json:"date"`
        Total float64 `json:"total"`
    }
    for rows.Next() {
        var item struct {
            Date  string  `json:"date"`
            Total float64 `json:"total"`
        }
        err := rows.Scan(&item.Date, &item.Total)
        if err != nil {
            return nil, err
        }
        trend = append(trend, item)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return trend, nil
}

func getDirectorTopProducts(db *sql.DB, startDate, endDate string) ([]struct {
    ProductName string  `json:"product_name"`
    Category    string  `json:"category"`
    UnitsSold   int     `json:"units_sold"`
    Revenue     float64 `json:"revenue"`
}, error) {
    query := `
        SELECT 
            p.name,
            p.category,
            SUM(si.quantity) as units_sold,
            SUM(si.subtotal) as revenue
        FROM sale_items si
        JOIN products p ON si.product_id = p.id
        JOIN sales s ON si.sale_id = s.id
        WHERE DATE(s.created_at) BETWEEN ? AND ?
        GROUP BY p.id, p.name, p.category
        ORDER BY units_sold DESC
        LIMIT 10
    `
    rows, err := db.Query(query, startDate, endDate)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var products []struct {
        ProductName string  `json:"product_name"`
        Category    string  `json:"category"`
        UnitsSold   int     `json:"units_sold"`
        Revenue     float64 `json:"revenue"`
    }
    for rows.Next() {
        var item struct {
            ProductName string  `json:"product_name"`
            Category    string  `json:"category"`
            UnitsSold   int     `json:"units_sold"`
            Revenue     float64 `json:"revenue"`
        }
        err := rows.Scan(&item.ProductName, &item.Category, &item.UnitsSold, &item.Revenue)
        if err != nil {
            return nil, err
        }
        products = append(products, item)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return products, nil
}

func getDirectorRecentTransactions(db *sql.DB, limit int) ([]struct {
    SaleID      int     `json:"sale_id"`
    ShopName    string  `json:"shop_name"`
    Amount      float64 `json:"amount"`
    PaymentType string  `json:"payment_type"`
    CreatedAt   string  `json:"created_at"`
}, error) {
    query := `
        SELECT 
            s.id,
            COALESCE(sh.name, 'Unknown Shop') as shop_name,
            s.total_amount,
            s.payment_type,
            DATE_FORMAT(s.created_at, '%Y-%m-%d %H:%i') as created_at
        FROM sales s
        LEFT JOIN shops sh ON s.shop_id = sh.id
        ORDER BY s.id DESC
        LIMIT ?
    `
    rows, err := db.Query(query, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var transactions []struct {
        SaleID      int     `json:"sale_id"`
        ShopName    string  `json:"shop_name"`
        Amount      float64 `json:"amount"`
        PaymentType string  `json:"payment_type"`
        CreatedAt   string  `json:"created_at"`
    }
    for rows.Next() {
        var t struct {
            SaleID      int     `json:"sale_id"`
            ShopName    string  `json:"shop_name"`
            Amount      float64 `json:"amount"`
            PaymentType string  `json:"payment_type"`
            CreatedAt   string  `json:"created_at"`
        }
        err := rows.Scan(&t.SaleID, &t.ShopName, &t.Amount, &t.PaymentType, &t.CreatedAt)
        if err != nil {
            return nil, err
        }
        transactions = append(transactions, t)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return transactions, nil
}

func getDirectorAlerts(db *sql.DB) ([]struct {
    ShopName     string `json:"shop_name"`
    ProductName  string `json:"product_name"`
    StockLevel   int    `json:"stock_level"`
    ReorderLevel int    `json:"reorder_level"`
    Severity     string `json:"severity"`
}, error) {
    query := `
        SELECT 
            COALESCE(sh.name, 'Main Shop') as shop_name,
            p.name as product_name,
            p.stock_quantity as stock_level,
            p.reorder_level,
            CASE 
                WHEN p.stock_quantity = 0 THEN 'critical'
                WHEN p.stock_quantity <= p.reorder_level/2 THEN 'high'
                ELSE 'medium'
            END as severity
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id
        LEFT JOIN shops sh ON ss.shop_id = sh.id
        WHERE p.stock_quantity <= p.reorder_level
        ORDER BY p.stock_quantity ASC
        LIMIT 20
    `
    rows, err := db.Query(query)
    if err != nil {
        // Return empty alerts if the query fails (e.g., missing tables)
        return []struct {
            ShopName     string `json:"shop_name"`
            ProductName  string `json:"product_name"`
            StockLevel   int    `json:"stock_level"`
            ReorderLevel int    `json:"reorder_level"`
            Severity     string `json:"severity"`
        }{}, nil
    }
    defer rows.Close()

    var alerts []struct {
        ShopName     string `json:"shop_name"`
        ProductName  string `json:"product_name"`
        StockLevel   int    `json:"stock_level"`
        ReorderLevel int    `json:"reorder_level"`
        Severity     string `json:"severity"`
    }
    for rows.Next() {
        var a struct {
            ShopName     string `json:"shop_name"`
            ProductName  string `json:"product_name"`
            StockLevel   int    `json:"stock_level"`
            ReorderLevel int    `json:"reorder_level"`
            Severity     string `json:"severity"`
        }
        err := rows.Scan(&a.ShopName, &a.ProductName, &a.StockLevel, &a.ReorderLevel, &a.Severity)
        if err != nil {
            return nil, err
        }
        alerts = append(alerts, a)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return alerts, nil
}


