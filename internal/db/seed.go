package db

import (
    "database/sql"
    "log"
)

func SeedUsers(db *sql.DB) error {
    // Check if admin exists
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
    if err != nil {
        return err
    }

    if count == 0 {
        log.Println("Seeding default users...")
        
        // Create default users
        users := []struct {
            username string
            password string
            name     string
            role     string
            shopID   int
        }{
            {"admin", "admin123", "Administrator", "director", 1},
            {"cashier1", "admin123", "Cashier One", "cashier", 1},
            {"cashier2", "admin123", "Cashier Two", "cashier", 2},
            {"manager1", "admin123", "Manager One", "manager", 1},
        }

        for _, u := range users {
            _, err := db.Exec(`
                INSERT INTO users (username, password, name, role, shop_id, is_active)
                VALUES (?, ?, ?, ?, ?, 1)
            `, u.username, u.password, u.name, u.role, u.shopID)
            if err != nil {
                log.Printf("Error creating user %s: %v", u.username, err)
            }
        }
        log.Println("✅ Default users created")
    }

    return nil
}
