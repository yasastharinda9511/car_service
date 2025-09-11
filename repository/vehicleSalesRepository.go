package repository

import (
	"car_service/dto/request"
	"car_service/entity"
	"context"
	"database/sql"

	"car_service/database"
)

type VehicleSalesRepository struct{}

func NewVehicleSalesRepository() *VehicleSalesRepository {
	return &VehicleSalesRepository{}
}

func (r *VehicleSalesRepository) InsertDefault(ctx context.Context, exec database.Executor, vehicleID int64) error {
	_, err := exec.ExecContext(ctx, `
        INSERT INTO vehicle_sales (vehicle_id, sale_status) 
        VALUES ($1, 'AVAILABLE')
    `, vehicleID)
	return err
}

func (r *VehicleSalesRepository) GetByVehicleID(ctx context.Context, exec database.Executor, vehicleID int64) (*entity.VehicleSales, error) {
	query := `
        SELECT id, vehicle_id, sold_date, revenue, profit,
        sold_to_name, sold_to_title, contact_number, customer_address, sale_status
        FROM vehicle_sales
        WHERE vehicle_id = $1
    `
	var vs entity.VehicleSales
	err := exec.QueryRowContext(ctx, query, vehicleID).Scan(
		&vs.ID, &vs.VehicleID, &vs.SoldDate, &vs.Revenue, &vs.Profit,
		&vs.SoldToName, &vs.SoldToTitle, &vs.ContactNumber,
		&vs.CustomerAddress, &vs.SaleStatus,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &vs, err
}

func (r *VehicleSalesRepository) UpdateSalesDetails(ctx context.Context, exec database.Executor, vehicleID int64, req *request.SalesDetailsRequest) error {
	query := `
       UPDATE vehicle_sales
       SET sold_date = $2,
           revenue = $3,
           profit = $4,
           sold_to_name = $5,
           sold_to_title = $6,
           contact_number = $7,
           customer_address = $8,
           other_contacts = $9,
           sale_remarks = $10,
           sale_status = $11,
           updated_at = CURRENT_TIMESTAMP
       WHERE vehicle_id = $1
   `
	_, err := exec.ExecContext(ctx, query, vehicleID, req.SoldDate, req.Revenue, req.Profit, req.SoldToName,
		req.SoldToTitle, req.ContactNumber, req.CustomerAddress, req.OtherContacts, req.SaleRemarks, req.SaleStatus)
	return err
}
