package repository

import (
	"car_service/database"
	"car_service/dto/request"
	"car_service/dto/response"
	"car_service/entity"
	"car_service/filters"
	"context"
	"database/sql"
)

type VehicleFinancialsRepository struct{}

func NewVehicleFinancialsRepository() *VehicleFinancialsRepository {
	return &VehicleFinancialsRepository{}
}

func (r *VehicleFinancialsRepository) InsertDefault(ctx context.Context, exec database.Executor, vehicleID int64) error {
	_, err := exec.ExecContext(ctx, `
        INSERT INTO cars.vehicle_financials (vehicle_id, charges_lkr, tt_lkr, duty_lkr,
        clearing_lkr, other_expenses_lkr, total_cost_lkr)
        VALUES ($1, 0, 0, 0, 0, '{}'::jsonb, 0)
    `, vehicleID)
	return err
}

func (r *VehicleFinancialsRepository) GetByVehicleID(ctx context.Context, exec database.Executor, vehicleID int64) (*entity.VehicleFinancials, error) {
	query := `
        SELECT id, vehicle_id, total_cost_lkr, charges_lkr,
        duty_lkr, clearing_lkr, other_expenses_lkr
        FROM cars.vehicle_financials
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
       UPDATE cars.vehicle_financials
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
func (r *VehicleFinancialsRepository) GetDetailedFinancialSummary(ctx context.Context, exec database.Executor, filter filters.Filter) (*response.DetailedFinancialSummary, error) {
	query := `SELECT
        COALESCE(SUM(charges_lkr), 0) as total_charges,
        COALESCE(SUM(tt_lkr), 0) as total_tt,
        COALESCE(SUM(duty_lkr), 0) as total_duty,
        COALESCE(SUM(clearing_lkr), 0) as total_clearing,
        COALESCE(SUM((
            SELECT SUM((value#>>'{}')::numeric)
            FROM jsonb_each(COALESCE(other_expenses_lkr, '{}'::jsonb))
        )), 0::numeric) as total_other_expenses,
        COALESCE(SUM(total_cost_lkr), 0) as total_investment
    FROM cars.vehicle_financials vf`

	query, args := filter.GetQuery(query, "", "", -1, -1)

	var summary response.DetailedFinancialSummary
	err := exec.QueryRowContext(ctx, query, args...).Scan(
		&summary.TotalCharges,
		&summary.TotalTT,
		&summary.TotalDuty,
		&summary.TotalClearing,
		&summary.TotalOtherExpenses,
		&summary.TotalInvestment,
	)

	if err != nil {
		return nil, err
	}
	return &summary, nil

}
