package db

import (
	"database/sql"
	"fmt"
)

type InventoryValuation struct {
	ShopID           int                 `json:"branch_id"`
	ShopName         string              `json:"branch_name"`
	TotalItems       int                 `json:"total_items"`
	TotalQuantity    int                 `json:"total_quantity"`
	TotalCostValue   float64             `json:"total_cost_value"`
	TotalRetailValue float64             `json:"total_retail_value"`
	PotentialProfit  float64             `json:"potential_profit"`
	Categories       []CategoryValuation `json:"categories"`
	Products         []ProductValuation  `json:"products"`
}

type CategoryValuation struct {
	Category         string  `json:"category"`
	ItemCount        int     `json:"item_count"`
	TotalQuantity    int     `json:"total_quantity"`
	TotalCostValue   float64 `json:"total_cost_value"`
	TotalRetailValue float64 `json:"total_retail_value"`
	PotentialProfit  float64 `json:"potential_profit"`
}

type ProductValuation struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Barcode     string  `json:"barcode"`
	Category    string  `json:"category"`
	Quantity    int     `json:"quantity"`
	CostPrice   float64 `json:"cost_price"`
	RetailPrice float64 `json:"retail_price"`
	CostValue   float64 `json:"cost_value"`
	RetailValue float64 `json:"retail_value"`
	Profit      float64 `json:"profit"`
}

func GetInventoryValuation(db *sql.DB, companyID int, shopID int) (*InventoryValuation, error) {
	valuation := &InventoryValuation{
		Categories: []CategoryValuation{},
		Products:   []ProductValuation{},
	}

	shopFilter := ""
	args := []interface{}{companyID}

	if shopID > 0 {
		shopFilter = " AND ss.shop_id = ?"
		args = append(args, shopID)
	}

	// ✅ Get product valuation
	query := `
        SELECT p.name, p.barcode, p.category, 
               COALESCE(ss.quantity, 0) as quantity,
               p.cost_price, p.retail_price,
               (COALESCE(ss.quantity, 0) * p.cost_price) as cost_value,
               (COALESCE(ss.quantity, 0) * p.retail_price) as retail_value,
               (COALESCE(ss.quantity, 0) * (p.retail_price - p.cost_price)) as profit
        FROM products p
        LEFT JOIN shop_stock ss ON p.id = ss.product_id AND ss.company_id = ?
        WHERE p.company_id = ?
    `
	// Add shop filter if specified
	query += shopFilter
	query += " ORDER BY p.category, p.name"

	rows, err := db.Query(query, companyID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totalQuantity int
	var totalCostValue float64
	var totalRetailValue float64
	var potentialProfit float64
	categoryMap := make(map[string]*CategoryValuation)

	for rows.Next() {
		var p ProductValuation
		err := rows.Scan(
			&p.ProductName, &p.Barcode, &p.Category,
			&p.Quantity, &p.CostPrice, &p.RetailPrice,
			&p.CostValue, &p.RetailValue, &p.Profit,
		)
		if err != nil {
			return nil, err
		}

		if p.Quantity > 0 {
			valuation.Products = append(valuation.Products, p)
			totalQuantity += p.Quantity
			totalCostValue += p.CostValue
			totalRetailValue += p.RetailValue
			potentialProfit += p.Profit

			// Add to category
			if _, exists := categoryMap[p.Category]; !exists {
				categoryMap[p.Category] = &CategoryValuation{
					Category: p.Category,
				}
			}
			categoryMap[p.Category].ItemCount++
			categoryMap[p.Category].TotalCostValue += p.CostValue
		}
		if err = rows.Err(); err != nil {
			return nil, err
		}
	}

	valuation.TotalItems = len(valuation.Products)
	valuation.TotalQuantity = totalQuantity
	valuation.TotalCostValue = totalCostValue
	valuation.TotalRetailValue = totalRetailValue
	valuation.PotentialProfit = potentialProfit

	// Convert category map to slice
	for _, cat := range categoryMap {
		valuation.Categories = append(valuation.Categories, *cat)
	}

	return valuation, nil
}

func GetProductValuation(db *sql.DB, branchID int) ([]ProductValuation, error) {
	var branchCondition string
	var args []interface{}
	if branchID > 0 {
		branchCondition = "AND ss.shop_id = ?"
		args = append(args, branchID)
	}

	query := fmt.Sprintf(`
        SELECT 
            p.id,
            p.name,
            p.barcode,
            COALESCE(p.category, 'Uncategorized') as category,
            COALESCE(ss.quantity, 0) as quantity,
            p.cost_price,
            p.retail_price,
            COALESCE(ss.quantity * p.cost_price, 0) as cost_value,
            COALESCE(ss.quantity * p.retail_price, 0) as retail_value
        FROM products p
        JOIN shop_stock ss ON p.id = ss.product_id
        WHERE p.is_active = 1 AND COALESCE(ss.quantity, 0) > 0 %s
        ORDER BY cost_value DESC
    `, branchCondition)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get product valuation: %w", err)
	}
	defer rows.Close()

	products := []ProductValuation{}
	for rows.Next() {
		var p ProductValuation
		err := rows.Scan(
			&p.ProductID,
			&p.ProductName,
			&p.Barcode,
			&p.Category,
			&p.Quantity,
			&p.CostPrice,
			&p.RetailPrice,
			&p.CostValue,
			&p.RetailValue,
		)
		if err != nil {
			return nil, err
		}
		p.Profit = p.RetailValue - p.CostValue
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}
