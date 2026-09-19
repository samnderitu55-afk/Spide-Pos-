package db

import (
	"database/sql"
	"time"
)

func GetDashboardStats(db *sql.DB, shopID int) (*DashboardStats, error) {
	stats := &DashboardStats{}
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Get today's sales stats (revenue, transactions, average ticket)
	query := `
        SELECT 
            COALESCE(SUM(s.total_amount), 0) AS total_revenue,
            COUNT(DISTINCT s.id)             AS transaction_count,
            COALESCE(AVG(s.total_amount), 0) AS average_ticket
        FROM sales s
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
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Get today's item count separately (separate query, no cross-multiplication)
	itemQuery := `
        SELECT COALESCE(SUM(si.quantity), 0)
        FROM sale_items si
        JOIN sales s ON si.sale_id = s.id
        WHERE DATE(s.created_at) = ?
    `
	itemArgs := []interface{}{today}
	if shopID > 0 {
		itemQuery += " AND s.shop_id = ?"
		itemArgs = append(itemArgs, shopID)
	}
	err = db.QueryRow(itemQuery, itemArgs...).Scan(&stats.TodaySales.TotalItems)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Get yesterday's sales
	query = `
        SELECT 
            COALESCE(SUM(s.total_amount), 0),
            COUNT(DISTINCT s.id)
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

	// Get quick stats — only products actually stocked in this shop
	var totalProducts int
	var totalStockValue float64
	if shopID > 0 {
		err = db.QueryRow(`
            SELECT 
                COUNT(DISTINCT p.id),
                COALESCE(SUM(ss.quantity * p.cost_price), 0)
            FROM products p
            JOIN shop_stock ss ON p.id = ss.product_id AND ss.shop_id = ?
            WHERE p.is_active = 1
        `, shopID).Scan(&totalProducts, &totalStockValue)
	} else {
		err = db.QueryRow(`
            SELECT 
                COUNT(DISTINCT p.id),
                COALESCE(SUM(ss.quantity * p.cost_price), 0)
            FROM products p
            JOIN shop_stock ss ON p.id = ss.product_id
            WHERE p.is_active = 1
        `).Scan(&totalProducts, &totalStockValue)
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	stats.QuickStats.TotalProducts = totalProducts
	stats.QuickStats.TotalStockValue = totalStockValue

	// Get sales trend (unchanged — no sale_items join, so no bug)
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

	trend := []struct {
		Hour   int     `json:"hour"`
		Amount float64 `json:"amount"`
		Count  int     `json:"count"`
	}{}
	for rows.Next() {
		var t struct {
			Hour   int     `json:"hour"`
			Amount float64 `json:"amount"`
			Count  int     `json:"count"`
		}
		if err := rows.Scan(&t.Hour, &t.Amount, &t.Count); err != nil {
			return nil, err
		}
		trend = append(trend, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	stats.SalesTrend = trend

	// Get top products (already correct — groups by product, no cross-multiplication issue)
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

	topProducts := []struct {
		ProductName string  `json:"product_name"`
		UnitsSold   int     `json:"units_sold"`
		Revenue     float64 `json:"revenue"`
	}{}
	for rows.Next() {
		var tp struct {
			ProductName string  `json:"product_name"`
			UnitsSold   int     `json:"units_sold"`
			Revenue     float64 `json:"revenue"`
		}
		if err := rows.Scan(&tp.ProductName, &tp.UnitsSold, &tp.Revenue); err != nil {
			return nil, err
		}
		topProducts = append(topProducts, tp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	stats.TopProducts = topProducts

	// Get recent sales (unchanged)
	query = `
        SELECT 
            s.id, s.total_amount, s.cash_amount, s.mpesa_amount,
            s.mpesa_code, s.payment_type, s.change_given,
            s.shop_id, s.created_at
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

	recentSales := []Sale{}
	for rows.Next() {
		var sale Sale
		if err := rows.Scan(
			&sale.ID, &sale.TotalAmount, &sale.CashAmount,
			&sale.MpesaAmount, &sale.MpesaCode, &sale.PaymentType,
			&sale.ChangeGiven, &sale.ShopID, &sale.CreatedAt,
		); err != nil {
			return nil, err
		}
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

	// Get low stock items — INNER JOIN, only products stocked in this shop
	lowStock := []LowStockReportItem{}
	var lowStockRows *sql.Rows
	if shopID > 0 {
		lowStockRows, err = db.Query(`
            SELECT 
                p.name,
                p.category,
                ss.quantity,
                COALESCE(ss.reorder_level, p.reorder_level) AS reorder_level,
                p.cost_price,
                ((COALESCE(ss.reorder_level, p.reorder_level) - ss.quantity) * p.cost_price) as restock_cost
            FROM shop_stock ss
            JOIN products p ON p.id = ss.product_id
            WHERE ss.shop_id = ? 
              AND p.is_active = 1 
              AND ss.quantity <= COALESCE(ss.reorder_level, p.reorder_level)
            ORDER BY ss.quantity ASC
            LIMIT 20
        `, shopID)
	} else {
		lowStockRows, err = db.Query(`
            SELECT 
                p.name,
                p.category,
                ss.quantity,
                COALESCE(ss.reorder_level, p.reorder_level) AS reorder_level,
                p.cost_price,
                ((COALESCE(ss.reorder_level, p.reorder_level) - ss.quantity) * p.cost_price) as restock_cost
            FROM shop_stock ss
            JOIN products p ON p.id = ss.product_id
            WHERE p.is_active = 1 
              AND ss.quantity <= COALESCE(ss.reorder_level, p.reorder_level)
            ORDER BY ss.quantity ASC
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
			if err := lowStockRows.Scan(
				&item.ProductName, &item.Category, &item.StockQuantity,
				&item.ReorderLevel, &item.CostPrice, &item.RestockCost,
			); err != nil {
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

type ShopStats struct {
	Revenue       float64 `json:"revenue"`
	TotalOrders   int     `json:"total_orders"`
	TodaySales    float64 `json:"today_sales"`
	TodayOrders   int     `json:"today_orders"`
	CashAmount    float64 `json:"cash_amount"`
	MpesaAmount   float64 `json:"mpesa_amount"`
	DepositAmount float64 `json:"deposit_amount"`
	CreditAmount  float64 `json:"credit_amount"`
	LowStockItems int     `json:"low_stock_items"`
}

func GetShopStats(db *sql.DB, shopID int) (*ShopStats, error) {
	stats := &ShopStats{}

	// Get revenue and orders
	query := `
        SELECT 
            COALESCE(SUM(total_amount), 0) as revenue,
            COUNT(*) as total_orders,
            COALESCE(SUM(CASE WHEN DATE(created_at) = CURDATE() THEN total_amount ELSE 0 END), 0) as today_sales,
            COUNT(CASE WHEN DATE(created_at) = CURDATE() THEN 1 END) as today_orders,
            COALESCE(SUM(CASE WHEN payment_type = 'cash' THEN total_amount ELSE 0 END), 0) as cash_amount,
            COALESCE(SUM(CASE WHEN payment_type = 'mpesa' THEN total_amount ELSE 0 END), 0) as mpesa_amount,
            COALESCE(SUM(CASE WHEN payment_type = 'deposit' THEN total_amount ELSE 0 END), 0) as deposit_amount,
            COALESCE(SUM(CASE WHEN payment_type = 'credit' THEN total_amount ELSE 0 END), 0) as credit_amount
        FROM sales
        WHERE shop_id = ?
    `
	err := db.QueryRow(query, shopID).Scan(
		&stats.Revenue,
		&stats.TotalOrders,
		&stats.TodaySales,
		&stats.TodayOrders,
		&stats.CashAmount,
		&stats.MpesaAmount,
		&stats.DepositAmount,
		&stats.CreditAmount,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Get low stock items
	err = db.QueryRow(`
        SELECT COUNT(*) 
        FROM shop_stock ss
        JOIN products p ON ss.product_id = p.id
        WHERE ss.shop_id = ? AND ss.quantity <= p.reorder_level
    `, shopID).Scan(&stats.LowStockItems)
	if err != nil && err != sql.ErrNoRows {
		stats.LowStockItems = 0
	}

	return stats, nil
}
