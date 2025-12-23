-- =====================================================
-- Migration: Add customer_id and remove legacy customer fields
-- Purpose: Normalize customer data by linking to customers table
-- Date: 2025-01-16
-- =====================================================

DROP VIEW IF EXISTS cars.vehicle_complete_info;

-- Add customer_id column to vehicle_sales table
ALTER TABLE cars.vehicle_sales
    ADD COLUMN IF NOT EXISTS customer_id BIGINT;

-- Add foreign key constraint to customers table
ALTER TABLE cars.vehicle_sales
    ADD CONSTRAINT fk_vehicle_sales_customer_id
        FOREIGN KEY (customer_id)
            REFERENCES cars.customers(id)
            ON DELETE SET NULL;

-- Create index for better query performance
CREATE INDEX IF NOT EXISTS idx_vehicle_sales_customer_id
    ON cars.vehicle_sales(customer_id);

-- Add comment to document the column purpose
COMMENT ON COLUMN cars.vehicle_sales.customer_id IS
    'Foreign key to customers table. Links the sale to an actual customer record.';

-- =====================================================
-- Remove legacy denormalized customer fields
-- =====================================================

-- Drop the legacy customer columns (data should be in customers table)
ALTER TABLE cars.vehicle_sales
    DROP COLUMN IF EXISTS sold_to_name,
    DROP COLUMN IF EXISTS sold_to_title,
    DROP COLUMN IF EXISTS contact_number,
    DROP COLUMN IF EXISTS customer_address,
    DROP COLUMN IF EXISTS other_contacts;

-- =====================================================
-- Optional: Migrate existing data
-- =====================================================
-- This attempts to link existing sales to customers by matching names
-- Uncomment and run if you want to migrate existing data

/*
UPDATE cars.vehicle_sales vs
SET customer_id = (
    SELECT c.id
    FROM cars.customers c
    WHERE LOWER(TRIM(c.customer_name)) = LOWER(TRIM(vs.sold_to_name))
    LIMIT 1
)
WHERE vs.sold_to_name IS NOT NULL
  AND vs.sold_to_name != ''
  AND vs.customer_id IS NULL;

-- Report migration results
SELECT
    COUNT(*) as total_sales,
    COUNT(customer_id) as sales_with_customer_id,
    COUNT(*) - COUNT(customer_id) as sales_without_customer_id
FROM cars.vehicle_sales;
*/

-- =====================================================
-- Verification Queries
-- =====================================================

-- Verify column was added
SELECT
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_schema = 'cars'
  AND table_name = 'vehicle_sales'
  AND column_name = 'customer_id';

-- Verify foreign key constraint
SELECT
    tc.constraint_name,
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name
FROM information_schema.table_constraints AS tc
         JOIN information_schema.key_column_usage AS kcu
              ON tc.constraint_name = kcu.constraint_name
                  AND tc.table_schema = kcu.table_schema
         JOIN information_schema.constraint_column_usage AS ccu
              ON ccu.constraint_name = tc.constraint_name
                  AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY'
  AND tc.table_schema = 'cars'
  AND tc.table_name = 'vehicle_sales'
  AND kcu.column_name = 'customer_id';
