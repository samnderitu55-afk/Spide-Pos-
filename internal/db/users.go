package db

import (
	"database/sql"
	"log"
	"time"
)

type User struct {
    ID          int       `json:"id"`
    Username    string    `json:"username"`
    Password    string    `json:"-"` // ← Hidden from JSON responses
    Email       sql.NullString `json:"email"`
    FullName    string    `json:"full_name"`
    Role        string    `json:"role"`
    ShopID      int       `json:"shop_id"`
    ShopName    string    `json:"shop_name,omitempty"`
    IsActive    bool      `json:"is_active"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    LastLogin   time.Time `json:"last_login,omitempty"`
}

func GetAllUsers(db *sql.DB) ([]User, error) {
    query := `
        SELECT u.id, u.username, u.email, u.full_name, u.role, u.shop_id, 
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

    var users []User
    for rows.Next() {
        var u User
        // ✅ Use sql.NullString for email (already defined in User struct)
        err := rows.Scan(
            &u.ID, 
            &u.Username, 
            &u.Email,      // ✅ sql.NullString handles NULL
            &u.FullName, 
            &u.Role, 
            &u.ShopID, 
            &u.ShopName, 
            &u.IsActive, 
            &u.CreatedAt, 
            &u.UpdatedAt,
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

func CreateUser(db *sql.DB, username, password, email, fullName, role string, shopID int) error {
    query := `
        INSERT INTO users (username, password, email, full_name, role, shop_id, is_active)
        VALUES (?, ?, ?, ?, ?, ?, 1)
    `
    // ✅ Convert empty string to NULL for email
    var emailValue interface{}
    if email == "" {
        emailValue = nil
    } else {
        emailValue = email
    }
    
    _, err := db.Exec(query, username, password, emailValue, fullName, role, shopID)
    return err
}

func UpdateUser(db *sql.DB, id int, username, password, email, fullName, role string, shopID int, isActive bool) error {
   // ✅ Convert empty string to NULL for email
    var emailValue interface{}
    if email == "" {
        emailValue = nil
    } else {
        emailValue = email
    }
   
    var query string
    var args []interface{}

    if password != "" {
        query = `
            UPDATE users 
            SET username = ?, password = ?, email = ?, full_name = ?, 
                role = ?, shop_id = ?, is_active = ?, updated_at = NOW()
            WHERE id = ?
        `
        args = []interface{}{username, password, emailValue, fullName, role, shopID, isActive, id}
    } else {
        query = `
            UPDATE users 
            SET username = ?, email = ?, full_name = ?, 
                role = ?, shop_id = ?, is_active = ?, updated_at = NOW()
            WHERE id = ?
        `
        args = []interface{}{username, emailValue, fullName, role, shopID, isActive, id}
    }

    _, err := db.Exec(query, args...)
    return err
}

func DeleteUser(db *sql.DB, id int) error {
    _, err := db.Exec("DELETE FROM users WHERE id = ?", id)
    return err
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
    query := `
        SELECT id, username, password, email, full_name, role, shop_id, is_active, created_at, updated_at
        FROM users
        WHERE username = ?
    `
    var u User
    var password string
    err := db.QueryRow(query, username).Scan(
        &u.ID, &u.Username, &password, &u.Email,
        &u.FullName, &u.Role, &u.ShopID, &u.IsActive,
        &u.CreatedAt, &u.UpdatedAt,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil // User not found
        }
        return nil, err
    }
    u.Password = password // Store the password hash for verification
    return &u, nil

  
}

func GetUserByID(db *sql.DB, id int) (*User, error) {
    var user User
    query := `
        SELECT id, username, password, full_name, role, shop_id, is_active 
        FROM users 
        WHERE id = ? AND is_active = 1
    `
    err := db.QueryRow(query, id).Scan(
        &user.ID,
        &user.Username,
        &user.Password,
        &user.FullName,
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
