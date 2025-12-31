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
        INSERT INTO cars.vehicle_makes (make_name, country_origin, logo_url, is_active)
        VALUES ($1, $2, $3, $4)
        RETURNING id, make_name, country_origin, logo_url, is_active, created_at
    `

	var make entity.VehicleMake
	err := exec.QueryRowContext(ctx, query, req.MakeName, req.CountryOrigin, req.LogoURL, req.IsActive).Scan(
		&make.ID, &make.MakeName, &make.CountryOrigin, &make.LogoURL, &make.IsActive, &make.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &make, nil
}

func (r *VehicleMakeRepository) GetAllVehicleMakes(ctx context.Context, exec database.Executor, activeOnly bool) ([]entity.VehicleMake, error) {
	query := `SELECT id, make_name, country_origin, logo_url, is_active, created_at FROM cars.vehicle_makes`
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
		err := rows.Scan(&make.ID, &make.MakeName, &make.CountryOrigin, &make.LogoURL, &make.IsActive, &make.CreatedAt)
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

// GetVehicleMakeByID retrieves a vehicle make by its ID
func (r *VehicleMakeRepository) GetVehicleMakeByID(ctx context.Context, exec database.Executor, id int) (*entity.VehicleMake, error) {
	query := `SELECT id, make_name, country_origin, logo_url, is_active, created_at FROM cars.vehicle_makes WHERE id = $1`

	var make entity.VehicleMake
	err := exec.QueryRowContext(ctx, query, id).Scan(
		&make.ID, &make.MakeName, &make.CountryOrigin, &make.LogoURL, &make.IsActive, &make.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &make, nil
}

// UpdateMakeLogo updates only the logo_url field for a vehicle make
func (r *VehicleMakeRepository) UpdateMakeLogo(ctx context.Context, exec database.Executor, id int, logoURL string) error {
	query := `UPDATE cars.vehicle_makes SET logo_url = $2 WHERE id = $1`
	_, err := exec.ExecContext(ctx, query, id, logoURL)
	return err
}
