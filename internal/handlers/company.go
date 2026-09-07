package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"spide-pos/internal/db"
	"spide-pos/internal/middleware"
	"strconv"
	"strings"
)

// ============================================
// COMPANY SETTINGS HANDLERS
// ============================================

func GetCompanySettingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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

	settings, err := db.GetCompanySettings(db.GetDB())
	if err != nil {
		log.Printf("❌ Error getting company settings: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func UpdateCompanySettingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil || (claims.Role != "admin" && claims.Role != "director") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ID                 int    `json:"id"`
		CompanyName        string `json:"company_name"`
		RegistrationNumber string `json:"registration_number"`
		TaxID              string `json:"tax_id"`
		Phone              string `json:"phone"`
		Email              string `json:"email"`
		Website            string `json:"website"`
		Address            string `json:"address"`
		LogoPath           string `json:"logo_path"`
		Currency           string `json:"currency"`
		ReceiptFooter      string `json:"receipt_footer"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("❌ Error decoding company settings: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request format: " + err.Error(),
		})
		return
	}

	// Convert to CompanySettings with NullString handling
	settings := db.CompanySettings{
		ID:                 req.ID,
		CompanyName:        req.CompanyName,
		RegistrationNumber: sql.NullString{String: req.RegistrationNumber, Valid: req.RegistrationNumber != ""},
		TaxID:              sql.NullString{String: req.TaxID, Valid: req.TaxID != ""},
		Phone:              sql.NullString{String: req.Phone, Valid: req.Phone != ""},
		Email:              sql.NullString{String: req.Email, Valid: req.Email != ""},
		Website:            sql.NullString{String: req.Website, Valid: req.Website != ""},
		Address:            sql.NullString{String: req.Address, Valid: req.Address != ""},
		LogoPath:           sql.NullString{String: req.LogoPath, Valid: req.LogoPath != ""},
		Currency:           req.Currency,
		ReceiptFooter:      sql.NullString{String: req.ReceiptFooter, Valid: req.ReceiptFooter != ""},
	}

	existing, err := db.GetCompanySettings(db.GetDB())
	if err != nil {
		log.Printf("❌ Error getting existing settings: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}
	settings.ID = existing.ID

	err = db.UpdateCompanySettings(db.GetDB(), settings)
	if err != nil {
		log.Printf("❌ Error updating company settings: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Company settings updated successfully",
	})
}

func UploadCompanyLogoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil || (claims.Role != "admin" && claims.Role != "director") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("logo")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	contentType := handler.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		http.Error(w, "Only image files are allowed", http.StatusBadRequest)
		return
	}

	uploadDir := "./static/uploads/logos"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}

	filename := "logo_" + handler.Filename
	filePath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	logoPath := "/static/uploads/logos/" + filename
	settings, err := db.GetCompanySettings(db.GetDB())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = db.UpdateCompanyLogo(db.GetDB(), logoPath, settings.ID)
	if err != nil {
		http.Error(w, "Failed to update logo in database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"message":   "Logo uploaded successfully",
		"logo_path": logoPath,
	})
}

// ============================================
// ✅ COMPANY MANAGEMENT HANDLERS
// ============================================

// GetCompaniesHandler - Get all companies (Admin/Director only)
func GetCompaniesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" && claims.Role != "director" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Forbidden: Only admins and directors can manage companies",
		})
		return
	}

	companies, err := db.GetAllCompanies(db.GetDB())
	if err != nil {
		log.Printf("❌ Error getting companies: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(companies)
}

// CreateCompanyHandler - Create a new company (Admin/Director only)
func CreateCompanyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil || (claims.Role != "admin" && claims.Role != "director") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		CompanyName        string `json:"company_name"`
		RegistrationNumber string `json:"registration_number"`
		TaxID              string `json:"tax_id"`
		Phone              string `json:"phone"`
		Email              string `json:"email"`
		Website            string `json:"website"`
		Address            string `json:"address"`
		LogoPath           string `json:"logo_path"`
		Currency           string `json:"currency"`
		ReceiptFooter      string `json:"receipt_footer"`
		IsActive           bool   `json:"is_active"`
		IsDemo             bool   `json:"is_demo"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("❌ Error decoding company: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request format: " + err.Error(),
		})
		return
	}

	if req.CompanyName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Company name is required",
		})
		return
	}

	company := db.Company{
		CompanyName:        req.CompanyName,
		RegistrationNumber: sql.NullString{String: req.RegistrationNumber, Valid: req.RegistrationNumber != ""},
		TaxID:              sql.NullString{String: req.TaxID, Valid: req.TaxID != ""},
		Phone:              sql.NullString{String: req.Phone, Valid: req.Phone != ""},
		Email:              sql.NullString{String: req.Email, Valid: req.Email != ""},
		Website:            sql.NullString{String: req.Website, Valid: req.Website != ""},
		Address:            sql.NullString{String: req.Address, Valid: req.Address != ""},
		LogoPath:           sql.NullString{String: req.LogoPath, Valid: req.LogoPath != ""},
		Currency:           req.Currency,
		ReceiptFooter:      sql.NullString{String: req.ReceiptFooter, Valid: req.ReceiptFooter != ""},
		IsActive:           req.IsActive,
		IsDemo:             req.IsDemo,
	}

	id, err := db.CreateCompany(db.GetDB(), company)
	if err != nil {
		log.Printf("❌ Error creating company: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	company.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Company created successfully",
		"company": company,
	})
}

// UpdateCompanyHandler - Update an existing company (Admin/Director only)
func UpdateCompanyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil || (claims.Role != "admin" && claims.Role != "director") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ID                 int    `json:"id"`
		CompanyName        string `json:"company_name"`
		RegistrationNumber string `json:"registration_number"`
		TaxID              string `json:"tax_id"`
		Phone              string `json:"phone"`
		Email              string `json:"email"`
		Website            string `json:"website"`
		Address            string `json:"address"`
		LogoPath           string `json:"logo_path"`
		Currency           string `json:"currency"`
		ReceiptFooter      string `json:"receipt_footer"`
		IsActive           bool   `json:"is_active"`
		IsDemo             bool   `json:"is_demo"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("❌ Error decoding company: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request format: " + err.Error(),
		})
		return
	}

	if req.ID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Company ID is required",
		})
		return
	}

	company := db.Company{
		ID:                 req.ID,
		CompanyName:        req.CompanyName,
		RegistrationNumber: sql.NullString{String: req.RegistrationNumber, Valid: req.RegistrationNumber != ""},
		TaxID:              sql.NullString{String: req.TaxID, Valid: req.TaxID != ""},
		Phone:              sql.NullString{String: req.Phone, Valid: req.Phone != ""},
		Email:              sql.NullString{String: req.Email, Valid: req.Email != ""},
		Website:            sql.NullString{String: req.Website, Valid: req.Website != ""},
		Address:            sql.NullString{String: req.Address, Valid: req.Address != ""},
		LogoPath:           sql.NullString{String: req.LogoPath, Valid: req.LogoPath != ""},
		Currency:           req.Currency,
		ReceiptFooter:      sql.NullString{String: req.ReceiptFooter, Valid: req.ReceiptFooter != ""},
		IsActive:           req.IsActive,
		IsDemo:             req.IsDemo,
	}

	err = db.UpdateCompany(db.GetDB(), company)
	if err != nil {
		log.Printf("❌ Error updating company: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Company updated successfully",
	})
}

// DeleteCompanyHandler - Delete a company (Admin/Director only)
func DeleteCompanyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil || (claims.Role != "admin" && claims.Role != "director") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// ✅ Use CompanyID with fallback
	companyID := claims.CompanyID
	if companyID == 0 {
		// If CompanyID is 0, try to get it from the user cookie
		userCookie, err := r.Cookie("spide_user")
		if err == nil {
			var userData map[string]interface{}
			decoded := strings.ReplaceAll(userCookie.Value, `'`, `"`)
			if err := json.Unmarshal([]byte(decoded), &userData); err == nil {
				if id, ok := userData["company_id"].(float64); ok {
					companyID = int(id)
				}
			}
		}
		// If still 0, default to 1
		if companyID == 0 {
			companyID = 1
		}
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Company ID required",
		})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid company ID",
		})
		return
	}

	// Prevent deleting your own company
	if id == claims.CompanyID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Cannot delete your own company",
		})
		return
	}

	err = db.DeleteCompany(db.GetDB(), id)
	if err != nil {
		log.Printf("❌ Error deleting company: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Company deleted successfully",
	})
}
