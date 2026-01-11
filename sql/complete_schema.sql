-- =====================================================
-- Complete Car Deals Database Schema
-- PostgreSQL Version
-- =====================================================
-- This file creates the complete database schema with all
-- tables, triggers, functions, views, and sample data.
-- All migrations have been applied to this schema.
-- =====================================================

-- Create database (uncomment if needed)
-- CREATE DATABASE car_deals_db;
-- \c car_deals_db;

-- =====================================================
-- CREATE CARS SCHEMA
-- =====================================================

CREATE SCHEMA IF NOT EXISTS cars;

-- Set search path to cars schema
SET search_path TO cars, public;

-- =====================================================
-- ENUMS AND CUSTOM TYPES
-- =====================================================

CREATE TYPE cars.condition_status_enum AS ENUM ('REGISTERED', 'UNREGISTERED');
CREATE TYPE cars.shipping_status_enum AS ENUM ('PROCESSING', 'SHIPPED', 'ARRIVED', 'CLEARED', 'DELIVERED');
CREATE TYPE cars.sale_status_enum AS ENUM ('AVAILABLE', 'RESERVED', 'SOLD', 'CANCELLED');
CREATE TYPE cars.customer_type_enum AS ENUM ('INDIVIDUAL', 'BUSINESS');
CREATE TYPE cars.supplier_type_enum AS ENUM ('AUCTION', 'DEALER', 'INDIVIDUAL');
CREATE TYPE cars.order_type_enum AS ENUM ('AUCTION', 'DIRECT', 'DEALER');
CREATE TYPE cars.priority_level_enum AS ENUM ('NORMAL', 'HIGH', 'URGENT');
CREATE TYPE cars.shipping_method_enum AS ENUM ('VESSEL', 'CONTAINER', 'RORO');
CREATE TYPE cars.payment_method_enum AS ENUM ('CASH', 'FINANCING', 'LEASE', 'INSTALLMENT');
CREATE TYPE cars.order_status_enum AS ENUM ('DRAFT', 'SUBMITTED', 'PROCESSING', 'MATCHED', 'COMPLETED', 'CANCELLED');
CREATE TYPE cars.document_type_enum AS ENUM ('INVOICE', 'SHIPPING', 'CUSTOMS', 'INSPECTION', 'REGISTRATION', 'OTHER', 'LC_DOCUMENT', 'RECEIPT', 'CONTRACT');
CREATE TYPE cars.audit_action_enum AS ENUM ('INSERT', 'UPDATE', 'DELETE');
CREATE TYPE cars.purchase_status_enum AS ENUM ('LC_PENDING', 'LC_OPENED', 'LC_RECEIVED', 'CANCELLED');

-- =====================================================
-- REFERENCE TABLES (Created first for foreign keys)
-- =====================================================

-- Vehicle Makes Reference Table
CREATE TABLE cars.vehicle_makes (
    id SERIAL PRIMARY KEY,
    make_name VARCHAR(50) UNIQUE NOT NULL,
    country_origin VARCHAR(50) DEFAULT 'Japan',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    logo_url VARCHAR(500)
);

-- Vehicle Models Reference Table
CREATE TABLE cars.vehicle_models (
    id SERIAL PRIMARY KEY,
    make_id INTEGER NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    body_type VARCHAR(50),
    fuel_type VARCHAR(30),
    transmission_type VARCHAR(30),
    engine_size_cc INTEGER,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_vehicle_models_make_id FOREIGN KEY (make_id) REFERENCES cars.vehicle_makes(id),
    CONSTRAINT unique_make_model UNIQUE (make_id, model_name)
);

CREATE INDEX idx_vehicle_models_make_id ON cars.vehicle_models(make_id);

-- Customers Table
CREATE TABLE cars.customers (
    id BIGSERIAL PRIMARY KEY,
    customer_title VARCHAR(10),
    customer_name VARCHAR(100) NOT NULL,
    contact_number VARCHAR(50),
    email VARCHAR(100),
    address TEXT,
    other_contacts TEXT,
    customer_type cars.customer_type_enum DEFAULT 'INDIVIDUAL',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_customers_customer_name ON cars.customers(customer_name);
CREATE INDEX idx_customers_contact_number ON cars.customers(contact_number);
CREATE INDEX idx_customers_email ON cars.customers(email);

-- Suppliers/Dealers Table
CREATE TABLE cars.suppliers (
    id BIGSERIAL PRIMARY KEY,
    supplier_name VARCHAR(100) NOT NULL,
    supplier_title VARCHAR(10),
    contact_number VARCHAR(50),
    email VARCHAR(100),
    address TEXT,
    other_contacts TEXT,
    supplier_type cars.supplier_type_enum DEFAULT 'AUCTION',
    country VARCHAR(50) DEFAULT 'Japan',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_suppliers_supplier_name ON cars.suppliers(supplier_name);
CREATE INDEX idx_suppliers_supplier_type ON cars.suppliers(supplier_type);

-- =====================================================
-- MAIN TABLES
-- =====================================================

-- Vehicles Master Table
CREATE TABLE cars.vehicles (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    make VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    trim_level VARCHAR(100),
    year_of_manufacture INTEGER NOT NULL,
    color VARCHAR(50) NOT NULL,
    mileage_km INTEGER,
    chassis_id VARCHAR(50) UNIQUE NOT NULL,
    condition_status cars.condition_status_enum DEFAULT 'UNREGISTERED',
    year_of_registration INTEGER,
    license_plate VARCHAR(20),
    auction_grade VARCHAR(10),
    auction_price DECIMAL(15,2),
    price_quoted DECIMAL(15,2),
    cif_value DECIMAL(15,2),
    currency VARCHAR(10) DEFAULT 'JPY',
    hs_code VARCHAR(20),
    invoice_fob_jpy DECIMAL(15,2),
    registration_number VARCHAR(20),
    record_date TIMESTAMP,
    is_featured BOOLEAN DEFAULT FALSE,
    featured_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add comments for featured fields
COMMENT ON COLUMN cars.vehicles.is_featured IS 'Whether this vehicle is featured/highlighted on the homepage or listings';
COMMENT ON COLUMN cars.vehicles.featured_at IS 'Timestamp when the vehicle was marked as featured';

-- Create indexes for vehicles table
CREATE INDEX idx_vehicles_chassis_id ON cars.vehicles(chassis_id);
CREATE INDEX idx_vehicles_make_model ON cars.vehicles(make, model);
CREATE INDEX idx_vehicles_year ON cars.vehicles(year_of_manufacture);
CREATE INDEX idx_vehicles_code ON cars.vehicles(code);
CREATE INDEX idx_vehicles_make_year ON cars.vehicles(make, year_of_manufacture);
CREATE INDEX idx_vehicles_is_featured ON cars.vehicles(is_featured) WHERE is_featured = true;
CREATE INDEX idx_vehicles_featured_at ON cars.vehicles(featured_at DESC) WHERE featured_at IS NOT NULL;

-- Purchase Information Table (with supplier_id foreign key)
CREATE TABLE cars.vehicle_purchases (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL,
    supplier_id BIGINT,
    purchase_remarks TEXT,
    lc_bank VARCHAR(100),
    lc_number VARCHAR(50),
    lc_cost_jpy DECIMAL(15,2),
    purchase_date TIMESTAMP,
    purchase_status cars.purchase_status_enum DEFAULT 'LC_PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_vehicle_purchases_vehicle_id
        FOREIGN KEY (vehicle_id)
        REFERENCES cars.vehicles(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_vehicle_purchases_supplier_id
        FOREIGN KEY (supplier_id)
        REFERENCES cars.suppliers(id)
        ON DELETE SET NULL
);

CREATE INDEX idx_vehicle_purchases_vehicle_id ON cars.vehicle_purchases(vehicle_id);
CREATE INDEX idx_vehicle_purchases_supplier_id ON cars.vehicle_purchases(supplier_id);
CREATE INDEX idx_vehicle_purchases_purchase_date ON cars.vehicle_purchases(purchase_date);
CREATE INDEX idx_vehicle_purchases_purchase_status ON cars.vehicle_purchases(purchase_status);

COMMENT ON COLUMN cars.vehicle_purchases.supplier_id IS 'Foreign key reference to suppliers table';

-- Shipping Information Table
CREATE TABLE cars.vehicle_shipping (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL,
    vessel_name VARCHAR(100),
    departure_harbour VARCHAR(50),
    shipment_date TIMESTAMP,
    arrival_date TIMESTAMP,
    clearing_date TIMESTAMP,
    shipping_status cars.shipping_status_enum DEFAULT 'PROCESSING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_vehicle_shipping_vehicle_id
        FOREIGN KEY (vehicle_id)
        REFERENCES cars.vehicles(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_vehicle_shipping_vehicle_id ON cars.vehicle_shipping(vehicle_id);
CREATE INDEX idx_vehicle_shipping_shipping_status ON cars.vehicle_shipping(shipping_status);
CREATE INDEX idx_vehicle_shipping_shipment_date ON cars.vehicle_shipping(shipment_date);
CREATE INDEX idx_vehicle_shipping_arrival_date ON cars.vehicle_shipping(arrival_date);
CREATE INDEX idx_vehicle_shipping_dates ON cars.vehicle_shipping(shipment_date, arrival_date);

-- Shipping History Table (for tracking status changes)
CREATE TABLE cars.vehicle_shipping_history (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL,
    old_status cars.shipping_status_enum,
    new_status cars.shipping_status_enum NOT NULL,
    vessel_name VARCHAR(100),
    departure_harbour VARCHAR(50),
    shipment_date TIMESTAMP,
    arrival_date TIMESTAMP,
    clearing_date TIMESTAMP,
    changed_by VARCHAR(100),
    change_remarks TEXT,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_shipping_history_vehicle_id
        FOREIGN KEY (vehicle_id)
        REFERENCES cars.vehicles(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_shipping_history_vehicle_id ON cars.vehicle_shipping_history(vehicle_id);
CREATE INDEX idx_shipping_history_changed_at ON cars.vehicle_shipping_history(changed_at);
CREATE INDEX idx_shipping_history_new_status ON cars.vehicle_shipping_history(new_status);

COMMENT ON TABLE cars.vehicle_shipping_history IS 'Tracks all changes to vehicle shipping status with timestamps and user information';

-- Purchase History Table
CREATE TABLE cars.vehicle_purchase_history (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL,
    old_status cars.purchase_status_enum,
    new_status cars.purchase_status_enum NOT NULL,
    supplier_id BIGINT,
    lc_bank VARCHAR(100),
    lc_number VARCHAR(50),
    lc_cost_jpy DECIMAL(15,2),
    purchase_date TIMESTAMP,
    purchase_remarks TEXT,
    changed_by VARCHAR(100),
    change_remarks TEXT,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_purchase_history_vehicle_id
        FOREIGN KEY (vehicle_id)
        REFERENCES cars.vehicles(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_purchase_history_supplier_id
        FOREIGN KEY (supplier_id)
        REFERENCES cars.suppliers(id)
        ON DELETE SET NULL
);

CREATE INDEX idx_purchase_history_vehicle_id ON cars.vehicle_purchase_history(vehicle_id);
CREATE INDEX idx_purchase_history_changed_at ON cars.vehicle_purchase_history(changed_at);
CREATE INDEX idx_purchase_history_new_status ON cars.vehicle_purchase_history(new_status);
CREATE INDEX idx_purchase_history_supplier_id ON cars.vehicle_purchase_history(supplier_id);

COMMENT ON TABLE cars.vehicle_purchase_history IS 'Tracks all changes to vehicle purchase information with timestamps and user information';

-- Financial Information Table
CREATE TABLE cars.vehicle_financials (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL,
    charges_lkr DECIMAL(15,2),
    tt_lkr DECIMAL(15,2),
    duty_lkr DECIMAL(15,2),
    clearing_lkr DECIMAL(15,2),
    other_expenses_lkr JSONB DEFAULT '{}',
    total_cost_lkr DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_vehicle_financials_vehicle_id
        FOREIGN KEY (vehicle_id)
        REFERENCES cars.vehicles(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_vehicle_financials_vehicle_id ON cars.vehicle_financials(vehicle_id);
CREATE INDEX idx_vehicle_financials_total_cost ON cars.vehicle_financials(total_cost_lkr);

-- Sales Information Table (with customer_id foreign key)
CREATE TABLE cars.vehicle_sales (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL,
    customer_id BIGINT,
    sold_date TIMESTAMP,
    revenue DECIMAL(15,2),
    profit DECIMAL(15,2),
    sale_remarks TEXT,
    sale_status cars.sale_status_enum DEFAULT 'AVAILABLE',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_vehicle_sales_vehicle_id
        FOREIGN KEY (vehicle_id)
        REFERENCES cars.vehicles(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_vehicle_sales_customer_id
        FOREIGN KEY (customer_id)
        REFERENCES cars.customers(id)
        ON DELETE SET NULL
);

CREATE INDEX idx_vehicle_sales_vehicle_id ON cars.vehicle_sales(vehicle_id);
CREATE INDEX idx_vehicle_sales_customer_id ON cars.vehicle_sales(customer_id);
CREATE INDEX idx_vehicle_sales_sold_date ON cars.vehicle_sales(sold_date);
CREATE INDEX idx_vehicle_sales_sale_status ON cars.vehicle_sales(sale_status);
CREATE INDEX idx_vehicle_sales_profit ON cars.vehicle_sales(profit);

COMMENT ON COLUMN cars.vehicle_sales.customer_id IS 'Foreign key to customers table. Links the sale to an actual customer record.';

-- =====================================================
-- ORDER MANAGEMENT TABLES
-- =====================================================

-- Customer Orders Table
CREATE TABLE cars.customer_orders (
    id BIGSERIAL PRIMARY KEY,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id BIGINT,

    -- Vehicle Requirements
    preferred_make VARCHAR(50),
    preferred_model VARCHAR(100),
    preferred_year_min INTEGER,
    preferred_year_max INTEGER,
    preferred_color VARCHAR(50),
    preferred_trim_level VARCHAR(100),
    max_mileage_km INTEGER,
    min_auction_grade VARCHAR(10),
    required_features JSONB,

    -- Order Information
    order_type cars.order_type_enum DEFAULT 'AUCTION',
    expected_delivery_date DATE,
    priority_level cars.priority_level_enum DEFAULT 'NORMAL',

    -- Shipping Preferences
    preferred_port VARCHAR(50),
    shipping_method cars.shipping_method_enum DEFAULT 'VESSEL',
    include_insurance BOOLEAN DEFAULT TRUE,

    -- Financial Information
    budget_min DECIMAL(15,2),
    budget_max DECIMAL(15,2),
    payment_method cars.payment_method_enum DEFAULT 'CASH',
    down_payment DECIMAL(15,2),

    -- Additional Information
    special_requests TEXT,
    internal_notes TEXT,

    -- Order Status
    order_status cars.order_status_enum DEFAULT 'DRAFT',
    is_draft BOOLEAN DEFAULT FALSE,

    -- Timestamps
    order_date TIMESTAMP NOT NULL,
    completed_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_customer_orders_customer_id
        FOREIGN KEY (customer_id)
        REFERENCES cars.customers(id)
);

CREATE INDEX idx_customer_orders_order_number ON cars.customer_orders(order_number);
CREATE INDEX idx_customer_orders_customer_id ON cars.customer_orders(customer_id);
CREATE INDEX idx_customer_orders_order_status ON cars.customer_orders(order_status);
CREATE INDEX idx_customer_orders_order_date ON cars.customer_orders(order_date);

-- Order-Vehicle Matching Table
CREATE TABLE cars.order_vehicle_matches (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    vehicle_id BIGINT NOT NULL,
    match_score DECIMAL(5,2),
    matched_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_selected BOOLEAN DEFAULT FALSE,

    CONSTRAINT fk_order_vehicle_matches_order_id
        FOREIGN KEY (order_id)
        REFERENCES cars.customer_orders(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_order_vehicle_matches_vehicle_id
        FOREIGN KEY (vehicle_id)
        REFERENCES cars.vehicles(id)
        ON DELETE CASCADE,
    CONSTRAINT unique_order_vehicle UNIQUE (order_id, vehicle_id)
);

CREATE INDEX idx_order_vehicle_matches_order_id ON cars.order_vehicle_matches(order_id);
CREATE INDEX idx_order_vehicle_matches_vehicle_id ON cars.order_vehicle_matches(vehicle_id);

-- =====================================================
-- AUDIT AND TRACKING TABLES
-- =====================================================

-- Audit Log Table
CREATE TABLE cars.audit_logs (
    id BIGSERIAL PRIMARY KEY,
    table_name VARCHAR(50) NOT NULL,
    record_id BIGINT NOT NULL,
    action cars.audit_action_enum NOT NULL,
    old_values JSONB,
    new_values JSONB,
    user_id VARCHAR(50),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_table_record ON cars.audit_logs(table_name, record_id);
CREATE INDEX idx_audit_logs_timestamp ON cars.audit_logs(timestamp);

-- Document Attachments Table
CREATE TABLE cars.vehicle_documents (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL,
    document_type cars.document_type_enum NOT NULL,
    document_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500),
    file_size_bytes BIGINT,
    mime_type VARCHAR(100),
    upload_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_vehicle_documents_vehicle_id
        FOREIGN KEY (vehicle_id)
        REFERENCES cars.vehicles(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_vehicle_documents_vehicle_id ON cars.vehicle_documents(vehicle_id);
CREATE INDEX idx_vehicle_documents_document_type ON cars.vehicle_documents(document_type);

-- Vehicle Images Table
CREATE TABLE cars.vehicle_images (
    id SERIAL PRIMARY KEY,
    vehicle_id INTEGER NOT NULL,
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255),
    file_path VARCHAR(500) NOT NULL,
    file_size INTEGER,
    mime_type VARCHAR(100),
    is_primary BOOLEAN DEFAULT false,
    upload_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    display_order INTEGER DEFAULT 0,

    CONSTRAINT fk_vehicle_images_vehicle_id
        FOREIGN KEY (vehicle_id)
        REFERENCES cars.vehicles(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_vehicle_images_vehicle_id ON cars.vehicle_images(vehicle_id);

-- =====================================================
-- TRIGGERS FOR AUTO-UPDATING TIMESTAMPS
-- =====================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION cars.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for updating updated_at columns
CREATE TRIGGER update_vehicles_updated_at
    BEFORE UPDATE ON cars.vehicles
    FOR EACH ROW EXECUTE FUNCTION cars.update_updated_at_column();

CREATE TRIGGER update_vehicle_purchases_updated_at
    BEFORE UPDATE ON cars.vehicle_purchases
    FOR EACH ROW EXECUTE FUNCTION cars.update_updated_at_column();

CREATE TRIGGER update_vehicle_shipping_updated_at
    BEFORE UPDATE ON cars.vehicle_shipping
    FOR EACH ROW EXECUTE FUNCTION cars.update_updated_at_column();

CREATE TRIGGER update_vehicle_financials_updated_at
    BEFORE UPDATE ON cars.vehicle_financials
    FOR EACH ROW EXECUTE FUNCTION cars.update_updated_at_column();

CREATE TRIGGER update_vehicle_sales_updated_at
    BEFORE UPDATE ON cars.vehicle_sales
    FOR EACH ROW EXECUTE FUNCTION cars.update_updated_at_column();

CREATE TRIGGER update_customers_updated_at
    BEFORE UPDATE ON cars.customers
    FOR EACH ROW EXECUTE FUNCTION cars.update_updated_at_column();

CREATE TRIGGER update_suppliers_updated_at
    BEFORE UPDATE ON cars.suppliers
    FOR EACH ROW EXECUTE FUNCTION cars.update_updated_at_column();

CREATE TRIGGER update_customer_orders_updated_at
    BEFORE UPDATE ON cars.customer_orders
    FOR EACH ROW EXECUTE FUNCTION cars.update_updated_at_column();

-- =====================================================
-- TRIGGERS FOR SHIPPING HISTORY
-- =====================================================

CREATE OR REPLACE FUNCTION cars.log_shipping_status_change()
RETURNS TRIGGER AS $$
BEGIN
    -- Only log if status actually changed
    IF (TG_OP = 'UPDATE' AND OLD.shipping_status IS DISTINCT FROM NEW.shipping_status) THEN
        INSERT INTO cars.vehicle_shipping_history (
            vehicle_id, old_status, new_status, vessel_name, departure_harbour,
            shipment_date, arrival_date, clearing_date, changed_by, changed_at
        ) VALUES (
            NEW.vehicle_id, OLD.shipping_status, NEW.shipping_status, NEW.vessel_name,
            NEW.departure_harbour, NEW.shipment_date, NEW.arrival_date, NEW.clearing_date,
            current_user, CURRENT_TIMESTAMP
        );
    END IF;

    -- Also log initial status when first created
    IF (TG_OP = 'INSERT') THEN
        INSERT INTO cars.vehicle_shipping_history (
            vehicle_id, old_status, new_status, vessel_name, departure_harbour,
            shipment_date, arrival_date, clearing_date, changed_by, changed_at
        ) VALUES (
            NEW.vehicle_id, NULL, NEW.shipping_status, NEW.vessel_name,
            NEW.departure_harbour, NEW.shipment_date, NEW.arrival_date, NEW.clearing_date,
            current_user, CURRENT_TIMESTAMP
        );
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER shipping_status_change_trigger
    AFTER INSERT OR UPDATE ON cars.vehicle_shipping
    FOR EACH ROW
    EXECUTE FUNCTION cars.log_shipping_status_change();

-- =====================================================
-- TRIGGERS FOR PURCHASE HISTORY
-- =====================================================

CREATE OR REPLACE FUNCTION cars.log_purchase_status_change()
RETURNS TRIGGER AS $$
BEGIN
    -- Log if any purchase field changed
    IF (TG_OP = 'UPDATE' AND (
        OLD.purchase_status IS DISTINCT FROM NEW.purchase_status OR
        OLD.supplier_id IS DISTINCT FROM NEW.supplier_id OR
        OLD.lc_bank IS DISTINCT FROM NEW.lc_bank OR
        OLD.lc_number IS DISTINCT FROM NEW.lc_number OR
        OLD.lc_cost_jpy IS DISTINCT FROM NEW.lc_cost_jpy OR
        OLD.purchase_date IS DISTINCT FROM NEW.purchase_date OR
        OLD.purchase_remarks IS DISTINCT FROM NEW.purchase_remarks
    )) THEN
        INSERT INTO cars.vehicle_purchase_history (
            vehicle_id, old_status, new_status, supplier_id, lc_bank,
            lc_number, lc_cost_jpy, purchase_date, purchase_remarks,
            changed_by, changed_at
        ) VALUES (
            NEW.vehicle_id, OLD.purchase_status, NEW.purchase_status, NEW.supplier_id,
            NEW.lc_bank, NEW.lc_number, NEW.lc_cost_jpy, NEW.purchase_date,
            NEW.purchase_remarks, current_user, CURRENT_TIMESTAMP
        );
    END IF;

    -- Log initial status on INSERT
    IF (TG_OP = 'INSERT') THEN
        INSERT INTO cars.vehicle_purchase_history (
            vehicle_id, old_status, new_status, supplier_id, lc_bank,
            lc_number, lc_cost_jpy, purchase_date, purchase_remarks,
            changed_by, changed_at
        ) VALUES (
            NEW.vehicle_id, NULL, NEW.purchase_status, NEW.supplier_id,
            NEW.lc_bank, NEW.lc_number, NEW.lc_cost_jpy, NEW.purchase_date,
            NEW.purchase_remarks, current_user, CURRENT_TIMESTAMP
        );
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER purchase_status_change_trigger
    AFTER INSERT OR UPDATE ON cars.vehicle_purchases
    FOR EACH ROW
    EXECUTE FUNCTION cars.log_purchase_status_change();

-- =====================================================
-- TRIGGERS FOR AUDIT LOGGING
-- =====================================================

CREATE OR REPLACE FUNCTION cars.audit_trigger_function()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO cars.audit_logs (table_name, record_id, action, new_values, user_id)
        VALUES (TG_TABLE_NAME, NEW.id, 'INSERT', to_jsonb(NEW), current_user);
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO cars.audit_logs (table_name, record_id, action, old_values, new_values, user_id)
        VALUES (TG_TABLE_NAME, NEW.id, 'UPDATE', to_jsonb(OLD), to_jsonb(NEW), current_user);
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO cars.audit_logs (table_name, record_id, action, old_values, user_id)
        VALUES (TG_TABLE_NAME, OLD.id, 'DELETE', to_jsonb(OLD), current_user);
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create audit triggers for main tables
CREATE TRIGGER vehicles_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON cars.vehicles
    FOR EACH ROW EXECUTE FUNCTION cars.audit_trigger_function();

CREATE TRIGGER vehicle_sales_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON cars.vehicle_sales
    FOR EACH ROW EXECUTE FUNCTION cars.audit_trigger_function();

-- =====================================================
-- VIEWS FOR COMMON QUERIES
-- =====================================================

-- Vehicle Shipping History View
CREATE OR REPLACE VIEW cars.vehicle_shipping_history_view AS
SELECT
    vsh.id,
    vsh.vehicle_id,
    v.code as vehicle_code,
    v.make,
    v.model,
    v.chassis_id,
    vsh.old_status,
    vsh.new_status,
    vsh.vessel_name,
    vsh.departure_harbour,
    vsh.shipment_date,
    vsh.arrival_date,
    vsh.clearing_date,
    vsh.changed_by,
    vsh.change_remarks,
    vsh.changed_at,
    EXTRACT(EPOCH FROM (vsh.changed_at - LAG(vsh.changed_at) OVER (PARTITION BY vsh.vehicle_id ORDER BY vsh.changed_at))) / 3600 as hours_in_previous_status
FROM cars.vehicle_shipping_history vsh
JOIN cars.vehicles v ON vsh.vehicle_id = v.id
ORDER BY vsh.vehicle_id, vsh.changed_at DESC;

-- Sales Summary View
CREATE VIEW cars.sales_summary AS
SELECT
    TO_CHAR(vsl.sold_date, 'YYYY-MM') as sale_month,
    COUNT(*) as vehicles_sold,
    SUM(vf.total_cost_lkr) as total_cost,
    SUM(vsl.revenue) as total_revenue,
    SUM(vsl.profit) as total_profit,
    AVG(vsl.profit) as avg_profit_per_vehicle,
    (SUM(vsl.profit) / NULLIF(SUM(vf.total_cost_lkr), 0)) * 100 as profit_margin_percentage
FROM cars.vehicle_sales vsl
JOIN cars.vehicle_financials vf ON vsl.vehicle_id = vf.vehicle_id
WHERE vsl.sold_date IS NOT NULL
GROUP BY TO_CHAR(vsl.sold_date, 'YYYY-MM')
ORDER BY sale_month DESC;

-- Inventory Status View
CREATE VIEW cars.inventory_status AS
SELECT
    vs.shipping_status,
    vsl.sale_status,
    COUNT(*) as vehicle_count,
    SUM(vf.total_cost_lkr) as total_investment
FROM cars.vehicles v
LEFT JOIN cars.vehicle_shipping vs ON v.id = vs.vehicle_id
LEFT JOIN cars.vehicle_sales vsl ON v.id = vsl.vehicle_id
LEFT JOIN cars.vehicle_financials vf ON v.id = vf.vehicle_id
GROUP BY vs.shipping_status, vsl.sale_status;

-- =====================================================
-- INITIAL DATA SETUP
-- =====================================================

-- Insert common vehicle makes
INSERT INTO cars.vehicle_makes (make_name, country_origin, log_url) VALUES
    ('Toyota', 'Japan', ''),
    ('Honda', 'Japan', ''),
    ('Nissan', 'Japan', ''),
    ('Mazda', 'Japan', ''),
    ('Suzuki', 'Japan', ''),
    ('Mitsubishi', 'Japan', ''),
    ('Subaru', 'Japan', ''),
    ('Lexus', 'Japan', ''),
    ('Infiniti', 'Japan', ''),
    ('Acura', 'Japan', '');

-- Insert common Toyota models
INSERT INTO cars.vehicle_models (make_id, model_name, body_type) VALUES
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Toyota'), 'Aqua', 'Hatchback'),
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Toyota'), 'Prius', 'Hatchback'),
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Toyota'), 'Vitz', 'Hatchback'),
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Toyota'), 'Axio', 'Sedan'),
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Toyota'), 'Fielder', 'Wagon'),
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Toyota'), 'Allion', 'Sedan'),
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Toyota'), 'Premio', 'Sedan'),
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Toyota'), 'Voxy', 'Minivan'),
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Toyota'), 'Noah', 'Minivan');

-- Insert common Honda models
INSERT INTO cars.vehicle_models (make_id, model_name, body_type) VALUES
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Honda'), 'Fit', 'Hatchback'),
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Honda'), 'Vezel', 'SUV'),
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Honda'), 'Grace', 'Sedan'),
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Honda'), 'Freed', 'Minivan'),
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Honda'), 'Shuttle', 'Wagon'),
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Honda'), 'CR-V', 'SUV'),
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Honda'), 'HR-V', 'SUV'),
    ((SELECT id FROM cars.vehicle_makes WHERE make_name = 'Honda'), 'Stepwgn', 'Minivan');

-- Insert sample suppliers
INSERT INTO cars.suppliers (supplier_name, supplier_title, supplier_type, country, contact_number, email, address) VALUES
    ('USS Auction', NULL, 'AUCTION', 'Japan', '+81-3-1234-5678', 'info@uss.co.jp', 'Tokyo, Japan'),
    ('Honda Japan', NULL, 'DEALER', 'Japan', '+81-3-2345-6789', 'export@honda.co.jp', 'Tokyo, Japan'),
    ('Toyota Motor Corporation', NULL, 'DEALER', 'Japan', '+81-3-3456-7890', 'export@toyota.co.jp', 'Toyota City, Japan'),
    ('JAA (Japan Auto Auction)', NULL, 'AUCTION', 'Japan', '+81-6-4567-8901', 'contact@jaa.co.jp', 'Osaka, Japan'),
    ('Nissan Export Division', NULL, 'DEALER', 'Japan', '+81-45-5678-9012', 'export@nissan.co.jp', 'Yokohama, Japan');

-- Insert sample customers
INSERT INTO cars.customers (customer_name, customer_title, contact_number, address, other_contacts, customer_type) VALUES
    ('Samanthika Perera', 'Ms', '717331843', 'Dondra', 'Amila: 0711492000', 'INDIVIDUAL'),
    ('KA Susantha', 'Mr', '716609525', 'Welegoda, Matara', NULL, 'INDIVIDUAL'),
    ('Dr Nalaka', 'Dr', '775093667', 'Colombo', NULL, 'INDIVIDUAL'),
    ('Sunil Perera', 'Mr', '771234567', 'Kandy', NULL, 'INDIVIDUAL'),
    ('Amara Silva', 'Ms', '778901234', 'Galle', NULL, 'INDIVIDUAL');

-- =====================================================
-- UPDATE SEQUENCES
-- =====================================================

SELECT setval('cars.vehicles_id_seq', 1);
SELECT setval('cars.vehicle_purchases_id_seq', 1);
SELECT setval('cars.vehicle_shipping_id_seq', 1);
SELECT setval('cars.vehicle_financials_id_seq', 1);
SELECT setval('cars.vehicle_sales_id_seq', 1);
SELECT setval('cars.customers_id_seq', (SELECT COALESCE(MAX(id), 1) FROM cars.customers));
SELECT setval('cars.suppliers_id_seq', (SELECT COALESCE(MAX(id), 1) FROM cars.suppliers));
SELECT setval('cars.customer_orders_id_seq', 1);

-- =====================================================
-- VERIFICATION
-- =====================================================

SELECT 'Schema creation completed successfully!' as status;

-- Show table counts
SELECT 'Vehicle Makes' as table_name, COUNT(*) as record_count FROM cars.vehicle_makes
UNION ALL
SELECT 'Vehicle Models', COUNT(*) FROM cars.vehicle_models
UNION ALL
SELECT 'Customers', COUNT(*) FROM cars.customers
UNION ALL
SELECT 'Suppliers', COUNT(*) FROM cars.suppliers;

-- =====================================================
-- VEHICLE SHARE TOKENS TABLE
-- =====================================================
CREATE TABLE cars.vehicle_share_tokens (
    id SERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL,
    token VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    include_details TEXT[] DEFAULT '{}',
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,

    CONSTRAINT fk_vehicle_share_tokens_vehicle
        FOREIGN KEY (vehicle_id)
            REFERENCES cars.vehicles(id)
            ON DELETE CASCADE
);

-- Indexes for vehicle_share_tokens
CREATE INDEX idx_vehicle_share_tokens_token
    ON cars.vehicle_share_tokens(token);

CREATE INDEX idx_vehicle_share_tokens_vehicle_id
    ON cars.vehicle_share_tokens(vehicle_id);

CREATE INDEX idx_vehicle_share_tokens_expires_at
    ON cars.vehicle_share_tokens(expires_at);

COMMENT ON TABLE cars.vehicle_share_tokens IS 'Stores shareable tokens for public vehicle data access';
COMMENT ON COLUMN cars.vehicle_share_tokens.token IS '64-character hex token for public access';
COMMENT ON COLUMN cars.vehicle_share_tokens.include_details IS 'Array of details to include: shipping, financial, purchase, images';
COMMENT ON COLUMN cars.vehicle_share_tokens.created_by IS 'User ID who created the share token';

