package repository

import (
	"car_service/dto/request"
	"car_service/entity"
	"car_service/filters"
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
        INSERT INTO cars.vehicle_sales (vehicle_id, sale_status)
        VALUES ($1, 'AVAILABLE')
    `, vehicleID)
	return err
}

func (r *VehicleSalesRepository) GetByVehicleID(ctx context.Context, exec database.Executor, vehicleID int64) (*entity.VehicleSales, error) {
	query := `
        SELECT id, vehicle_id, customer_id, sold_date, revenue, profit,
        sale_remarks, sale_status
        FROM cars.vehicle_sales
        WHERE vehicle_id = $1
    `
	var vs entity.VehicleSales
	err := exec.QueryRowContext(ctx, query, vehicleID).Scan(
		&vs.ID, &vs.VehicleID, &vs.CustomerID, &vs.SoldDate, &vs.Revenue, &vs.Profit,
		&vs.SaleRemarks, &vs.SaleStatus,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &vs, err
}

func (r *VehicleSalesRepository) UpdateSalesDetails(ctx context.Context, exec database.Executor, vehicleID int64, req *request.SalesDetailsRequest) error {
	query := `
       UPDATE cars.vehicle_sales
       SET customer_id = $2,
           sold_date = $3,
           revenue = $4,
           profit = $5,
           sale_remarks = $6,
           sale_status = $7,
           updated_at = CURRENT_TIMESTAMP
       WHERE vehicle_id = $1
   `
	_, err := exec.ExecContext(ctx, query, vehicleID, req.CustomerID, req.SoldDate, req.Revenue, req.Profit,
		req.SaleRemarks, req.SaleStatus)
	return err
}

func (r *VehicleSalesRepository) GetSalesStustVehicleCount(ctx context.Context, exec database.Executor, filter filters.Filter) (map[string]int, error) {
	query := `SELECT
    sale_status,
    COUNT(*) as vehicle_count
	FROM cars.vehicle_sales vsl`

	query, args := filter.GetQuery(query, "vsl.sale_status", "", -1, -1)
	rows, err := exec.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]int)

	for rows.Next() {
		var status string
		var count int

		err := rows.Scan(&status, &count)
		if err != nil {
			return nil, err
		}

		results[status] = count
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// AssignCustomer assigns a customer to a vehicle sale
func (r *VehicleSalesRepository) AssignCustomer(ctx context.Context, exec database.Executor, vehicleID int64, customerID int64) error {
	query := `
        UPDATE cars.vehicle_sales
        SET customer_id = $2,
            updated_at = CURRENT_TIMESTAMP
        WHERE vehicle_id = $1
    `
	_, err := exec.ExecContext(ctx, query, vehicleID, customerID)
	return err
}

// RemoveCustomer removes the customer assignment from a vehicle sale
func (r *VehicleSalesRepository) RemoveCustomer(ctx context.Context, exec database.Executor, vehicleID int64) error {
	query := `
        UPDATE cars.vehicle_sales
        SET customer_id = NULL,
            updated_at = CURRENT_TIMESTAMP
        WHERE vehicle_id = $1
    `
	_, err := exec.ExecContext(ctx, query, vehicleID)
	return err
}

// GetVehiclesByCustomerID retrieves all vehicles associated with a specific customer
func (r *VehicleSalesRepository) GetVehiclesByCustomerID(ctx context.Context, exec database.Executor, customerID int64) ([]int64, error) {
	query := `
        SELECT vehicle_id
        FROM cars.vehicle_sales
        WHERE customer_id = $1
        ORDER BY created_at DESC
    `
	rows, err := exec.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicleIDs []int64
	for rows.Next() {
		var vehicleID int64
		if err := rows.Scan(&vehicleID); err != nil {
			return nil, err
		}
		vehicleIDs = append(vehicleIDs, vehicleID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return vehicleIDs, nil
}
