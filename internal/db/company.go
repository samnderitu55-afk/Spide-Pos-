package db

import (
	"database/sql"
	"fmt"
	"time"
)

// ============================================
// COMPANY SETTINGS STRUCT
// ============================================

type CompanySettings struct {
	ID                 int            `json:"id"`
	CompanyName        string         `json:"company_name"`
	RegistrationNumber sql.NullString `json:"registration_number"`
	TaxID              sql.NullString `json:"tax_id"`
	Phone              sql.NullString `json:"phone"`
	Email              sql.NullString `json:"email"`
	Website            sql.NullString `json:"website"`
	Address            sql.NullString `json:"address"`
	LogoPath           sql.NullString `json:"logo_path"`
	Currency           string         `json:"currency"`
	ReceiptFooter      sql.NullString `json:"receipt_footer"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

// ============================================
// GET COMPANY SETTINGS
// ============================================

func GetCompanySettings(db *sql.DB) (*CompanySettings, error) {
	var settings CompanySettings
	query := `
        SELECT id, company_name, registration_number, tax_id, phone, email, 
               website, address, logo_path, currency, receipt_footer, created_at, updated_at
        FROM company_settings
        LIMIT 1
    `
	err := db.QueryRow(query).Scan(
		&settings.ID, &settings.CompanyName, &settings.RegistrationNumber,
		&settings.TaxID, &settings.Phone, &settings.Email, &settings.Website,
		&settings.Address, &settings.LogoPath, &settings.Currency,
		&settings.ReceiptFooter, &settings.CreatedAt, &settings.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		// Return default settings if not found
		return &CompanySettings{
			CompanyName: "Spide POS",
			Currency:    "KES",
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// ============================================
// UPDATE COMPANY SETTINGS
// ============================================

func UpdateCompanySettings(db *sql.DB, settings CompanySettings) error {
	// Check if record exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM company_settings").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Insert new record
		query := `
            INSERT INTO company_settings (company_name, registration_number, tax_id, phone, 
                email, website, address, logo_path, currency, receipt_footer)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `
		_, err := db.Exec(query,
			settings.CompanyName, settings.RegistrationNumber, settings.TaxID,
			settings.Phone, settings.Email, settings.Website, settings.Address,
			settings.LogoPath, settings.Currency, settings.ReceiptFooter,
		)
		return err
	}

	// Update existing record
	query := `
        UPDATE company_settings 
        SET company_name = ?, registration_number = ?, tax_id = ?, 
            phone = ?, email = ?, website = ?, address = ?, 
            logo_path = ?, currency = ?, receipt_footer = ?,
            updated_at = NOW()
        WHERE id = ?
    `
	_, err = db.Exec(query,
		settings.CompanyName, settings.RegistrationNumber, settings.TaxID,
		settings.Phone, settings.Email, settings.Website, settings.Address,
		settings.LogoPath, settings.Currency, settings.ReceiptFooter,
		settings.ID,
	)
	return err
}

// ============================================
// UPDATE COMPANY LOGO
// ============================================

func UpdateCompanyLogo(db *sql.DB, logoPath string, companyID int) error {
	query := `
        UPDATE company_settings 
        SET logo_path = ?, updated_at = NOW()
        WHERE id = ?
    `
	_, err := db.Exec(query, logoPath, companyID)
	return err
}

// ============================================
// MULTI-COMPANY SUPPORT (Optional - for future use)
// ============================================

type Company struct {
	ID                 int            `json:"id"`
	CompanyName        string         `json:"company_name"`
	RegistrationNumber sql.NullString `json:"registration_number"`
	TaxID              sql.NullString `json:"tax_id"`
	Phone              sql.NullString `json:"phone"`
	Email              sql.NullString `json:"email"`
	Website            sql.NullString `json:"website"`
	Address            sql.NullString `json:"address"`
	LogoPath           sql.NullString `json:"logo_path"`
	Currency           string         `json:"currency"`
	ReceiptFooter      sql.NullString `json:"receipt_footer"`
	IsActive           bool           `json:"is_active"`
	IsDemo             bool           `json:"is_demo"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

func GetAllCompanies(db *sql.DB) ([]Company, error) {
	query := `
        SELECT id, company_name, registration_number, tax_id, phone, email, 
               website, address, logo_path, currency, receipt_footer, 
               is_active, is_demo, created_at, updated_at
        FROM companies
        ORDER BY company_name
    `
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []Company
	for rows.Next() {
		var c Company
		err := rows.Scan(
			&c.ID, &c.CompanyName, &c.RegistrationNumber, &c.TaxID,
			&c.Phone, &c.Email, &c.Website, &c.Address,
			&c.LogoPath, &c.Currency, &c.ReceiptFooter,
			&c.IsActive, &c.IsDemo, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		companies = append(companies, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return companies, nil
}

func GetCompanyByID(db *sql.DB, id int) (*Company, error) {
	var c Company
	query := `
        SELECT id, company_name, registration_number, tax_id, phone, email, 
               website, address, logo_path, currency, receipt_footer, 
               is_active, is_demo, created_at, updated_at
        FROM companies
        WHERE id = ?
    `
	err := db.QueryRow(query, id).Scan(
		&c.ID, &c.CompanyName, &c.RegistrationNumber, &c.TaxID,
		&c.Phone, &c.Email, &c.Website, &c.Address,
		&c.LogoPath, &c.Currency, &c.ReceiptFooter,
		&c.IsActive, &c.IsDemo, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func CreateCompany(db *sql.DB, company Company) (int64, error) {
	query := `
        INSERT INTO companies (company_name, registration_number, tax_id, phone, email, 
            website, address, logo_path, currency, receipt_footer, is_active, is_demo)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	result, err := db.Exec(query,
		company.CompanyName, company.RegistrationNumber, company.TaxID,
		company.Phone, company.Email, company.Website, company.Address,
		company.LogoPath, company.Currency, company.ReceiptFooter,
		company.IsActive, company.IsDemo,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func UpdateCompany(db *sql.DB, company Company) error {
	query := `
        UPDATE companies 
        SET company_name = ?, registration_number = ?, tax_id = ?, 
            phone = ?, email = ?, website = ?, address = ?, 
            logo_path = ?, currency = ?, receipt_footer = ?,
            is_active = ?, is_demo = ?, updated_at = NOW()
        WHERE id = ?
    `
	_, err := db.Exec(query,
		company.CompanyName, company.RegistrationNumber, company.TaxID,
		company.Phone, company.Email, company.Website, company.Address,
		company.LogoPath, company.Currency, company.ReceiptFooter,
		company.IsActive, company.IsDemo, company.ID,
	)
	return err
}

func DeleteCompany(db *sql.DB, id int) error {
	// Prevent deleting the last active company
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM companies WHERE is_active = 1").Scan(&count)
	if err != nil {
		return err
	}
	if count <= 1 {
		return fmt.Errorf("cannot delete the last active company")
	}

	_, err = db.Exec("DELETE FROM companies WHERE id = ?", id)
	return err
}
