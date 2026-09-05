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
	"spide-pos/internal/middleware"
)

var startTime = time.Now()

func main() {
	// Load .env file
	envPath := filepath.Join(".", ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("No .env file found, using environment variables")
	} else {
		log.Printf("✅ .env file loaded")
	}

	// Database configuration
	dbConfig := db.Config{
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", "Tende@2016"),
		Host:     getEnv("DB_HOST", "127.0.0.1"),
		Port:     getEnv("DB_PORT", "3306"),
		Name:     getEnv("DB_NAME", "spide_pos"),
	}

	log.Printf("Connecting to database...")

	// Connect to database
	if _, err := db.ConnectDB(dbConfig); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.CloseDB()

	// Setup routes
	mux := http.NewServeMux()

	// Favicon
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "favicon.ico")
	})

	// Health check (public)
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

	// Login routes (public)
	mux.HandleFunc("/login", handlers.ServeLogin)
	mux.HandleFunc("/api/login", handlers.LoginHandler)

	// Protected routes (require authentication)
	mux.HandleFunc("/", middleware.AuthMiddleware(handlers.ServeDashboard))
	mux.HandleFunc("/pos", middleware.AuthMiddleware(handlers.ServePOS))
	mux.HandleFunc("/director", middleware.AuthMiddleware(handlers.ServeDirector))
	mux.HandleFunc("/api/logout", middleware.AuthMiddleware(handlers.LogoutHandler))

	// API Routes - Protected
	mux.HandleFunc("/api/dashboard/stats", middleware.AuthMiddleware(handlers.DashboardStatsHandler))
	mux.HandleFunc("/api/products", middleware.AuthMiddleware(handlers.GetProductsHandler))
	mux.HandleFunc("/api/products/create", middleware.AuthMiddleware(handlers.CreateProductHandler))
	mux.HandleFunc("/api/products/update", middleware.AuthMiddleware(handlers.UpdateProductHandler))
	mux.HandleFunc("/api/products/search", middleware.AuthMiddleware(handlers.SearchProductsHandler))
	mux.HandleFunc("/api/products/scan-html", middleware.AuthMiddleware(handlers.ScanProductHandler))
	mux.HandleFunc("/api/sales/checkout", middleware.AuthMiddleware(handlers.CheckoutHandler))
	mux.HandleFunc("/api/sales/recent", middleware.AuthMiddleware(handlers.RecentSalesHandler))
	mux.HandleFunc("/api/sales/z-report", middleware.AuthMiddleware(handlers.ZReportHandler))
	mux.HandleFunc("/api/reports/product-sales", middleware.AuthMiddleware(handlers.ProductSalesReportHandler))
	mux.HandleFunc("/api/reports/low-stock", middleware.AuthMiddleware(handlers.LowStockReportHandler))
	mux.HandleFunc("/api/reports/inventory-valuation", middleware.AuthMiddleware(handlers.InventoryValuationHandler))
	mux.HandleFunc("/api/reports/inventory-valuation/details", middleware.AuthMiddleware(handlers.CategoryDrilldownHandler))
	mux.HandleFunc("/api/expenses/create", middleware.AuthMiddleware(handlers.CreateExpenseHandler))
	mux.HandleFunc("/api/expenses/report", middleware.AuthMiddleware(handlers.ExpenseReportHandler))
	mux.HandleFunc("/api/expenses/categories", middleware.AuthMiddleware(handlers.ExpenseCategoriesHandler))
	mux.HandleFunc("/api/expenses/delete", middleware.AuthMiddleware(handlers.DeleteExpenseHandler))
	mux.HandleFunc("/api/transfers/create", middleware.AuthMiddleware(handlers.CreateTransferHandler))
	mux.HandleFunc("/api/transfers", middleware.AuthMiddleware(handlers.GetTransfersHandler))
	mux.HandleFunc("/api/transfers/detail", middleware.AuthMiddleware(handlers.GetTransferDetailHandler))
	mux.HandleFunc("/api/purchases/create", middleware.AuthMiddleware(handlers.CreatePurchaseHandler))
	mux.HandleFunc("/api/purchases/report", middleware.AuthMiddleware(handlers.PurchaseReportHandler))
	mux.HandleFunc("/api/purchases/items", middleware.AuthMiddleware(handlers.GetPurchaseItemsHandler))
	mux.HandleFunc("/api/suppliers", middleware.AuthMiddleware(handlers.GetSuppliersHandler))
	mux.HandleFunc("/api/suppliers/create", middleware.AuthMiddleware(handlers.CreateSupplierHandler))
	mux.HandleFunc("/api/categories", middleware.AuthMiddleware(handlers.GetCategoriesHandler))
	mux.HandleFunc("/api/categories/create", middleware.AuthMiddleware(handlers.CreateCategoryHandler))
	mux.HandleFunc("/api/branches", middleware.AuthMiddleware(handlers.BranchesHandler))

	// Director Dashboard
	mux.HandleFunc("/api/director/dashboard", middleware.AuthMiddleware(handlers.DirectorDashboardHandler))

	// User management (protected - director only)
	mux.HandleFunc("/api/users", middleware.AuthMiddleware(handlers.GetUsersHandler))
	mux.HandleFunc("/api/users/create", middleware.AuthMiddleware(handlers.CreateUserHandler))
	mux.HandleFunc("/api/users/update", middleware.AuthMiddleware(handlers.UpdateUserHandler))
	mux.HandleFunc("/api/users/delete", middleware.AuthMiddleware(handlers.DeleteUserHandler))

	// Customer deposit routes
	mux.HandleFunc("/api/customers/deposit", middleware.AuthMiddleware(handlers.AddCustomerDepositHandler))
	mux.HandleFunc("/api/customers/balance", middleware.AuthMiddleware(handlers.GetCustomerBalanceHandler))
	mux.HandleFunc("/api/customers/transactions", middleware.AuthMiddleware(handlers.GetCustomerTransactionsHandler))

	// Customer routes
	mux.HandleFunc("/api/customers", middleware.AuthMiddleware(handlers.GetCustomersHandler))
	mux.HandleFunc("/api/customers/get", middleware.AuthMiddleware(handlers.GetCustomerHandler))
	mux.HandleFunc("/api/customers/create", middleware.AuthMiddleware(handlers.CreateCustomerHandler))
	mux.HandleFunc("/api/customers/update", middleware.AuthMiddleware(handlers.UpdateCustomerHandler))
	mux.HandleFunc("/api/customers/delete", middleware.AuthMiddleware(handlers.DeleteCustomerHandler))
	mux.HandleFunc("/api/customers/search", middleware.AuthMiddleware(handlers.SearchCustomersHandler))
	

	// Credit sales routes
	mux.HandleFunc("/api/credit-sales", middleware.AuthMiddleware(handlers.GetCreditSalesHandler))
	mux.HandleFunc("/api/credit-sales/create", middleware.AuthMiddleware(handlers.CreateCreditSaleHandler))
	mux.HandleFunc("/api/credit-sales/payments", middleware.AuthMiddleware(handlers.GetCreditPaymentsHandler))
	mux.HandleFunc("/api/credit-sales/add-payment", middleware.AuthMiddleware(handlers.AddCreditPaymentHandler))
	
	
	//customer statement route
	mux.HandleFunc("/api/customers/statement", middleware.AuthMiddleware(handlers.GetCustomerStatementHandler))
	

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
		log.Printf("🔐 Login: http://localhost:%s/login", port)
		log.Printf("📊 Dashboard: http://localhost:%s/", port)
		log.Printf("🛒 POS: http://localhost:%s/pos", port)
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
