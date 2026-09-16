# Spide POS — Project Context

## Stack
- **Backend**: Go 1.x, MySQL (`spide_pos` database)
- **Frontend**: HTML + Tailwind (CDN) + vanilla JS. No build step, no framework.
- **Server**: `cmd/server/main.go`, runs on `:8081`, serves static files from `static/`
- **Repo**: https://github.com/samnderitu55-afk/Spide-Pos-

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
│   ├── handlers/               # HTTP handlers, one file per domain
│   ├── middleware/             # JWT auth middleware
│   └── templates/              # HTML pages served by Go
│       ├── login.html
│       ├── pos.html
│       ├── dashboard.html
│       ├── director.html
│       └── modals.html         # source of truth for modal HTML
├── static/
│   └── js/
│       ├── spide-modals.js     # generated from modals.html, injected at runtime
│       └── spide-loaders.js    # all shared data loaders and save handlers
├── go.mod / go.sum
└── PROJECT_CONTEXT.md          # this file
```

## Architecture rules (learned the hard way)

### Frontend
1. **All modals live in `spide-modals.js`** — a single `MODAL_HTML` template literal + open/close functions. Nothing modal-related lives in page files. If you need a new modal, add it to `MODAL_HTML` and add `window.openXModal` / `window.closeXModal`.
2. **All data loaders live in `spide-loaders.js`**, wrapped as `window.loadX = async function () { ... }`. Page files should NOT define `loadUsers`, `loadProducts`, `saveUser`, etc. — those are shared.
3. **Page files keep only page-specific logic**: cart state, search dropdown, checkout flow, dashboard KPI rendering, director outlet tables. If it needs to work on more than one page, it goes in `spide-loaders.js`.
4. **Modals auto-mount**: `SpideModals.mount()` runs on `DOMContentLoaded` and injects all modal HTML into `document.body`.
5. **`spide-modals.js` loads BEFORE `spide-loaders.js`** in each page's footer. Both load after page-specific inline scripts.

### Backend
1. **Every handler reads `company_id` from JWT claims** via `middleware.GetUserFromContext(r)`. Never trust client-provided `company_id` or `shop_id`.
2. **Every DB write filters by `company_id`** in the WHERE clause. `UpdateProduct`, `DeleteUser`, etc. all have `AND company_id = ?`.
3. **Handlers return JSON errors**, not `http.Error` plain text. Use `writeJSONError(w, status, msg)` so the frontend can parse.
4. **DB functions return `[]T{}` not `var x []T`** so empty results serialize to `[]` not `null`. The frontend iterates these directly.
5. **SQL uses snake_case column names** (`category_id`, not `categoryID`). Go variables use camelCase. Never mix them inside a SQL string.
6. **`COALESCE` for nullable columns** when scanning into plain `string`/`int`. Specifically: `email`, `shop_id`, `last_login`, `mpesa_code`. Or use `sql.NullX` types in the struct.
7. **Handlers return the full updated object** on success (not empty body). Frontend expects `result.name`, `result.id`, etc.

### Database
- `users.company_id` — NOT NULL for active users
- `products.category` (string) AND `products.category_id` (int, nullable) — two columns, both used. Frontend sends the name; handler resolves to id via `GetCategoryIDByName`.
- `sales.deposit_amount` — added mid-project; used by Z-Report
- `shops` table has NO `is_active` column — don't reference it in queries
- Foreign keys are strict — always set `company_id` before insert

### Frontend ↔ Backend contract
- Login: POST `/api/login` → sets `spide_token` and `spide_user` cookies (URL-encoded JSON)
- Every authenticated fetch sends `Authorization: Bearer <token>` header
- Update endpoints use `PUT`, create uses `POST`, delete uses `DELETE`
- Error responses: `{"error": "message"}` with appropriate HTTP status

## What's working (as of last session)

- ✅ Login / logout with JWT
- ✅ POS cart, checkout, receipts (thermal print)
- ✅ Dashboard KPIs, sales trend, top products
- ✅ Director dashboard (outlets, 30-day trend, alerts)
- ✅ Shared modals + loaders across all 3 pages
- ✅ Products: create, edit, per-shop Quick Stock Update
- ✅ User Management: create, edit (pointer-based partial update), delete (self-delete + last-admin protected)
- ✅ Low Stock report
- ✅ Inventory Valuation report (per-shop filter works for directors)
- ✅ Product Sales report (shop / category / product filters)
- ✅ Z-Report with correct payment breakdown
- ✅ Recent Sales modal + reprint

## What's pending

- ⏳ Expenses: add + report (modal exists, loaders/handlers probably need same fixes)
- ⏳ Purchases: create, report, suppliers
- ⏳ Transfers: create, history, detail
- ⏳ Customers: list, quick add, deposit modal, search
- ⏳ Customer Statement: `loadCustomerStatement` may not exist
- ⏳ Credit Sale modal: `submitCreditSale` may not exist
- ⏳ Company Settings: load/save current company
- ⏳ Company Management: list/add/edit companies
- ⏳ Company Setup Wizard: multi-step new-company creation
- ⏳ Import Products: Excel preview + import

## When adding a new feature — checklist

1. **Frontend**
   - Modal HTML → `spide-modals.js` in `MODAL_HTML`
   - `window.openXModal` / `window.closeXModal` → `spide-modals.js`
   - `window.loadX` (data fetch) → `spide-loaders.js`
   - `window.saveX` / `window.submitX` (form submit) → `spide-loaders.js`
   - Delete any duplicate definitions from page files
   - Verify with `Select-String -Path .\internal\templates\pos.html -Pattern "function loadX"` → should be empty

2. **Backend**
   - Handler reads `claims.CompanyID`, refuses if 0
   - Handler sets `item.CompanyID = claims.CompanyID` (never trusts payload)
   - Handler returns JSON on both success and error
   - DB query filters by `company_id`
   - DB returns `[]T{}` not `var x []T` for lists
   - SQL uses snake_case columns
   - `COALESCE` on nullable columns

3. **Test**
   - Open modal from POS, dashboard, director — behavior identical
   - Verify row in DB with `SELECT ... WHERE id = X` (confirm `company_id` is correct)
   - Verify cross-company access is blocked (try editing another company's row via curl)

## Diagnostic commands

```powershell
# Find a handler
Get-ChildItem -Recurse -Filter "*.go" | Select-String -Pattern "func CreateXHandler"

# Find a DB function
Get-ChildItem -Recurse -Filter "*.go" | Select-String -Pattern "func GetX"

# Find a JS function
Select-String -Path .\static\js\spide-loaders.js -Pattern "window\.loadX"
Select-String -Path .\internal\templates\pos.html -Pattern "function loadX"

# Find duplicate definitions across pages
Select-String -Path .\internal\templates\*.html -Pattern "function (loadX|saveX)"

# Build
go build ./...
go run ./cmd/server

# Git
git status
git log --oneline -10
```

## Known gotchas / recent bugs fixed

- **NULL vs `[]`**: Go `nil` slices encode as JSON `null`, which crashes frontend loops. Always initialize as `[]T{}`.
- **`categoryID` vs `category_id`**: the SQL column is `category_id`. A camelCase typo in the SQL string throws `Unknown column`.
- **`is_active` on `shops`**: doesn't exist. Only `products` has it.
- **Missing JSON fields**: if the frontend omits a field, Go decodes it as the zero value. For update operations, use `*bool` / `*string` pointers in the request struct so missing means "preserve current value".
- **Method mismatch**: frontend must send `PUT` for updates. If it sends `POST`, handler returns 405 plain text, and `response.json()` throws "Network error".
- **`http.Error` vs JSON errors**: `http.Error` writes plain text. Frontend `response.json()` fails. Always use `writeJSONError` helpers.
- **Path is `internal/templates/`, not `templates/`**.

## Environment

- `.env` file at project root (gitignored):
  ```
  JWT_SECRET=<hex>
  DB_USER=root
  DB_PASSWORD=<password>
  DB_HOST=127.0.0.1
  DB_PORT=3306
  DB_NAME=spide_pos
  PORT=8081
  ```
- Demo user: `demo_admin` / `admin123`, company_id `6`, shop_id `20`

## Conventions

- Commit messages: short imperative, e.g. "Fix Z-report split payment aggregation"
- Tags at milestones: `v0.1-working`, `v0.2-products`, `v0.3-users`, `v0.4-reports`
- Always `git push` before ending a session
- Run `go build ./...` before committing Go changes