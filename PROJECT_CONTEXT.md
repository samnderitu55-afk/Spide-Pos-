# Spide POS — Project Context

## Stack
- **Backend**: Go 1.23+, MySQL 8 (`spide_pos` database)
- **Frontend**: HTML + Tailwind (CDN) + vanilla JS. No build step, no framework.
- **Server**: `cmd/server/main.go`, runs on `:8081`, serves static files from `static/`
- **Repo**: https://github.com/samnderitu55-afk/Spide-Pos-

## Current deployment (as of 2026-09-18)

- **Production server**: `64.226.124.37` (DigitalOcean Frankfurt, `Spide-V2`, Ubuntu 24.04)
- **Domain**: `pos.spide.co.ke` (also `pos2.spide.co.ke`)
- **App path**: `/var/www/spide/spide-pos` (Linux binary)
- **Working directory**: `/var/www/spide`
- **Systemd unit**: `/etc/systemd/system/spide-pos.service` (runs as root)
- **nginx**: reverse proxy on 443 → localhost:8081, gzip enabled, HTTP/2 enabled
- **MySQL**: localhost, user `root`, password `Tende@2016`, DB `spide_pos`
- **Backups**: `/root/backups/` — multiple tarballs and dumps from the migration

### Active companies in DB
- **id 1**: Demo Retail Ltd (demo data, 5 products, 6 demo users)
- **id 2**: Glow Age Beauty & Cosmetics (live, ~1690 products, 2 users: David, Precious)

## Local dev environment (Windows)

- **Path**: `C:\Users\sam\spide-pos`
- **MySQL**: local 8.0, root password matches server (`Tende@2016`)
- **DB**: `spide_pos` — imported from server dump, refreshed manually when needed
- **Build for local run**: `go build -o spide-pos.exe cmd/server/main.go`
- **Live reload tool**: `air` (config in `.air.toml`, watches Go files only)
- **`.env`**: local copy in project root, points at `127.0.0.1`, port 8081

## Directory layout
```
spide-pos/
├── cmd/
│   ├── server/main.go          # entry point, route registration
│   ├── checkpw/                # dev tool: verify password hashes
│   └── hashgen/                # dev tool: generate bcrypt hashes
├── internal/
│   ├── auth/                   # JWT generation/validation, bcrypt helpers
│   ├── db/                     # all SQL queries, one file per domain
│   │   ├── models.go           # structs + Product.RenderRowHTML()
│   │   ├── sales.go            # sale queries, GetRecentSales*
│   │   ├── dashboard.go        # dashboard stats, getSaleItems() helper
│   │   ├── shops.go            # shop queries (COALESCE on nullable cols)
│   │   └── ...
│   ├── handlers/               # HTTP handlers, one file per domain
│   ├── middleware/             # JWT auth middleware
│   └── templates/              # HTML pages served by Go (read from disk)
│       ├── login.html
│       ├── pos.html            # biggest page — has inline JS for cart, checkout, print
│       ├── dashboard.html
│       ├── director.html
│       └── modals.html         # source of truth for modal HTML
├── static/
│   └── js/
│       ├── spide-modals.js     # all modals + open/close functions
│       └── spide-loaders.js    # shared loaders, saveProduct, printThermalReceipt
├── go.mod / go.sum
└── PROJECT_CONTEXT.md          # this file
```

## Architecture rules (learned the hard way)

### Frontend
1. **All modals live in `spide-modals.js`** — single `MODAL_HTML` template literal + open/close functions. New modal → add to `MODAL_HTML` + add `window.openXModal` / `window.closeXModal`.
2. **All shared data loaders live in `spide-loaders.js`**, wrapped as `window.loadX = async function () { ... }`. Page files should NOT define `loadUsers`, `loadProducts`, `saveUser`, etc.
3. **Page files keep only page-specific logic**: cart state, search dropdown, checkout flow, dashboard KPI rendering.
4. **Modals auto-mount**: `SpideModals.mount()` runs on `DOMContentLoaded`.
5. **`spide-modals.js` loads BEFORE `spide-loaders.js`**. Both load after page-specific inline scripts (so loaders override page definitions).

### Backend
1. **Every handler reads `company_id` from JWT claims** via `middleware.GetUserFromContext(r)`.
2. **Every DB write filters by `company_id`** in the WHERE clause.
3. **Handlers return JSON errors**, not `http.Error` plain text. Use `writeJSONError` helpers.
4. **DB functions return `[]T{}` not `var x []T`** so empty results serialize to `[]` not `null`.
5. **SQL uses snake_case column names**.
6. **`COALESCE` for nullable columns** when scanning into plain `string`/`int`. Specifically: `email`, `shop_id`, `last_login`, `mpesa_code`, `location`, `phone`.
7. **Handlers return the full updated object** on success.
8. **HTML escaping in templates**: use `html.EscapeString()` on any user-provided string that goes into an HTML attribute (product names, barcodes, shop names).

### Database
- `users.company_id` — NOT NULL for active users
- `products.category` (string) AND `products.category_id` (int, nullable) — both used
- `products.barcode` — was nullable, now every product has one (real or `SKU-XXXXXXXX` synthetic)
- `sales.deposit_amount` — added mid-project
- `shops` table has NO `is_active` column
- Foreign keys are strict — always set `company_id` before insert

### Frontend ↔ Backend contract
- Login: POST `/api/login` → sets `spide_token` and `spide_user` cookies
- Every authenticated fetch sends `Authorization: Bearer <token>` header
- Update endpoints use `PUT`, create uses `POST`, delete uses `DELETE`
- Error responses: `{"error": "message"}`
- Sale checkout: POST `/api/sales/checkout` returns `{ sale_id, ... }`
- Product HTML injection: GET `/api/products/scan-html?barcode=X&qty=N&shop_id=S` returns a `<tr>` row to inject into the cart

## What's working (as of 2026-09-18)

- ✅ Login / logout with JWT (users: David/director, Precious/cashier on company 2)
- ✅ POS cart with +/− qty buttons, working delete, select-all on price click
- ✅ Checkout with split payments (cash/mpesa/deposit/credit)
- ✅ Payment modal opens with all fields at 0.00
- ✅ Receipts: bold items, Branch line, "You were served by: [username]", date+receipt# side by side
- ✅ Thermal receipt auto-prints on every sale
- ✅ Reprint from Recent Sales shows items correctly
- ✅ Dashboard KPIs, sales trend, top products
- ✅ Director dashboard (outlets, 30-day trend, alerts)
- ✅ Products: create, edit, per-shop Quick Stock Update
- ✅ User Management
- ✅ Low Stock report
- ✅ Inventory Valuation report
- ✅ Product Sales report
- ✅ Z-Report
- ✅ Recent Sales modal + reprint
- ✅ gzip compression on nginx (560 KB → 57 KB on wire)
- ✅ HTTP/2 on nginx

## What's pending

- ⏳ **"Not found → add new + drop to cart"** (POS search/scan miss — user wants an inline add flow)
- ⏳ **Cloudflare setup** (would cut Nairobi → Frankfurt latency dramatically; ~15 min effort)
- ⏳ Expenses: add + report
- ⏳ Purchases: create, report, suppliers
- ⏳ Transfers: create, history, detail
- ⏳ Customers: list, quick add, deposit modal, search
- ⏳ Customer Statement
- ⏳ Credit Sale modal
- ⏳ Company Settings
- ⏳ Company Management
- ⏳ Company Setup Wizard
- ⏳ Import Products: Excel preview + import (QuickBooks format)

## Deployment workflow

### Local dev loop
1. Edit code in VS Code
2. If Go change: air rebuilds automatically (2-3 sec)
3. If template/JS change: just save, hard-refresh browser (`Ctrl+F5`)
4. Test at http://localhost:8081

### Deploying Go changes to production
```powershell
cd C:\Users\sam\spide-pos
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
go build -o spide-pos cmd/server/main.go
scp .\spide-pos root@64.226.124.37:/tmp/spide-pos.new
```
Then on server:
```bash
cp /var/www/spide/spide-pos /root/backups/spide-pos.$(date +%Y%m%d_%H%M)
mv /tmp/spide-pos.new /var/www/spide/spide-pos
chmod +x /var/www/spide/spide-pos
systemctl restart spide-pos
systemctl status spide-pos --no-pager
```

### Deploying template/JS changes (no rebuild, no restart)
```powershell
scp .\internal\templates\pos.html root@64.226.124.37:/tmp/pos.html.new
scp .\static\js\spide-loaders.js  root@64.226.124.37:/tmp/spide-loaders.js.new
# also spide-modals.js if changed
```
Then on server:
```bash
cp /var/www/spide/internal/templates/pos.html /root/backups/pos.html.$(date +%Y%m%d_%H%M)
mv /tmp/pos.html.new /var/www/spide/internal/templates/pos.html
# same for JS files
```

### IMPORTANT: template and JS changes do NOT travel with the binary
The Go binary embeds Go code. Templates and JS are read from disk at request time. If you `go build` and deploy, the JS/template files on the server stay at their old versions. **You must SCP them separately if they changed.**

## Diagnostic commands

### Server-side
```bash
# App log (live)
journalctl -u spide-pos -f

# Recent errors
journalctl -u spide-pos --since "5 minutes ago" --no-pager | grep -E "❌|Error|panic"

# MySQL process list (spot stuck queries)
mysql -u root -p'Tende@2016' -e "SHOW PROCESSLIST;"

# Check gzip is working
curl -s -D - -o /dev/null -H "Accept-Encoding: gzip" \
  -H "Authorization: Bearer <token>" \
  https://pos.spide.co.ke/api/products?shop_id=4 | grep -i content-encoding

# Tail MySQL general log (turn on first — see below)
mysql -u root -p'Tende@2016' -e "SET GLOBAL general_log_file='/tmp/mysql.log'; SET GLOBAL general_log='ON';"
# ...do the thing...
mysql -u root -p'Tende@2016' -e "SET GLOBAL general_log='OFF';"
grep "Query\|Execute" /tmp/mysql.log | tail -40
```

### Local (Windows PowerShell)
```powershell
# Find a handler
Get-ChildItem -Recurse -Filter "*.go" | Select-String -Pattern "func CreateXHandler"

# Find a DB function
Get-ChildItem -Recurse -Filter "*.go" | Select-String -Pattern "func GetX"

# Find a JS function
Select-String -Path .\static\js\spide-loaders.js -Pattern "window\.loadX"
Select-String -Path .\internal\templates\pos.html -Pattern "function loadX"

# Build + run locally
air
# or:
go build -o spide-pos.exe cmd/server/main.go
.\spide-pos.exe
```

## Known gotchas / recent bugs fixed

- **NULL vs `[]`**: Go `nil` slices encode as JSON `null`. Always initialize as `[]T{}`.
- **`categoryID` vs `category_id`**: the SQL column is `category_id`.
- **`is_active` on `shops`**: doesn't exist.
- **Missing JSON fields**: use pointer types in request structs for partial updates.
- **Method mismatch**: frontend must send `PUT` for updates, `POST` for creates.
- **`http.Error` vs JSON errors**: always use `writeJSONError`.
- **Path is `internal/templates/`, not `templates/`**.
- **Receipt `window.onload` only fires once**: don't rely on it for auto-print on window reuse. Use explicit `printWindow.print()` in a `setTimeout`.
- **`/api/sales/recent` returns no items unless you attach them**: `GetRecentSalesForShopAndCompany` must call `getSaleItems(db, sale.ID)` per row.
- **Cart row HTML comes from Go, not JS**: `Product.RenderRowHTML()` in `models.go` generates the `<tr>`. Changes to the cart row layout require a Go rebuild.
- **Two `printThermalReceipt` functions exist**: one in `pos.html`, one in `spide-loaders.js`. Loaders wins (loaded later). Edit loaders' version for behavior changes.
- **Barcode duplicates**: fixed by adding `SKU-XXXXXXXX` synthetic barcodes to NULL-barcode products. The one-time backfill was `UPDATE products SET barcode = CONCAT('SKU-', LPAD(id, 8, '0')) WHERE barcode IS NULL OR barcode = '';`. Going forward, product creation should assign a barcode if missing (not yet implemented in handler).
Symptom: JS change works in editor but not in browser — even after cache clear, incognito, and server restart. Check the editor's save state. VS Code (or any editor) can hold an unsaved buffer indefinitely. Before diving into caching/path/binary diagnostics, verify with Get-Item <file> that the on-disk mtime matches when you last saved. Enable files.autoSave in VS Code to prevent this class of confusion entirely.

## Environment

- `.env` at project root (gitignored):
  ```
  JWT_SECRET=<hex>
  DB_USER=root
  DB_PASSWORD=Tende@2016
  DB_HOST=127.0.0.1
  DB_PORT=3306
  DB_NAME=spide_pos
  PORT=8081
  ```
- Demo user (company 1): `demo_admin` / `admin123`
- Live users (company 2): `David` (director), `Precious` (cashier)

## Network / performance notes (2026-09-18)

- **Server location**: Frankfurt, Germany (DigitalOcean)
- **Client location**: Nairobi, Kenya — RTT ~250ms
- **Problem**: 560 KB uncompressed JSON products list took ~38s to transfer over Kenyan DSL
- **Fix applied**: gzip on nginx → 57 KB → ~5-6s transfer
- **Remaining issue**: at ~1.5 Mbps, some transfers still take 20+ seconds
- **Future fix**: Cloudflare (Nairobi PoP) would bring this to ~1s

## Conventions

- Commit messages: short imperative
- Tags at milestones: `v0.1-working`, `v0.2-products`, etc.
- Always `git push` before ending a session
- Run `go build ./...` before committing Go changes

## When in doubt

- **Symptom: change works locally but not on server** → did you SCP the JS/template? The binary doesn't carry them.
- **Symptom: JSON field is `null` when it should be `[]`** → Go `nil` slice, initialize with `[]T{}`.
- **Symptom: server returns 200 but response is empty** → check `journalctl -u spide-pos -n 50`.
- **Symptom: page slow to load over VPN/hotspot** → check gzip is on, check payload size in DevTools Network tab.
- **Symptom: function undefined error in browser console** → check load order (modals before loaders), check the function is `window.`-prefixed if defined in loaders.js.