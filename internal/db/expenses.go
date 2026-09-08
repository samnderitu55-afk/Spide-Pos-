package db

import (
	"database/sql"
	"fmt"
)

func CreateExpense(db *sql.DB, expense *Expense) error {
	query := `
        INSERT INTO expenses (category, description, amount, expense_date, 
                              payment_method, reference, notes, shop_id, created_by, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
    `
	_, err := db.Exec(query,
		expense.Category,
		expense.Description,
		expense.Amount,
		expense.ExpenseDate,
		expense.PaymentMethod,
		expense.Reference,
		expense.Notes,
		expense.ShopID,
		expense.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to create expense: %w", err)
	}

	return nil
}

func GetExpenseReport(db *sql.DB, startDate, endDate string, shopID int) ([]Expense, float64, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "DATE(created_at) BETWEEN ? AND ?")
	args = append(args, startDate, endDate)

	if shopID > 0 {
		conditions = append(conditions, "shop_id = ?")
		args = append(args, shopID)
	}

	whereClause := ""
	for i, cond := range conditions {
		if i == 0 {
			whereClause += " WHERE " + cond
		} else {
			whereClause += " AND " + cond
		}
	}

	// Get expenses
	query := `
        SELECT id, category, description, amount, expense_date, 
               payment_method, reference, notes, shop_id, created_by, created_at
        FROM expenses ` + whereClause + `
        ORDER BY created_at DESC
    `

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get expenses: %w", err)
	}
	defer rows.Close()

	var expenses []Expense
	for rows.Next() {
		var e Expense
		err := rows.Scan(
			&e.ID,
			&e.Category,
			&e.Description,
			&e.Amount,
			&e.ExpenseDate,
			&e.PaymentMethod,
			&e.Reference,
			&e.Notes,
			&e.ShopID,
			&e.CreatedBy,
			&e.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// Get shop name
		if e.ShopID > 0 {
			var shopName string
			db.QueryRow("SELECT name FROM shops WHERE id = ?", e.ShopID).Scan(&shopName)
			e.ShopName = shopName
		}

		expenses = append(expenses, e)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating expenses: %w", err)
	}

	// Get total
	totalQuery := `SELECT COALESCE(SUM(amount), 0) FROM expenses ` + whereClause
	var total float64
	err = db.QueryRow(totalQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get expense total: %w", err)
	}

	return expenses, total, nil
}

func GetExpenseCategories(db *sql.DB, shopID int) ([]string, error) {
	var conditions []string
	var args []interface{}

	if shopID > 0 {
		conditions = append(conditions, "shop_id = ?")
		args = append(args, shopID)
	}

	whereClause := ""
	for i, cond := range conditions {
		if i == 0 {
			whereClause += " WHERE " + cond
		} else {
			whereClause += " AND " + cond
		}
	}

	query := `
        SELECT DISTINCT category 
        FROM expenses 
        WHERE category != '' ` + whereClause + `
        ORDER BY category ASC
    `

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense categories: %w", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var cat string
		err := rows.Scan(&cat)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	return categories, nil
}

func DeleteExpense(db *sql.DB, id int) error {
	query := `DELETE FROM expenses WHERE id = ?`
	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("expense not found")
	}

	return nil
}
