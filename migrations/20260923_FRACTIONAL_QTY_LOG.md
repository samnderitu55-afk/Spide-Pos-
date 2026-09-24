# Fractional Quantities Migration — 2026-09-23

## Goal
Allow Perfume Re-fill products (category_id 46, company 2) to be sold
in fractional ml quantities (0.5, 1, 1.5, 3, 6 ml), while all other
products remain whole-unit.

## Status
- Local (Windows): COMPLETE and verified end-to-end
- Production (pos.spide.co.ke): PENDING — scheduled 2026-09-23 22:00 EAT

---

## DDL changes applied to LOCAL database

Migration file 1: `migrations/20260923_fractional_qty.sql`
Migration file 2: `migrations/20260923_fractional_totals.sql`

### Type changes (INT → DECIMAL(10,3))
| Table | Column |
|---|---|
| shop_stock | quantity |
| sale_items | quantity |
| transfer_items | quantity |
| purchase_items | quantity |
| products | stock_quantity |
| products | wholesale_min_qty |
| purchases | total_items |
| stock_transfers | total_items |

### New columns on `products`
- `unit_type` ENUM('package','unit') NOT NULL DEFAULT 'package'
- `unit_label` VARCHAR(20) NULL

### New constraint
- `chk_unit_label` — enforces (unit_type='unit' ↔ unit_label NOT NULL)
  and (unit_type='package' ↔ unit_label IS NULL)

### Data backfill
- 97 rows with company_id=2 AND category_id=46 → unit_type='unit', unit_label='ml'
- All other company-2 products remain unit_type='package', unit_label=NULL
- No drift between `category` (varchar) and `category_id` (int) for the refill set

---

## Go changes

### `internal/db/models.go`
- `FormatQty(q float64) string` helper — trims trailing zeros (`1.5`, `3`, `0.5`)
- `Product` struct: all quantity fields widened to `float64`
  (`WholesaleMinQty`, `StockQuantity`, `ReorderLevel`, `StockCap`, `ShopReorderLevel`)
- `Product` struct: added `UnitType string` and `UnitLabel *string`
- `ShopStock`, `SaleItem`, `SaleRequest.Items`, `Purchase`, `StockTransfer`,
  `TransferItem`, `StockTransferRequest.Items`, `TransferItemDetail`,
  `PurchaseRequest.Items` — all quantity fields widened to `float64`
- `RenderRowHTML()` — added `data-unit-type`, `data-unit-label`, conditional
  `min`/`step` (`0.5` for unit products, `1` for packages), widened input to `w-20`,
  passes step to `adjustCartQty`

### `internal/db/*.go` — quantity scan targets widened
- `dashboard.go`: `UnitsSold` ×2
- `director.go`: `TodayItems` ×2, `UnitsSold` ×2, `StockLevel` ×2, `ReorderLevel` ×2
- `inventory_valuation.go`: `var totalQuantity` local
- `models.go`: `UnitsSold`, `StockQuantity`, `TotalQuantity`, `UnitsSold`
- `products.go`: `GetProductByBarcodeAndShop` now selects + scans `unit_type`, `unit_label`
- `purchases.go`: `var totalItems`, `var currentStock` widened
- `reports.go`: `StockQuantity`

### `internal/handlers/*.go`
- `products.go`: `qty` and `req.Quantity` widened; `StockCap`/`ReorderLevel`
  in inline struct → `*float64`; format strings fixed (`%d` → `%s` + `FormatQty`)
- `purchases.go`: inline `PurchaseRequest.Items[].Quantity` → `float64`
- `reports.go`: `StockQuantity` → `float64`

### Frontend — `internal/templates/pos.html`
All cart quantity `parseInt` → `parseFloat` (lines 586, 587, 595, 614, 619,
686, 722, 860). Two `< 1` clamps → `< 0.5` (lines 688, 725).
Line 685 `parseInt` on `data-stock` → `parseFloat`.

---

## Verified locally (2026-09-23)
- 1.5ml refill sale: cart computes correctly, checkout succeeds,
  `sale_items.quantity = 1.500`, `shop_stock.quantity` decrements by 1.5
- Receipt prints `1.5 x KES 33.40`
- Recent Sales + reprint display `1.5` correctly
- Package product sale: `2.000`, receipt shows `2`
- `go build ./...` clean, `go vet ./...` clean
- Director dashboard loads without scan errors

---

## KNOWN GAPS — must be resolved before production deploy
### GAP-1: RESOLVED 2026-09-24
All 10 product-loading functions in internal/db/products.go now
select and scan unit_type/unit_label. Verified via POS search
dropdown, product catalogue, and 1.5ml checkout. Commit: <your new SHA>

### GAP-2: 'ml' label not displayed on cart rows or receipts
The data is available (`data-unit-label="ml"` on the cart row,
`item.unit_label` in the API response). The frontend just doesn't
read it yet. Receipt shows `1.5 x KES 33.40` — should be `1.5 ml x ...`.
Cart row shows `1.5` — should show `1.5 ml`.

### GAP-3: spide-loaders.js transfer flow still uses parseInt
Lines 714, 720, 723, 749, 800, 824 (approximately). Only matters if
fractional stock transfers between shops are needed. Deferred.

---

## Production deploy plan (2026-09-23, 22:00 EAT)

### Pre-deploy (no downtime)
1. mysqldump production → /root/backups/spide_pos_pre_fractional_<ts>.sql
2. Build Linux binary locally:
   $env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
   go build -o spide-pos cmd/server/main.go
3. scp binary → root@64.226.124.37:/tmp/spide-pos.new
4. scp changed templates/JS → /tmp/*.new
5. scp migration SQL files → /tmp/

### Deploy (downtime ~2-3 min)
6. systemctl stop spide-pos
7. Second mysqldump (point-in-time) → /root/backups/
8. mysql spide_pos < /tmp/20260923_fractional_qty.sql
9. mysql spide_pos < /tmp/20260923_fractional_totals.sql
10. Verify:
    SELECT COLUMN_TYPE FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA='spide_pos' AND COLUMN_NAME='quantity'
       AND TABLE_NAME IN ('shop_stock','sale_items','transfer_items','purchase_items');
    -- all should be decimal(10,3)
    SELECT unit_type, unit_label, COUNT(*) FROM products
     WHERE company_id=2 GROUP BY unit_type, unit_label;
    -- expect: unit|ml|97 and package|NULL|~1651
11. Backup old binary, swap new binary, swap templates
12. systemctl start spide-pos
13. systemctl status spide-pos --no-pager
14. journalctl -u spide-pos -n 50 --no-pager

### Post-deploy verification
15. Log in on production
16. 1.5ml refill sale via scan → verify receipt, verify DB
17. Package sale qty 2 → verify still works

### Rollback (if step 13 or 16 fails)
systemctl stop spide-pos
mysql spide_pos < /root/backups/spide_pos_pre_fractional_<ts>.sql
cp /root/backups/spide-pos.<ts> /var/www/spide/spide-pos
cp /root/backups/pos.html.<ts> /var/www/spide/internal/templates/pos.html
# ...other changed templates...
systemctl start spide-pos

## Post-deploy backlog (not blocking tonight)
- GAP-2: 'ml' label on cart rows and receipts
- GAP-3: fractional transfer flow
- Refill pricing tiers (3ml=100, 6ml=150) — deferred per handoff