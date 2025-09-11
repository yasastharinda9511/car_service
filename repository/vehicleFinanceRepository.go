package repository

import (
	"car_service/database"
	"car_service/dto/request"
	"car_service/entity"
	"context"
	"database/sql"
)

type VehicleFinancialsRepository struct{}

func NewVehicleFinancialsRepository() *VehicleFinancialsRepository {
	return &VehicleFinancialsRepository{}
}

func (r *VehicleFinancialsRepository) InsertDefault(ctx context.Context, exec database.Executor, vehicleID int64) error {
	_, err := exec.ExecContext(ctx, `
        INSERT INTO vehicle_financials (vehicle_id, charges_lkr, tt_lkr, duty_lkr, 
        clearing_lkr, other_expenses_lkr, total_cost_lkr) 
        VALUES ($1, 0, 0, 0, 0, 0, 0)
    `, vehicleID)
	return err
}

func (r *VehicleFinancialsRepository) GetByVehicleID(ctx context.Context, exec database.Executor, vehicleID int64) (*entity.VehicleFinancials, error) {
	query := `
        SELECT id, vehicle_id, total_cost_lkr, charges_lkr,
        duty_lkr, clearing_lkr, other_expenses_lkr
        FROM vehicle_financials
        WHERE vehicle_id = $1
    `
	var vf entity.VehicleFinancials
	err := exec.QueryRowContext(ctx, query, vehicleID).Scan(
		&vf.ID, &vf.VehicleID, &vf.TotalCostLKR, &vf.ChargesLKR,
		&vf.DutyLKR, &vf.ClearingLKR, &vf.OtherExpensesLKR,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &vf, err
}

func (r *VehicleFinancialsRepository) UpdateFinancialDetails(ctx context.Context, exec database.Executor, vehicleID int64, request *request.FinancialDetailsRequest) error {

	query := `
       UPDATE vehicle_financials
       SET charges_lkr = $2,
           tt_lkr = $3,
           duty_lkr = $4,
           clearing_lkr = $5,
           other_expenses_lkr = $6,
           total_cost_lkr = $7,
           updated_at = CURRENT_TIMESTAMP
       WHERE vehicle_id = $1
   `

	_, err := exec.ExecContext(ctx, query, vehicleID, request.ChargesLKR, request.TTLKR, request.DutyLKR,
		request.ClearingLKR, request.OtherExpensesLKR, request.TotalCostLKR)
	return err

}
