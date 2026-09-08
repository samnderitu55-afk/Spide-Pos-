package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"spide-pos/internal/db"
	"spide-pos/internal/middleware"
	"strconv"
)

// ============================================
// GET SHOPS
// ============================================

func GetShopsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// ✅ Get company ID from claims
	companyID := claims.CompanyID
	if companyID == 0 {
		companyID = 1 // Default fallback
	}

	// ✅ Only get shops for the user's company
	shops, err := db.GetShopsByCompany(db.GetDB(), companyID)
	if err != nil {
		log.Printf("❌ Error getting shops: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Check if user should only see their assigned shop
	if claims.Role == "cashier" || claims.Role == "manager" {
		if claims.ShopID > 0 {
			// Filter to only their shop
			var filteredShops []db.Shop
			for _, shop := range shops {
				if shop.ID == claims.ShopID {
					filteredShops = append(filteredShops, shop)
					break
				}
			}
			shops = filteredShops
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shops)
}

// ============================================
// CREATE SHOP
// ============================================

func CreateShopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admin and director can create shops
	if claims.Role != "admin" && claims.Role != "director" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var shop db.Shop
	if err := json.NewDecoder(r.Body).Decode(&shop); err != nil {
		log.Printf("❌ Error decoding shop: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if shop.Name == "" {
		http.Error(w, "Shop name is required", http.StatusBadRequest)
		return
	}

	// Use the user's company ID if not provided
	if shop.CompanyID == 0 {
		shop.CompanyID = claims.CompanyID
		if shop.CompanyID == 0 {
			shop.CompanyID = 1 // Default to company 1
		}
	}

	log.Printf("📝 Creating shop: %s, Company: %d", shop.Name, shop.CompanyID)

	id, err := db.CreateShop(db.GetDB(), shop)
	if err != nil {
		log.Printf("❌ Error creating shop: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle empty strings as NULL
	if shop.Location.String == "" {
		shop.Location.Valid = false
	}
	if shop.Phone.String == "" {
		shop.Phone.Valid = false
	}
	if shop.Email.String == "" {
		shop.Email.Valid = false
	}

	shop.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Shop created successfully",
		"shop":    shop,
	})
}

// ============================================
// UPDATE SHOP
// ============================================

func UpdateShopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" && claims.Role != "director" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var shop db.Shop
	if err := json.NewDecoder(r.Body).Decode(&shop); err != nil {
		log.Printf("❌ Error decoding shop: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if shop.ID == 0 {
		http.Error(w, "Shop ID is required", http.StatusBadRequest)
		return
	}

	if shop.Name == "" {
		http.Error(w, "Shop name is required", http.StatusBadRequest)
		return
	}

	shop.CompanyID = claims.CompanyID
	if shop.CompanyID == 0 {
		shop.CompanyID = 1
	}

	err := db.UpdateShop(db.GetDB(), shop)
	if err != nil {
		log.Printf("❌ Error updating shop: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Shop updated successfully",
	})
}

// ============================================
// DELETE SHOP
// ============================================

func DeleteShopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" && claims.Role != "director" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	shopID := r.URL.Query().Get("id")
	if shopID == "" {
		http.Error(w, "Shop ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(shopID)
	if err != nil {
		http.Error(w, "Invalid shop ID", http.StatusBadRequest)
		return
	}

	companyID := claims.CompanyID
	if companyID == 0 {
		companyID = 1
	}

	err = db.DeleteShop(db.GetDB(), id, companyID)
	if err != nil {
		log.Printf("❌ Error deleting shop: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Shop deleted successfully",
	})
}

// ============================================
// BRANCHES HANDLER (Alias for GetShops)
// ============================================

func ShopsHandler(w http.ResponseWriter, r *http.Request) {
	GetShopsHandler(w, r)
}

// ============================================
// GET SHOPS BY COMPANY
// ============================================

func GetShopsByCompanyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	companyID := claims.CompanyID
	if companyID == 0 {
		companyID = 1
	}

	shops, err := db.GetShopsByCompany(db.GetDB(), companyID)
	if err != nil {
		log.Printf("❌ Error getting shops by company: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shops)
}
