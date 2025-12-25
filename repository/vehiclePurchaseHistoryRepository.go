package repository

import (
	"car_service/database"
	"car_service/entity"
	"context"
	"database/sql"
)

type VehiclePurchaseHistoryRepository struct{}

func NewVehiclePurchaseHistoryRepository() *VehiclePurchaseHistoryRepository {
	return &VehiclePurchaseHistoryRepository{}
}

// GetHistoryByVehicleID retrieves all purchase changes for a specific vehicle
func (r *VehiclePurchaseHistoryRepository) GetHistoryByVehicleID(ctx context.Context, exec database.Executor, vehicleID int64) ([]entity.VehiclePurchaseHistoryWithDetails, error) {
	query := `
		SELECT
			vph.id,
			vph.vehicle_id,
			v.code as vehicle_code,
			v.make,
			v.model,
			v.chassis_id,
			vph.old_status,
			vph.new_status,
			vph.supplier_id,
			vph.lc_bank,
			vph.lc_number,
			vph.lc_cost_jpy,
			vph.purchase_date,
			vph.purchase_remarks,
			vph.changed_by,
			vph.change_remarks,
			vph.changed_at,
			EXTRACT(EPOCH FROM (vph.changed_at - LAG(vph.changed_at) OVER (PARTITION BY vph.vehicle_id ORDER BY vph.changed_at))) / 3600 as hours_in_previous_status
		FROM cars.vehicle_purchase_history vph
		JOIN cars.vehicles v ON vph.vehicle_id = v.id
		WHERE vph.vehicle_id = $1
		ORDER BY vph.changed_at DESC
	`

	rows, err := exec.QueryContext(ctx, query, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []entity.VehiclePurchaseHistoryWithDetails
	for rows.Next() {
		var h entity.VehiclePurchaseHistoryWithDetails
		var supplierID sql.NullInt64

		err := rows.Scan(
			&h.ID, &h.VehicleID, &h.VehicleCode, &h.Make, &h.Model, &h.ChassisID,
			&h.OldStatus, &h.NewStatus, &supplierID,
			&h.LCBank, &h.LCNumber, &h.LCCostJPY, &h.PurchaseDate,
			&h.PurchaseRemarks, &h.ChangedBy, &h.ChangeRemarks, &h.ChangedAt,
			&h.HoursInPreviousStatus,
		)
		if err != nil {
			return nil, err
		}

		// Convert sql.NullInt64 to *int64
		if supplierID.Valid {
			h.SupplierID = &supplierID.Int64
		}

		history = append(history, h)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

// GetRecentHistory retrieves recent purchase changes across all vehicles
func (r *VehiclePurchaseHistoryRepository) GetRecentHistory(ctx context.Context, exec database.Executor, limit int) ([]entity.VehiclePurchaseHistoryWithDetails, error) {
	query := `
		SELECT
			vph.id,
			vph.vehicle_id,
			v.code as vehicle_code,
			v.make,
			v.model,
			v.chassis_id,
			vph.old_status,
			vph.new_status,
			vph.supplier_id,
			vph.lc_bank,
			vph.lc_number,
			vph.lc_cost_jpy,
			vph.purchase_date,
			vph.purchase_remarks,
			vph.changed_by,
			vph.change_remarks,
			vph.changed_at,
			NULL as hours_in_previous_status
		FROM cars.vehicle_purchase_history vph
		JOIN cars.vehicles v ON vph.vehicle_id = v.id
		ORDER BY vph.changed_at DESC
		LIMIT $1
	`

	rows, err := exec.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []entity.VehiclePurchaseHistoryWithDetails
	for rows.Next() {
		var h entity.VehiclePurchaseHistoryWithDetails
		var supplierID sql.NullInt64

		err := rows.Scan(
			&h.ID, &h.VehicleID, &h.VehicleCode, &h.Make, &h.Model, &h.ChassisID,
			&h.OldStatus, &h.NewStatus, &supplierID,
			&h.LCBank, &h.LCNumber, &h.LCCostJPY, &h.PurchaseDate,
			&h.PurchaseRemarks, &h.ChangedBy, &h.ChangeRemarks, &h.ChangedAt,
			&h.HoursInPreviousStatus,
		)
		if err != nil {
			return nil, err
		}

		// Convert sql.NullInt64 to *int64
		if supplierID.Valid {
			h.SupplierID = &supplierID.Int64
		}

		history = append(history, h)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

// GetHistoryByStatus retrieves all vehicles that have been in a specific purchase status
func (r *VehiclePurchaseHistoryRepository) GetHistoryByStatus(ctx context.Context, exec database.Executor, status string) ([]entity.VehiclePurchaseHistoryWithDetails, error) {
	query := `
		SELECT
			vph.id,
			vph.vehicle_id,
			v.code as vehicle_code,
			v.make,
			v.model,
			v.chassis_id,
			vph.old_status,
			vph.new_status,
			vph.supplier_id,
			vph.lc_bank,
			vph.lc_number,
			vph.lc_cost_jpy,
			vph.purchase_date,
			vph.purchase_remarks,
			vph.changed_by,
			vph.change_remarks,
			vph.changed_at,
			NULL as hours_in_previous_status
		FROM cars.vehicle_purchase_history vph
		JOIN cars.vehicles v ON vph.vehicle_id = v.id
		WHERE vph.new_status = $1
		ORDER BY vph.changed_at DESC
	`

	rows, err := exec.QueryContext(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []entity.VehiclePurchaseHistoryWithDetails
	for rows.Next() {
		var h entity.VehiclePurchaseHistoryWithDetails
		var supplierID sql.NullInt64

		err := rows.Scan(
			&h.ID, &h.VehicleID, &h.VehicleCode, &h.Make, &h.Model, &h.ChassisID,
			&h.OldStatus, &h.NewStatus, &supplierID,
			&h.LCBank, &h.LCNumber, &h.LCCostJPY, &h.PurchaseDate,
			&h.PurchaseRemarks, &h.ChangedBy, &h.ChangeRemarks, &h.ChangedAt,
			&h.HoursInPreviousStatus,
		)
		if err != nil {
			return nil, err
		}

		// Convert sql.NullInt64 to *int64
		if supplierID.Valid {
			h.SupplierID = &supplierID.Int64
		}

		history = append(history, h)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

// GetHistoryBySupplier retrieves all purchase history entries for a specific supplier
func (r *VehiclePurchaseHistoryRepository) GetHistoryBySupplier(ctx context.Context, exec database.Executor, supplierID int64) ([]entity.VehiclePurchaseHistoryWithDetails, error) {
	query := `
		SELECT
			vph.id,
			vph.vehicle_id,
			v.code as vehicle_code,
			v.make,
			v.model,
			v.chassis_id,
			vph.old_status,
			vph.new_status,
			vph.supplier_id,
			vph.lc_bank,
			vph.lc_number,
			vph.lc_cost_jpy,
			vph.purchase_date,
			vph.purchase_remarks,
			vph.changed_by,
			vph.change_remarks,
			vph.changed_at,
			NULL as hours_in_previous_status
		FROM cars.vehicle_purchase_history vph
		JOIN cars.vehicles v ON vph.vehicle_id = v.id
		WHERE vph.supplier_id = $1
		ORDER BY vph.changed_at DESC
	`

	rows, err := exec.QueryContext(ctx, query, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []entity.VehiclePurchaseHistoryWithDetails
	for rows.Next() {
		var h entity.VehiclePurchaseHistoryWithDetails
		var supplierIDNullable sql.NullInt64

		err := rows.Scan(
			&h.ID, &h.VehicleID, &h.VehicleCode, &h.Make, &h.Model, &h.ChassisID,
			&h.OldStatus, &h.NewStatus, &supplierIDNullable,
			&h.LCBank, &h.LCNumber, &h.LCCostJPY, &h.PurchaseDate,
			&h.PurchaseRemarks, &h.ChangedBy, &h.ChangeRemarks, &h.ChangedAt,
			&h.HoursInPreviousStatus,
		)
		if err != nil {
			return nil, err
		}

		// Convert sql.NullInt64 to *int64
		if supplierIDNullable.Valid {
			h.SupplierID = &supplierIDNullable.Int64
		}

		history = append(history, h)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

// GetCurrentStatusForAllVehicles gets the latest purchase status for all vehicles
func (r *VehiclePurchaseHistoryRepository) GetCurrentStatusForAllVehicles(ctx context.Context, exec database.Executor) (map[int64]string, error) {
	query := `
		SELECT DISTINCT ON (vehicle_id)
			vehicle_id,
			new_status
		FROM cars.vehicle_purchase_history
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
func (r *VehiclePurchaseHistoryRepository) InsertManualHistoryEntry(ctx context.Context, exec database.Executor, vehicleID int64, oldStatus, newStatus *string, changedBy, remarks string) error {
	query := `
		INSERT INTO cars.vehicle_purchase_history
		(vehicle_id, old_status, new_status, changed_by, change_remarks, changed_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
	`

	_, err := exec.ExecContext(ctx, query, vehicleID, oldStatus, newStatus, changedBy, remarks)
	return err
}
