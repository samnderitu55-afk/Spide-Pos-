-- ============================================
-- Seed Data for Testing
-- ============================================

-- Insert a test product
INSERT IGNORE INTO products 
(barcode, name, category, cost_price, retail_price, wholesale_price, wholesale_min_qty, stock_quantity, reorder_level)
VALUES 
('COSM-001', 'Matte Lipstick - Ruby Red', 'Lipsticks', 250.00, 450.00, 350.00, 6, 50, 5),
('COSM-002', 'Liquid Foundation - Ivory', 'Foundations', 350.00, 650.00, 500.00, 6, 30, 5),
('COSM-003', 'Mascara - Black', 'Mascara', 180.00, 350.00, 280.00, 6, 40, 5),
('COSM-004', 'Eyeshadow Palette - Neutral', 'Eyeshadows', 400.00, 750.00, 600.00, 6, 20, 5),
('COSM-005', 'Blush - Pink', 'Blush', 200.00, 400.00, 320.00, 6, 35, 5);

-- Insert a test supplier
INSERT IGNORE INTO suppliers (name, contact_person, phone, email, address)
VALUES 
('ABC Cosmetics Ltd', 'John Mwangi', '+254 712 345 678', 'john@abccosmetics.com', 'Nairobi, Kenya'),
('Beauty Wholesale KE', 'Mary Wanjiru', '+254 722 345 678', 'mary@beautywholesale.co.ke', 'Mombasa, Kenya');

-- Insert shop stock (main shop)
INSERT IGNORE INTO shop_stock (shop_id, product_id, quantity)
SELECT 1, id, stock_quantity FROM products;
