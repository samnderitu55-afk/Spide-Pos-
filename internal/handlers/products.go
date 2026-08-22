package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "strings"
    "spide-pos/internal/db"
)

func GetProductsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    products, err := db.GetAllProducts(db.GetDB())
    if err != nil {
        http.Error(w, `{"error":"Failed to load products"}`, http.StatusInternalServerError)
        return
    }
    if products == nil {
        products = []db.Product{}
    }
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
    q := r.URL.Query().Get("q")
    if strings.TrimSpace(q) == "" {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode([]db.Product{})
        return
    }

    products, err := db.SearchProducts(db.GetDB(), q)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if products == nil {
        products = []db.Product{}
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(products)
}

func ScanProductHandler(w http.ResponseWriter, r *http.Request) {
    barcode := r.URL.Query().Get("barcode")
    qtyStr := r.URL.Query().Get("qty")

    qty := 1
    if parsedQty, err := strconv.Atoi(qtyStr); err == nil && parsedQty > 0 {
        qty = parsedQty
    }

    product, err := db.GetProductByBarcode(db.GetDB(), barcode, qty)
    if err != nil {
        w.WriteHeader(http.StatusNotFound)
        w.Write([]byte(`<tr><td colspan="7" class="p-3 text-center text-red-500 font-medium">⚠️ Product not found</td></tr>`))
        return
    }

    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(product.RenderRowHTML()))
}

