// internal/handlers/import.go

package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"spide-pos/internal/db"
	"spide-pos/internal/middleware"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ============================================
// TYPES
// ============================================

type ImportResult struct {
	TotalRows        int      `json:"total_rows"`
	ImportedRows     int      `json:"imported_rows"`
	SkippedRows      int      `json:"skipped_rows"`
	Errors           []string `json:"errors"`
	ImportedProducts []string `json:"imported_products"`
	DuplicateCount   int      `json:"duplicate_count"`
}

type ImportPreviewItem struct {
	RowNumber   int     `json:"row_number"`
	Name        string  `json:"name"`
	Barcode     string  `json:"barcode"`
	Category    string  `json:"category"`
	CostPrice   float64 `json:"cost_price"`
	RetailPrice float64 `json:"retail_price"`
	StockQty    int     `json:"stock_qty"`
	Supplier    string  `json:"supplier"`
	Description string  `json:"description"`
	IsValid     bool    `json:"is_valid"`
	Error       string  `json:"error,omitempty"`
	Action      string  `json:"action"` // new, skip, update
}

type ImportPreviewResponse struct {
	TotalRows    int                 `json:"total_rows"`
	ValidRows    int                 `json:"valid_rows"`
	InvalidRows  int                 `json:"invalid_rows"`
	PreviewItems []ImportPreviewItem `json:"preview_items"`
	Categories   []string            `json:"categories"`
}

type ProductImportData struct {
	Name        string
	Barcode     string
	Description string
	Size        string
	CostPrice   float64
	RetailPrice float64
	StockQty    int
	Department  string
	Vendor      string
}

// ============================================
// HELPER FUNCTIONS
// ============================================

func mapQuickBooksColumns(headers []string) map[string]int {
	mapping := make(map[string]int)
	for i, header := range headers {
		header = strings.TrimSpace(header)
		headerLower := strings.ToLower(header)

		// Product Name
		if headerLower == "item name" || headerLower == "itemname" ||
			headerLower == "name" || headerLower == "product name" ||
			headerLower == "productname" {
			mapping["ItemName"] = i
			continue
		}

		// ✅ BARCODE - PRIMARY (Your "Alternate Lookup" column)
		if headerLower == "alternate lookup" || headerLower == "alt lookup" ||
			headerLower == "alternate" || headerLower == "barcode" ||
			headerLower == "bar code" || headerLower == "alu" {
			mapping["Barcode"] = i
			mapping["AlternateLookup"] = i
			continue
		}

		// UPC - SECONDARY
		if headerLower == "upc" || headerLower == "upc code" ||
			headerLower == "upc#" || headerLower == "ean" {
			mapping["UPC"] = i
			continue
		}

		// Item Number - TERTIARY
		if headerLower == "item number" || headerLower == "itemnumber" ||
			headerLower == "sku" || headerLower == "sku number" ||
			headerLower == "sku#" {
			mapping["ItemNumber"] = i
			continue
		}

		// Prices
		if headerLower == "average unit cost" || headerLower == "cost" ||
			headerLower == "unit cost" || headerLower == "cost price" ||
			headerLower == "costprice" || headerLower == "buying price" {
			mapping["CostPrice"] = i
			continue
		}
		if headerLower == "regular price" || headerLower == "price" ||
			headerLower == "retail price" || headerLower == "retailprice" ||
			headerLower == "selling price" || headerLower == "sale price" {
			mapping["RetailPrice"] = i
			continue
		}

		// Quantity
		if headerLower == "qty 1" || headerLower == "qty" ||
			headerLower == "quantity" || headerLower == "stock" ||
			headerLower == "on hand" || headerLower == "inventory" {
			mapping["StockQty"] = i
			continue
		}

		// Category
		if headerLower == "department name" || headerLower == "department" ||
			headerLower == "category" || headerLower == "type" {
			mapping["Department"] = i
			continue
		}

		// Vendor
		if headerLower == "vendor name" || headerLower == "vendor" ||
			headerLower == "supplier" || headerLower == "supplier name" {
			mapping["Vendor"] = i
			continue
		}

		// Description
		if headerLower == "item description" || headerLower == "description" ||
			headerLower == "notes" {
			mapping["Description"] = i
			continue
		}

		// Size
		if headerLower == "size" || headerLower == "size/color" ||
			headerLower == "variant" {
			mapping["Size"] = i
			continue
		}
	}
	return mapping
}

func extractProductData(row []string, mapping map[string]int) ProductImportData {
	data := ProductImportData{}

	// Get product name
	if idx, ok := mapping["ItemName"]; ok && idx < len(row) {
		data.Name = strings.TrimSpace(row[idx])
	}

	// ✅ BARCODE PRIORITY ORDER:
	// 1. Alternate Lookup (Your Excel column)
	// 2. Barcode
	// 3. UPC
	// 4. Item Number

	// 1️⃣ PRIORITY: Alternate Lookup
	if idx, ok := mapping["AlternateLookup"]; ok && idx < len(row) {
		data.Barcode = strings.TrimSpace(row[idx])
	}

	// 2️⃣ SECONDARY: Barcode
	if data.Barcode == "" {
		if idx, ok := mapping["Barcode"]; ok && idx < len(row) {
			data.Barcode = strings.TrimSpace(row[idx])
		}
	}

	// 3️⃣ TERTIARY: UPC
	if data.Barcode == "" {
		if idx, ok := mapping["UPC"]; ok && idx < len(row) {
			data.Barcode = strings.TrimSpace(row[idx])
		}
	}

	// 4️⃣ LAST: Item Number
	if data.Barcode == "" {
		if idx, ok := mapping["ItemNumber"]; ok && idx < len(row) {
			data.Barcode = strings.TrimSpace(row[idx])
		}
	}

	// Clean up barcode
	if data.Barcode != "" {
		data.Barcode = strings.Trim(data.Barcode, `"'`)
		data.Barcode = strings.TrimSpace(data.Barcode)
	}

	// ✅ Log if barcode is missing (only for debugging)
	if data.Barcode == "" && data.Name != "" {
		log.Printf("⚠️ No barcode found for product: %s", data.Name)
	}

	// Description
	if idx, ok := mapping["Description"]; ok && idx < len(row) {
		data.Description = strings.TrimSpace(row[idx])
	}

	// Size
	if idx, ok := mapping["Size"]; ok && idx < len(row) {
		data.Size = strings.TrimSpace(row[idx])
		if data.Description != "" {
			data.Description = data.Description + " | Size: " + data.Size
		} else {
			data.Description = "Size: " + data.Size
		}
	}

	// Prices
	if idx, ok := mapping["CostPrice"]; ok && idx < len(row) {
		data.CostPrice = parsePrice(row[idx])
	}

	if idx, ok := mapping["RetailPrice"]; ok && idx < len(row) {
		data.RetailPrice = parsePrice(row[idx])
	}

	// Stock Quantity
	if idx, ok := mapping["StockQty"]; ok && idx < len(row) {
		data.StockQty = parseQuantity(row[idx])
	}

	// Department/Category
	if idx, ok := mapping["Department"]; ok && idx < len(row) {
		data.Department = strings.TrimSpace(row[idx])
	}
	if data.Department == "" {
		data.Department = "General"
	}

	// Vendor
	if idx, ok := mapping["Vendor"]; ok && idx < len(row) {
		data.Vendor = strings.TrimSpace(row[idx])
	}

	return data
}

func parsePrice(val string) float64 {
	val = strings.ReplaceAll(val, ",", "")
	val = strings.ReplaceAll(val, "$", "")
	val = strings.ReplaceAll(val, "KSh", "")
	val = strings.ReplaceAll(val, "KES", "")
	val = strings.TrimSpace(val)

	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0
	}
	return f
}

func parseQuantity(val string) int {
	val = strings.ReplaceAll(val, ",", "")
	val = strings.TrimSpace(val)

	i, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return i
}

// ============================================
// PREVIEW HANDLER
// ============================================

func PreviewImportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	companyID := claims.CompanyID
	if companyID == 0 {
		companyID = 1
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !strings.HasSuffix(handler.Filename, ".xlsx") && !strings.HasSuffix(handler.Filename, ".xls") {
		http.Error(w, "Only Excel files (.xlsx, .xls) are supported", http.StatusBadRequest)
		return
	}

	f, err := excelize.OpenReader(file)
	if err != nil {
		http.Error(w, "Failed to read Excel file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		http.Error(w, "No sheets found in Excel file", http.StatusBadRequest)
		return
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		http.Error(w, "Failed to read sheet: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(rows) < 2 {
		http.Error(w, "No data rows found (need header row + data rows)", http.StatusBadRequest)
		return
	}

	mapping := mapQuickBooksColumns(rows[0])

	requiredCols := []string{"ItemName", "CostPrice", "RetailPrice"}
	for _, col := range requiredCols {
		if _, ok := mapping[col]; !ok {
			http.Error(w, "Missing required column: "+col, http.StatusBadRequest)
			return
		}
	}

	response := &ImportPreviewResponse{
		PreviewItems: []ImportPreviewItem{},
		Categories:   []string{},
	}

	categorySet := make(map[string]bool)

	// Process data rows for preview (limit to 2000 rows for performance)
	maxRows := len(rows)
	if maxRows > 2000 {
		maxRows = 2000
	}

	for i := 1; i < maxRows; i++ {
		row := rows[i]
		if len(row) == 0 {
			continue
		}

		item := ImportPreviewItem{
			RowNumber: i + 1,
			Action:    "new",
		}

		productData := extractProductData(row, mapping)

		isValid := true
		var errors []string

		if productData.Name == "" {
			isValid = false
			errors = append(errors, "Product name is required")
		}

		if productData.CostPrice <= 0 {
			isValid = false
			errors = append(errors, "Cost price must be greater than 0")
		}

		if productData.RetailPrice <= 0 {
			isValid = false
			errors = append(errors, "Retail price must be greater than 0")
		}

		// Check for duplicates
		if productData.Barcode != "" {
			existing, _ := db.GetProductByBarcode(db.GetDB(), productData.Barcode, companyID)
			if existing != nil {
				item.Action = "skip"
				isValid = false
				errors = append(errors, fmt.Sprintf("Duplicate barcode: already exists as '%s'", existing.Name))
			}
		}

		item.Name = productData.Name
		item.Barcode = productData.Barcode
		item.Category = productData.Department
		item.CostPrice = productData.CostPrice
		item.RetailPrice = productData.RetailPrice
		item.StockQty = productData.StockQty
		item.Supplier = productData.Vendor
		item.Description = productData.Description
		item.IsValid = isValid

		if !isValid {
			item.Error = strings.Join(errors, "; ")
			response.InvalidRows++
		} else {
			response.ValidRows++
			if item.Category != "" {
				categorySet[item.Category] = true
			}
		}

		response.PreviewItems = append(response.PreviewItems, item)
	}

	response.TotalRows = len(response.PreviewItems)

	for cat := range categorySet {
		response.Categories = append(response.Categories, cat)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ============================================
// CONFIRM IMPORT HANDLER
// ============================================

func ConfirmImportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	companyID := claims.CompanyID
	if companyID == 0 {
		companyID = 1
	}

	var req struct {
		ShopID   int                 `json:"shop_id"`
		Products []ImportPreviewItem `json:"products"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("❌ Failed to decode request: %v", err)
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("📊 Import Request - Company: %d, ShopID: %d, Products: %d",
		companyID, req.ShopID, len(req.Products))

	result := &ImportResult{
		Errors:           []string{},
		ImportedProducts: []string{},
	}

	// If ShopID is 0, use the user's shop
	shopID := req.ShopID
	if shopID == 0 {
		shopID = claims.ShopID
		if shopID == 0 {
			shopID = 1
		}
	}

	// ✅ Get only valid products
	validProducts := []ImportPreviewItem{}
	skippedInvalid := 0
	for _, item := range req.Products {
		if item.IsValid {
			validProducts = append(validProducts, item)
		} else {
			skippedInvalid++
			result.SkippedRows++
			result.Errors = append(result.Errors,
				fmt.Sprintf("Row %d: %s - %s", item.RowNumber, item.Name, item.Error))
		}
	}

	totalValid := len(validProducts)
	log.Printf("📊 Valid products to import: %d, Invalid: %d", totalValid, skippedInvalid)

	if totalValid == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"imported_rows":   0,
			"skipped_rows":    result.SkippedRows,
			"total_processed": 0,
			"errors":          result.Errors,
			"message":         "No valid products to import",
		})
		return
	}

	// ✅ BATCH PROCESSING - Process in batches of 100
	const batchSize = 100
	totalBatches := (totalValid + batchSize - 1) / batchSize

	log.Printf("📦 Processing %d products in %d batches of %d",
		totalValid, totalBatches, batchSize)

	// ✅ Track skipped products
	skippedNoBarcode := 0
	skippedDuplicate := 0
	skippedCategoryError := 0
	skippedProductError := 0

	for batchNum := 0; batchNum < totalBatches; batchNum++ {
		start := batchNum * batchSize
		end := start + batchSize
		if end > totalValid {
			end = totalValid
		}

		batch := validProducts[start:end]
		log.Printf("📦 Batch %d/%d: Processing products %d-%d",
			batchNum+1, totalBatches, start+1, end)

		// Process each item in the batch
		for _, item := range batch {
			// ✅ Check if product has barcode (log warning but continue)
			if item.Barcode == "" {
				log.Printf("📝 Product '%s' - No barcode (will be NULL)", item.Name)
				// Keep empty - CreateProduct will convert to NULL
			}

			// ✅ Check for duplicate barcode (only if barcode is not empty)
			if item.Barcode != "" {
				existing, err := db.GetProductByBarcode(db.GetDB(), item.Barcode, companyID)
				if err != nil {
					log.Printf("⚠️ Error checking duplicate barcode for '%s': %v", item.Name, err)
				}
				if existing != nil {
					log.Printf("⚠️ Duplicate barcode '%s' for '%s', skipping", item.Barcode, item.Name)
					skippedDuplicate++
					result.SkippedRows++
					result.Errors = append(result.Errors,
						fmt.Sprintf("Row %d: Duplicate barcode '%s' - already exists as '%s'",
							item.RowNumber, item.Barcode, existing.Name))
					continue
				}
			}

			// 1️⃣ Get or create category
			categoryID, err := db.GetOrCreateCategory(db.GetDB(), item.Category, companyID)
			if err != nil {
				log.Printf("❌ Failed to create category for '%s': %v", item.Name, err)
				skippedCategoryError++
				result.SkippedRows++
				result.Errors = append(result.Errors,
					fmt.Sprintf("Row %d: Failed to create category: %v", item.RowNumber, err))
				continue
			}
			var barcodePtr *string
			if item.Barcode == "" {
				barcodePtr = nil // This will insert NULL in the database
			} else {
				barcodePtr = &item.Barcode // Pointer to the string
			}
			wholesalePrice := item.RetailPrice * 0.9
			// 2️⃣ Create product with company_id
			product := db.Product{
				CompanyID:       companyID,
				Barcode:         barcodePtr,
				Name:            item.Name,
				Category:        item.Category,
				CategoryID:      categoryID,
				CostPrice:       item.CostPrice,
				RetailPrice:     item.RetailPrice,
				WholesalePrice:  wholesalePrice,
				WholesaleMinQty: 0,
				ReorderLevel:    5,
			}

			productID, err := db.CreateProduct(db.GetDB(), &product)
			if err != nil {
				// ✅ Check if it's a duplicate barcode error
				if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "barcode") {
					log.Printf("⚠️ Duplicate barcode for '%s', skipping", item.Name)
					skippedDuplicate++
					result.SkippedRows++
					result.Errors = append(result.Errors,
						fmt.Sprintf("Row %d: Duplicate barcode - '%s' already exists", item.RowNumber, item.Barcode))
					continue
				}

				log.Printf("❌ Failed to create product '%s': %v", item.Name, err)
				skippedProductError++
				result.SkippedRows++
				result.Errors = append(result.Errors,
					fmt.Sprintf("Row %d: Failed to create product: %v", item.RowNumber, err))
				continue
			}

			// 3️⃣ Add stock to shop if quantity > 0
			if shopID > 0 && item.StockQty > 0 {
				err = db.CreateShopStock(db.GetDB(), shopID, int(productID), item.StockQty, companyID)
				if err != nil {
					log.Printf("⚠️ Warning: Failed to create shop stock for '%s': %v", item.Name, err)
					// Don't fail the import, just log the warning
				}
			}

			result.ImportedRows++
			result.ImportedProducts = append(result.ImportedProducts, item.Name)
		}

		// ✅ Log batch completion
		log.Printf("✅ Batch %d/%d complete: %d products imported so far",
			batchNum+1, totalBatches, result.ImportedRows)
	}

	log.Printf("📊 Import Complete - Imported: %d, Skipped: %d (No Barcode: %d, Duplicate: %d, Category Error: %d, Product Error: %d)",
		result.ImportedRows, result.SkippedRows, skippedNoBarcode, skippedDuplicate, skippedCategoryError, skippedProductError)

	// ✅ Return detailed result
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"imported_rows":     result.ImportedRows,
		"skipped_rows":      result.SkippedRows,
		"total_processed":   result.ImportedRows + result.SkippedRows,
		"errors":            result.Errors,
		"imported_products": result.ImportedProducts,
		"skipped_details": map[string]int{
			"no_barcode":     skippedNoBarcode,
			"duplicate":      skippedDuplicate,
			"category_error": skippedCategoryError,
			"product_error":  skippedProductError,
		},
		"message": fmt.Sprintf("Successfully imported %d of %d products",
			result.ImportedRows, totalValid),
	}

	if result.SkippedRows > 0 {
		response["warning"] = fmt.Sprintf("%d products were skipped. Check errors for details.", result.SkippedRows)
	}

	json.NewEncoder(w).Encode(response)
}

// ============================================
// BARCODE UPDATE FUNCTIONS
// ============================================

// extractBarcodeData - Extract ONLY barcode data from Excel (for barcode updates)
func extractBarcodeData(row []string, mapping map[string]int) ProductImportData {
	data := ProductImportData{}

	// Get product name (required for matching)
	if idx, ok := mapping["ItemName"]; ok && idx < len(row) {
		data.Name = strings.TrimSpace(row[idx])
	}

	// ✅ Get barcode from mapping (handles both "Barcode" and "AlternateLookup")
	if idx, ok := mapping["Barcode"]; ok && idx < len(row) {
		data.Barcode = strings.TrimSpace(row[idx])
	}

	// If no Barcode mapping, try AlternateLookup
	if data.Barcode == "" {
		if idx, ok := mapping["AlternateLookup"]; ok && idx < len(row) {
			data.Barcode = strings.TrimSpace(row[idx])
		}
	}

	// If still no barcode, try UPC
	if data.Barcode == "" {
		if idx, ok := mapping["UPC"]; ok && idx < len(row) {
			data.Barcode = strings.TrimSpace(row[idx])
		}
	}

	// If still no barcode, try Item Number
	if data.Barcode == "" {
		if idx, ok := mapping["ItemNumber"]; ok && idx < len(row) {
			data.Barcode = strings.TrimSpace(row[idx])
		}
	}

	// Clean up barcode - remove quotes, spaces
	if data.Barcode != "" {
		data.Barcode = strings.Trim(data.Barcode, `"'`)
		data.Barcode = strings.TrimSpace(data.Barcode)
	}

	return data
}

func UpdateBarcodesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	companyID := claims.CompanyID
	if companyID == 0 {
		companyID = 1
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !strings.HasSuffix(handler.Filename, ".xlsx") && !strings.HasSuffix(handler.Filename, ".xls") {
		http.Error(w, "Only Excel files (.xlsx, .xls) are supported", http.StatusBadRequest)
		return
	}

	f, err := excelize.OpenReader(file)
	if err != nil {
		http.Error(w, "Failed to read Excel file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		http.Error(w, "No sheets found in Excel file", http.StatusBadRequest)
		return
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		http.Error(w, "Failed to read sheet: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(rows) < 2 {
		http.Error(w, "No data rows found (need header row + data rows)", http.StatusBadRequest)
		return
	}

	// Map headers
	mapping := mapQuickBooksColumns(rows[0])

	// ✅ Debug: Log what was found
	log.Printf("📋 Mapping results:")
	for key, idx := range mapping {
		if idx < len(rows[0]) {
			log.Printf("  %s → Column %d: '%s'", key, idx, rows[0][idx])
		}
	}

	// ✅ Only require Item Name for barcode updates
	if _, ok := mapping["ItemName"]; !ok {
		log.Printf("❌ ERROR: 'Item Name' column not found!")
		log.Printf("Available columns: %v", rows[0])
		http.Error(w, "Missing required column: Item Name", http.StatusBadRequest)
		return
	}

	// ✅ Check if we found a barcode column
	if _, ok := mapping["Barcode"]; !ok {
		if _, ok := mapping["AlternateLookup"]; !ok {
			log.Printf("⚠️ WARNING: No barcode column found! Looking for 'Alternate Lookup' or 'Barcode'")
			log.Printf("Available columns: %v", rows[0])
		}
	}

	log.Printf("📊 Barcode Update - Company: %d, File: %s, Rows: %d",
		companyID, handler.Filename, len(rows)-1)

	result := &ImportResult{
		Errors:           []string{},
		ImportedProducts: []string{},
	}

	// Process each row
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue
		}

		// ✅ Use the barcode-only extraction function
		productData := extractBarcodeData(row, mapping)

		// Skip if no product name
		if productData.Name == "" {
			log.Printf("⚠️ Row %d: No product name, skipping", i+1)
			result.SkippedRows++
			continue
		}

		// ✅ Skip if no barcode to update with
		if productData.Barcode == "" {
			log.Printf("⚠️ Row %d: No barcode for '%s', skipping", i+1, productData.Name)
			result.SkippedRows++
			result.Errors = append(result.Errors,
				fmt.Sprintf("Row %d: No barcode for '%s'", i+1, productData.Name))
			continue
		}

		log.Printf("📝 Processing: '%s' → Barcode: '%s'", productData.Name, productData.Barcode)

		// Find product by name (exact match first)
		var product *db.Product
		var err error

		// Try exact match first
		product, err = db.GetProductByNameExact(db.GetDB(), productData.Name, companyID)
		if err != nil {
			log.Printf("❌ Error searching for '%s': %v", productData.Name, err)
			result.Errors = append(result.Errors,
				fmt.Sprintf("Row %d: Error searching for '%s': %v", i+1, productData.Name, err))
			result.SkippedRows++
			continue
		}

		// If not found, try fuzzy match (contains)
		if product == nil {
			products, err := db.GetProductsByNameFuzzy(db.GetDB(), productData.Name, companyID)
			if err != nil {
				log.Printf("❌ Error fuzzy searching for '%s': %v", productData.Name, err)
				continue
			}
			if len(products) == 1 {
				product = &products[0]
				log.Printf("🔍 Fuzzy matched '%s' to '%s'", productData.Name, product.Name)
			} else if len(products) > 1 {
				log.Printf("⚠️ Multiple matches for '%s', skipping", productData.Name)
				result.Errors = append(result.Errors,
					fmt.Sprintf("Row %d: Multiple matches for '%s'", i+1, productData.Name))
				result.SkippedRows++
				continue
			}
		}

		if product == nil {
			log.Printf("⚠️ Product not found: '%s'", productData.Name)
			result.SkippedRows++
			result.Errors = append(result.Errors,
				fmt.Sprintf("Row %d: Product not found: '%s'", i+1, productData.Name))
			continue
		}

		// Check if barcode is already in use (by another product)
		existing, _ := db.GetProductByBarcode(db.GetDB(), productData.Barcode, companyID)
		if existing != nil && existing.ID != product.ID {
			log.Printf("⚠️ Barcode '%s' already used by '%s', skipping",
				productData.Barcode, existing.Name)
			result.SkippedRows++
			result.Errors = append(result.Errors,
				fmt.Sprintf("Row %d: Barcode '%s' already used by '%s'",
					i+1, productData.Barcode, existing.Name))
			continue
		}

		// Update the barcode
		err = db.UpdateProductBarcode(db.GetDB(), product.ID, productData.Barcode)
		if err != nil {
			log.Printf("❌ Failed to update barcode for '%s': %v", product.Name, err)
			result.SkippedRows++
			result.Errors = append(result.Errors,
				fmt.Sprintf("Row %d: Failed to update barcode: %v", i+1, err))
			continue
		}

		result.ImportedRows++
		result.ImportedProducts = append(result.ImportedProducts,
			fmt.Sprintf("%s → %s", product.Name, productData.Barcode))

		log.Printf("✅ Updated barcode for '%s': %s", product.Name, productData.Barcode)
	}

	log.Printf("📊 Barcode Update Complete - Updated: %d, Skipped: %d, Errors: %d",
		result.ImportedRows, result.SkippedRows, len(result.Errors))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"updated_rows":     result.ImportedRows,
		"skipped_rows":     result.SkippedRows,
		"total_processed":  result.ImportedRows + result.SkippedRows,
		"errors":           result.Errors,
		"updated_products": result.ImportedProducts,
		"message": fmt.Sprintf("Successfully updated %d product barcodes",
			result.ImportedRows),
	})
}
