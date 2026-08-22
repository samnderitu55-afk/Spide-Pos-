package db

import (
    "database/sql"
    "fmt"
    "time"
)

func CreateExpense(db *sql.DB, e *Expense) (int64, error) {
    query := `
        INSERT INTO expenses (category, description, amount, expense_date, 
                             payment_method, reference, notes, created_by, shop_id, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
    `
    result, err := db.Exec(query,
        e.Category, e.Description, e.Amount, e.ExpenseDate,
        e.PaymentMethod, e.Reference, e.Notes, e.CreatedBy, e.ShopID,
    )
    if err != nil {
        return 0, fmt.Errorf("failed to create expense: %w", err)
    }
    return result.LastInsertId()
}

func GetExpensesByDateRange(db *sql.DB, startDate, endDate string, shopID int) ([]Expense, error) {
    query := `
        SELECT id, category, description, amount, expense_date, 
               payment_method, reference, notes, created_by, shop_id, created_at
        FROM expenses
        WHERE DATE(expense_date) BETWEEN ? AND ? AND shop_id = ?
        ORDER BY expense_date DESC, id DESC
    `
    rows, err := db.Query(query, startDate, endDate, shopID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var expenses []Expense
    for rows.Next() {
        var e Expense
        err := rows.Scan(
            &e.ID, &e.Category, &e.Description, &e.Amount,
            &e.ExpenseDate, &e.PaymentMethod, &e.Reference,
            &e.Notes, &e.CreatedBy, &e.ShopID, &e.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        expenses = append(expenses, e)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating expenses: %w", err)
    }

    return expenses, nil
}

type ExpenseReport struct {
    TotalExpenses     float64            `json:"total_expenses"`
    TotalCount        int                `json:"total_count"`
    AveragePerDay     float64            `json:"average_per_day"`
    CategoryBreakdown map[string]float64 `json:"category_breakdown"`
    DailyBreakdown    []struct {
        Date          string  `json:"date"`
        TotalExpenses float64 `json:"total_expenses"`
        Count         int     `json:"count"`
    } `json:"daily_breakdown"`
    RecentExpenses   []Expense `json:"recent_expenses"`
}

func GetExpenseReport(db *sql.DB, startDate, endDate string, shopID int) (*ExpenseReport, error) {
    if startDate == "" {
        startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
    }
    if endDate == "" {
        endDate = time.Now().Format("2006-01-02")
    }

    report := &ExpenseReport{
        CategoryBreakdown: make(map[string]float64),
    }

    query := `
        SELECT COALESCE(SUM(amount), 0), COUNT(*)
        FROM expenses
        WHERE DATE(expense_date) BETWEEN ? AND ? AND shop_id = ?
    `
    err := db.QueryRow(query, startDate, endDate, shopID).Scan(&report.TotalExpenses, &report.TotalCount)
    if err != nil {
        return nil, fmt.Errorf("failed to get total expenses: %w", err)
    }

    catQuery := `
        SELECT category, COALESCE(SUM(amount), 0)
        FROM expenses
        WHERE DATE(expense_date) BETWEEN ? AND ? AND shop_id = ?
        GROUP BY category
        ORDER BY SUM(amount) DESC
    `
    rows, err := db.Query(catQuery, startDate, endDate, shopID)
    if err != nil {
        return nil, fmt.Errorf("failed to get category breakdown: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var category string
        var amount float64
        if err := rows.Scan(&category, &amount); err != nil {
            continue
        }
        report.CategoryBreakdown[category] = amount
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating category breakdown: %w", err)
    }

    days := getDaysBetween(startDate, endDate)
    if days > 0 {
        report.AveragePerDay = report.TotalExpenses / float64(days)
    }

    return report, nil
}

func GetExpenseCategories(db *sql.DB, shopID int) ([]string, error) {
    query := `SELECT DISTINCT category FROM expenses WHERE shop_id = ? ORDER BY category ASC`
    rows, err := db.Query(query, shopID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var categories []string
    for rows.Next() {
        var cat string
        if err := rows.Scan(&cat); err != nil {
            continue
        }
        categories = append(categories, cat)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating categories: %w", err)
    }

    return categories, nil
}

func UpdateExpense(db *sql.DB, e *Expense) error {
    query := `
        UPDATE expenses 
        SET category = ?, description = ?, amount = ?, expense_date = ?,
            payment_method = ?, reference = ?, notes = ?, shop_id = ?
        WHERE id = ?
    `
    _, err := db.Exec(query,
        e.Category, e.Description, e.Amount, e.ExpenseDate,
        e.PaymentMethod, e.Reference, e.Notes, e.ShopID, e.ID,
    )
    return err
}

func DeleteExpense(db *sql.DB, id int) error {
    _, err := db.Exec("DELETE FROM expenses WHERE id = ?", id)
    return err
}

func getDaysBetween(start, end string) int {
    t1, _ := time.Parse("2006-01-02", start)
    t2, _ := time.Parse("2006-01-02", end)
    diff := t2.Sub(t1)
    return int(diff.Hours()/24) + 1
}


