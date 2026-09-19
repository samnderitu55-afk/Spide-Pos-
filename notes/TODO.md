# Spide POS — TODO

## In progress
- (nothing)

## Next up
- [ ] POS: "not found" → add new product inline + drop to cart
  - Chosen approach: (TBD — Option B or C from previous discussion)
- [ ] Cloudflare in front of pos.spide.co.ke
  - Would cut Nairobi → Frankfurt latency from 250ms to ~15ms
  - Free tier, ~15 min setup

## Backlog
- [ ] Expenses: add + report (modal exists, loaders need wiring)
- [ ] Purchases: create, report, suppliers
- [ ] Transfers: create, history, detail (loaders exist, handlers may need fixes)
- [ ] Customers: list, quick add, deposit modal, search
- [ ] Customer Statement
- [ ] Credit Sale modal
- [ ] Company Settings
- [ ] Company Management
- [ ] Company Setup Wizard
- [ ] Import Products (QuickBooks format)
- [ ] Handler for product creation should auto-assign SKU barcode if missing
- [ ] Fix `company_id: 0` cosmetic bug in `/api/sales/recent` response
- [ ] Fix `mpesa_code` returning `{String, Valid}` instead of plain string

## Done (recent)
- [x] Migrated to new code + DB (2026-09-17)
- [x] Deduplicated products, filled NULL barcodes (2026-09-18)
- [x] gzip + HTTP/2 on nginx (2026-09-18)
- [x] POS cart: +/− buttons, delete, select-all on prices (2026-09-18)
- [x] Payment modal: zeroed fields + select-all (2026-09-18)
- [x] Receipts: bold items, Branch, served-by, auto-print, reprint fix (2026-09-18)

## Ideas / nice-to-have
- [ ] Product count in POS header ("1690 products")
- [ ] Cache product list in localStorage (5-min TTL) for instant reloads
- [ ] Barcode on receipt (dropped for now — overkill)
- [ ] Keyboard shortcut for "complete sale" (Enter on payment modal)
- [ ] Quick-sale mode: single-item checkout without cart