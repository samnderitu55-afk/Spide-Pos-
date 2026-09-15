package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	db, err := sql.Open("mysql", "root:Tende@2016@tcp(127.0.0.1:3306)/spide_pos")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var hash string
	err = db.QueryRow("SELECT password FROM users WHERE username = ?", "demo_admin").Scan(&hash)
	if err != nil {
		log.Fatalf("user not found: %v", err)
	}

	fmt.Println("stored hash:", hash)
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte("admin123"))
	if err != nil {
		fmt.Println("❌ PASSWORD MISMATCH:", err)
	} else {
		fmt.Println("✅ Password matches")
	}
}
