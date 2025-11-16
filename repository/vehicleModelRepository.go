package repository

import (
	"car_service/database"
	"car_service/dto/request"
	"car_service/entity"
	"context"
	"strings"
)

type VehicleModelRepository struct{}

func NewVehicleModelRepository() *VehicleModelRepository {
	return &VehicleModelRepository{}
}

func (r *VehicleModelRepository) CreateVehicleModel(ctx context.Context, exec database.Executor, req request.CreateVehicleModel) (*entity.VehicleModel, error) {
	query := `
        INSERT INTO cars.vehicle_models (make_id, model_name, body_type, fuel_type, transmission_type, engine_size_cc, is_active)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, make_id, model_name, body_type, fuel_type, transmission_type, engine_size_cc, is_active, created_at
    `

	var model entity.VehicleModel
	err := exec.QueryRowContext(ctx, query, req.MakeID, req.ModelName, req.BodyType, req.FuelType, req.TransmissionType, req.EngineSizeCC, req.IsActive).Scan(
		&model.ID, &model.MakeID, &model.ModelName, &model.BodyType, &model.FuelType,
		&model.TransmissionType, &model.EngineSizeCC, &model.IsActive, &model.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &model, nil
}

func (r *VehicleModelRepository) GetVehicleModels(ctx context.Context, exec database.Executor, makeID *int, activeOnly bool) ([]entity.VehicleModelWithMake, error) {
	query := `
        SELECT vm.id, vm.make_id, vm.model_name, vm.body_type, vm.fuel_type,
               vm.transmission_type, vm.engine_size_cc, vm.is_active, vm.created_at,
               vma.make_name
        FROM cars.vehicle_models vm
        JOIN cars.vehicle_makes vma ON vm.make_id = vma.id
    `

	var args []interface{}
	var conditions []string

	if makeID != nil {
		conditions = append(conditions, "vm.make_id = $1")
		args = append(args, *makeID)
	}

	if activeOnly {
		if len(args) > 0 {
			conditions = append(conditions, "vm.is_active = $2")
		} else {
			conditions = append(conditions, "vm.is_active = $1")
		}
		args = append(args, true)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY vma.make_name, vm.model_name"

	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []entity.VehicleModelWithMake
	for rows.Next() {
		var model entity.VehicleModelWithMake
		err := rows.Scan(&model.ID, &model.MakeID, &model.ModelName, &model.BodyType,
			&model.FuelType, &model.TransmissionType, &model.EngineSizeCC, &model.IsActive,
			&model.CreatedAt, &model.MakeName)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return models, nil
}

func (r *VehicleModelRepository) GetVehicleModelByID(ctx context.Context, exec database.Executor, id int) (*entity.VehicleModelWithMake, error) {
	query := `
        SELECT vm.id, vm.make_id, vm.model_name, vm.body_type, vm.fuel_type,
               vm.transmission_type, vm.engine_size_cc, vm.is_active, vm.created_at,
               vma.make_name
        FROM cars.vehicle_models vm
        JOIN cars.vehicle_makes vma ON vm.make_id = vma.id
        WHERE vm.id = $1
    `

	var model entity.VehicleModelWithMake
	err := exec.QueryRowContext(ctx, query, id).Scan(
		&model.ID, &model.MakeID, &model.ModelName, &model.BodyType,
		&model.FuelType, &model.TransmissionType, &model.EngineSizeCC,
		&model.IsActive, &model.CreatedAt, &model.MakeName,
	)
	if err != nil {
		return nil, err
	}

	return &model, nil
}

func (r *VehicleModelRepository) UpdateVehicleModel(ctx context.Context, exec database.Executor, id int, modelName, bodyType, fuelType, transmissionType *string, engineSizeCC *int, isActive *bool) error {
	query := `
        UPDATE cars.vehicle_models
        SET model_name = COALESCE($2, model_name),
            body_type = COALESCE($3, body_type),
            fuel_type = COALESCE($4, fuel_type),
            transmission_type = COALESCE($5, transmission_type),
            engine_size_cc = COALESCE($6, engine_size_cc),
            is_active = COALESCE($7, is_active)
        WHERE id = $1
    `

	_, err := exec.ExecContext(ctx, query, id, modelName, bodyType, fuelType, transmissionType, engineSizeCC, isActive)
	return err
}
