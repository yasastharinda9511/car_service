package repository

import (
	"car_service/database"
	"car_service/entity"
	"context"
)

type VehicleShippingHistoryRepository struct{}

func NewVehicleShippingHistoryRepository() *VehicleShippingHistoryRepository {
	return &VehicleShippingHistoryRepository{}
}

// GetHistoryByVehicleID retrieves all shipping status changes for a specific vehicle
func (r *VehicleShippingHistoryRepository) GetHistoryByVehicleID(ctx context.Context, exec database.Executor, vehicleID int64) ([]entity.VehicleShippingHistoryWithDetails, error) {
	query := `
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
        WHERE vsh.vehicle_id = $1
        ORDER BY vsh.changed_at DESC
    `

	rows, err := exec.QueryContext(ctx, query, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []entity.VehicleShippingHistoryWithDetails
	for rows.Next() {
		var h entity.VehicleShippingHistoryWithDetails
		err := rows.Scan(
			&h.ID, &h.VehicleID, &h.VehicleCode, &h.Make, &h.Model, &h.ChassisID,
			&h.OldStatus, &h.NewStatus, &h.VesselName, &h.DepartureHarbour,
			&h.ShipmentDate, &h.ArrivalDate, &h.ClearingDate,
			&h.ChangedBy, &h.ChangeRemarks, &h.ChangedAt, &h.HoursInPreviousStatus,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

// GetRecentHistory retrieves recent shipping status changes across all vehicles
func (r *VehicleShippingHistoryRepository) GetRecentHistory(ctx context.Context, exec database.Executor, limit int) ([]entity.VehicleShippingHistoryWithDetails, error) {
	query := `
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
            NULL as hours_in_previous_status
        FROM cars.vehicle_shipping_history vsh
        JOIN cars.vehicles v ON vsh.vehicle_id = v.id
        ORDER BY vsh.changed_at DESC
        LIMIT $1
    `

	rows, err := exec.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []entity.VehicleShippingHistoryWithDetails
	for rows.Next() {
		var h entity.VehicleShippingHistoryWithDetails
		err := rows.Scan(
			&h.ID, &h.VehicleID, &h.VehicleCode, &h.Make, &h.Model, &h.ChassisID,
			&h.OldStatus, &h.NewStatus, &h.VesselName, &h.DepartureHarbour,
			&h.ShipmentDate, &h.ArrivalDate, &h.ClearingDate,
			&h.ChangedBy, &h.ChangeRemarks, &h.ChangedAt, &h.HoursInPreviousStatus,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

// GetHistoryByStatus retrieves all vehicles that have been in a specific status
func (r *VehicleShippingHistoryRepository) GetHistoryByStatus(ctx context.Context, exec database.Executor, status string) ([]entity.VehicleShippingHistoryWithDetails, error) {
	query := `
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
            NULL as hours_in_previous_status
        FROM cars.vehicle_shipping_history vsh
        JOIN cars.vehicles v ON vsh.vehicle_id = v.id
        WHERE vsh.new_status = $1
        ORDER BY vsh.changed_at DESC
    `

	rows, err := exec.QueryContext(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []entity.VehicleShippingHistoryWithDetails
	for rows.Next() {
		var h entity.VehicleShippingHistoryWithDetails
		err := rows.Scan(
			&h.ID, &h.VehicleID, &h.VehicleCode, &h.Make, &h.Model, &h.ChassisID,
			&h.OldStatus, &h.NewStatus, &h.VesselName, &h.DepartureHarbour,
			&h.ShipmentDate, &h.ArrivalDate, &h.ClearingDate,
			&h.ChangedBy, &h.ChangeRemarks, &h.ChangedAt, &h.HoursInPreviousStatus,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

// GetCurrentStatusForAllVehicles gets the latest status for all vehicles
func (r *VehicleShippingHistoryRepository) GetCurrentStatusForAllVehicles(ctx context.Context, exec database.Executor) (map[int64]string, error) {
	query := `
        SELECT DISTINCT ON (vehicle_id)
            vehicle_id,
            new_status
        FROM cars.vehicle_shipping_history
        ORDER BY vehicle_id, changed_at DESC
    `

	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statusMap := make(map[int64]string)
	for rows.Next() {
		var vehicleID int64
		var status string
		if err := rows.Scan(&vehicleID, &status); err != nil {
			return nil, err
		}
		statusMap[vehicleID] = status
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return statusMap, nil
}

// InsertManualHistoryEntry allows manual insertion of history with remarks
func (r *VehicleShippingHistoryRepository) InsertManualHistoryEntry(ctx context.Context, exec database.Executor, vehicleID int64, oldStatus, newStatus *string, changedBy, remarks string) error {
	query := `
        INSERT INTO cars.vehicle_shipping_history
        (vehicle_id, old_status, new_status, changed_by, change_remarks, changed_at)
        VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
    `

	_, err := exec.ExecContext(ctx, query, vehicleID, oldStatus, newStatus, changedBy, remarks)
	return err
}
