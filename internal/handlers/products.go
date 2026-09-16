package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"spide-pos/internal/db"
	"spide-pos/internal/middleware"
	"strconv"
)

// GetProductsHandler - Get all products for the user's company with stock
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

	// ✅ Get company ID from claims
	companyID := claims.CompanyID
	if companyID == 0 {
		companyID = 1
	}

	// ✅ Get shop_id from query parameter (for stock info)
	shopIDParam := r.URL.Query().Get("shop_id")
	var shopID int
	if shopIDParam != "" {
		id, err := strconv.Atoi(shopIDParam)
		if err == nil && id > 0 {
			shopID = id
		}
	}
	if shopID == 0 {
		shopID = claims.ShopID
	}
	if shopID == 0 {
		shopID = 1
	}

	// ✅ Get products for this company with stock for the shop
	products, err := db.GetProductsByCompanyWithStock(db.GetDB(), companyID, shopID)
	if err != nil {
		log.Printf("❌ Error getting products: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ If no products, return empty array
	if products == nil {
		products = []db.Product{}
	}

	log.Printf("📦 Retrieved %d products for company %d, shop %d", len(products), companyID, shopID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var product db.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// ✅ Always override with the authenticated user's company
	product.CompanyID = claims.CompanyID
	if product.CompanyID == 0 {
		http.Error(w, `{"error":"User has no company assigned"}`, http.StatusBadRequest)
		return
	}

	// ✅ Resolve category name → category_id
	if product.Category != "" {
		catID, err := db.GetCategoryIDByName(db.GetDB(), product.Category, product.CompanyID)
		if err != nil {
			log.Printf("⚠️  Could not resolve category %q: %v", product.Category, err)
			// Don't fail the whole request — just leave category_id NULL
			product.CategoryID = 0
		} else {
			product.CategoryID = catID
		}
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

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var product db.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Always use the authenticated user's company
	product.CompanyID = claims.CompanyID
	if product.CompanyID == 0 {
		http.Error(w, `{"error":"User has no company assigned"}`, http.StatusBadRequest)
		return
	}

	// Resolve category name → category_id
	if product.Category != "" {
		catID, err := db.GetCategoryIDByName(db.GetDB(), product.Category, product.CompanyID)
		if err != nil {
			log.Printf("⚠️  Could not resolve category %q: %v", product.Category, err)
			product.CategoryID = 0
		} else {
			product.CategoryID = catID
		}
	}

	if err := db.UpdateProduct(db.GetDB(), &product); err != nil {
		log.Printf("❌ UpdateProduct error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// ✅ Respond with the updated product
	w.WriteHeader(http.StatusOK)
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

	// ✅ Get company ID from claims
	companyID := claims.CompanyID
	if companyID == 0 {
		companyID = 1 // Default fallback
	}

	// ✅ Get shop_id from query param, default to user's shop
	shopID := claims.ShopID
	if shopIDParam := r.URL.Query().Get("shop_id"); shopIDParam != "" {
		if id, err := strconv.Atoi(shopIDParam); err == nil && id > 0 {
			shopID = id
		}
	}
	if shopID == 0 {
		shopID = 1
	}

	// ✅ Search products by company AND shop
	products, err := db.SearchProductsByCompany(db.GetDB(), query, companyID, shopID)
	if err != nil {
		log.Printf("❌ Error searching products: %v", err)
		http.Error(w, `{"error":"Failed to search products: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// ✅ Return empty array if no products found (not null)
	if products == nil {
		products = []db.Product{}
	}

	json.NewEncoder(w).Encode(products)
}

// ScanProductHandler - Scan barcode for a specific company and shop
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

	// ✅ Get shop and company from claims
	shopID := claims.ShopID
	if shopIDParam := r.URL.Query().Get("shop_id"); shopIDParam != "" {
		if id, err := strconv.Atoi(shopIDParam); err == nil && id > 0 {
			shopID = id
		}
	}
	if shopID == 0 {
		shopID = 1
	}

	companyID := claims.CompanyID
	if companyID == 0 {
		companyID = 1
	}

	log.Printf("🔍 Scanning barcode: %s, company: %d, shop: %d, qty: %d", barcode, companyID, shopID, qty)

	// ✅ Get product from master catalog (company_id)
	// ✅ Also gets stock quantity for this specific shop
	product, err := db.GetProductByBarcodeAndShop(db.GetDB(), barcode, companyID, shopID)
	if err != nil {
		log.Printf("❌ Database error: %v", err)
		http.Error(w, `<div class="text-red-500 text-center p-2">❌ Database error</div>`, http.StatusInternalServerError)
		return
	}

	if product == nil {
		log.Printf("❌ Product not found for barcode: %s (company: %d)", barcode, companyID)
		// ✅ Clear error message with suggestion
		html := fmt.Sprintf(`
            <div class="text-center p-4 bg-red-50 rounded-lg border border-red-200">
                <div class="text-3xl mb-2">🔍</div>
                <div class="text-red-600 font-bold">Product not found</div>
                <div class="text-xs text-gray-500 mt-1">Barcode: <span class="font-mono">%s</span></div>
                <div class="text-xs text-gray-400 mt-0.5">This product is not in your company's catalog</div>
                <div class="text-xs text-gray-400">Try searching by name or add it first</div>
                <button onclick="openAddProductModal()" 
                        class="mt-2 text-purple-600 hover:text-purple-800 text-xs font-semibold underline">
                    ➕ Add New Product
                </button>
            </div>
        `, barcode)
		http.Error(w, html, http.StatusNotFound)
		return
	}

	// ✅ Check stock for this shop
	if product.StockQuantity <= 0 {
		html := fmt.Sprintf(`
            <div class="text-center p-4 bg-amber-50 rounded-lg border border-amber-200">
                <div class="text-3xl mb-2">📦</div>
                <div class="text-amber-600 font-bold">Out of Stock at this shop!</div>
                <div class="text-sm text-gray-600 mt-1">%s</div>
                <div class="text-xs text-gray-400 mt-1">No stock available at this location</div>
                <div class="text-xs text-gray-400">Consider transferring stock from another shop</div>
            </div>
        `, product.Name)
		http.Error(w, html, http.StatusBadRequest)
		return
	}

	if product.StockQuantity < qty {
		html := fmt.Sprintf(`
            <div class="text-center p-4 bg-amber-50 rounded-lg border border-amber-200">
                <div class="text-3xl mb-2">⚠️</div>
                <div class="text-amber-600 font-bold">Insufficient Stock</div>
                <div class="text-sm text-gray-600 mt-1">%s</div>
                <div class="text-sm font-bold text-amber-700">Only %d items available</div>
                <div class="text-xs text-gray-400 mt-1">Requested: %d</div>
                <div class="text-xs text-gray-400">Reduce quantity or transfer stock</div>
            </div>
        `, product.Name, product.StockQuantity, qty)
		http.Error(w, html, http.StatusBadRequest)
		return
	}

	// ✅ Render HTML row for the product
	w.Write([]byte(product.RenderRowHTML()))
}

// UpdateProductStockHandler - Update stock for a product
func UpdateProductStockHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	// Only managers and above can update stock
	if claims.Role != "admin" && claims.Role != "director" && claims.Role != "manager" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Forbidden - Only managers and above can update stock",
		})
		return
	}

	var req struct {
		ProductID      int    `json:"product_id"`
		Quantity       int    `json:"quantity"`
		AdjustmentType string `json:"adjustment_type"`
		ShopID         int    `json:"shop_id"`
	}

	// ✅ Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("❌ Failed to decode request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	log.Printf("📦 Stock update request - ProductID: %d, Quantity: %d, Type: %s, ShopID: %d",
		req.ProductID, req.Quantity, req.AdjustmentType, req.ShopID)

	if req.ProductID <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Product ID is required",
		})
		return
	}

	if req.Quantity < 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Quantity cannot be negative",
		})
		return
	}

	companyID := claims.CompanyID
	if companyID == 0 {
		companyID = 1
	}

	shopID := req.ShopID
	if shopID == 0 {
		shopID = claims.ShopID
	}
	if shopID == 0 {
		shopID = 1
	}

	// ✅ Verify product belongs to company
	var productName string
	err := db.GetDB().QueryRow(`
        SELECT name FROM products WHERE id = ? AND company_id = ?
    `, req.ProductID, companyID).Scan(&productName)
	if err != nil {
		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Product not found",
			})
			return
		}
		log.Printf("❌ Failed to verify product: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to verify product",
		})
		return
	}

	// ✅ Get current stock
	var currentStock int
	err = db.GetDB().QueryRow(`
        SELECT COALESCE(quantity, 0) 
        FROM shop_stock 
        WHERE product_id = ? AND shop_id = ? AND company_id = ?
    `, req.ProductID, shopID, companyID).Scan(&currentStock)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("❌ Failed to get current stock: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to get current stock",
		})
		return
	}

	// ✅ Calculate new quantity
	var newQuantity int
	switch req.AdjustmentType {
	case "set":
		newQuantity = req.Quantity
	case "add":
		newQuantity = currentStock + req.Quantity
	case "subtract":
		newQuantity = currentStock - req.Quantity
		if newQuantity < 0 {
			newQuantity = 0
		}
	default:
		newQuantity = req.Quantity
	}

	log.Printf("📦 Stock update - Product: %s, Current: %d, New: %d",
		productName, currentStock, newQuantity)

	// ✅ Update stock
	query := `
        INSERT INTO shop_stock (shop_id, product_id, quantity, company_id, created_at, updated_at)
        VALUES (?, ?, ?, ?, NOW(), NOW())
        ON DUPLICATE KEY UPDATE quantity = ?, updated_at = NOW()
    `
	_, err = db.GetDB().Exec(query, shopID, req.ProductID, newQuantity, companyID, newQuantity)
	if err != nil {
		log.Printf("❌ Failed to update stock: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to update stock: " + err.Error(),
		})
		return
	}

	log.Printf("✅ Stock updated - Product: %s, Old: %d, New: %d", productName, currentStock, newQuantity)

	// ✅ Always return JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"message":      "Stock updated successfully",
		"product_name": productName,
		"old_quantity": currentStock,
		"new_quantity": newQuantity,
		"product_id":   req.ProductID,
		"shop_id":      shopID,
		"adjustment":   req.AdjustmentType,
	})
}
