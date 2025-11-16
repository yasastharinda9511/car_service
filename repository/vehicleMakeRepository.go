package repository

import (
	"car_service/database"
	"car_service/dto/request"
	"car_service/entity"
	"context"
)

type VehicleMakeRepository struct{}

func NewVehicleMakeRepository() *VehicleMakeRepository {
	return &VehicleMakeRepository{}
}

func (r *VehicleMakeRepository) CreateVehicleMake(ctx context.Context, exec database.Executor, req request.CreateVehicleMake) (*entity.VehicleMake, error) {
	query := `
        INSERT INTO cars.vehicle_makes (make_name, country_origin, is_active)
        VALUES ($1, $2, $3)
        RETURNING id, make_name, country_origin, is_active, created_at
    `

	var make entity.VehicleMake
	err := exec.QueryRowContext(ctx, query, req.MakeName, req.CountryOrigin, req.IsActive).Scan(
		&make.ID, &make.MakeName, &make.CountryOrigin, &make.IsActive, &make.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &make, nil
}

func (r *VehicleMakeRepository) GetAllVehicleMakes(ctx context.Context, exec database.Executor, activeOnly bool) ([]entity.VehicleMake, error) {
	query := `SELECT id, make_name, country_origin, is_active, created_at FROM cars.vehicle_makes`
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY make_name`

	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var makes []entity.VehicleMake
	for rows.Next() {
		var make entity.VehicleMake
		err := rows.Scan(&make.ID, &make.MakeName, &make.CountryOrigin, &make.IsActive, &make.CreatedAt)
		if err != nil {
			return nil, err
		}
		makes = append(makes, make)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return makes, nil
}

func (r *VehicleMakeRepository) UpdateVehicleMake(ctx context.Context, exec database.Executor, id int, makeName, countryOrigin *string, isActive *bool) error {
	query := `
        UPDATE cars.vehicle_makes
        SET make_name = COALESCE($2, make_name),
            country_origin = COALESCE($3, country_origin),
            is_active = COALESCE($4, is_active)
        WHERE id = $1
    `

	_, err := exec.ExecContext(ctx, query, id, makeName, countryOrigin, isActive)
	return err
}
