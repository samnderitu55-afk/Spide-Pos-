// internal/db/categories.go

package db

import (
	"database/sql"
	"strings"
)

func GetOrCreateCategory(db *sql.DB, name string, companyID int) (int, error) {
	if name == "" {
		name = "General"
	}

	name = strings.TrimSpace(name)

	// ✅ Check if category exists for this company
	var id int
	err := db.QueryRow(`
        SELECT id FROM categories 
        WHERE name = ? AND company_id = ?
    `, name, companyID).Scan(&id)

	if err == nil {
		return id, nil
	}

	if err != sql.ErrNoRows {
		return 0, err
	}

	// ✅ Create new category WITHOUT is_active
	result, err := db.Exec(`
        INSERT INTO categories (name, company_id, created_at,updated_at)
        VALUES (?, ?, NOW()NOW())
    `, name, companyID)
	if err != nil {
		return 0, err
	}

	id64, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id64), nil
}
