package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	CompanyID int       `json:"company_id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	ShopID    int       `json:"shop_id"`
	ShopName  string    `json:"shop_name,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastLogin time.Time `json:"last_login,omitempty"`
}

func GetUsersByCompany(db *sql.DB, companyID int) ([]User, error) {
	query := `
        SELECT u.id, u.company_id, u.username, COALESCE(u.email, '') as email,
               u.full_name, u.role, COALESCE(u.shop_id, 0) as shop_id,
               COALESCE(s.name, '') as shop_name, u.is_active, u.created_at, u.updated_at
        FROM users u
        LEFT JOIN shops s ON u.shop_id = s.id
        WHERE u.company_id = ?
        ORDER BY u.username
    `
	rows, err := db.Query(query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		err := rows.Scan(
			&u.ID, &u.CompanyID, &u.Username, &u.Email, &u.FullName,
			&u.Role, &u.ShopID, &u.ShopName, &u.IsActive,
			&u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func GetAllUsers(db *sql.DB) ([]User, error) {
	query := `
        SELECT u.id, u.company_id, u.username, COALESCE(u.email, '') as email,
               u.full_name, u.role, COALESCE(u.shop_id, 0) as shop_id,
               COALESCE(s.name, '') as shop_name, u.is_active, u.created_at, u.updated_at
        FROM users u
        LEFT JOIN shops s ON u.shop_id = s.id
        ORDER BY u.username
    `
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		err := rows.Scan(
			&u.ID, &u.CompanyID, &u.Username, &u.Email,
			&u.FullName, &u.Role, &u.ShopID, &u.ShopName,
			&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			log.Printf("❌ Error scanning user row: %v", err)
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func CreateUser(db *sql.DB, username, password, email, fullName, role string, shopID, companyID int) error {
	var emailValue interface{}
	if email == "" {
		emailValue = nil
	} else {
		emailValue = email
	}

	var shopIDValue interface{}
	if shopID == 0 {
		shopIDValue = nil
	} else {
		shopIDValue = shopID
	}

	query := `
        INSERT INTO users (username, password, email, full_name, role, shop_id, company_id, is_active)
        VALUES (?, ?, ?, ?, ?, ?, ?, 1)
    `
	_, err := db.Exec(query, username, password, emailValue, fullName, role, shopIDValue, companyID)
	return err
}

func UpdateUser(db *sql.DB, id int, username, password, email, fullName, role string, shopID, companyID int, isActive bool) error {
	var emailValue interface{}
	if email == "" {
		emailValue = nil
	} else {
		emailValue = email
	}

	var shopIDValue interface{}
	if shopID == 0 {
		shopIDValue = nil
	} else {
		shopIDValue = shopID
	}

	var query string
	var args []interface{}

	if password != "" {
		query = `
            UPDATE users 
            SET username = ?, password = ?, email = ?, full_name = ?, 
                role = ?, shop_id = ?, is_active = ?, updated_at = NOW()
            WHERE id = ? AND company_id = ?
        `
		args = []interface{}{username, password, emailValue, fullName, role, shopIDValue, isActive, id, companyID}
	} else {
		query = `
            UPDATE users 
            SET username = ?, email = ?, full_name = ?, 
                role = ?, shop_id = ?, is_active = ?, updated_at = NOW()
            WHERE id = ? AND company_id = ?
        `
		args = []interface{}{username, emailValue, fullName, role, shopIDValue, isActive, id, companyID}
	}

	result, err := db.Exec(query, args...)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func DeleteUser(db *sql.DB, id, companyID int) error {
	result, err := db.Exec("DELETE FROM users WHERE id = ? AND company_id = ?", id, companyID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	query := `
        SELECT id, username, password, COALESCE(email, '') as email,
               full_name, role, COALESCE(shop_id, 0) as shop_id,
               company_id, is_active, created_at, updated_at
        FROM users
        WHERE username = ?
    `
	var u User
	var password string
	err := db.QueryRow(query, username).Scan(
		&u.ID, &u.Username, &password, &u.Email,
		&u.FullName, &u.Role, &u.ShopID, &u.CompanyID,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	u.Password = password
	return &u, nil
}

// ✅ GetUserByID — returns full record including inactive users, with company check
func GetUserByID(db *sql.DB, id int) (*User, error) {
	query := `
        SELECT id, company_id, username, COALESCE(email, '') as email,
               full_name, role, COALESCE(shop_id, 0) as shop_id,
               is_active, created_at, updated_at
        FROM users
        WHERE id = ?
    `
	var u User
	err := db.QueryRow(query, id).Scan(
		&u.ID, &u.CompanyID, &u.Username, &u.Email,
		&u.FullName, &u.Role, &u.ShopID,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	if u.ShopID > 0 {
		var shopName string
		err := db.QueryRow("SELECT name FROM shops WHERE id = ?", u.ShopID).Scan(&shopName)
		if err == nil {
			u.ShopName = shopName
		}
	}

	return &u, nil
}

func UpdateUserPassword(db *sql.DB, id, companyID int, hashedPassword string) error {
	_, err := db.Exec(
		"UPDATE users SET password = ? WHERE id = ? AND company_id = ?",
		hashedPassword, id, companyID,
	)
	return err
}

func UpdateLastLogin(db *sql.DB, userID int) error {
	_, err := db.Exec("UPDATE users SET last_login = NOW() WHERE id = ?", userID)
	return err
}
