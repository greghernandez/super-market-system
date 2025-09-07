-- Database migrations for Aurora PostgreSQL
-- Run this after deployment to set up initial data

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create departments table
CREATE TABLE IF NOT EXISTS departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(255),
    image VARCHAR(255),
    slug VARCHAR(100) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create categories table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES categories(id),
    level INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku VARCHAR(100) UNIQUE NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    original_price DECIMAL(10,2),
    images TEXT[],
    category_id UUID NOT NULL REFERENCES categories(id),
    department_id UUID NOT NULL,
    brand VARCHAR(100),
    unit VARCHAR(50),
    stock INTEGER NOT NULL DEFAULT 0,
    min_stock INTEGER NOT NULL DEFAULT 0,
    weight DECIMAL(10,3),
    weight_unit VARCHAR(20),
    dim_length DECIMAL(10,2),
    dim_width DECIMAL(10,2),
    dim_height DECIMAL(10,2),
    is_on_sale BOOLEAN DEFAULT FALSE,
    discount DECIMAL(5,2),
    rating DECIMAL(3,2) DEFAULT 0,
    reviews INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    tags TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    date_of_birth DATE,
    is_active BOOLEAN DEFAULT TRUE,
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create user_addresses table
CREATE TABLE IF NOT EXISTS user_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    address_type VARCHAR(20) DEFAULT 'shipping', -- 'shipping', 'billing'
    street_address VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'confirmed', 'processing', 'shipped', 'delivered', 'cancelled'
    total_amount DECIMAL(10,2) NOT NULL,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    shipping_amount DECIMAL(10,2) DEFAULT 0,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    shipping_address_id UUID REFERENCES user_addresses(id),
    billing_address_id UUID REFERENCES user_addresses(id),
    payment_status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'paid', 'failed', 'refunded'
    payment_method VARCHAR(50),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create order_items table
CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_department_id ON products(department_id);
CREATE INDEX IF NOT EXISTS idx_products_brand ON products(brand);
CREATE INDEX IF NOT EXISTS idx_products_is_on_sale ON products(is_on_sale);
CREATE INDEX IF NOT EXISTS idx_products_is_active ON products(is_active);
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
CREATE INDEX IF NOT EXISTS idx_products_slug ON products(slug);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);

-- Insert sample departments
INSERT INTO departments (name, description, slug, icon) VALUES
('Produce', 'Fresh fruits and vegetables', 'produce', '🥬'),
('Dairy', 'Milk, cheese, yogurt and dairy products', 'dairy', '🥛'),
('Meat & Seafood', 'Fresh meat, poultry and seafood', 'meat-seafood', '🥩'),
('Bakery', 'Fresh bread, pastries and baked goods', 'bakery', '🍞'),
('Pantry', 'Canned goods, grains, pasta and staples', 'pantry', '🥫'),
('Frozen', 'Frozen foods and ice cream', 'frozen', '🧊'),
('Beverages', 'Soft drinks, juices, water and beverages', 'beverages', '🥤'),
('Health & Beauty', 'Personal care and health products', 'health-beauty', '🧴')
ON CONFLICT (slug) DO NOTHING;

-- Insert sample categories
INSERT INTO categories (name, slug, description, level) VALUES
('Fruits', 'fruits', 'Fresh seasonal fruits', 1),
('Vegetables', 'vegetables', 'Fresh vegetables and greens', 1),
('Organic Produce', 'organic-produce', 'Certified organic fruits and vegetables', 1),
('Milk & Cream', 'milk-cream', 'Various types of milk and cream', 1),
('Cheese', 'cheese', 'Domestic and imported cheeses', 1),
('Yogurt', 'yogurt', 'Greek, regular and specialty yogurts', 1),
('Beef', 'beef', 'Fresh beef cuts', 1),
('Chicken', 'chicken', 'Fresh chicken and poultry', 1),
('Seafood', 'seafood', 'Fresh fish and seafood', 1)
ON CONFLICT (slug) DO NOTHING;

-- Insert sample products
INSERT INTO products (sku, slug, name, description, price, category_id, department_id, brand, stock, min_stock, unit)
SELECT 
    'PROD-' || LPAD(generate_series::text, 6, '0'),
    'product-' || generate_series,
    'Sample Product ' || generate_series,
    'This is a sample product for testing purposes',
    (random() * 50 + 5)::DECIMAL(10,2),
    (SELECT id FROM categories ORDER BY random() LIMIT 1),
    (SELECT id FROM departments ORDER BY random() LIMIT 1),
    'Sample Brand',
    (random() * 100 + 10)::INTEGER,
    5,
    'each'
FROM generate_series(1, 20)
ON CONFLICT (sku) DO NOTHING;