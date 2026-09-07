package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"spide-pos/internal/db"
	"spide-pos/internal/middleware"
	"strconv"
)

func GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// ✅ Get shop_id from query parameter
	shopIDParam := r.URL.Query().Get("shop_id")
	var shopID int

	if shopIDParam != "" {
		id, err := strconv.Atoi(shopIDParam)
		if err == nil && id > 0 {
			shopID = id
		}
	}

	// If no shop_id provided, use user's assigned shop
	if shopID == 0 {
		shopID = claims.ShopID
	}

	// ✅ Verify the shop belongs to the user's company
	if shopID > 0 {
		shop, err := db.GetShopByID(db.GetDB(), shopID)
		if err != nil || shop == nil || shop.CompanyID != claims.CompanyID {
			http.Error(w, "Forbidden: Invalid shop for this company", http.StatusForbidden)
			return
		}
	}

	// ✅ Get products filtered by shop and company
	products, err := db.GetProductsByShopAndCompany(db.GetDB(), shopID, claims.CompanyID)
	if err != nil {
		log.Printf("❌ Error getting products: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var product db.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	id, err := db.CreateProduct(db.GetDB(), &product)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	product.ID = id
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPut {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var product db.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if err := db.UpdateProduct(db.GetDB(), &product); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(product)
}

func SearchProductsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, `{"error":"Search query required"}`, http.StatusBadRequest)
		return
	}

	// Get shop_id from query param, default to user's shop
	shopID := claims.ShopID
	if shopIDParam := r.URL.Query().Get("shop_id"); shopIDParam != "" {
		if id, err := strconv.Atoi(shopIDParam); err == nil && id > 0 {
			shopID = id
		}
	}

	products, err := db.SearchProductsByShop(db.GetDB(), query, shopID)
	if err != nil {
		http.Error(w, `{"error":"Failed to search products: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(products)
}

func ScanProductHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, `Unauthorized`, http.StatusUnauthorized)
		return
	}

	barcode := r.URL.Query().Get("barcode")
	if barcode == "" {
		http.Error(w, `<div class="text-red-500 text-center p-2">❌ Barcode required</div>`, http.StatusBadRequest)
		return
	}

	qty := 1
	if qtyParam := r.URL.Query().Get("qty"); qtyParam != "" {
		if q, err := strconv.Atoi(qtyParam); err == nil && q > 0 {
			qty = q
		}
	}

	shopID := claims.ShopID
	if shopIDParam := r.URL.Query().Get("shop_id"); shopIDParam != "" {
		if id, err := strconv.Atoi(shopIDParam); err == nil && id > 0 {
			shopID = id
		}
	}

	log.Printf("🔍 Scanning barcode: %s, shop: %d, qty: %d", barcode, shopID, qty)

	product, err := db.GetProductByBarcode(db.GetDB(), barcode, shopID)
	if err != nil {
		log.Printf("❌ Database error: %v", err)
		http.Error(w, `<div class="text-red-500 text-center p-2">❌ Product not found</div>`, http.StatusNotFound)
		return
	}

	if product == nil {
		log.Printf("❌ Product not found for barcode: %s", barcode)
		http.Error(w, `<div class="text-red-500 text-center p-2">❌ Product not found</div>`, http.StatusNotFound)
		return
	}

	log.Printf("✅ Product found: %s, stock: %d", product.Name, product.StockQuantity)

	// Check if stock is available
	if product.StockQuantity <= 0 {
		http.Error(w, `<div class="text-red-500 text-center p-2 font-bold">❌ Out of Stock!</div>`, http.StatusBadRequest)
		return
	}

	if product.StockQuantity < qty {
		msg := fmt.Sprintf(`<div class="text-amber-500 text-center p-2 font-bold">⚠️ Only %d items available</div>`, product.StockQuantity)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Render HTML row for the product
	w.Write([]byte(product.RenderRowHTML()))
}
