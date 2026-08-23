package db

import (
    "database/sql"
    "fmt"
    "log"
    "time"
)

type DirectorDashboard struct {
    TotalRevenueToday   float64          `json:"total_revenue_today"`
    TotalRevenueMonth   float64          `json:"total_revenue_month"`
    TotalStores         int              `json:"total_stores"`
    ActiveStores        int              `json:"active_stores"`
    LowStockItems       int              `json:"low_stock_items"`
    TodayTransactions   int              `json:"today_transactions"`
    OutletStats         []OutletStats    `json:"outlet_stats"`
    SalesTrend          []SalesTrendItem `json:"sales_trend"`
    TopProducts         []TopProductItem `json:"top_products"`
    RecentTransactions  []RecentTxItem   `json:"recent_transactions"`
    Alerts              []AlertItem      `json:"alerts"`
}

type OutletStats struct {
    ShopID             int     `json:"shop_id"`
    ShopName           string  `json:"shop_name"`
    Location           string  `json:"location"`
    Manager            string  `json:"manager"`
    TodayRevenue       float64 `json:"today_revenue"`
    TodayTransactions  int     `json:"today_transactions"`
    TodayItems         int     `json:"today_items"`
    MonthRevenue       float64 `json:"month_revenue"`
    MonthTransactions  int     `json:"month_transactions"`
    Status             string  `json:"status"`
}

type SalesTrendItem struct {
    Date  string  `json:"date"`
    Total float64 `json:"total"`
}

type TopProductItem struct {
    ProductName string  `json:"product_name"`
    Category    string  `json:"category"`
    UnitsSold   int     `json:"units_sold"`
    Revenue     float64 `json:"revenue"`
}

type RecentTxItem struct {
    SaleID      int     `json:"sale_id"`
    ShopName    string  `json:"shop_name"`
    Amount      float64 `json:"amount"`
    PaymentType string  `json:"payment_type"`
    CreatedAt   string  `json:"created_at"`
}

type AlertItem struct {
    ShopName     string `json:"shop_name"`
    ProductName  string `json:"product_name"`
    StockLevel   int    `json:"stock_level"`
    ReorderLevel int    `json:"reorder_level"`
    Severity     string `json:"severity"`
}

func GetDirectorDashboard(db *sql.DB) (*DirectorDashboard, error) {
    dashboard := &DirectorDashboard{}
    today := time.Now().Format("2006-01-02")
    monthStart := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

    log.Println("🔍 Starting GetDirectorDashboard...")

    // 1. Get total stores
    err := db.QueryRow("SELECT COUNT(*) FROM branches WHERE is_active = 1").Scan(&dashboard.TotalStores)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get total stores: %w", err)
    }
    log.Printf("✅ Total stores: %d", dashboard.TotalStores)

    // 2. Get active stores (with sales today)
    err = db.QueryRow(`
        SELECT COUNT(DISTINCT b.id) 
        FROM branches b
        INNER JOIN sales s ON s.shop_id = b.id
        WHERE DATE(s.created_at) = ? AND b.is_active = 1
    `, today).Scan(&dashboard.ActiveStores)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get active stores: %w", err)
    }
    log.Printf("✅ Active stores: %d", dashboard.ActiveStores)

    // 3. Get today's total revenue
    err = db.QueryRow(`
        SELECT COALESCE(SUM(total_amount), 0)
        FROM sales
        WHERE DATE(created_at) = ?
    `, today).Scan(&dashboard.TotalRevenueToday)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get today's revenue: %w", err)
    }
    log.Printf("✅ Today's revenue: %.2f", dashboard.TotalRevenueToday)

    // 4. Get month revenue (last 30 days)
    err = db.QueryRow(`
        SELECT COALESCE(SUM(total_amount), 0)
        FROM sales
        WHERE DATE(created_at) >= ?
    `, monthStart).Scan(&dashboard.TotalRevenueMonth)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get month revenue: %w", err)
    }
    log.Printf("✅ Month revenue: %.2f", dashboard.TotalRevenueMonth)

    // 5. Get today's transaction count
    err = db.QueryRow(`
        SELECT COUNT(*)
        FROM sales
        WHERE DATE(created_at) = ?
    `, today).Scan(&dashboard.TodayTransactions)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get today's transactions: %w", err)
    }
    log.Printf("✅ Today's transactions: %d", dashboard.TodayTransactions)

    // 6. Get low stock items (from shop_stock with branch_inventory reorder_level)
    err = db.QueryRow(`
        SELECT COUNT(DISTINCT ss.product_id)
        FROM shop_stock ss
        LEFT JOIN branch_inventory bi ON ss.shop_id = bi.branch_id AND ss.product_id = bi.product_id
        LEFT JOIN products p ON ss.product_id = p.id
        WHERE p.is_active = 1 
        AND ss.quantity <= COALESCE(bi.reorder_level, 5)
    `).Scan(&dashboard.LowStockItems)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("failed to get low stock items: %w", err)
    }
    log.Printf("✅ Low stock items: %d", dashboard.LowStockItems)

    // 7. Get outlet stats
    outletStats, err := getOutletStats(db, today, monthStart)
    if err != nil {
        return nil, fmt.Errorf("failed to get outlet stats: %w", err)
    }
    dashboard.OutletStats = outletStats
    log.Printf("✅ Outlet stats: %d outlets", len(outletStats))

    // 8. Get sales trend
    salesTrend, err := getDirectorSalesTrend(db, monthStart, today)
    if err != nil {
        return nil, fmt.Errorf("failed to get sales trend: %w", err)
    }
    dashboard.SalesTrend = salesTrend
    log.Printf("✅ Sales trend: %d days", len(salesTrend))

    // 9. Get top products
    topProducts, err := getTopProductsAllOutlets(db, monthStart, today)
    if err != nil {
        return nil, fmt.Errorf("failed to get top products: %w", err)
    }
    dashboard.TopProducts = topProducts
    log.Printf("✅ Top products: %d", len(topProducts))

    // 10. Get recent transactions
    recentTransactions, err := getRecentTransactions(db, 10)
    if err != nil {
        return nil, fmt.Errorf("failed to get recent transactions: %w", err)
    }
    dashboard.RecentTransactions = recentTransactions
    log.Printf("✅ Recent transactions: %d", len(recentTransactions))

    // 11. Get alerts
    alerts, err := getAlerts(db)
    if err != nil {
        return nil, fmt.Errorf("failed to get alerts: %w", err)
    }
    dashboard.Alerts = alerts
    log.Printf("✅ Alerts: %d", len(alerts))

    log.Println("🎉 DirectorDashboard complete!")
    return dashboard, nil
}

func getOutletStats(db *sql.DB, today, monthStart string) ([]OutletStats, error) {
    query := `
        SELECT 
            b.id,
            b.name,
            COALESCE(b.location, 'N/A') as location,
            COALESCE(u.username, 'N/A') as manager,
            COALESCE((
                SELECT SUM(total_amount) 
                FROM sales 
                WHERE shop_id = b.id AND DATE(created_at) = ?
            ), 0) as today_revenue,
            COALESCE((
                SELECT COUNT(*) 
                FROM sales 
                WHERE shop_id = b.id AND DATE(created_at) = ?
            ), 0) as today_transactions,
            COALESCE((
                SELECT SUM(si.quantity) 
                FROM sales s
                JOIN sale_items si ON s.id = si.sale_id
                WHERE s.shop_id = b.id AND DATE(s.created_at) = ?
            ), 0) as today_items,
            COALESCE((
                SELECT SUM(total_amount) 
                FROM sales 
                WHERE shop_id = b.id AND DATE(created_at) >= ?
            ), 0) as month_revenue,
            COALESCE((
                SELECT COUNT(*) 
                FROM sales 
                WHERE shop_id = b.id AND DATE(created_at) >= ?
            ), 0) as month_transactions,
            CASE 
                WHEN EXISTS (
                    SELECT 1 FROM sales 
                    WHERE shop_id = b.id AND DATE(created_at) = ?
                ) THEN 'active'
                ELSE 'inactive'
            END as status
        FROM branches b
        LEFT JOIN users u ON b.manager_id = u.id
        WHERE b.is_active = 1
        ORDER BY today_revenue DESC
    `
    rows, err := db.Query(query, today, today, today, monthStart, monthStart, today)
    if err != nil {
        log.Printf("❌ getOutletStats query error: %v", err)
        return nil, err
    }
    defer rows.Close()

    var stats []OutletStats
    for rows.Next() {
        var stat OutletStats
        err := rows.Scan(
            &stat.ShopID, &stat.ShopName, &stat.Location, &stat.Manager,
            &stat.TodayRevenue, &stat.TodayTransactions, &stat.TodayItems,
            &stat.MonthRevenue, &stat.MonthTransactions, &stat.Status,
        )
        if err != nil {
            log.Printf("❌ getOutletStats scan error: %v", err)
            return nil, err
        }
        stats = append(stats, stat)
    }

    if err := rows.Err(); err != nil {
        log.Printf("❌ getOutletStats rows error: %v", err)
        return nil, err
    }

    log.Printf("✅ getOutletStats: found %d outlets", len(stats))
    return stats, nil
}

func getDirectorSalesTrend(db *sql.DB, startDate, endDate string) ([]SalesTrendItem, error) {
    query := `
        SELECT DATE(created_at) as date, COALESCE(SUM(total_amount), 0) as total
        FROM sales
        WHERE DATE(created_at) BETWEEN ? AND ?
        GROUP BY DATE(created_at)
        ORDER BY date ASC
    `
    rows, err := db.Query(query, startDate, endDate)
    if err != nil {
        log.Printf("❌ getDirectorSalesTrend query error: %v", err)
        return nil, err
    }
    defer rows.Close()

    var trend []SalesTrendItem
    for rows.Next() {
        var item SalesTrendItem
        err := rows.Scan(&item.Date, &item.Total)
        if err != nil {
            log.Printf("❌ getDirectorSalesTrend scan error: %v", err)
            return nil, err
        }
        trend = append(trend, item)
    }

    if err := rows.Err(); err != nil {
        log.Printf("❌ getDirectorSalesTrend rows error: %v", err)
        return nil, err
    }

    return trend, nil
}

func getTopProductsAllOutlets(db *sql.DB, startDate, endDate string) ([]TopProductItem, error) {
    query := `
        SELECT 
            p.name,
            COALESCE(p.category, 'Uncategorized') as category,
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
        log.Printf("❌ getTopProductsAllOutlets query error: %v", err)
        return nil, err
    }
    defer rows.Close()

    var products []TopProductItem
    for rows.Next() {
        var item TopProductItem
        err := rows.Scan(&item.ProductName, &item.Category, &item.UnitsSold, &item.Revenue)
        if err != nil {
            log.Printf("❌ getTopProductsAllOutlets scan error: %v", err)
            return nil, err
        }
        products = append(products, item)
    }

    if err := rows.Err(); err != nil {
        log.Printf("❌ getTopProductsAllOutlets rows error: %v", err)
        return nil, err
    }

    return products, nil
}

func getRecentTransactions(db *sql.DB, limit int) ([]RecentTxItem, error) {
    query := `
        SELECT 
            s.id,
            COALESCE(b.name, 'Unknown Shop') as shop_name,
            s.total_amount,
            COALESCE(s.payment_type, 'cash') as payment_type,
            DATE_FORMAT(s.created_at, '%Y-%m-%d %H:%i') as created_at
        FROM sales s
        LEFT JOIN branches b ON s.shop_id = b.id
        ORDER BY s.id DESC
        LIMIT ?
    `
    rows, err := db.Query(query, limit)
    if err != nil {
        log.Printf("❌ getRecentTransactions query error: %v", err)
        return nil, err
    }
    defer rows.Close()

    var transactions []RecentTxItem
    for rows.Next() {
        var t RecentTxItem
        err := rows.Scan(&t.SaleID, &t.ShopName, &t.Amount, &t.PaymentType, &t.CreatedAt)
        if err != nil {
            log.Printf("❌ getRecentTransactions scan error: %v", err)
            return nil, err
        }
        transactions = append(transactions, t)
    }

    if err := rows.Err(); err != nil {
        log.Printf("❌ getRecentTransactions rows error: %v", err)
        return nil, err
    }

    return transactions, nil
}

func getAlerts(db *sql.DB) ([]AlertItem, error) {
    query := `
        SELECT 
            COALESCE(b.name, 'Main Shop') as shop_name,
            p.name as product_name,
            COALESCE(ss.quantity, 0) as stock_level,
            COALESCE(bi.reorder_level, 5) as reorder_level,
            CASE 
                WHEN COALESCE(ss.quantity, 0) = 0 THEN 'critical'
                WHEN COALESCE(ss.quantity, 0) <= COALESCE(bi.reorder_level, 5) / 2 THEN 'high'
                ELSE 'medium'
            END as severity
        FROM products p
        JOIN shop_stock ss ON p.id = ss.product_id
        LEFT JOIN branches b ON ss.shop_id = b.id
        LEFT JOIN branch_inventory bi ON b.id = bi.branch_id AND p.id = bi.product_id
        WHERE p.is_active = 1 
        AND COALESCE(ss.quantity, 0) <= COALESCE(bi.reorder_level, 5)
        ORDER BY stock_level ASC
        LIMIT 20
    `
    rows, err := db.Query(query)
    if err != nil {
        log.Printf("❌ getAlerts query error: %v", err)
        return nil, err
    }
    defer rows.Close()

    var alerts []AlertItem
    for rows.Next() {
        var a AlertItem
        err := rows.Scan(&a.ShopName, &a.ProductName, &a.StockLevel, &a.ReorderLevel, &a.Severity)
        if err != nil {
            log.Printf("❌ getAlerts scan error: %v", err)
            return nil, err
        }
        alerts = append(alerts, a)
    }

    if err := rows.Err(); err != nil {
        log.Printf("❌ getAlerts rows error: %v", err)
        return nil, err
    }

    return alerts, nil
}
