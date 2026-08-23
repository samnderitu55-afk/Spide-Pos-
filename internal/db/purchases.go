package db

import (
    "database/sql"
    "fmt"
    "time"
)

type PurchaseReportItem struct {
    ID             int     `json:"id"`
    PurchaseNumber string  `json:"purchase_number"`
    SupplierName   string  `json:"supplier_name"`
    TotalItems     int     `json:"total_items"`
    TotalCost      float64 `json:"total_cost"`
    PurchaseDate   string  `json:"purchase_date"`
    Notes          string  `json:"notes"`
    CreatedBy      string  `json:"created_by"`
    CreatedAt      string  `json:"created_at"`
    Items          []PurchaseItemDetail `json:"items,omitempty"`
}

type PurchaseItemDetail struct {
    ProductName string  `json:"product_name"`
    Barcode     string  `json:"barcode"`
    Quantity    int     `json:"quantity"`
    CostPrice   float64 `json:"cost_price"`
    Subtotal    float64 `json:"subtotal"`
}

func CreatePurchase(db *sql.DB, purchase *Purchase) error {
    tx, err := db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Generate purchase number
    purchaseNumber := "PO-" + time.Now().Format("20060102") + "-" + fmt.Sprintf("%04d", time.Now().UnixNano()%10000)

    // Calculate total
    var totalItems int
    var totalCost float64
    for _, item := range purchase.Items {
        totalItems += item.Quantity
        totalCost += float64(item.Quantity) * item.CostPrice
    }

    // Insert purchase - WITHOUT shop_id
    query := `
        INSERT INTO purchases (purchase_number, supplier_id, total_items, total_cost, 
                               purchase_date, notes, created_by, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
    `
    result, err := tx.Exec(query,
        purchaseNumber,
        purchase.SupplierID,
        totalItems,
        totalCost,
        purchase.PurchaseDate,
        purchase.Notes,
        purchase.CreatedBy,
    )
    if err != nil {
        return fmt.Errorf("failed to insert purchase: %w", err)
    }

    purchaseID, err := result.LastInsertId()
    if err != nil {
        return fmt.Errorf("failed to get purchase ID: %w", err)
    }
    purchase.ID = int(purchaseID)
    purchase.PurchaseNumber = purchaseNumber
    purchase.TotalItems = totalItems
    purchase.TotalCost = totalCost

    // Always add stock to MAIN SHOP (shop_id = 1)
    const mainShopID = 1

    // Insert purchase items and update stock
    for _, item := range purchase.Items {
        // Insert purchase item
        itemQuery := `
            INSERT INTO purchase_items (purchase_id, product_id, quantity, cost_price, subtotal)
            VALUES (?, ?, ?, ?, ?)
        `
        subtotal := float64(item.Quantity) * item.CostPrice
        _, err := tx.Exec(itemQuery, purchaseID, item.ProductID, item.Quantity, item.CostPrice, subtotal)
        if err != nil {
            return fmt.Errorf("failed to insert purchase item: %w", err)
        }

        // Add to MAIN SHOP stock (shop_id = 1)
        _, err = tx.Exec(`
            INSERT INTO shop_stock (shop_id, product_id, quantity, created_at, updated_at)
            VALUES (?, ?, ?, NOW(), NOW())
            ON DUPLICATE KEY UPDATE quantity = quantity + ?, updated_at = NOW()
        `, mainShopID, item.ProductID, item.Quantity, item.Quantity)
        if err != nil {
            return fmt.Errorf("failed to update stock: %w", err)
        }

        // Update product stock quantity and cost price (weighted average)
        var currentStock int
        var currentCost float64
        err = tx.QueryRow(`
            SELECT COALESCE(stock_quantity, 0), COALESCE(cost_price, 0) 
            FROM products WHERE id = ?
        `, item.ProductID).Scan(&currentStock, &currentCost)
        if err != nil && err != sql.ErrNoRows {
            return fmt.Errorf("failed to get product current values: %w", err)
        }

        // Calculate new weighted average cost
        newTotalCost := (currentCost * float64(currentStock)) + (item.CostPrice * float64(item.Quantity))
        newStock := currentStock + item.Quantity
        newCost := newTotalCost / float64(newStock)

        // Update product
        _, err = tx.Exec(`
            UPDATE products 
            SET cost_price = ?, stock_quantity = ?
            WHERE id = ?
        `, newCost, newStock, item.ProductID)
        if err != nil {
            return fmt.Errorf("failed to update product: %w", err)
        }
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit purchase: %w", err)
    }

    return nil
}

func GetPurchaseReport(db *sql.DB, startDate, endDate string, supplierID int) ([]PurchaseReportItem, float64, error) {
    var conditions []string
    var args []interface{}

    conditions = append(conditions, "DATE(p.created_at) BETWEEN ? AND ?")
    args = append(args, startDate, endDate)

    if supplierID > 0 {
        conditions = append(conditions, "p.supplier_id = ?")
        args = append(args, supplierID)
    }

    whereClause := ""
    for i, cond := range conditions {
        if i == 0 {
            whereClause += " WHERE " + cond
        } else {
            whereClause += " AND " + cond
        }
    }

    query := `
        SELECT 
            p.id,
            p.purchase_number,
            s.name as supplier_name,
            p.total_items,
            p.total_cost,
            p.purchase_date,
            p.notes,
            p.created_by,
            p.created_at
        FROM purchases p
        JOIN suppliers s ON p.supplier_id = s.id
    ` + whereClause + `
        ORDER BY p.created_at DESC
    `

    rows, err := db.Query(query, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to get purchase report: %w", err)
    }
    defer rows.Close()

    var purchases []PurchaseReportItem
    var grandTotal float64

    for rows.Next() {
        var p PurchaseReportItem
        err := rows.Scan(
            &p.ID,
            &p.PurchaseNumber,
            &p.SupplierName,
            &p.TotalItems,
            &p.TotalCost,
            &p.PurchaseDate,
            &p.Notes,
            &p.CreatedBy,
            &p.CreatedAt,
        )
        if err != nil {
            return nil, 0, err
        }
        grandTotal += p.TotalCost
        purchases = append(purchases, p)
    }

    if err := rows.Err(); err != nil {
        return nil, 0, fmt.Errorf("error iterating purchases: %w", err)
    }

    return purchases, grandTotal, nil
}

func GetPurchaseItems(db *sql.DB, purchaseID int) ([]PurchaseItemDetail, error) {
    query := `
        SELECT 
            p.name as product_name,
            COALESCE(p.barcode, '') as barcode,
            pi.quantity,
            pi.cost_price,
            pi.subtotal
        FROM purchase_items pi
        JOIN products p ON pi.product_id = p.id
        WHERE pi.purchase_id = ?
    `
    rows, err := db.Query(query, purchaseID)
    if err != nil {
        return nil, fmt.Errorf("failed to get purchase items: %w", err)
    }
    defer rows.Close()

    var items []PurchaseItemDetail
    for rows.Next() {
        var item PurchaseItemDetail
        err := rows.Scan(
            &item.ProductName,
            &item.Barcode,
            &item.Quantity,
            &item.CostPrice,
            &item.Subtotal,
        )
        if err != nil {
            return nil, err
        }
        items = append(items, item)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating purchase items: %w", err)
    }

    return items, nil
}

func GetSuppliers(db *sql.DB) ([]Supplier, error) {
    query := `
        SELECT id, name, contact_person, phone, email, address, notes, is_active
        FROM suppliers
        WHERE is_active = 1
        ORDER BY name
    `
    rows, err := db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("failed to get suppliers: %w", err)
    }
    defer rows.Close()

    var suppliers []Supplier
    for rows.Next() {
        var s Supplier
        var contactPerson, phone, email, address, notes sql.NullString
        
        err := rows.Scan(
            &s.ID,
            &s.Name,
            &contactPerson,
            &phone,
            &email,
            &address,
            &notes,
            &s.IsActive,
        )
        if err != nil {
            return nil, err
        }
        
        s.ContactPerson = contactPerson.String
        s.Phone = phone.String
        s.Email = email.String
        s.Address = address.String
        s.Notes = notes.String
        
        suppliers = append(suppliers, s)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating suppliers: %w", err)
    }

    return suppliers, nil
}

func CreateSupplier(db *sql.DB, supplier *Supplier) error {
    query := `
        INSERT INTO suppliers (name, contact_person, phone, email, address, notes, is_active)
        VALUES (?, ?, ?, ?, ?, ?, 1)
    `
    result, err := db.Exec(query,
        supplier.Name,
        supplier.ContactPerson,
        supplier.Phone,
        supplier.Email,
        supplier.Address,
        supplier.Notes,
    )
    if err != nil {
        return fmt.Errorf("failed to create supplier: %w", err)
    }

    id, err := result.LastInsertId()
    if err == nil {
        supplier.ID = int(id)
    }

    return nil
}
