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
        INSERT INTO vehicle_purchases (vehicle_id) 
        VALUES ($1)
    `, vehicleID)
	return err
}

func (r *VehiclePurchaseRepository) GetByVehicleID(ctx context.Context, exec database.Executor, vehicleID int64) (*entity.VehiclePurchase, error) {
	query := `
        SELECT id, vehicle_id, bought_from_name, bought_from_title,
               bought_from_contact, bought_from_address, bought_from_other_contacts,
               purchase_remarks, lc_bank, lc_number, lc_cost_jpy, purchase_date
        FROM vehicle_purchases
        WHERE vehicle_id = $1
    `
	var vp entity.VehiclePurchase
	err := exec.QueryRowContext(ctx, query, vehicleID).Scan(
		&vp.ID, &vp.VehicleID, &vp.BoughtFromName, &vp.BoughtFromTitle,
		&vp.BoughtFromContact, &vp.BoughtFromAddress, &vp.BoughtFromOtherContacts,
		&vp.PurchaseRemarks, &vp.LCBank, &vp.LCNumber, &vp.LCCostJPY, &vp.PurchaseDate,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &vp, err
}

func (r *VehiclePurchaseRepository) UpdateVehiclePurchase(ctx context.Context, exec database.Executor, vehicleID int64, request *request.PurchaseRequest) error {

	query := `
       UPDATE vehicle_purchases
       SET bought_from_name = $2,
           bought_from_title = $3,
           bought_from_contact = $4,
           bought_from_address = $5,
           bought_from_other_contacts = $6,
           purchase_remarks = $7,
           lc_bank = $8,
           lc_number = $9,
           lc_cost_jpy = $10,
           purchase_date = $11,
           updated_at = CURRENT_TIMESTAMP
       WHERE vehicle_id = $1
   `

	_, err := exec.ExecContext(ctx, query, vehicleID, request.BoughtFromName, request.BoughtFromTitle,
		request.BoughtFromContact, request.BoughtFromAddress, request.BoughtFromOtherContacts, request.PurchaseRemarks,
		request.LCBank, request.LCNumber, request.LCCostJPY, request.PurchaseDate)
	return err
}
