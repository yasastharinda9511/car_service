package repository

import (
	"car_service/dto/request"
	"car_service/entity"
	"context"
	"database/sql"

	"car_service/database"
)

type VehiclePurchaseRepository struct{}

func NewVehiclePurchaseRepository() *VehiclePurchaseRepository {
	return &VehiclePurchaseRepository{}
}

func (r *VehiclePurchaseRepository) InsertDefault(ctx context.Context, exec database.Executor, vehicleID int64) error {
	_, err := exec.ExecContext(ctx, `
        INSERT INTO cars.vehicle_purchases (vehicle_id)
        VALUES ($1)
    `, vehicleID)
	return err
}

func (r *VehiclePurchaseRepository) GetByVehicleID(ctx context.Context, exec database.Executor, vehicleID int64) (*entity.VehiclePurchase, error) {
	query := `
        SELECT
            vp.id, vp.vehicle_id, vp.supplier_id,
            vp.purchase_remarks, vp.lc_bank, vp.lc_number, vp.lc_cost_jpy, vp.purchase_date,
            COALESCE(vp.purchase_status, 'LC_PENDING') as purchase_status,
            vp.created_at, vp.updated_at
        FROM cars.vehicle_purchases vp
        WHERE vp.vehicle_id = $1
    `
	var vp entity.VehiclePurchase

	err := exec.QueryRowContext(ctx, query, vehicleID).Scan(
		&vp.ID, &vp.VehicleID, &vp.SupplierID,
		&vp.PurchaseRemarks, &vp.LCBank, &vp.LCNumber, &vp.LCCostJPY, &vp.PurchaseDate,
		&vp.PurchaseStatus,
		&vp.CreatedAt, &vp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &vp, nil
}

func (r *VehiclePurchaseRepository) UpdateVehiclePurchase(ctx context.Context, exec database.Executor, vehicleID int64, request *request.PurchaseRequest) error {

	query := `
       UPDATE cars.vehicle_purchases
       SET supplier_id = COALESCE($2, supplier_id),
           purchase_remarks = COALESCE($3, purchase_remarks),
           lc_bank = COALESCE($4, lc_bank),
           lc_number = COALESCE($5, lc_number),
           lc_cost_jpy = COALESCE($6, lc_cost_jpy),
           purchase_date = COALESCE($7, purchase_date),
           purchase_status = COALESCE($8, purchase_status),
           updated_at = CURRENT_TIMESTAMP
       WHERE vehicle_id = $1
   `

	_, err := exec.ExecContext(ctx, query, vehicleID, request.SupplierID,
		request.PurchaseRemarks, request.LCBank, request.LCNumber, request.LCCostJPY,
		request.PurchaseDate, request.PurchaseStatus)
	return err
}
