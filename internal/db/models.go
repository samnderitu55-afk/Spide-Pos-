package db

import (
	"database/sql"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"
)

func FormatQty(q float64) string {
	s := strconv.FormatFloat(q, 'f', 3, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

type Product struct {
	ID               int64     `json:"id"`
	CompanyID        int       `json:"company_id"`
	Barcode          *string   `json:"barcode"`
	Name             string    `json:"name"`
	Category         string    `json:"category"`
	CategoryID       int       `json:"category_id"`
	CostPrice        float64   `json:"cost_price"`
	RetailPrice      float64   `json:"retail_price"`
	WholesalePrice   float64   `json:"wholesale_price"`
	WholesaleMinQty  float64   `json:"wholesale_min_qty"`
	StockQuantity    float64   `json:"stock_quantity"`
	ReorderLevel     float64   `json:"reorder_level"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	StockCap         float64   `json:"stock_cap"`
	ShopReorderLevel float64   `json:"shop_reorder_level"`
	UnitType         string    `json:"unit_type"`
	UnitLabel        *string   `json:"unit_label"`
}

func (p *Product) RenderRowHTML() string {
	barcodeDisplay := "N/A"
	if p.Barcode != nil && *p.Barcode != "" {
		barcodeDisplay = *p.Barcode
	}

	nameEsc := html.EscapeString(p.Name)
	barcodeEsc := html.EscapeString(barcodeDisplay)

	stockClass := "text-green-600"
	if p.StockQuantity <= 0 {
		stockClass = "text-red-500 font-bold"
	} else if p.StockQuantity <= p.ReorderLevel {
		stockClass = "text-amber-500 font-bold"
	}

	// Fractional step for unit-based products (e.g. refills sold in ml)
	step := "1"
	if p.UnitType == "unit" {
		step = "0.5"
	}

	// Unit label for display (empty for packages)
	unitLabel := ""
	if p.UnitLabel != nil {
		unitLabel = *p.UnitLabel
	}

	return fmt.Sprintf(`<tr data-product-id="%d" data-retail-price="%.2f" data-wholesale-price="%.2f" data-wholesale-threshold="%s" data-barcode="%s" data-stock="%s" data-unit-type="%s" data-unit-label="%s">
        <td class="p-3 font-medium text-gray-800 text-sm min-w-0 max-w-[220px]">%s</td>
        <td class="p-3 text-center font-mono %s">%s</td>
        <td class="p-3 text-right font-mono text-gray-500 text-xs whitespace-nowrap">KES %.2f</td>
        <td class="p-3 text-center">
            <div class="inline-flex items-center gap-1">
                <button type="button" onclick="adjustCartQty(this, -%s)"
                        class="w-7 h-7 rounded bg-gray-100 hover:bg-gray-200 text-gray-700 font-bold leading-none">−</button>
                <input type="number" class="cart-qty-input w-20 px-2 py-1 border rounded text-center focus:ring-2 focus:ring-purple-500 focus:outline-none"
                       value="1" min="%s" step="%s" max="%s" oninput="updateRowTotals(this, false)">
                <button type="button" onclick="adjustCartQty(this, %s)"
                        class="w-7 h-7 rounded bg-gray-100 hover:bg-gray-200 text-gray-700 font-bold leading-none">+</button>
            </div>
        </td>
        <td class="p-3 text-right">
            <div class="flex items-center justify-end gap-2">
                <span class="price-badge bg-gray-200 text-gray-700 text-xs px-2 py-0.5 rounded whitespace-nowrap">Retail</span>
                <input type="number" step="0.01" class="cart-price-input w-24 px-2 py-1 border rounded text-right focus:ring-2 focus:ring-purple-500 focus:outline-none"
                       value="%.2f" onfocus="this.select()" oninput="updateRowTotals(this, true)">
            </div>
        </td>
        <td class="subtotal-td p-3 text-right font-bold text-purple-900 whitespace-nowrap">KES %.2f</td>
        <td class="p-3 text-center">
            <button type="button" onclick="removeCartRow(this)"
                    class="text-red-500 hover:text-red-700 font-bold text-lg leading-none">✕</button>
        </td>
    </tr>`,
		p.ID,
		p.RetailPrice,
		p.WholesalePrice,
		FormatQty(p.WholesaleMinQty),
		barcodeEsc,
		FormatQty(p.StockQuantity),
		p.UnitType,
		unitLabel,
		nameEsc,
		stockClass,
		FormatQty(p.StockQuantity),
		p.CostPrice,
		step,
		step,
		step,
		FormatQty(p.StockQuantity),
		step,
		p.RetailPrice,
		p.RetailPrice,
	)
}

type ShopStock struct {
	ID        int       `json:"id"`
	CompanyID int       `json:"company_id"`
	ShopID    int       `json:"shop_id"`
	ProductID int       `json:"product_id"`
	Quantity  float64   `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Category struct {
	ID        int       `json:"id"`
	CompanyID int       `json:"company_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SaleItem struct {
	ID          int64   `json:"id"`
	SaleID      int64   `json:"sale_id"`
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Subtotal    float64 `json:"subtotal"`
}

type Sale struct {
	ID            int64          `json:"id"`
	CompanyID     int            `json:"company_id"`
	CustomerID    int            `json:"customer_id"`
	TotalAmount   float64        `json:"total_amount"`
	CashAmount    float64        `json:"cash_amount"`
	MpesaAmount   float64        `json:"mpesa_amount"`
	MpesaCode     sql.NullString `json:"mpesa_code"`
	DepositAmount float64        `json:"deposit_amount"`
	CreditAmount  float64        `json:"dcredit_amount"`
	PaymentType   string         `json:"payment_type"`
	ChangeGiven   float64        `json:"change_given,omitempty"`
	CreatedAt     string         `json:"created_at"`
	ShopID        int            `json:"shop_id"`
	Items         []SaleItem     `json:"items,omitempty"`
}

type SaleRequest struct {
	TotalAmount   float64 `json:"total_amount"`
	CashAmount    float64 `json:"cash_amount"`
	MpesaAmount   float64 `json:"mpesa_amount"`
	MpesaCode     string  `json:"mpesa_code"`
	PaymentType   string  `json:"payment_type"`
	ChangeGiven   float64 `json:"change_given,omitempty"`
	ShopID        int     `json:"shop_id"`
	CreditAmount  float64 `json:"credit_amount"`
	CustomerID    int     `json:"customer_id"`
	DepositAmount float64 `json:"deposit_amount"`
	Items         []struct {
		ProductID int64   `json:"product_id"`
		Quantity  float64 `json:"quantity"`
		UnitPrice float64 `json:"unit_price"`
		Subtotal  float64 `json:"subtotal"`
	} `json:"items"`
}

type Expense struct {
	ID            int     `json:"id"`
	Category      string  `json:"category"`
	Description   string  `json:"description"`
	Amount        float64 `json:"amount"`
	ExpenseDate   string  `json:"expense_date"`
	PaymentMethod string  `json:"payment_method"`
	Reference     string  `json:"reference"`
	Notes         string  `json:"notes"`
	ShopID        int     `json:"shop_id"`
	ShopName      string  `json:"shop_name"`
	CreatedBy     string  `json:"created_by"`
	CreatedAt     string  `json:"created_at"`
}

type ZReport struct {
	ReportDate      string  `json:"report_date"`
	TotalRevenue    float64 `json:"total_revenue"`
	TotalCost       float64 `json:"total_cost"`
	TotalProfit     float64 `json:"total_profit"`
	MarginPercent   float64 `json:"margin_percent"`
	TotalSalesCount int     `json:"total_sales_count"`

	// ✅ Add these fields for deposit and credit
	TotalCash    float64 `json:"total_cash"`
	TotalMpesa   float64 `json:"total_mpesa"`
	TotalDeposit float64 `json:"total_deposit"`
	TotalCredit  float64 `json:"total_credit"`

	CashSalesCount    int `json:"cash_sales_count"`
	MpesaSalesCount   int `json:"mpesa_sales_count"`
	DepositSalesCount int `json:"deposit_sales_count"`
	CreditSalesCount  int `json:"credit_sales_count"`

	TotalExpenses    float64            `json:"total_expenses"`
	ExpenseCount     int                `json:"expense_count"`
	ExpenseBreakdown map[string]float64 `json:"expense_breakdown"`
	NetProfit        float64            `json:"net_profit"`

	// Shop info
	ShopName string `json:"shop_name,omitempty"`
	ShopID   int    `json:"shop_id,omitempty"`
}

type ProductSalesReportItem struct {
	ProductID    int     `json:"product_id"`
	ProductName  string  `json:"product_name"`
	Category     string  `json:"category"`
	UnitsSold    float64 `json:"units_sold"`
	TotalCost    float64 `json:"total_cost"`
	TotalRevenue float64 `json:"total_revenue"`
	NetProfit    float64 `json:"net_profit"`
	MarginPct    float64 `json:"margin_pct"`
	CurrentStock float64 `json:"current_stock"`
	StockCap     int     `json:"stock_cap"`
}

type LowStockReportItem struct {
	ProductName   string  `json:"product_name"`
	Category      string  `json:"category"`
	StockQuantity float64 `json:"stock_quantity"`
	ReorderLevel  float64 `json:"reorder_level"`
	CostPrice     float64 `json:"cost_price"`
	RestockCost   float64 `json:"restock_cost"`
}

type InventoryValuationItem struct {
	Category        string  `json:"category"`
	TotalItems      int     `json:"total_items"`
	TotalQuantity   int     `json:"total_quantity"`
	TotalCost       float64 `json:"total_cost"`
	TotalRetail     float64 `json:"total_retail"`
	PotentialProfit float64 `json:"potential_profit"`
}

type StockTransferRequest struct {
	ToShopID     int    `json:"to_shop_id"`
	TransferDate string `json:"transfer_date"`
	Notes        string `json:"notes"`
	Items        []struct {
		ProductID int     `json:"product_id"`
		Quantity  int     `json:"quantity"`
		CostPrice float64 `json:"cost_price"`
	} `json:"items"`
}

type StockTransfer struct {
	ID             int                  `json:"id"`
	TransferNumber string               `json:"transfer_number"`
	CompanyID      int                  `json:"company_id"`
	FromShopID     int                  `json:"from_shop_id"`
	ToShopID       int                  `json:"to_shop_id"`
	TotalItems     int                  `json:"total_items"`
	TotalCost      float64              `json:"total_cost"`
	TransferDate   string               `json:"transfer_date"`
	Status         string               `json:"status"`
	Notes          string               `json:"notes"`
	CreatedBy      string               `json:"created_by"`
	CreatedAt      string               `json:"created_at"`
	Items          []TransferItemDetail `json:"items,omitempty"`
	FromShopName   string               `json:"from_shop_name,omitempty"`
	ToShopName     string               `json:"to_shop_name,omitempty"`
}

type TransferItemDetail struct {
	ID          int     `json:"id"`
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Barcode     string  `json:"barcode,omitempty"`
	Quantity    int     `json:"quantity"`
	CostPrice   float64 `json:"cost_price"`
	Subtotal    float64 `json:"subtotal"`
}

type PurchaseRequest struct {
	SupplierID   int    `json:"supplier_id"`
	PurchaseDate string `json:"purchase_date"`
	Notes        string `json:"notes"`
	Items        []struct {
		ProductID int     `json:"product_id"`
		Quantity  int     `json:"quantity"`
		CostPrice float64 `json:"cost_price"`
	} `json:"items"`
}

type Purchase struct {
	ID             int     `json:"id"`
	PurchaseNumber string  `json:"purchase_number"`
	SupplierID     int     `json:"supplier_id"`
	SupplierName   string  `json:"supplier_name"`
	TotalItems     float64 `json:"total_items"`
	TotalCost      float64 `json:"total_cost"`
	PurchaseDate   string  `json:"purchase_date"`
	Notes          string  `json:"notes"`
	CreatedBy      string  `json:"created_by"`
	CreatedAt      string  `json:"created_at"`
	ShopID         int     `json:"shop_id"`
	Items          []struct {
		ProductID int     `json:"product_id"`
		Quantity  float64 `json:"quantity"`
		CostPrice float64 `json:"cost_price"`
	} `json:"items,omitempty"`
}

type Supplier struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	ContactPerson string `json:"contact_person"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Notes         string `json:"notes"`
	IsActive      bool   `json:"is_active"`
}

type DashboardStats struct {
	TodaySales struct {
		TotalRevenue     float64 `json:"total_revenue"`
		TransactionCount int     `json:"transaction_count"`
		AverageTicket    float64 `json:"average_ticket"`
		TotalItems       int     `json:"total_items"`
	} `json:"today_sales"`
	YesterdaySales struct {
		TotalRevenue     float64 `json:"total_revenue"`
		TransactionCount int     `json:"transaction_count"`
	} `json:"yesterday_sales"`
	QuickStats struct {
		TotalProducts   int     `json:"total_products"`
		TotalStockValue float64 `json:"total_stock_value"`
	} `json:"quick_stats"`
	SalesTrend []struct {
		Hour   int     `json:"hour"`
		Amount float64 `json:"amount"`
		Count  int     `json:"count"`
	} `json:"sales_trend"`
	TopProducts []struct {
		ProductName string  `json:"product_name"`
		UnitsSold   float64 `json:"units_sold"`
		Revenue     float64 `json:"revenue"`
	} `json:"top_products"`
	RecentSales   []Sale               `json:"recent_sales"`
	LowStockItems []LowStockReportItem `json:"low_stock_items"`
}
