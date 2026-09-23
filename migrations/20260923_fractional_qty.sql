-- ============================================================
-- Migration: fractional quantities for Perfume Re-fill
-- Date: 2026-09-23
-- Company: 2 (Glow Age Beauty & Cosmetics)
-- Category: 46 (Perfume Re-fill) -- values already stored in ml
-- Scope confirmed: 97 products, category_id=46
-- ============================================================

-- 1. Widen quantity columns INT -> DECIMAL(10,3)
ALTER TABLE shop_stock     MODIFY quantity          DECIMAL(10,3) NOT NULL DEFAULT 0;
ALTER TABLE sale_items     MODIFY quantity          DECIMAL(10,3) NOT NULL;
ALTER TABLE products       MODIFY stock_quantity    DECIMAL(10,3) NOT NULL DEFAULT 0;
ALTER TABLE products       MODIFY wholesale_min_qty DECIMAL(10,3) NOT NULL DEFAULT 0;
ALTER TABLE transfer_items MODIFY quantity          DECIMAL(10,3) NOT NULL;
ALTER TABLE purchase_items MODIFY quantity          DECIMAL(10,3) NOT NULL;

-- 2. New columns on products
ALTER TABLE products
  ADD COLUMN unit_type  ENUM('package','unit') NOT NULL DEFAULT 'package' AFTER wholesale_min_qty,
  ADD COLUMN unit_label VARCHAR(20) NULL AFTER unit_type;

-- 3. Constraint: unit products must have a label; packages must not
ALTER TABLE products
  ADD CONSTRAINT chk_unit_label
  CHECK (
    (unit_type = 'unit'    AND unit_label IS NOT NULL) OR
    (unit_type = 'package' AND unit_label IS NULL)
  );

-- 4. Flag only the refill category as unit-based (ml)
UPDATE products
   SET unit_type = 'unit', unit_label = 'ml'
 WHERE company_id = 2 AND category_id = 46;