package db

import (
    "database/sql"
)

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
    var user User
    query := `
        SELECT id, username, password, name, role, shop_id, is_active 
        FROM users 
        WHERE username = ? AND is_active = 1
    `
    err := db.QueryRow(query, username).Scan(
        &user.ID,
        &user.Username,
        &user.Password,
        &user.Name,
        &user.Role,
        &user.ShopID,
        &user.IsActive,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }

    // Get shop name if shop_id > 0
    if user.ShopID > 0 {
        var shopName string
        err := db.QueryRow("SELECT name FROM branches WHERE id = ?", user.ShopID).Scan(&shopName)
        if err == nil {
            user.ShopName = shopName
        }
    }

    return &user, nil
}

func GetUserByID(db *sql.DB, id int) (*User, error) {
    var user User
    query := `
        SELECT id, username, password, name, role, shop_id, is_active 
        FROM users 
        WHERE id = ? AND is_active = 1
    `
    err := db.QueryRow(query, id).Scan(
        &user.ID,
        &user.Username,
        &user.Password,
        &user.Name,
        &user.Role,
        &user.ShopID,
        &user.IsActive,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }

    // Get shop name if shop_id > 0
    if user.ShopID > 0 {
        var shopName string
        err := db.QueryRow("SELECT name FROM branches WHERE id = ?", user.ShopID).Scan(&shopName)
        if err == nil {
            user.ShopName = shopName
        }
    }

    return &user, nil
}

func GetAllUsers(db *sql.DB) ([]User, error) {
    query := `
        SELECT id, username, name, role, shop_id, is_active 
        FROM users 
        WHERE is_active = 1
        ORDER BY username
    `
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var user User
        err := rows.Scan(
            &user.ID,
            &user.Username,
            &user.Name,
            &user.Role,
            &user.ShopID,
            &user.IsActive,
        )
        if err != nil {
            return nil, err
        }

        // Get shop name if shop_id > 0
        if user.ShopID > 0 {
            var shopName string
            err := db.QueryRow("SELECT name FROM branches WHERE id = ?", user.ShopID).Scan(&shopName)
            if err == nil {
                user.ShopName = shopName
            }
        }

        users = append(users, user)
    }

    // Check for errors after iterating
    if err := rows.Err(); err != nil {
        return nil, err
    }

    return users, nil
}

func CreateUser(db *sql.DB, user *User) error {
    query := `
        INSERT INTO users (username, password, name, role, shop_id, is_active)
        VALUES (?, ?, ?, ?, ?, ?)
    `
    result, err := db.Exec(query, user.Username, user.Password, user.Name, user.Role, user.ShopID, user.IsActive)
    if err != nil {
        return err
    }

    id, err := result.LastInsertId()
    if err == nil {
        user.ID = int(id)
    }

    return nil
}

func UpdateUser(db *sql.DB, user *User) error {
    query := `
        UPDATE users 
        SET username = ?, name = ?, role = ?, shop_id = ?, is_active = ?
        WHERE id = ?
    `
    _, err := db.Exec(query, user.Username, user.Name, user.Role, user.ShopID, user.IsActive, user.ID)
    return err
}

func DeleteUser(db *sql.DB, id int) error {
    query := `UPDATE users SET is_active = 0 WHERE id = ?`
    _, err := db.Exec(query, id)
    return err
}

func UpdateUserPassword(db *sql.DB, id int, hashedPassword string) error {
    query := `UPDATE users SET password = ? WHERE id = ?`
    _, err := db.Exec(query, hashedPassword, id)
    return err
}

func UpdateLastLogin(db *sql.DB, userID int) error {
    query := `UPDATE users SET last_login = NOW() WHERE id = ?`
    _, err := db.Exec(query, userID)
    return err
}
