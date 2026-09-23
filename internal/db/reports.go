package db

import (
	"database/sql"
	"fmt"
	"time"
)

type LowStockItem struct {
	ProductID     int     `json:"product_id"`
	ProductName   string  `json:"product_name"`
	Category      string  `json:"category"`
	StockQuantity float64 `json:"stock_quantity"`
	ReorderLevel  float64 `json:"reorder_level"`
	RestockCost   float64 `json:"restock_cost"`
	ShopID        int     `json:"shop_id"`
	ShopName      string  `json:"shop_name"`
}

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

func GetProductSalesReport(
	db *sql.DB,
	companyID int,
	startDate, endDate string,
	shopID int,
	category string,
	productID int,
) ([]ProductSalesReportItem, error) {
	if startDate == "" || endDate == "" {
		return nil, fmt.Errorf("start and end dates are required")
	}

	var query string
	var args []interface{}

	if shopID > 0 {
		query = `
			SELECT p.id, p.name, p.category,
			       SUM(si.quantity),
			       SUM(si.quantity * p.cost_price),
			       SUM(si.subtotal),
			       SUM(si.subtotal - si.quantity * p.cost_price),
			       COALESCE(ss.quantity, 0)  AS current_stock,
			       COALESCE(ss.stock_cap, 0) AS stock_cap
			FROM sale_items si
			JOIN products p ON si.product_id = p.id
			JOIN sales s    ON si.sale_id = s.id
			LEFT JOIN shop_stock ss
			       ON ss.product_id = p.id AND ss.shop_id = ?
			WHERE s.company_id = ?
			  AND DATE(s.created_at) BETWEEN ? AND ?
			  AND s.shop_id = ?
		`
		args = []interface{}{shopID, companyID, startDate, endDate, shopID}
	} else {
		query = `
			SELECT p.id, p.name, p.category,
			       SUM(si.quantity),
			       SUM(si.quantity * p.cost_price),
			       SUM(si.subtotal),
			       SUM(si.subtotal - si.quantity * p.cost_price),
			       COALESCE(agg.total_qty, 0)  AS current_stock,
			       COALESCE(agg.total_cap, 0)  AS stock_cap
			FROM sale_items si
			JOIN products p ON si.product_id = p.id
			JOIN sales s    ON si.sale_id = s.id
			LEFT JOIN (
			    SELECT ss.product_id,
			           SUM(ss.quantity)  AS total_qty,
			           SUM(ss.stock_cap) AS total_cap
			    FROM shop_stock ss
			    JOIN shops sh ON sh.id = ss.shop_id
			    WHERE sh.company_id = ?
			    GROUP BY ss.product_id
			) agg ON agg.product_id = p.id
			WHERE s.company_id = ?
			  AND DATE(s.created_at) BETWEEN ? AND ?
		`
		args = []interface{}{companyID, companyID, startDate, endDate}
	}

	if category != "" {
		query += " AND p.category = ?"
		args = append(args, category)
	}
	if productID > 0 {
		query += " AND si.product_id = ?"
		args = append(args, productID)
	}

	query += " GROUP BY p.id, p.name, p.category"
	if shopID > 0 {
		query += ", ss.quantity, ss.stock_cap"
	} else {
		query += ", agg.total_qty, agg.total_cap"
	}
	query += " ORDER BY SUM(si.subtotal) DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []ProductSalesReportItem{}
	for rows.Next() {
		var row ProductSalesReportItem
		if err := rows.Scan(
			&row.ProductID,
			&row.ProductName,
			&row.Category,
			&row.UnitsSold,
			&row.TotalCost,
			&row.TotalRevenue,
			&row.NetProfit,
			&row.CurrentStock,
			&row.StockCap,
		); err != nil {
			return nil, err
		}
		if row.TotalRevenue > 0 {
			row.MarginPct = (row.NetProfit / row.TotalRevenue) * 100
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating product sales: %w", err)
	}
	return results, nil
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

	items := []InventoryValuationItem{}
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

	details := []struct {
		ProductName string  `json:"product_name"`
		InStock     int     `json:"in_stock"`
		CostPrice   float64 `json:"cost_price"`
		RetailPrice float64 `json:"retail_price"`
		TotalCost   float64 `json:"total_cost"`
	}{}
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

func GetLowStockItems(db *sql.DB, companyID int, shopID int) ([]LowStockItem, error) {
	shopFilter := ""
	args := []interface{}{companyID}

	if shopID > 0 {
		shopFilter = " AND ss.shop_id = ?"
		args = append(args, shopID)
	}

	query := `
        SELECT p.id, p.name, p.category, ss.quantity, p.reorder_level,
               (p.cost_price * (p.reorder_level - ss.quantity)) as restock_cost,
               ss.shop_id, s.name as shop_name
        FROM shop_stock ss
        JOIN products p ON ss.product_id = p.id
        JOIN shops s ON ss.shop_id = s.id
        WHERE ss.company_id = ? AND ss.quantity <= p.reorder_level` + shopFilter + `
        ORDER BY ss.quantity ASC, p.name
    `

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []LowStockItem{}
	for rows.Next() {
		var item LowStockItem
		err := rows.Scan(
			&item.ProductID, &item.ProductName, &item.Category,
			&item.StockQuantity, &item.ReorderLevel, &item.RestockCost,
			&item.ShopID, &item.ShopName,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetProductSalesCategories returns distinct categories for the given shop
// (or all company categories if shopID is 0).
func GetProductSalesCategories(db *sql.DB, companyID, shopID int) ([]string, error) {
	var query string
	var args []interface{}

	if shopID > 0 {
		query = `
			SELECT DISTINCT p.category
			FROM shop_stock ss
			JOIN products p ON p.id = ss.product_id
			WHERE ss.shop_id = ? AND p.company_id = ?
			  AND p.category IS NOT NULL AND p.category != ''
			ORDER BY p.category
		`
		args = []interface{}{shopID, companyID}
	} else {
		query = `
			SELECT DISTINCT category
			FROM products
			WHERE company_id = ? AND is_active = 1
			  AND category IS NOT NULL AND category != ''
			ORDER BY category
		`
		args = []interface{}{companyID}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []string{}
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

// ProductPick is a minimal product projection for dropdowns.
type ProductPick struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GetProductSalesProducts returns id+name pairs, filtered by shop and/or category.
func GetProductSalesProducts(db *sql.DB, companyID, shopID int, category string) ([]ProductPick, error) {
	var query string
	var args []interface{}

	if shopID > 0 && category != "" {
		query = `SELECT DISTINCT p.id, p.name
		         FROM shop_stock ss
		         JOIN products p ON p.id = ss.product_id
		         WHERE ss.shop_id = ? AND p.company_id = ? AND p.is_active = 1 AND p.category = ?
		         ORDER BY p.name`
		args = []interface{}{shopID, companyID, category}
	} else if shopID > 0 {
		query = `SELECT DISTINCT p.id, p.name
		         FROM shop_stock ss
		         JOIN products p ON p.id = ss.product_id
		         WHERE ss.shop_id = ? AND p.company_id = ? AND p.is_active = 1
		         ORDER BY p.name`
		args = []interface{}{shopID, companyID}
	} else if category != "" {
		query = `SELECT p.id, p.name
		         FROM products p
		         WHERE p.company_id = ? AND p.is_active = 1 AND p.category = ?
		         ORDER BY p.name`
		args = []interface{}{companyID, category}
	} else {
		query = `SELECT p.id, p.name
		         FROM products p
		         WHERE p.company_id = ? AND p.is_active = 1
		         ORDER BY p.name`
		args = []interface{}{companyID}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []ProductPick{}
	for rows.Next() {
		var p ProductPick
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}
