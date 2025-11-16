-- =====================================================
-- Migration: Create vehicle_shipping_history table
-- Purpose: Track all shipment status changes with timestamps
-- Date: 2025-01-16
-- =====================================================

-- Create the shipping history table
CREATE TABLE IF NOT EXISTS cars.vehicle_shipping_history (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL,
    old_status cars.shipping_status_enum,
    new_status cars.shipping_status_enum NOT NULL,
    vessel_name VARCHAR(100),
    departure_harbour VARCHAR(50),
    shipment_date TIMESTAMP,
    arrival_date TIMESTAMP,
    clearing_date TIMESTAMP,
    changed_by VARCHAR(100),  -- User who made the change
    change_remarks TEXT,       -- Optional notes about the change
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_shipping_history_vehicle_id
        FOREIGN KEY (vehicle_id)
        REFERENCES cars.vehicles(id)
        ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_shipping_history_vehicle_id
    ON cars.vehicle_shipping_history(vehicle_id);

CREATE INDEX IF NOT EXISTS idx_shipping_history_changed_at
    ON cars.vehicle_shipping_history(changed_at);

CREATE INDEX IF NOT EXISTS idx_shipping_history_new_status
    ON cars.vehicle_shipping_history(new_status);

-- Add comments
COMMENT ON TABLE cars.vehicle_shipping_history IS
    'Tracks all changes to vehicle shipping status with timestamps and user information';

COMMENT ON COLUMN cars.vehicle_shipping_history.old_status IS
    'Previous shipping status before the change';

COMMENT ON COLUMN cars.vehicle_shipping_history.new_status IS
    'New shipping status after the change';

COMMENT ON COLUMN cars.vehicle_shipping_history.changed_by IS
    'Username or ID of the person who made the change';

-- =====================================================
-- Create trigger function to automatically log changes
-- =====================================================

CREATE OR REPLACE FUNCTION cars.log_shipping_status_change()
RETURNS TRIGGER AS $$
BEGIN
    -- Only log if status actually changed
    IF (TG_OP = 'UPDATE' AND OLD.shipping_status IS DISTINCT FROM NEW.shipping_status) THEN
        INSERT INTO cars.vehicle_shipping_history (
            vehicle_id,
            old_status,
            new_status,
            vessel_name,
            departure_harbour,
            shipment_date,
            arrival_date,
            clearing_date,
            changed_by,
            changed_at
        ) VALUES (
            NEW.vehicle_id,
            OLD.shipping_status,
            NEW.shipping_status,
            NEW.vessel_name,
            NEW.departure_harbour,
            NEW.shipment_date,
            NEW.arrival_date,
            NEW.clearing_date,
            current_user,
            CURRENT_TIMESTAMP
        );
    END IF;

    -- Also log initial status when first created
    IF (TG_OP = 'INSERT') THEN
        INSERT INTO cars.vehicle_shipping_history (
            vehicle_id,
            old_status,
            new_status,
            vessel_name,
            departure_harbour,
            shipment_date,
            arrival_date,
            clearing_date,
            changed_by,
            changed_at
        ) VALUES (
            NEW.vehicle_id,
            NULL,
            NEW.shipping_status,
            NEW.vessel_name,
            NEW.departure_harbour,
            NEW.shipment_date,
            NEW.arrival_date,
            NEW.clearing_date,
            current_user,
            CURRENT_TIMESTAMP
        );
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger on vehicle_shipping table
DROP TRIGGER IF EXISTS shipping_status_change_trigger ON cars.vehicle_shipping;

CREATE TRIGGER shipping_status_change_trigger
    AFTER INSERT OR UPDATE ON cars.vehicle_shipping
    FOR EACH ROW
    EXECUTE FUNCTION cars.log_shipping_status_change();

-- =====================================================
-- Create view for easy history querying
-- =====================================================

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

-- =====================================================
-- Sample queries for verification
-- =====================================================

-- Get history for a specific vehicle
-- SELECT * FROM cars.vehicle_shipping_history WHERE vehicle_id = 1 ORDER BY changed_at DESC;

-- Get all vehicles that changed status today
-- SELECT * FROM cars.vehicle_shipping_history WHERE DATE(changed_at) = CURRENT_DATE;

-- Get vehicles currently in a specific status
-- SELECT DISTINCT vehicle_id, new_status
-- FROM cars.vehicle_shipping_history
-- WHERE (vehicle_id, changed_at) IN (
--     SELECT vehicle_id, MAX(changed_at)
--     FROM cars.vehicle_shipping_history
--     GROUP BY vehicle_id
-- );

-- Get average time spent in each status
-- SELECT
--     new_status,
--     AVG(hours_in_previous_status) as avg_hours
-- FROM cars.vehicle_shipping_history_view
-- WHERE hours_in_previous_status IS NOT NULL
-- GROUP BY new_status;

-- =====================================================
-- Verification
-- =====================================================

SELECT
    table_name,
    column_name,
    data_type
FROM information_schema.columns
WHERE table_schema = 'cars'
  AND table_name = 'vehicle_shipping_history'
ORDER BY ordinal_position;
