package repository

import (
	"car_service/entity"
	"context"
	"database/sql"

	"car_service/database"
	"car_service/dto/request"
	"car_service/filters"
)

type VehicleShippingRepository struct{}

func NewVehicleShippingRepository() *VehicleShippingRepository {
	return &VehicleShippingRepository{}
}

func (r *VehicleShippingRepository) GetByVehicleID(ctx context.Context, exec database.Executor, vehicleID int64) (*entity.VehicleShipping, error) {
	query := `
        SELECT id, vehicle_id, vessel_name, departure_harbour,
               shipment_date, arrival_date, clearing_date, shipping_status
        FROM vehicle_shipping
        WHERE vehicle_id = $1
    `
	var vs entity.VehicleShipping
	err := exec.QueryRowContext(ctx, query, vehicleID).Scan(
		&vs.ID, &vs.VehicleID, &vs.VesselName, &vs.DepartureHarbour,
		&vs.ShipmentDate, &vs.ArrivalDate, &vs.ClearingDate, &vs.ShippingStatus,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &vs, err
}

func (r *VehicleShippingRepository) InsertDefault(ctx context.Context, exec database.Executor, vehicleID int64) error {
	_, err := exec.ExecContext(ctx, `
        INSERT INTO vehicle_shipping (vehicle_id, shipping_status) 
        VALUES ($1, 'PROCESSING')
    `, vehicleID)
	return err
}

func (r *VehicleShippingRepository) UpdateShippingStatus(ctx context.Context, exec database.Executor, vehicleID int64, detailsRequest request.ShippingDetailsRequest) error {
	query := `
       UPDATE vehicle_shipping
       SET vessel_name = $2,
           departure_harbour = $3,
           shipment_date = $4,
           arrival_date = $5,
           clearing_date = $6,
           shipping_status = $7,
           updated_at = CURRENT_TIMESTAMP
       WHERE vehicle_id = $1
   `

	_, err := exec.ExecContext(ctx, query, vehicleID, detailsRequest.VesselName, detailsRequest.DepartureHarbour,
		detailsRequest.ShipmentDate, detailsRequest.ArrivalDate, detailsRequest.ClearingDate, detailsRequest.ShippingStatus)
	return err
}

func (r *VehicleShippingRepository) GetShippingStustVehicleCount(ctx context.Context, exec database.Executor, filter filters.Filter) (map[string]int, error) {
	query := `SELECT 
    shipping_status,
    COUNT(*) as vehicle_count
	FROM vehicle_shipping vs`

	query, args := filter.GetQuery(query, "vs.shipping_status", "", -1, -1)
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
