package db

import (
	"database/sql"
	"time"
)

func GetDashboardStats(db *sql.DB, shopID int) (*DashboardStats, error) {
    stats := &DashboardStats{}
    today := time.Now().Format("2006-01-02")
    yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

    // Get today's sales stats
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
    args := []interface{}{today}
    
    if shopID > 0 {
        query += " AND s.shop_id = ?"
        args = append(args, shopID)
    }
    
    err := db.QueryRow(query, args...).Scan(
        &stats.TodaySales.TotalRevenue,
        &stats.TodaySales.TransactionCount,
        &stats.TodaySales.AverageTicket,
        &stats.TodaySales.TotalItems,
    )
    if err != nil && err != sql.ErrNoRows {
        return nil, err
    }

    // Get yesterday's sales
    query = `
        SELECT 
            COALESCE(SUM(total_amount), 0),
            COUNT(*)
        FROM sales s
        WHERE DATE(s.created_at) = ? 
    `
    args = []interface{}{yesterday}
    
    if shopID > 0 {
        query += " AND s.shop_id = ?"
        args = append(args, shopID)
    }
    
    err = db.QueryRow(query, args...).Scan(
        &stats.YesterdaySales.TotalRevenue,
        &stats.YesterdaySales.TransactionCount,
    )
    if err != nil && err != sql.ErrNoRows {
        return nil, err
    }

    // Get quick stats
    var totalProducts int
    var totalStockValue float64
    
    if shopID > 0 {
        err = db.QueryRow(`
            SELECT 
                COUNT(DISTINCT p.id),
                COALESCE(SUM(ss.quantity * p.cost_price), 0)
            FROM products p
            LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = ?
            WHERE p.is_active = 1
        `, shopID).Scan(&totalProducts, &totalStockValue)
    } else {
        err = db.QueryRow(`
            SELECT 
                COUNT(DISTINCT p.id),
                COALESCE(SUM(ss.quantity * p.cost_price), 0)
            FROM products p
            LEFT JOIN shop_stock ss ON p.id = ss.product_id
            WHERE p.is_active = 1
        `).Scan(&totalProducts, &totalStockValue)
    }
    if err != nil && err != sql.ErrNoRows {
        return nil, err
    }
    stats.QuickStats.TotalProducts = totalProducts
    stats.QuickStats.TotalStockValue = totalStockValue

    // Get sales trend
    query = `
        SELECT 
            HOUR(created_at) as hour,
            COALESCE(SUM(total_amount), 0) as amount,
            COUNT(*) as count
        FROM sales s
        WHERE DATE(s.created_at) = ? 
    `
    args = []interface{}{today}
    
    if shopID > 0 {
        query += " AND s.shop_id = ?"
        args = append(args, shopID)
    }
    
    query += " GROUP BY HOUR(created_at) ORDER BY hour ASC"
    
    rows, err := db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var trend []struct {
        Hour   int     `json:"hour"`
        Amount float64 `json:"amount"`
        Count  int     `json:"count"`
    }
    for rows.Next() {
        var t struct {
            Hour   int     `json:"hour"`
            Amount float64 `json:"amount"`
            Count  int     `json:"count"`
        }
        err := rows.Scan(&t.Hour, &t.Amount, &t.Count)
        if err != nil {
            return nil, err
        }
        trend = append(trend, t)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    stats.SalesTrend = trend

    // Get top products
    query = `
        SELECT 
            p.name,
            SUM(si.quantity) as units_sold,
            SUM(si.subtotal) as revenue
        FROM sale_items si
        JOIN products p ON si.product_id = p.id
        JOIN sales s ON si.sale_id = s.id
        WHERE DATE(s.created_at) = ? 
    `
    args = []interface{}{today}
    
    if shopID > 0 {
        query += " AND s.shop_id = ?"
        args = append(args, shopID)
    }
    
    query += " GROUP BY p.id, p.name ORDER BY units_sold DESC LIMIT 10"
    
    rows, err = db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var topProducts []struct {
        ProductName string  `json:"product_name"`
        UnitsSold   int     `json:"units_sold"`
        Revenue     float64 `json:"revenue"`
    }
    for rows.Next() {
        var tp struct {
            ProductName string  `json:"product_name"`
            UnitsSold   int     `json:"units_sold"`
            Revenue     float64 `json:"revenue"`
        }
        err := rows.Scan(&tp.ProductName, &tp.UnitsSold, &tp.Revenue)
        if err != nil {
            return nil, err
        }
        topProducts = append(topProducts, tp)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    stats.TopProducts = topProducts

    // Get recent sales
    query = `
        SELECT 
            s.id,
            s.total_amount,
            s.cash_amount,
            s.mpesa_amount,
            s.mpesa_code,
            s.payment_type,
            s.change_given,
            s.shop_id,
            s.created_at
        FROM sales s
        WHERE 1=1 
    `
    args = []interface{}{}
    
    if shopID > 0 {
        query += " AND s.shop_id = ?"
        args = append(args, shopID)
    }
    
    query += " ORDER BY s.id DESC LIMIT 10"
    
    rows, err = db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var recentSales []Sale
    for rows.Next() {
        var sale Sale
        err := rows.Scan(
            &sale.ID, &sale.TotalAmount, &sale.CashAmount,
            &sale.MpesaAmount, &sale.MpesaCode, &sale.PaymentType,
            &sale.ChangeGiven, &sale.ShopID, &sale.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        // Get items for this sale
        items, err := getSaleItems(db, sale.ID)
        if err == nil {
            sale.Items = items
        }
        recentSales = append(recentSales, sale)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    stats.RecentSales = recentSales

    // Get low stock items
    var lowStock []LowStockReportItem
    var lowStockRows *sql.Rows
    
    if shopID > 0 {
        lowStockRows, err = db.Query(`
            SELECT 
                p.name,
                p.category,
                COALESCE(ss.quantity, 0) as stock_quantity,
                p.reorder_level,
                p.cost_price,
                COALESCE(ss.quantity * p.cost_price, 0) as restock_cost
            FROM products p
            LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = ?
            WHERE p.is_active = 1 AND COALESCE(ss.quantity, 0) <= p.reorder_level
            ORDER BY stock_quantity ASC
            LIMIT 20
        `, shopID)
    } else {
        lowStockRows, err = db.Query(`
            SELECT 
                p.name,
                p.category,
                COALESCE(ss.quantity, 0) as stock_quantity,
                p.reorder_level,
                p.cost_price,
                COALESCE(ss.quantity * p.cost_price, 0) as restock_cost
            FROM products p
            LEFT JOIN shop_stock ss ON p.id = ss.product_id
            WHERE p.is_active = 1 AND COALESCE(ss.quantity, 0) <= p.reorder_level
            ORDER BY stock_quantity ASC
            LIMIT 20
        `)
    }
    if err != nil && err != sql.ErrNoRows {
        return nil, err
    }
    if lowStockRows != nil {
        defer lowStockRows.Close()
        for lowStockRows.Next() {
            var item LowStockReportItem
            err := lowStockRows.Scan(
                &item.ProductName,
                &item.Category,
                &item.StockQuantity,
                &item.ReorderLevel,
                &item.CostPrice,
                &item.RestockCost,
            )
            if err != nil {
                return nil, err
            }
            lowStock = append(lowStock, item)
        }
        if err := lowStockRows.Err(); err != nil {
            return nil, err
        }
    }
    stats.LowStockItems = lowStock

    return stats, nil
}