package db

import (
    "database/sql"
    "fmt"
)

// GetUserByUsername retrieves a user by username
func GetUserByUsername(db *sql.DB, username string) (*User, error) {
    query := `
        SELECT u.id, u.username, u.password, u.name, u.role, u.shop_id, b.name as shop_name, u.is_active
        FROM users u
        LEFT JOIN branches b ON u.shop_id = b.id
        WHERE u.username = ? AND u.is_active = 1
    `
    var user User
    err := db.QueryRow(query, username).Scan(
        &user.ID, &user.Username, &user.Password, &user.Name,
        &user.Role, &user.ShopID, &user.ShopName, &user.IsActive,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("user not found")
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return &user, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(db *sql.DB, id int) (*User, error) {
    query := `
        SELECT u.id, u.username, u.password, u.name, u.role, u.shop_id, b.name as shop_name, u.is_active
        FROM users u
        LEFT JOIN branches b ON u.shop_id = b.id
        WHERE u.id = ? AND u.is_active = 1
    `
    var user User
    err := db.QueryRow(query, id).Scan(
        &user.ID, &user.Username, &user.Password, &user.Name,
        &user.Role, &user.ShopID, &user.ShopName, &user.IsActive,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("user not found")
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return &user, nil
}

// UpdateLastLogin updates the last login time for a user
func UpdateLastLogin(db *sql.DB, userID int) error {
    query := `UPDATE users SET last_login = NOW() WHERE id = ?`
    _, err := db.Exec(query, userID)
    return err
}


