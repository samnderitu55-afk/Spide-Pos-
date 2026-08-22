package db

import (
    "database/sql"
    "fmt"
    "spide-pos/internal/auth"
)

// User struct is already defined in models.go, so we don't redeclare it here

func CreateUser(db *sql.DB, username, password, name, role string, shopID int) (*User, error) {
    hashedPassword, err := auth.HashPassword(password)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }

    query := `
        INSERT INTO users (username, password, name, role, shop_id, is_active, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, 1, NOW(), NOW())
    `
    result, err := db.Exec(query, username, hashedPassword, name, role, shopID)
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    id, err := result.LastInsertId()
    if err != nil {
        return nil, fmt.Errorf("failed to get user ID: %w", err)
    }

    return GetUserByID(db, int(id))
}

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

func GetAllUsers(db *sql.DB) ([]User, error) {
    query := `
        SELECT u.id, u.username, u.name, u.role, u.shop_id, b.name as shop_name, u.is_active
        FROM users u
        LEFT JOIN branches b ON u.shop_id = b.id
        WHERE u.is_active = 1
        ORDER BY u.name ASC
    `
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var u User
        err := rows.Scan(
            &u.ID, &u.Username, &u.Name, &u.Role,
            &u.ShopID, &u.ShopName, &u.IsActive,
        )
        if err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    return users, nil
}

func UpdateUser(db *sql.DB, id int, name, role string, shopID int) error {
    query := `
        UPDATE users 
        SET name = ?, role = ?, shop_id = ?, updated_at = NOW()
        WHERE id = ?
    `
    _, err := db.Exec(query, name, role, shopID, id)
    return err
}

func UpdateUserPassword(db *sql.DB, id int, newPassword string) error {
    hashedPassword, err := auth.HashPassword(newPassword)
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }

    query := `
        UPDATE users 
        SET password = ?, updated_at = NOW()
        WHERE id = ?
    `
    _, err = db.Exec(query, hashedPassword, id)
    return err
}

func DeleteUser(db *sql.DB, id int) error {
    query := `UPDATE users SET is_active = 0, updated_at = NOW() WHERE id = ?`
    _, err := db.Exec(query, id)
    return err
}

func UpdateLastLogin(db *sql.DB, userID int) error {
    query := `UPDATE users SET last_login = NOW() WHERE id = ?`
    _, err := db.Exec(query, userID)
    return err
}
