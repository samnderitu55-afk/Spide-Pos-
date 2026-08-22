CREATE DATABASE IF NOT EXISTS cosmetics_pos;
USE cosmetics_pos;

-- 1. Branches Table
CREATE TABLE branches (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed initial 3 branches
INSERT INTO branches (name, address) VALUES 
('Branch 1 - Central Hub', 'Main Street'),
('Branch 2', 'Westside Mall'),
('Branch 3', 'Eastside Plaza');

-- 2. Master Product Catalog (With Wholesale Support)
CREATE TABLE products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    barcode VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(150) NOT NULL,
    category VARCHAR(50),
    retail_price DECIMAL(10, 2) NOT NULL,
    wholesale_price DECIMAL(10, 2) NOT NULL,
    wholesale_min_qty INT NOT NULL DEFAULT 6,
    cost_price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed sample cosmetics item
INSERT INTO products (barcode, name, category, retail_price, wholesale_price, wholesale_min_qty, cost_price)
VALUES ('600123456789', 'Matte Liquid Lipstick - Ruby Red', 'Lipstick', 12.00, 8.50, 6, 5.00);

-- 3. Branch Inventory
CREATE TABLE branch_inventory (
    branch_id INT NOT NULL,
    product_id INT NOT NULL,
    stock_quantity INT NOT NULL DEFAULT 0,
    PRIMARY KEY (branch_id, product_id),
    FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

-- Seed initial stock for Branch 1
INSERT INTO branch_inventory (branch_id, product_id, stock_quantity) VALUES (1, 1, 50);