-- Car Deals Database Schema - PostgreSQL Version
-- Based on analysis of Deals.xlsx file structure
-- Comprehensive schema for Japanese car import business

-- Create database (uncomment if needed)
-- CREATE DATABASE car_deals_db;
-- \c car_deals_db;

-- =====================================================
-- ENUMS AND CUSTOM TYPES
-- =====================================================

CREATE TYPE condition_status_enum AS ENUM ('REGISTERED', 'UNREGISTERED');
CREATE TYPE shipping_status_enum AS ENUM ('PROCESSING', 'SHIPPED', 'ARRIVED', 'CLEARED', 'DELIVERED');
CREATE TYPE sale_status_enum AS ENUM ('AVAILABLE', 'RESERVED', 'SOLD', 'CANCELLED');
CREATE TYPE customer_type_enum AS ENUM ('INDIVIDUAL', 'BUSINESS');
CREATE TYPE supplier_type_enum AS ENUM ('AUCTION', 'DEALER', 'INDIVIDUAL');
CREATE TYPE order_type_enum AS ENUM ('AUCTION', 'DIRECT', 'DEALER');
CREATE TYPE priority_level_enum AS ENUM ('NORMAL', 'HIGH', 'URGENT');
CREATE TYPE shipping_method_enum AS ENUM ('VESSEL', 'CONTAINER', 'RORO');
CREATE TYPE payment_method_enum AS ENUM ('CASH', 'FINANCING', 'LEASE', 'INSTALLMENT');
CREATE TYPE order_status_enum AS ENUM ('DRAFT', 'SUBMITTED', 'PROCESSING', 'MATCHED', 'COMPLETED', 'CANCELLED');
CREATE TYPE document_type_enum AS ENUM ('INVOICE', 'SHIPPING', 'CUSTOMS', 'INSPECTION', 'REGISTRATION', 'OTHER');
CREATE TYPE audit_action_enum AS ENUM ('INSERT', 'UPDATE', 'DELETE');

-- =====================================================
-- MAIN TABLES
-- =====================================================

-- Vehicles Master Table
CREATE TABLE vehicles (
                          id BIGSERIAL PRIMARY KEY,
                          code INTEGER UNIQUE NOT NULL,
                          make VARCHAR(50) NOT NULL,
                          model VARCHAR(100) NOT NULL,
                          trim_level VARCHAR(100),
                          year_of_manufacture INTEGER NOT NULL,
                          color VARCHAR(50) NOT NULL,
                          mileage_km INTEGER,
                          chassis_id VARCHAR(50) UNIQUE NOT NULL,
                          condition_status condition_status_enum DEFAULT 'UNREGISTERED',
                          year_of_registration INTEGER,
                          license_plate VARCHAR(20),
                          auction_grade VARCHAR(10),
                          auction_price DECIMAL(15,2),
                          cif_value DECIMAL(15,2),
                          currency VARCHAR(10) DEFAULT 'JPY',
                          hs_code VARCHAR(20),
                          invoice_fob_jpy DECIMAL(15,2),
                          registration_number VARCHAR(20),
                          record_date TIMESTAMP,
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for vehicles table
CREATE INDEX idx_vehicles_chassis_id ON vehicles(chassis_id);
CREATE INDEX idx_vehicles_make_model ON vehicles(make, model);
CREATE INDEX idx_vehicles_year ON vehicles(year_of_manufacture);
CREATE INDEX idx_vehicles_code ON vehicles(code);
CREATE INDEX idx_vehicles_make_year ON vehicles(make, year_of_manufacture);

-- Purchase Information Table
CREATE TABLE vehicle_purchases (
                                   id BIGSERIAL PRIMARY KEY,
                                   vehicle_id BIGINT NOT NULL,
                                   bought_from_name VARCHAR(100),
                                   bought_from_title VARCHAR(10),
                                   bought_from_contact VARCHAR(50),
                                   bought_from_address TEXT,
                                   bought_from_other_contacts TEXT,
                                   purchase_remarks TEXT,
                                   lc_bank VARCHAR(100),
                                   lc_number VARCHAR(50),
                                   lc_cost_jpy DECIMAL(15,2),
                                   purchase_date TIMESTAMP,
                                   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                                   CONSTRAINT fk_vehicle_purchases_vehicle_id FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

CREATE INDEX idx_vehicle_purchases_vehicle_id ON vehicle_purchases(vehicle_id);
CREATE INDEX idx_vehicle_purchases_purchase_date ON vehicle_purchases(purchase_date);

-- Shipping Information Table
CREATE TABLE vehicle_shipping (
                                  id BIGSERIAL PRIMARY KEY,
                                  vehicle_id BIGINT NOT NULL,
                                  vessel_name VARCHAR(100),
                                  departure_harbour VARCHAR(50),
                                  shipment_date TIMESTAMP,
                                  arrival_date TIMESTAMP,
                                  clearing_date TIMESTAMP,
                                  shipping_status shipping_status_enum DEFAULT 'PROCESSING',
                                  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                                  CONSTRAINT fk_vehicle_shipping_vehicle_id FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

CREATE INDEX idx_vehicle_shipping_vehicle_id ON vehicle_shipping(vehicle_id);
CREATE INDEX idx_vehicle_shipping_shipping_status ON vehicle_shipping(shipping_status);
CREATE INDEX idx_vehicle_shipping_shipment_date ON vehicle_shipping(shipment_date);
CREATE INDEX idx_vehicle_shipping_arrival_date ON vehicle_shipping(arrival_date);
CREATE INDEX idx_vehicle_shipping_dates ON vehicle_shipping(shipment_date, arrival_date);

-- Financial Information Table
CREATE TABLE vehicle_financials (
                                    id BIGSERIAL PRIMARY KEY,
                                    vehicle_id BIGINT NOT NULL,
                                    charges_lkr DECIMAL(15,2),
                                    tt_lkr DECIMAL(15,2),
                                    duty_lkr DECIMAL(15,2),
                                    clearing_lkr DECIMAL(15,2),
                                    other_expenses_lkr DECIMAL(15,2),
                                    total_cost_lkr DECIMAL(15,2) NOT NULL,
                                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                                    CONSTRAINT fk_vehicle_financials_vehicle_id FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

CREATE INDEX idx_vehicle_financials_vehicle_id ON vehicle_financials(vehicle_id);
CREATE INDEX idx_vehicle_financials_total_cost ON vehicle_financials(total_cost_lkr);

-- Sales Information Table
CREATE TABLE vehicle_sales (
                               id BIGSERIAL PRIMARY KEY,
                               vehicle_id BIGINT NOT NULL,
                               sold_date TIMESTAMP,
                               revenue DECIMAL(15,2),
                               profit DECIMAL(15,2),
                               sold_to_name VARCHAR(100),
                               sold_to_title VARCHAR(10),
                               contact_number VARCHAR(50),
                               customer_address TEXT,
                               other_contacts TEXT,
                               sale_remarks TEXT,
                               sale_status sale_status_enum DEFAULT 'AVAILABLE',
                               created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                               updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                               CONSTRAINT fk_vehicle_sales_vehicle_id FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

CREATE INDEX idx_vehicle_sales_vehicle_id ON vehicle_sales(vehicle_id);
CREATE INDEX idx_vehicle_sales_sold_date ON vehicle_sales(sold_date);
CREATE INDEX idx_vehicle_sales_sale_status ON vehicle_sales(sale_status);
CREATE INDEX idx_vehicle_sales_customer_name ON vehicle_sales(sold_to_name);
CREATE INDEX idx_vehicle_sales_profit ON vehicle_sales(profit);

-- =====================================================
-- REFERENCE TABLES
-- =====================================================

-- Vehicle Makes Reference Table
CREATE TABLE vehicle_makes (
                               id SERIAL PRIMARY KEY,
                               make_name VARCHAR(50) UNIQUE NOT NULL,
                               country_origin VARCHAR(50) DEFAULT 'Japan',
                               is_active BOOLEAN DEFAULT TRUE,
                               created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Vehicle Models Reference Table
CREATE TABLE vehicle_models (
                                id SERIAL PRIMARY KEY,
                                make_id INTEGER NOT NULL,
                                model_name VARCHAR(100) NOT NULL,
                                body_type VARCHAR(50),
                                fuel_type VARCHAR(30),
                                transmission_type VARCHAR(30),
                                engine_size_cc INTEGER,
                                is_active BOOLEAN DEFAULT TRUE,
                                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                                CONSTRAINT fk_vehicle_models_make_id FOREIGN KEY (make_id) REFERENCES vehicle_makes(id),
                                CONSTRAINT unique_make_model UNIQUE (make_id, model_name)
);

CREATE INDEX idx_vehicle_models_make_id ON vehicle_models(make_id);

-- Customers Table
CREATE TABLE customers (
                           id BIGSERIAL PRIMARY KEY,
                           customer_title VARCHAR(10),
                           customer_name VARCHAR(100) NOT NULL,
                           contact_number VARCHAR(50),
                           email VARCHAR(100),
                           address TEXT,
                           other_contacts TEXT,
                           customer_type customer_type_enum DEFAULT 'INDIVIDUAL',
                           is_active BOOLEAN DEFAULT TRUE,
                           created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                           updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_customers_customer_name ON customers(customer_name);
CREATE INDEX idx_customers_contact_number ON customers(contact_number);
CREATE INDEX idx_customers_email ON customers(email);

-- Suppliers/Dealers Table
CREATE TABLE suppliers (
                           id BIGSERIAL PRIMARY KEY,
                           supplier_name VARCHAR(100) NOT NULL,
                           supplier_title VARCHAR(10),
                           contact_number VARCHAR(50),
                           email VARCHAR(100),
                           address TEXT,
                           other_contacts TEXT,
                           supplier_type supplier_type_enum DEFAULT 'AUCTION',
                           country VARCHAR(50) DEFAULT 'Japan',
                           is_active BOOLEAN DEFAULT TRUE,
                           created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                           updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_suppliers_supplier_name ON suppliers(supplier_name);
CREATE INDEX idx_suppliers_supplier_type ON suppliers(supplier_type);

-- =====================================================
-- ORDER MANAGEMENT TABLES
-- =====================================================

-- Customer Orders Table (for new order system)
CREATE TABLE customer_orders (
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
                                 order_type order_type_enum DEFAULT 'AUCTION',
                                 expected_delivery_date DATE,
                                 priority_level priority_level_enum DEFAULT 'NORMAL',

    -- Shipping Preferences
                                 preferred_port VARCHAR(50),
                                 shipping_method shipping_method_enum DEFAULT 'VESSEL',
                                 include_insurance BOOLEAN DEFAULT TRUE,

    -- Financial Information
                                 budget_min DECIMAL(15,2),
                                 budget_max DECIMAL(15,2),
                                 payment_method payment_method_enum DEFAULT 'CASH',
                                 down_payment DECIMAL(15,2),

    -- Additional Information
                                 special_requests TEXT,
                                 internal_notes TEXT,

    -- Order Status
                                 order_status order_status_enum DEFAULT 'DRAFT',
                                 is_draft BOOLEAN DEFAULT FALSE,

    -- Timestamps
                                 order_date TIMESTAMP NOT NULL,
                                 completed_date TIMESTAMP,
                                 created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                 updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                                 CONSTRAINT fk_customer_orders_customer_id FOREIGN KEY (customer_id) REFERENCES customers(id)
);

CREATE INDEX idx_customer_orders_order_number ON customer_orders(order_number);
CREATE INDEX idx_customer_orders_customer_id ON customer_orders(customer_id);
CREATE INDEX idx_customer_orders_order_status ON customer_orders(order_status);
CREATE INDEX idx_customer_orders_order_date ON customer_orders(order_date);

-- Order-Vehicle Matching Table
CREATE TABLE order_vehicle_matches (
                                       id BIGSERIAL PRIMARY KEY,
                                       order_id BIGINT NOT NULL,
                                       vehicle_id BIGINT NOT NULL,
                                       match_score DECIMAL(5,2), -- Percentage match
                                       matched_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                       is_selected BOOLEAN DEFAULT FALSE,

                                       CONSTRAINT fk_order_vehicle_matches_order_id FOREIGN KEY (order_id) REFERENCES customer_orders(id) ON DELETE CASCADE,
                                       CONSTRAINT fk_order_vehicle_matches_vehicle_id FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE,
                                       CONSTRAINT unique_order_vehicle UNIQUE (order_id, vehicle_id)
);

CREATE INDEX idx_order_vehicle_matches_order_id ON order_vehicle_matches(order_id);
CREATE INDEX idx_order_vehicle_matches_vehicle_id ON order_vehicle_matches(vehicle_id);

-- =====================================================
-- AUDIT AND TRACKING TABLES
-- =====================================================

-- Audit Log Table
CREATE TABLE audit_logs (
                            id BIGSERIAL PRIMARY KEY,
                            table_name VARCHAR(50) NOT NULL,
                            record_id BIGINT NOT NULL,
                            action audit_action_enum NOT NULL,
                            old_values JSONB,
                            new_values JSONB,
                            user_id VARCHAR(50),
                            timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_table_record ON audit_logs(table_name, record_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);

-- Document Attachments Table
CREATE TABLE vehicle_documents (
                                   id BIGSERIAL PRIMARY KEY,
                                   vehicle_id BIGINT NOT NULL,
                                   document_type document_type_enum NOT NULL,
                                   document_name VARCHAR(255) NOT NULL,
                                   file_path VARCHAR(500),
                                   file_size_bytes BIGINT,
                                   mime_type VARCHAR(100),
                                   upload_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                                   CONSTRAINT fk_vehicle_documents_vehicle_id FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

CREATE INDEX idx_vehicle_documents_vehicle_id ON vehicle_documents(vehicle_id);
CREATE INDEX idx_vehicle_documents_document_type ON vehicle_documents(document_type);

-- =====================================================
-- VIEWS FOR COMMON QUERIES
-- =====================================================

-- Complete Vehicle Information View
CREATE VIEW vehicle_complete_info AS
SELECT
    v.id,
    v.code,
    v.make,
    v.model,
    v.trim_level,
    v.year_of_manufacture,
    v.color,
    v.mileage_km,
    v.chassis_id,
    v.condition_status,
    v.auction_grade,
    v.cif_value,
    v.currency,

    -- Purchase info
    vp.bought_from_name,
    vp.bought_from_title,
    vp.purchase_date,

    -- Shipping info
    vs.vessel_name,
    vs.departure_harbour,
    vs.shipment_date,
    vs.arrival_date,
    vs.clearing_date,
    vs.shipping_status,

    -- Financial info
    vf.total_cost_lkr,
    vf.charges_lkr,
    vf.duty_lkr,
    vf.clearing_lkr,
    vf.other_expenses_lkr,

    -- Sales info
    vsl.sold_date,
    vsl.revenue,
    vsl.profit,
    vsl.sold_to_name,
    vsl.sold_to_title,
    vsl.contact_number,
    vsl.customer_address,
    vsl.sale_status,

    v.created_at,
    v.updated_at
FROM vehicles v
         LEFT JOIN vehicle_purchases vp ON v.id = vp.vehicle_id
         LEFT JOIN vehicle_shipping vs ON v.id = vs.vehicle_id
         LEFT JOIN vehicle_financials vf ON v.id = vf.vehicle_id
         LEFT JOIN vehicle_sales vsl ON v.id = vsl.vehicle_id;

-- Sales Summary View
CREATE VIEW sales_summary AS
SELECT
    TO_CHAR(vsl.sold_date, 'YYYY-MM') as sale_month,
    COUNT(*) as vehicles_sold,
    SUM(vf.total_cost_lkr) as total_cost,
    SUM(vsl.revenue) as total_revenue,
    SUM(vsl.profit) as total_profit,
    AVG(vsl.profit) as avg_profit_per_vehicle,
    (SUM(vsl.profit) / SUM(vf.total_cost_lkr)) * 100 as profit_margin_percentage
FROM vehicle_sales vsl
         JOIN vehicle_financials vf ON vsl.vehicle_id = vf.vehicle_id
WHERE vsl.sold_date IS NOT NULL
GROUP BY TO_CHAR(vsl.sold_date, 'YYYY-MM')
ORDER BY sale_month DESC;

-- Inventory Status View
CREATE VIEW inventory_status AS
SELECT
    vs.shipping_status,
    vsl.sale_status,
    COUNT(*) as vehicle_count,
    SUM(vf.total_cost_lkr) as total_investment
FROM vehicles v
         LEFT JOIN vehicle_shipping vs ON v.id = vs.vehicle_id
         LEFT JOIN vehicle_sales vsl ON v.id = vsl.vehicle_id
         LEFT JOIN vehicle_financials vf ON v.id = vf.vehicle_id
GROUP BY vs.shipping_status, vsl.sale_status;

-- =====================================================
-- FUNCTIONS (PostgreSQL equivalent of stored procedures)
-- =====================================================

-- Function to get vehicle complete information
CREATE OR REPLACE FUNCTION get_vehicle_details(vehicle_code INTEGER)
RETURNS TABLE (
    id BIGINT,
    code INTEGER,
    make VARCHAR(50),
    model VARCHAR(100),
    trim_level VARCHAR(100),
    year_of_manufacture INTEGER,
    color VARCHAR(50),
    mileage_km INTEGER,
    chassis_id VARCHAR(50),
    condition_status condition_status_enum,
    auction_grade VARCHAR(10),
    cif_value DECIMAL(15,2),
    currency VARCHAR(10),
    bought_from_name VARCHAR(100),
    bought_from_title VARCHAR(10),
    purchase_date TIMESTAMP,
    vessel_name VARCHAR(100),
    departure_harbour VARCHAR(50),
    shipment_date TIMESTAMP,
    arrival_date TIMESTAMP,
    clearing_date TIMESTAMP,
    shipping_status shipping_status_enum,
    total_cost_lkr DECIMAL(15,2),
    charges_lkr DECIMAL(15,2),
    duty_lkr DECIMAL(15,2),
    clearing_lkr DECIMAL(15,2),
    other_expenses_lkr DECIMAL(15,2),
    sold_date TIMESTAMP,
    revenue DECIMAL(15,2),
    profit DECIMAL(15,2),
    sold_to_name VARCHAR(100),
    sold_to_title VARCHAR(10),
    contact_number VARCHAR(50),
    customer_address TEXT,
    sale_status sale_status_enum,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
) AS $$
BEGIN
RETURN QUERY
SELECT * FROM vehicle_complete_info WHERE vehicle_complete_info.code = vehicle_code;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate profit margins
CREATE OR REPLACE FUNCTION calculate_profit_margins(start_date DATE, end_date DATE)
RETURNS TABLE (
    make VARCHAR(50),
    model VARCHAR(100),
    vehicles_sold BIGINT,
    avg_profit DECIMAL(15,2),
    avg_profit_margin_percent DECIMAL(15,2)
) AS $$
BEGIN
RETURN QUERY
SELECT
    v.make,
    v.model,
    COUNT(*) as vehicles_sold,
    AVG(vsl.profit) as avg_profit,
    AVG((vsl.profit / vf.total_cost_lkr) * 100) as avg_profit_margin_percent
FROM vehicles v
         JOIN vehicle_sales vsl ON v.id = vsl.vehicle_id
         JOIN vehicle_financials vf ON v.id = vf.vehicle_id
WHERE vsl.sold_date BETWEEN start_date AND end_date
GROUP BY v.make, v.model
ORDER BY avg_profit_margin_percent DESC;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- TRIGGERS FOR AUTO-UPDATING timestamps
-- =====================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for updating updated_at columns
CREATE TRIGGER update_vehicles_updated_at BEFORE UPDATE ON vehicles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vehicle_purchases_updated_at BEFORE UPDATE ON vehicle_purchases
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vehicle_shipping_updated_at BEFORE UPDATE ON vehicle_shipping
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vehicle_financials_updated_at BEFORE UPDATE ON vehicle_financials
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vehicle_sales_updated_at BEFORE UPDATE ON vehicle_sales
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_suppliers_updated_at BEFORE UPDATE ON suppliers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_customer_orders_updated_at BEFORE UPDATE ON customer_orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- TRIGGERS FOR AUDIT LOGGING
-- =====================================================

-- Function for audit logging
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_logs (table_name, record_id, action, new_values, user_id)
        VALUES (TG_TABLE_NAME, NEW.id, 'INSERT', to_jsonb(NEW), current_user);
RETURN NEW;
ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_logs (table_name, record_id, action, old_values, new_values, user_id)
        VALUES (TG_TABLE_NAME, NEW.id, 'UPDATE', to_jsonb(OLD), to_jsonb(NEW), current_user);
RETURN NEW;
ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit_logs (table_name, record_id, action, old_values, user_id)
        VALUES (TG_TABLE_NAME, OLD.id, 'DELETE', to_jsonb(OLD), current_user);
RETURN OLD;
END IF;
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create audit triggers for main tables
CREATE TRIGGER vehicles_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON vehicles
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER vehicle_sales_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON vehicle_sales
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- =====================================================
-- INITIAL DATA SETUP
-- =====================================================

-- Insert common vehicle makes
INSERT INTO vehicle_makes (make_name, country_origin) VALUES
                                                          ('Toyota', 'Japan'),
                                                          ('Honda', 'Japan'),
                                                          ('Nissan', 'Japan'),
                                                          ('Mazda', 'Japan'),
                                                          ('Suzuki', 'Japan'),
                                                          ('Mitsubishi', 'Japan'),
                                                          ('Subaru', 'Japan'),
                                                          ('Lexus', 'Japan'),
                                                          ('Infiniti', 'Japan'),
                                                          ('Acura', 'Japan');

-- Insert common Toyota models
INSERT INTO vehicle_models (make_id, model_name, body_type) VALUES
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Toyota'), 'Aqua', 'Hatchback'),
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Toyota'), 'Prius', 'Hatchback'),
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Toyota'), 'Vitz', 'Hatchback'),
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Toyota'), 'Axio', 'Sedan'),
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Toyota'), 'Fielder', 'Wagon'),
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Toyota'), 'Allion', 'Sedan'),
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Toyota'), 'Premio', 'Sedan'),
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Toyota'), 'Voxy', 'Minivan'),
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Toyota'), 'Noah', 'Minivan');

-- Insert common Honda models
INSERT INTO vehicle_models (make_id, model_name, body_type) VALUES
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Honda'), 'Fit', 'Hatchback'),
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Honda'), 'Vezel', 'SUV'),
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Honda'), 'Grace', 'Sedan'),
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Honda'), 'Freed', 'Minivan'),
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Honda'), 'Shuttle', 'Wagon'),
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Honda'), 'CR-V', 'SUV'),
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Honda'), 'HR-V', 'SUV'),
                                                                ((SELECT id FROM vehicle_makes WHERE make_name = 'Honda'), 'Stepwgn', 'Minivan');

-- =====================================================
-- USER MANAGEMENT (Adjust as needed)
-- =====================================================

-- Create application user (uncomment and modify as needed)
-- CREATE USER car_deals_app WITH PASSWORD 'secure_password_here';
-- GRANT CONNECT ON DATABASE car_deals_db TO car_deals_app;
-- GRANT USAGE ON SCHEMA public TO car_deals_app;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO car_deals_app;
-- GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO car_deals_app;

-- Grant permissions for future tables
-- ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO car_deals_app;
-- ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO car_deals_app;



---------------------- Mock data for testing -----------------------------------------------

-- PostgreSQL INSERT Queries for Car Deals Database
-- Based on Excel data structure analysis

-- =====================================================
-- INSERT VEHICLES DATA
-- =====================================================

-- Sample data based on your Excel analysis
INSERT INTO vehicles (code, make, model, trim_level, year_of_manufacture, color, mileage_km, chassis_id, condition_status, auction_grade, cif_value, currency, record_date) VALUES
                                                                                                                                                                                (1, 'Toyota', 'Aqua', 'S', 2012, 'Silver', 16600, 'NHP10-6096289', 'UNREGISTERED', 'A/B', 1500000, 'JPY', '2013-12-26'),
                                                                                                                                                                                (2, 'Toyota', 'Aqua', 'S', 2012, 'White', 16000, 'NHP10-6050119', 'UNREGISTERED', '5/A', NULL, 'JPY', '2013-12-31'),
                                                                                                                                                                                (3, 'Honda', 'Fit', 'She''s', 2013, 'Pearl White', 4704, 'GP1-1206999', 'UNREGISTERED', '5/A', 1590000, 'JPY', '2013-12-26'),
                                                                                                                                                                                (4, 'Toyota', 'Prius', 'S', 2012, 'Pearl White', 8520, 'ZVW30-5563533', 'UNREGISTERED', '5AA', 2080000, 'JPY', '2014-01-16'),
                                                                                                                                                                                (5, 'Toyota', 'Aqua', 'G', 2013, 'Pearl White', 2328, 'NHP10-2176888', 'UNREGISTERED', '6AA', 1910000, 'JPY', '2013-12-20'),
                                                                                                                                                                                (6, 'Toyota', 'Aqua', 'S-LED', 2012, 'Grey', 11000, 'NHP10-2009138', 'UNREGISTERED', '4.5B', 1880000, 'JPY', '2014-01-28'),
                                                                                                                                                                                (7, 'Toyota', 'Aqua', 'G', 2012, 'Silver', 3000, 'NHP10-2127667', 'UNREGISTERED', NULL, 1810000, 'JPY', '2014-01-30'),
                                                                                                                                                                                (8, 'Toyota', 'Aqua', 'S', 2014, 'Pearl White', 4, 'NHP10-2284450', 'UNREGISTERED', 'SA', 1940000, 'JPY', '2014-02-25'),
                                                                                                                                                                                (9, 'Honda', 'Fit', 'Smart Selection', 2013, 'Silver', 18000, 'GP1-1234558', 'UNREGISTERED', '5A', 1560000, 'JPY', '2014-04-15'),
                                                                                                                                                                                (10, 'Nissan', 'Note', 'X', 2013, 'Black', 12000, 'E12-567890', 'UNREGISTERED', '4/B', 1350000, 'JPY', '2014-05-10'),
                                                                                                                                                                                (11, 'Mazda', 'Demio', '13S', 2014, 'Red', 3500, 'DJ5FS-800123', 'UNREGISTERED', '4.5/B', 1450000, 'JPY', '2014-03-15'),
                                                                                                                                                                                (12, 'Honda', 'Vezel', 'Hybrid Z', 2014, 'Blue', 8900, 'RU3-1100456', 'UNREGISTERED', '5/A', 2200000, 'JPY', '2014-04-20'),
                                                                                                                                                                                (13, 'Toyota', 'Vitz', 'F', 2013, 'White', 25000, 'NCP131-2345678', 'UNREGISTERED', '4/B', 1200000, 'JPY', '2014-06-01'),
                                                                                                                                                                                (14, 'Honda', 'Grace', 'Hybrid DX', 2014, 'Silver', 15000, 'GM4-1111111', 'UNREGISTERED', '5/A', 1750000, 'JPY', '2014-07-15'),
                                                                                                                                                                                (15, 'Toyota', 'Fielder', 'X', 2013, 'Dark Blue', 22000, 'NZE161G-3456789', 'UNREGISTERED', '4.5/B', 1400000, 'JPY', '2014-08-20');

-- =====================================================
-- INSERT VEHICLE PURCHASES DATA
-- =====================================================

INSERT INTO vehicle_purchases (vehicle_id, bought_from_name, bought_from_title, bought_from_contact, bought_from_address, lc_cost_jpy, purchase_date, purchase_remarks) VALUES
                                                                                                                                                                            (1, NULL, NULL, NULL, NULL, 1384900, '2013-12-25', 'Auction purchase'),
                                                                                                                                                                            (2, 'Mr Duminda', 'Mr', NULL, NULL, 1272700, '2013-12-31', 'Bought by Mr Duminda for his customer'),
                                                                                                                                                                            (3, NULL, NULL, NULL, NULL, 1384900, '2013-12-25', 'Auction purchase'),
                                                                                                                                                                            (4, NULL, NULL, NULL, NULL, 2064250, '2014-01-15', 'Auction purchase'),
                                                                                                                                                                            (5, NULL, NULL, NULL, NULL, 1402900, '2013-12-19', 'Auction purchase'),
                                                                                                                                                                            (6, NULL, NULL, NULL, NULL, 1428750, '2014-01-28', 'Auction purchase'),
                                                                                                                                                                            (7, NULL, NULL, NULL, NULL, 1428750, '2014-01-30', 'Auction purchase'),
                                                                                                                                                                            (8, NULL, NULL, NULL, NULL, 1662920, '2014-02-25', 'Auction purchase'),
                                                                                                                                                                            (9, NULL, NULL, NULL, NULL, 1410094, '2014-04-15', 'Auction purchase'),
                                                                                                                                                                            (10, NULL, NULL, NULL, NULL, 1500000, '2014-05-10', 'Auction purchase'),
                                                                                                                                                                            (11, NULL, NULL, NULL, NULL, 1600000, '2014-03-15', 'Auction purchase'),
                                                                                                                                                                            (12, NULL, NULL, NULL, NULL, 1800000, '2014-04-20', 'Auction purchase'),
                                                                                                                                                                            (13, NULL, NULL, NULL, NULL, 1100000, '2014-06-01', 'Auction purchase'),
                                                                                                                                                                            (14, NULL, NULL, NULL, NULL, 1650000, '2014-07-15', 'Auction purchase'),
                                                                                                                                                                            (15, NULL, NULL, NULL, NULL, 1350000, '2014-08-20', 'Auction purchase');

-- =====================================================
-- INSERT VEHICLE SHIPPING DATA
-- =====================================================

INSERT INTO vehicle_shipping (vehicle_id, vessel_name, departure_harbour, shipment_date, arrival_date, clearing_date, shipping_status) VALUES
                                                                                                                                           (1, 'Delphinus Leader v24', 'Nagoya', '2013-12-26', '2013-12-13', '2014-01-21', 'CLEARED'),
                                                                                                                                           (2, NULL, NULL, NULL, '2014-01-01', '2014-01-01', 'CLEARED'),
                                                                                                                                           (3, 'Delphinus Leader v24', 'Nagoya', '2013-12-26', '2013-12-26', '2014-01-21', 'CLEARED'),
                                                                                                                                           (4, 'Noble Ace', 'Yokohama', '2014-01-16', '2014-01-29', '2014-02-06', 'CLEARED'),
                                                                                                                                           (5, 'Opal Ace v0020A', 'Kobe', '2013-12-20', NULL, '2014-01-21', 'CLEARED'),
                                                                                                                                           (6, 'Chang Chuan', 'Osaka', '2014-01-29', '2014-02-11', '2014-02-19', 'CLEARED'),
                                                                                                                                           (7, 'Swift Ace', 'Yokohama', '2014-01-31', '2014-02-12', '2014-02-19', 'CLEARED'),
                                                                                                                                           (8, 'Grand Race V7', 'Kobe', '2014-02-26', '2014-03-14', '2014-03-21', 'CLEARED'),
                                                                                                                                           (9, 'Jasper Arrow', 'Osaka', '2014-04-16', '2014-04-30', '2014-04-30', 'CLEARED'),
                                                                                                                                           (10, 'Pacific Breeze', 'Kobe', '2014-05-11', NULL, NULL, 'SHIPPED'),
                                                                                                                                           (11, 'Morning Crystal', 'Yokohama', '2014-03-16', '2014-03-31', '2014-04-11', 'CLEARED'),
                                                                                                                                           (12, 'Asian Majesty', 'Nagoya', '2014-04-21', '2014-05-06', NULL, 'ARRIVED'),
                                                                                                                                           (13, 'Ocean Pioneer', 'Tokyo', '2014-06-02', '2014-06-18', '2014-06-25', 'CLEARED'),
                                                                                                                                           (14, 'Sea Master', 'Yokohama', '2014-07-16', '2014-08-01', '2014-08-08', 'CLEARED'),
                                                                                                                                           (15, 'Blue Wave', 'Kobe', '2014-08-21', NULL, NULL, 'SHIPPED');

-- =====================================================
-- INSERT VEHICLE FINANCIALS DATA
-- =====================================================

INSERT INTO vehicle_financials (vehicle_id, charges_lkr, tt_lkr, duty_lkr, clearing_lkr, other_expenses_lkr, total_cost_lkr) VALUES
                                                                                                                                 (1, 7274, 514000, 1158520, 22000, 1000, 3087694),
                                                                                                                                 (2, 15000, NULL, 1157200, NULL, 100000, 2544900),
                                                                                                                                 (3, 7216, 629650, 1194995, 22000, 500, 3239261),
                                                                                                                                 (4, 12158, 619200, 1749410, 22000, 3000, 4470018),
                                                                                                                                 (5, 7234, 1040850, 1280145, 26183, 2500, 3759812),
                                                                                                                                 (6, 5329, 1014000, 1178995, 22000, 104500, 3753574),
                                                                                                                                 (7, 5329, 923000, 1181300, 22000, 17000, 3577379),
                                                                                                                                 (8, 10658, 827520, 1316515, 22000, 209670, 4049283),
                                                                                                                                 (9, 5223, 593400, 1099075, 22000, 0, 3129792),
                                                                                                                                 (10, 8000, 650000, 1200000, 25000, 5000, 2888000),
                                                                                                                                 (11, 7500, 720000, 1150000, 24000, 8000, 2909500),
                                                                                                                                 (12, 9000, 850000, 1800000, 28000, 15000, 4502000),
                                                                                                                                 (13, 6000, 580000, 950000, 20000, 3000, 2659000),
                                                                                                                                 (14, 8500, 780000, 1400000, 26000, 12000, 3876500),
                                                                                                                                 (15, 7000, 650000, 1100000, 22000, 6000, 3135000);

-- =====================================================
-- INSERT VEHICLE SALES DATA
-- =====================================================

INSERT INTO vehicle_sales (vehicle_id, sold_date, revenue, profit, sold_to_name, sold_to_title, contact_number, customer_address, other_contacts, sale_remarks, sale_status) VALUES
                                                                                                                                                                                 (1, '2014-02-12', 3250000, 162306, 'Samanthika Perera', 'Ms', '717331843/0412221319', 'Dondra', 'Amila: 0711492000, Mr Rohana: 0718577460', 'Cheque received from Siyapatha Finance', 'SOLD'),
                                                                                                                                                                                 (2, NULL, 2835100, 290200, NULL, NULL, NULL, NULL, 'Mr Duminda', NULL, 'AVAILABLE'),
                                                                                                                                                                                 (3, '2014-01-28', 3350000, 110739, 'Mohan (Com Bank)', 'Mr', NULL, NULL, NULL, NULL, 'SOLD'),
                                                                                                                                                                                 (4, '2014-02-10', 4625000, 154982, 'KA Susantha', 'Mr', '716609525', 'Welegoda, Matara', NULL, NULL, 'SOLD'),
                                                                                                                                                                                 (5, '2014-01-21', 3850000, 90188, 'Nalaka', 'Dr', '775093667', NULL, NULL, NULL, 'SOLD'),
                                                                                                                                                                                 (6, '2014-02-28', 3820000, 66426, 'Kasun', 'Mr', '717226641', NULL, NULL, 'Broker (Kotawaya maama) fee: 100,000 (added to other expenses)', 'SOLD'),
                                                                                                                                                                                 (7, '2014-02-24', 3740000, 162621, 'Jagath Mendis', 'Mr', '773823293', 'SLIC', NULL, 'Broker (Ishan SLIC) Fee: 15,000 (added to other expenses)', 'SOLD'),
                                                                                                                                                                                 (8, '2014-04-04', 4175330, 126047, 'Pinnapol', 'Mr', '773790999', NULL, NULL, 'Brokers(Dilruk, Malinda)fee: 90,000', 'SOLD'),
                                                                                                                                                                                 (9, '2014-05-05', 3200000, 70208, 'Ravindra', 'Dr', '714043230', NULL, NULL, NULL, 'SOLD'),
                                                                                                                                                                                 (10, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'In transit', 'AVAILABLE'),
                                                                                                                                                                                 (11, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'Still in stock', 'AVAILABLE'),
                                                                                                                                                                                 (12, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'Clearing in progress', 'AVAILABLE'),
                                                                                                                                                                                 (13, '2014-07-15', 2800000, 141000, 'Sunil Perera', 'Mr', '771234567', 'Kandy', NULL, NULL, 'SOLD'),
                                                                                                                                                                                 (14, '2014-08-20', 4100000, 223500, 'Amara Silva', 'Ms', '778901234', 'Galle', NULL, NULL, 'SOLD'),
                                                                                                                                                                                 (15, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'Recently shipped', 'AVAILABLE');

-- =====================================================
-- INSERT CUSTOMERS DATA (Extract from sales)
-- =====================================================

INSERT INTO customers (customer_name, customer_title, contact_number, address, other_contacts, customer_type) VALUES
                                                                                                                  ('Samanthika Perera', 'Ms', '717331843/0412221319', 'Dondra', 'Amila: 0711492000, Mr Rohana: 0718577460', 'INDIVIDUAL'),
                                                                                                                  ('Mohan', 'Mr', NULL, 'Commercial Bank', NULL, 'BUSINESS'),
                                                                                                                  ('KA Susantha', 'Mr', '716609525', 'Welegoda, Matara', NULL, 'INDIVIDUAL'),
                                                                                                                  ('Nalaka', 'Dr', '775093667', NULL, NULL, 'INDIVIDUAL'),
                                                                                                                  ('Kasun', 'Mr', '717226641', NULL, NULL, 'INDIVIDUAL'),
                                                                                                                  ('Jagath Mendis', 'Mr', '773823293', 'SLIC', NULL, 'INDIVIDUAL'),
                                                                                                                  ('Pinnapol', 'Mr', '773790999', NULL, NULL, 'INDIVIDUAL'),
                                                                                                                  ('Ravindra', 'Dr', '714043230', NULL, NULL, 'INDIVIDUAL'),
                                                                                                                  ('Sunil Perera', 'Mr', '771234567', 'Kandy', NULL, 'INDIVIDUAL'),
                                                                                                                  ('Amara Silva', 'Ms', '778901234', 'Galle', NULL, 'INDIVIDUAL');

-- =====================================================
-- INSERT SAMPLE CUSTOMER ORDERS
-- =====================================================

INSERT INTO customer_orders (order_number, customer_id, preferred_make, preferred_model, preferred_year_min, preferred_year_max, preferred_color, preferred_trim_level, max_mileage_km, min_auction_grade, required_features, order_type, expected_delivery_date, priority_level, preferred_port, shipping_method, include_insurance, budget_min, budget_max, payment_method, special_requests, order_status, is_draft, order_date) VALUES
                                                                                                                                                                                                                                                                                                                                                                                                                                        ('ORD-2024-001', 1, 'Toyota', 'Aqua', 2020, 2024, 'Pearl White', 'G', 20000, '5/A', '["Navigation System", "Reverse Camera", "Smart Key"]', 'AUCTION', '2024-12-31', 'NORMAL', 'Nagoya', 'VESSEL', true, 3000000, 4000000, 'CASH', 'Low mileage preferred', 'SUBMITTED', false, '2024-09-01'),
                                                                                                                                                                                                                                                                                                                                                                                                                                        ('ORD-2024-002', 3, 'Honda', 'Vezel', 2021, 2024, 'Blue', 'Hybrid Z', 15000, '5AA', '["Navigation System", "Alloy Wheels", "Auto AC"]', 'AUCTION', '2024-11-30', 'HIGH', 'Yokohama', 'CONTAINER', true, 4500000, 6000000, 'FINANCING', 'Hybrid model only', 'PROCESSING', false, '2024-09-05'),
                                                                                                                                                                                                                                                                                                                                                                                                                                        ('ORD-2024-003', 5, 'Toyota', 'Prius', 2019, 2023, 'Silver', 'S', 30000, '4/B', '["ETC", "Power Steering", "ABS"]', 'DIRECT', '2024-10-15', 'URGENT', 'Kobe', 'VESSEL', false, 3500000, 4500000, 'CASH', 'Urgent delivery required', 'MATCHED', false, '2024-08-20'),
                                                                                                                                                                                                                                                                                                                                                                                                                                        ('ORD-2024-004', NULL, 'Nissan', 'Note', 2020, 2024, 'Black', 'X', 25000, '5/A', '["Navigation System", "Reverse Camera"]', 'AUCTION', '2024-12-15', 'NORMAL', 'Tokyo', 'VESSEL', true, 2500000, 3500000, 'INSTALLMENT', NULL, 'DRAFT', true, '2024-09-06');

-- =====================================================
-- INSERT SAMPLE SUPPLIERS
-- =====================================================

INSERT INTO suppliers (supplier_name, supplier_title, supplier_type, country, contact_number, email, address) VALUES
                                                                                                                  ('USS Auction', NULL, 'AUCTION', 'Japan', '+81-3-1234-5678', 'info@uss.co.jp', 'Tokyo, Japan'),
                                                                                                                  ('Honda Japan', NULL, 'DEALER', 'Japan', '+81-3-2345-6789', 'export@honda.co.jp', 'Tokyo, Japan'),
                                                                                                                  ('Toyota Motor Corporation', NULL, 'DEALER', 'Japan', '+81-3-3456-7890', 'export@toyota.co.jp', 'Toyota City, Japan'),
                                                                                                                  ('JAA (Japan Auto Auction)', NULL, 'AUCTION', 'Japan', '+81-6-4567-8901', 'contact@jaa.co.jp', 'Osaka, Japan'),
                                                                                                                  ('Nissan Export Division', NULL, 'DEALER', 'Japan', '+81-45-5678-9012', 'export@nissan.co.jp', 'Yokohama, Japan');

-- =====================================================
-- UPDATE SEQUENCES (to ensure proper ID continuation)
-- =====================================================

-- Reset sequences to continue from current max values
SELECT setval('vehicles_id_seq', (SELECT COALESCE(MAX(id), 1) FROM vehicles));
SELECT setval('vehicle_purchases_id_seq', (SELECT COALESCE(MAX(id), 1) FROM vehicle_purchases));
SELECT setval('vehicle_shipping_id_seq', (SELECT COALESCE(MAX(id), 1) FROM vehicle_shipping));
SELECT setval('vehicle_financials_id_seq', (SELECT COALESCE(MAX(id), 1) FROM vehicle_financials));
SELECT setval('vehicle_sales_id_seq', (SELECT COALESCE(MAX(id), 1) FROM vehicle_sales));
SELECT setval('customers_id_seq', (SELECT COALESCE(MAX(id), 1) FROM customers));
SELECT setval('customer_orders_id_seq', (SELECT COALESCE(MAX(id), 1) FROM customer_orders));
SELECT setval('suppliers_id_seq', (SELECT COALESCE(MAX(id), 1) FROM suppliers));

-- =====================================================
-- VERIFY DATA INSERTION
-- =====================================================

-- Check inserted data
SELECT 'Vehicles' as table_name, COUNT(*) as record_count FROM vehicles
UNION ALL
SELECT 'Vehicle Purchases', COUNT(*) FROM vehicle_purchases
UNION ALL
SELECT 'Vehicle Shipping', COUNT(*) FROM vehicle_shipping
UNION ALL
SELECT 'Vehicle Financials', COUNT(*) FROM vehicle_financials
UNION ALL
SELECT 'Vehicle Sales', COUNT(*) FROM vehicle_sales
UNION ALL
SELECT 'Customers', COUNT(*) FROM customers
UNION ALL
SELECT 'Customer Orders', COUNT(*) FROM customer_orders
UNION ALL
SELECT 'Suppliers', COUNT(*) FROM suppliers;

-- Sample query to test the complete view
SELECT
    v.code,
    v.make,
    v.model,
    v.color,
    v.year_of_manufacture,
    vs.shipping_status,
    vf.total_cost_lkr,
    vsl.revenue,
    vsl.profit,
    vsl.sold_to_name
FROM vehicles v
         LEFT JOIN vehicle_shipping vs ON v.id = vs.vehicle_id
         LEFT JOIN vehicle_financials vf ON v.id = vf.vehicle_id
         LEFT JOIN vehicle_sales vsl ON v.id = vsl.vehicle_id
ORDER BY v.code
    LIMIT 10;

CREATE TABLE vehicle_images (
                                id SERIAL PRIMARY KEY,
                                vehicle_id INTEGER NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
                                filename VARCHAR(255) NOT NULL,
                                original_name VARCHAR(255),
                                file_path VARCHAR(500) NOT NULL,
                                file_size INTEGER,
                                mime_type VARCHAR(100),
                                is_primary BOOLEAN DEFAULT false,
                                upload_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                display_order INTEGER DEFAULT 0
);

CREATE INDEX idx_vehicle_images_vehicle_id ON vehicle_images(vehicle_id);