package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	users := []struct {
		Username string
		Password string
	}{
		{"demo_admin", "admin123"},
		{"demo_manager", "manager123"},
		{"demo_cashier", "cashier123"},
		{"demo_cashier2", "cashier123"},
	}

	for _, u := range users {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("hash failed for %s: %v", u.Username, err)
		}
		fmt.Printf("%-15s %s\n", u.Username, string(hash))
	}
}
