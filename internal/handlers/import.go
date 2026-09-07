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
		switch header {
		case "Item Name":
			mapping["ItemName"] = i
		case "Item Number":
			mapping["ItemNumber"] = i
		case "Alternate Lookup":
			mapping["AlternateLookup"] = i
		case "UPC":
			mapping["UPC"] = i
		case "Average Unit Cost":
			mapping["CostPrice"] = i
		case "Regular Price":
			mapping["RetailPrice"] = i
		case "Qty 1":
			mapping["StockQty"] = i
		case "Department Name":
			mapping["Department"] = i
		case "Vendor Name":
			mapping["Vendor"] = i
		case "Item Description":
			mapping["Description"] = i
		case "Size":
			mapping["Size"] = i
		case "MSRP":
			mapping["MSRP"] = i
		}
	}
	return mapping
}

func extractProductData(row []string, mapping map[string]int) ProductImportData {
	data := ProductImportData{}

	if idx, ok := mapping["ItemName"]; ok && idx < len(row) {
		data.Name = strings.TrimSpace(row[idx])
	}

	// Barcode: Try UPC first, then Alternate Lookup
	if idx, ok := mapping["UPC"]; ok && idx < len(row) {
		data.Barcode = strings.TrimSpace(row[idx])
	}
	if data.Barcode == "" {
		if idx, ok := mapping["AlternateLookup"]; ok && idx < len(row) {
			data.Barcode = strings.TrimSpace(row[idx])
		}
	}
	if data.Barcode == "" {
		if idx, ok := mapping["ItemNumber"]; ok && idx < len(row) {
			data.Barcode = strings.TrimSpace(row[idx])
		}
	}

	if idx, ok := mapping["Description"]; ok && idx < len(row) {
		data.Description = strings.TrimSpace(row[idx])
	}

	if idx, ok := mapping["Size"]; ok && idx < len(row) {
		data.Size = strings.TrimSpace(row[idx])
		if data.Description != "" {
			data.Description = data.Description + " | Size: " + data.Size
		} else {
			data.Description = "Size: " + data.Size
		}
	}

	if idx, ok := mapping["CostPrice"]; ok && idx < len(row) {
		data.CostPrice = parsePrice(row[idx])
	}

	if idx, ok := mapping["RetailPrice"]; ok && idx < len(row) {
		data.RetailPrice = parsePrice(row[idx])
	}

	if idx, ok := mapping["StockQty"]; ok && idx < len(row) {
		data.StockQty = parseQuantity(row[idx])
	}

	if idx, ok := mapping["Department"]; ok && idx < len(row) {
		data.Department = strings.TrimSpace(row[idx])
	}
	if data.Department == "" {
		data.Department = "General"
	}

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
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	result := &ImportResult{
		Errors:           []string{},
		ImportedProducts: []string{},
	}

	for _, item := range req.Products {
		if !item.IsValid {
			result.SkippedRows++
			continue
		}

		// Get or create category
		categoryID, err := db.GetOrCreateCategory(db.GetDB(), item.Category, companyID)
		if err != nil {
			result.SkippedRows++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: Failed to create category: %v", item.RowNumber, err))
			continue
		}

		// Create product
		product := db.Product{
			CompanyID:       companyID,
			Barcode:         item.Barcode,
			Name:            item.Name,
			Category:        item.Category,
			CategoryID:      categoryID,
			CostPrice:       item.CostPrice,
			RetailPrice:     item.RetailPrice,
			WholesalePrice:  0,
			WholesaleMinQty: 0,
			ReorderLevel:    5,
		}

		productID, err := db.CreateProduct(db.GetDB(), &product)
		if err != nil {
			result.SkippedRows++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: Failed to create product: %v", item.RowNumber, err))
			continue
		}

		// Add stock if shop selected
		if req.ShopID > 0 && item.StockQty > 0 {
			err = db.CreateShopStock(db.GetDB(), req.ShopID, int(productID), item.StockQty, companyID)
			if err != nil {
				log.Printf("Warning: Failed to create shop stock for %s: %v", item.Name, err)
			}
		}

		result.ImportedRows++
		result.ImportedProducts = append(result.ImportedProducts, item.Name)
		result.TotalRows++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
