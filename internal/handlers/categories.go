package handlers

import (
	"encoding/json"
	"net/http"
	"spide-pos/internal/db"
	"spide-pos/internal/middleware"
)

func GetCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dbConn := db.GetDB()

	rows, err := dbConn.Query("SELECT id, name FROM categories ORDER BY name ASC")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch categories"})
		return
	}
	defer rows.Close()

	categories := []db.Category{}
	for rows.Next() {
		var c db.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			continue
		}
		categories = append(categories, c)
	}

	// Check for errors after the loop
	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error reading categories"})
		return
	}

	json.NewEncoder(w).Encode(categories)
}

func CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dbConn := db.GetDB()

	var c db.Category
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil || c.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Category name is required"})
		return
	}

	// ✅ Add company_id to the insert
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}
	c.CompanyID = claims.CompanyID
	if c.CompanyID == 0 {
		c.CompanyID = 1
	}

	result, err := dbConn.Exec("INSERT INTO categories (name, company_id, is_active) VALUES (?, ?, 1)", c.Name, c.CompanyID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Category already exists or database error: " + err.Error()})
		return
	}

	// ✅ Convert int64 to int
	id64, err := result.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get category ID"})
		return
	}
	c.ID = int(id64) // ✅ Convert to int

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}
