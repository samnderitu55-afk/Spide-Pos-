-- ============================================================
-- Migration: widen total_items columns for fractional quantities
-- Date: 2026-09-23
-- Companion to 20260923_fractional_qty.sql
-- ============================================================

ALTER TABLE purchases       MODIFY total_items DECIMAL(10,3) NOT NULL DEFAULT 0;
ALTER TABLE stock_transfers MODIFY total_items DECIMAL(10,3) NOT NULL DEFAULT 0;