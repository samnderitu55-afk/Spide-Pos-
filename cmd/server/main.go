package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "os/signal"
    "path/filepath"
    "syscall"
    "time"

    "github.com/joho/godotenv"

    "spide-pos/internal/db"
    "spide-pos/internal/handlers"
)

var startTime = time.Now()

func main() {
    // Load .env file
    envPath := filepath.Join(".", ".env")
    if err := godotenv.Load(envPath); err != nil {
        log.Printf("⚠️ No .env file found at %s, using environment variables", envPath)
    } else {
        log.Printf("✅ .env file loaded from %s", envPath)
    }

    // Database configuration
    dbConfig := db.Config{
        User:     getEnv("DB_USER", "root"),
        Password: getEnv("DB_PASSWORD", "Tende@2016"),
        Host:     getEnv("DB_HOST", "127.0.0.1"),
        Port:     getEnv("DB_PORT", "3306"),
        Name:     getEnv("DB_NAME", "spide_pos"),
    }

    log.Printf("🔐 Connecting to database: %s@%s:%s/%s", dbConfig.User, dbConfig.Host, dbConfig.Port, dbConfig.Name)

    // Connect to database
    if _, err := db.ConnectDB(dbConfig); err != nil {
        log.Fatalf("Database connection failed: %v", err)
    }
    defer db.CloseDB()

    // Setup routes - NO AUTH
    mux := http.NewServeMux()

    // Favicon
    mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "favicon.ico")
    })

    // Health check
    mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        if err := db.GetDB().Ping(); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "status":    "unhealthy",
                "error":     err.Error(),
                "timestamp": time.Now().Format(time.RFC3339),
            })
            return
        }
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status":    "healthy",
            "timestamp": time.Now().Format(time.RFC3339),
            "uptime":    time.Since(startTime).String(),
        })
    })

    // Login routes (keep for later)
    mux.HandleFunc("/login", handlers.ServeLogin)
    mux.HandleFunc("/api/login", handlers.LoginHandler)
    mux.HandleFunc("/api/logout", handlers.LogoutHandler)

    // HTML Pages - NO AUTH
    mux.HandleFunc("/", handlers.ServeDashboard)
    mux.HandleFunc("/pos", handlers.ServePOS)
    mux.HandleFunc("/director", handlers.ServeDirector)

    // API Routes - NO AUTH - DIRECT HANDLERS
    mux.HandleFunc("/api/dashboard/stats", handlers.DashboardStatsHandler)
    mux.HandleFunc("/api/products", handlers.GetProductsHandler)
    mux.HandleFunc("/api/products/create", handlers.CreateProductHandler)
    mux.HandleFunc("/api/products/update", handlers.UpdateProductHandler)
    mux.HandleFunc("/api/products/search", handlers.SearchProductsHandler)
    mux.HandleFunc("/api/products/scan-html", handlers.ScanProductHandler)
    mux.HandleFunc("/api/sales/checkout", handlers.CheckoutHandler)
    mux.HandleFunc("/api/sales/recent", handlers.RecentSalesHandler)
    mux.HandleFunc("/api/sales/z-report", handlers.ZReportHandler)
    mux.HandleFunc("/api/reports/product-sales", handlers.ProductSalesReportHandler)
    mux.HandleFunc("/api/reports/low-stock", handlers.LowStockReportHandler)
    mux.HandleFunc("/api/reports/inventory-valuation", handlers.InventoryValuationHandler)
    mux.HandleFunc("/api/reports/inventory-valuation/details", handlers.CategoryDrilldownHandler)
    mux.HandleFunc("/api/expenses/create", handlers.CreateExpenseHandler)
    mux.HandleFunc("/api/expenses/report", handlers.ExpenseReportHandler)
    mux.HandleFunc("/api/expenses/categories", handlers.ExpenseCategoriesHandler)
    mux.HandleFunc("/api/expenses/delete", handlers.DeleteExpenseHandler)
    mux.HandleFunc("/api/transfers/create", handlers.CreateTransferHandler)
    mux.HandleFunc("/api/transfers", handlers.GetTransfersHandler)
    mux.HandleFunc("/api/transfers/detail", handlers.GetTransferDetailHandler)
    mux.HandleFunc("/api/purchases/create", handlers.CreatePurchaseHandler)
    mux.HandleFunc("/api/suppliers", handlers.GetSuppliersHandler)
    mux.HandleFunc("/api/suppliers/create", handlers.CreateSupplierHandler)
    mux.HandleFunc("/api/categories", handlers.GetCategoriesHandler)
    mux.HandleFunc("/api/categories/create", handlers.CreateCategoryHandler)

    // Server configuration
    port := getEnv("PORT", "8081")
    server := &http.Server{
        Addr:         ":" + port,
        Handler:      mux,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Start server
    go func() {
        log.Printf("🕷️ Spide POS running on http://localhost:%s", port)
        log.Printf("📊 Dashboard: http://localhost:%s/", port)
        log.Printf("🛒 POS: http://localhost:%s/pos", port)
        log.Printf("🔐 Login: http://localhost:%s/login (disabled)", port)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("🛑 Shutting down Spide POS...")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server shutdown error: %v", err)
    }
    log.Println("✅ Spide POS stopped gracefully")
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
