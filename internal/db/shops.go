package db

import (
	"database/sql"
	"time"
)

type Shop struct {
	ID        int            `json:"id"`
	CompanyID int            `json:"company_id"`
	Name      string         `json:"name"`
	Location  sql.NullString `json:"location"`
	Phone     sql.NullString `json:"phone"`
	Email     sql.NullString `json:"email"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func GetShopsByCompany(db *sql.DB, companyID int) ([]Shop, error) {
	query := `
        SELECT id, company_id, name, location, phone, email, created_at, updated_at
        FROM shops
        WHERE company_id = ?
        ORDER BY name
    `
	rows, err := db.Query(query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shops := []Shop{}
	for rows.Next() {
		var s Shop
		err := rows.Scan(
			&s.ID, &s.CompanyID, &s.Name,
			&s.Location, // ✅ sql.NullString
			&s.Phone,    // ✅ sql.NullString
			&s.Email,    // ✅ sql.NullString
			&s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		shops = append(shops, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return shops, nil
}

func CreateShop(db *sql.DB, shop Shop) (int64, error) {
	query := `
        INSERT INTO shops (company_id, name, location, phone, email)
        VALUES (?, ?, ?, ?, ?)
    `
	result, err := db.Exec(query,
		shop.CompanyID, shop.Name,
		shop.Location, // ✅ sql.NullString
		shop.Phone,    // ✅ sql.NullString
		shop.Email,    // ✅ sql.NullString
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func UpdateShop(db *sql.DB, shop Shop) error {
	query := `
        UPDATE shops 
        SET name = ?, location = ?, phone = ?, email = ?, updated_at = NOW()
        WHERE id = ? AND company_id = ?
    `
	_, err := db.Exec(query,
		shop.Name,
		shop.Location, // ✅ sql.NullString
		shop.Phone,    // ✅ sql.NullString
		shop.Email,    // ✅ sql.NullString
		shop.ID, shop.CompanyID,
	)
	return err
}

func GetAllShops(db *sql.DB) ([]Shop, error) {
	query := `
        SELECT id, company_id, name, location, phone, email, created_at, updated_at
        FROM shops
        ORDER BY name
    `
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shops := []Shop{}
	for rows.Next() {
		var s Shop
		err := rows.Scan(
			&s.ID, &s.CompanyID, &s.Name,
			&s.Location, // ✅ sql.NullString
			&s.Phone,    // ✅ sql.NullString
			&s.Email,    // ✅ sql.NullString
			&s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		shops = append(shops, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return shops, nil
}

func DeleteShop(db *sql.DB, id int, companyID int) error {
	result, err := db.Exec("DELETE FROM shops WHERE id = ? AND company_id = ?", id, companyID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func GetShopByID(db *sql.DB, id int) (*Shop, error) {
	var shop Shop
	query := `
        SELECT id, company_id, name, location, phone, email, created_at, updated_at
        FROM shops
        WHERE id = ?
    `
	err := db.QueryRow(query, id).Scan(
		&shop.ID, &shop.CompanyID, &shop.Name,
		&shop.Location, // ✅ sql.NullString
		&shop.Phone,    // ✅ sql.NullString
		&shop.Email,    // ✅ sql.NullString
		&shop.CreatedAt, &shop.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &shop, nil
}
