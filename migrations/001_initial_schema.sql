-- ============================================
-- COSMETICS POS - Initial Database Schema
-- ============================================

-- 1. Categories
CREATE TABLE IF NOT EXISTS categories (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at DATETIME DEFAULT NOW()
);

-- 2. Products
CREATE TABLE IF NOT EXISTS products (
    id INT PRIMARY KEY AUTO_INCREMENT,
    barcode VARCHAR(100) UNIQUE,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    cost_price DECIMAL(10,2) DEFAULT 0,
    retail_price DECIMAL(10,2) DEFAULT 0,
    wholesale_price DECIMAL(10,2) DEFAULT 0,
    wholesale_min_qty INT DEFAULT 0,
    stock_quantity INT DEFAULT 0,
    reorder_level INT DEFAULT 5,
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT NOW(),
    updated_at DATETIME DEFAULT NOW()
);

-- 3. Suppliers
CREATE TABLE IF NOT EXISTS suppliers (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    contact_person VARCHAR(255),
    phone VARCHAR(50),
    email VARCHAR(255),
    address TEXT,
    notes TEXT,
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT NOW()
);

-- 4. Sales
CREATE TABLE IF NOT EXISTS sales (
    id INT PRIMARY KEY AUTO_INCREMENT,
    total_amount DECIMAL(10,2) NOT NULL,
    cash_amount DECIMAL(10,2) DEFAULT 0,
    mpesa_amount DECIMAL(10,2) DEFAULT 0,
    mpesa_code VARCHAR(100),
    payment_type ENUM('cash','mpesa','split') DEFAULT 'cash',
    change_given DECIMAL(10,2) DEFAULT 0,
    created_at DATETIME DEFAULT NOW()
);

-- 5. Sale Items
CREATE TABLE IF NOT EXISTS sale_items (
    id INT PRIMARY KEY AUTO_INCREMENT,
    sale_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    subtotal DECIMAL(10,2) NOT NULL,
    FOREIGN KEY (sale_id) REFERENCES sales(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id)
);

-- 6. Expenses
CREATE TABLE IF NOT EXISTS expenses (
    id INT PRIMARY KEY AUTO_INCREMENT,
    category VARCHAR(100) NOT NULL,
    description VARCHAR(255) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    expense_date DATE,
    payment_method VARCHAR(50),
    reference VARCHAR(100),
    notes TEXT,
    created_by VARCHAR(100),
    created_at DATETIME DEFAULT NOW()
);

-- 7. Shop Stock (Multi-shop support)
CREATE TABLE IF NOT EXISTS shop_stock (
    id INT PRIMARY KEY AUTO_INCREMENT,
    shop_id INT NOT NULL DEFAULT 1,
    product_id INT NOT NULL,
    quantity INT DEFAULT 0,
    created_at DATETIME DEFAULT NOW(),
    updated_at DATETIME DEFAULT NOW(),
    UNIQUE KEY (shop_id, product_id)
);

-- 8. Stock Transfers
CREATE TABLE IF NOT EXISTS stock_transfers (
    id INT PRIMARY KEY AUTO_INCREMENT,
    transfer_number VARCHAR(50) UNIQUE,
    from_shop_id INT NOT NULL,
    to_shop_id INT NOT NULL,
    total_items INT DEFAULT 0,
    total_cost DECIMAL(10,2) DEFAULT 0,
    transfer_date DATE,
    status ENUM('pending','completed','cancelled') DEFAULT 'pending',
    notes TEXT,
    created_by VARCHAR(100),
    created_at DATETIME DEFAULT NOW()
);

-- 9. Transfer Items
CREATE TABLE IF NOT EXISTS transfer_items (
    id INT PRIMARY KEY AUTO_INCREMENT,
    transfer_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL,
    cost_price DECIMAL(10,2) NOT NULL,
    subtotal DECIMAL(10,2) NOT NULL,
    FOREIGN KEY (transfer_id) REFERENCES stock_transfers(id) ON DELETE CASCADE
);

-- 10. Purchases
CREATE TABLE IF NOT EXISTS purchases (
    id INT PRIMARY KEY AUTO_INCREMENT,
    purchase_number VARCHAR(50) UNIQUE,
    supplier_id INT NOT NULL,
    total_items INT DEFAULT 0,
    total_cost DECIMAL(10,2) DEFAULT 0,
    purchase_date DATE,
    notes TEXT,
    created_by VARCHAR(100),
    created_at DATETIME DEFAULT NOW(),
    FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
);

-- 11. Purchase Items
CREATE TABLE IF NOT EXISTS purchase_items (
    id INT PRIMARY KEY AUTO_INCREMENT,
    purchase_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL,
    cost_price DECIMAL(10,2) NOT NULL,
    subtotal DECIMAL(10,2) NOT NULL,
    FOREIGN KEY (purchase_id) REFERENCES purchases(id) ON DELETE CASCADE
);

-- 12. Insert default categories
INSERT IGNORE INTO categories (name) VALUES 
    ('Lipsticks'),
    ('Foundations'),
    ('Mascara'),
    ('Eyeshadows'),
    ('Blush'),
    ('Concealer'),
    ('Skincare'),
    ('Fragrances'),
    ('Accessories'),
    ('Other');
