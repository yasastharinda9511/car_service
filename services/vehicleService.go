package services

import (
	"car_service/dto/request"
	"car_service/dto/response"
	"car_service/entity"

	//"car_service/database"
	//"car_service/dto/request"
	//"car_service/entity"
	"car_service/filters"
	"car_service/repository"
	//"strings"
	//"time"
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type VehicleService struct {
	db                               *sql.DB
	vehicleRepository                *repository.VehicleRepository
	vehicleIMageRepository           *repository.VehicleImageRepository
	vehicleFinancialsRepository      *repository.VehicleFinancialsRepository
	vehicleShippingRepository        *repository.VehicleShippingRepository
	vehiclePurchaseRepository        *repository.VehiclePurchaseRepository
	vehicleSalesRepository           *repository.VehicleSalesRepository
	vehicleShippingHistoryRepository *repository.VehicleShippingHistoryRepository
}

func NewVehicleService(db *sql.DB) *VehicleService {
	return &VehicleService{db: db,
		vehicleRepository:                repository.NewVehicleRepository(),
		vehicleIMageRepository:           repository.NewVehicleImageRepository(),
		vehicleFinancialsRepository:      repository.NewVehicleFinancialsRepository(),
		vehiclePurchaseRepository:        repository.NewVehiclePurchaseRepository(),
		vehicleSalesRepository:           repository.NewVehicleSalesRepository(),
		vehicleShippingRepository:        repository.NewVehicleShippingRepository(),
		vehicleShippingHistoryRepository: repository.NewVehicleShippingHistoryRepository(),
	}
}

func (s *VehicleService) GetAllVehicles(ctx context.Context, limit int, offset int, filter filters.Filter) (*response.VehiclesResponse, error) {

	vehicles, err := s.vehicleRepository.GetAllVehicles(ctx, s.db, limit, offset, filter)
	if err != nil {
		return nil, err
	}

	vehicleCount, err := s.vehicleRepository.GetAllVehicleCount(ctx, s.db, filter)
	if err != nil {
		return nil, err
	}
	var vehiclesResponse response.VehiclesResponse
	vehiclesResponse.Vehicles = vehicles
	vehiclesResponse.Meta.Limit = limit
	vehiclesResponse.Meta.Total = int(vehicleCount)
	return &vehiclesResponse, nil
}

func (s *VehicleService) GetAllVehicleCount(ctx context.Context, filter filters.Filter) (int64, error) {

	count, err := s.vehicleRepository.GetAllVehicleCount(ctx, s.db, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *VehicleService) GetVehicleByID(ctx context.Context, id int64) (*entity.VehicleComplete, error) {
	vehicle, err := s.vehicleRepository.GetVehicleByID(ctx, s.db, id)

	if err != nil {
		return nil, err
	}

	shipping, err := s.vehicleShippingRepository.GetByVehicleID(ctx, s.db, vehicle.ID)
	if err != nil {
		return nil, err
	}

	purchase, err := s.vehiclePurchaseRepository.GetByVehicleID(ctx, s.db, vehicle.ID)
	if err != nil {
		return nil, err
	}

	finance, err := s.vehicleFinancialsRepository.GetByVehicleID(ctx, s.db, vehicle.ID)
	if err != nil {
		return nil, err
	}

	sales, err := s.vehicleSalesRepository.GetByVehicleID(ctx, s.db, vehicle.ID)
	if err != nil {
		return nil, err
	}

	images, err := s.vehicleIMageRepository.GetByVehicleID(ctx, s.db, vehicle.ID)
	var vehicleComplete entity.VehicleComplete

	vehicleComplete.Vehicle = *vehicle
	vehicleComplete.VehicleShipping = *shipping
	vehicleComplete.VehiclePurchase = *purchase
	vehicleComplete.VehicleFinancials = *finance
	vehicleComplete.VehicleSales = *sales
	vehicleComplete.VehicleImages = images

	return &vehicleComplete, nil
}

func (s *VehicleService) InsertVehicleImage(ctx context.Context, vehicleImage []entity.VehicleImage) ([]entity.VehicleImage, error) {

	var vehicleImages []entity.VehicleImage
	for _, image := range vehicleImage {
		vehicleImage, err := s.vehicleIMageRepository.InsertVehicleImage(ctx, s.db, &image)
		vehicleImages = append(vehicleImages, *vehicleImage)
		if err != nil {
			return nil, err
		}
	}

	return vehicleImages, nil
}

func (s *VehicleService) CreateVehicle(ctx context.Context, req request.CreateVehicleRequest) (*entity.Vehicle, error) {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // Will be ignored if tx is committed

	vehicleID, err := s.vehicleRepository.Insert(ctx, tx, req)

	if err != nil {
		return nil, err
	}

	err = s.vehicleShippingRepository.InsertDefault(ctx, tx, vehicleID)
	if err != nil {
		return nil, err
	}

	// Insert default financial record with zero costs
	err = s.vehicleFinancialsRepository.InsertDefault(ctx, tx, vehicleID)
	if err != nil {
		return nil, err
	}

	err = s.vehicleSalesRepository.InsertDefault(ctx, tx, vehicleID)
	if err != nil {
		return nil, err
	}

	// Insert default purchase record (can be updated later)
	err = s.vehiclePurchaseRepository.InsertDefault(ctx, tx, vehicleID)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	vehicle, err := s.vehicleRepository.GetVehicleByID(ctx, s.db, vehicleID)
	if err != nil {
		return nil, err
	}

	return vehicle, nil
}

func (s *VehicleService) UpdateShippingStatus(ctx context.Context, vehicleID int64, detailsRequest request.ShippingDetailsRequest) error {
	err := s.vehicleShippingRepository.UpdateShippingStatus(ctx, s.db, vehicleID, detailsRequest)
	return err

}

func (s *VehicleService) UpdatePurchaseDetails(ctx context.Context, id int64, purchaseRequest *request.PurchaseRequest) error {

	err := s.vehiclePurchaseRepository.UpdateVehiclePurchase(ctx, s.db, id, purchaseRequest)
	return err
}

func (s *VehicleService) UpdateFinancialDetails(ctx context.Context, vehicleID int64, detailsRequest *request.FinancialDetailsRequest) error {

	err := s.vehicleFinancialsRepository.UpdateFinancialDetails(ctx, s.db, vehicleID, detailsRequest)
	return err
}

func (s *VehicleService) UpdateSalesDetails(ctx context.Context, vehicleID int64, req *request.SalesDetailsRequest) error {

	err := s.vehicleSalesRepository.UpdateSalesDetails(ctx, s.db, vehicleID, req)
	return err

}

func (s *VehicleService) UpdateVehicleDetails(ctx context.Context, vehicleID int64, req *request.UpdateVehicleRequest) error {

	err := s.vehicleRepository.UpdateVehicleDetails(ctx, s.db, vehicleID, req)
	return err
}

func (s *VehicleService) GetDropdownOptions(ctx context.Context) (*repository.DropdownOptions, error) {
	return s.vehicleRepository.GetDropdownOptions(ctx, s.db)
}

// AssignCustomerToVehicle assigns a customer to a vehicle sale
func (s *VehicleService) AssignCustomerToVehicle(ctx context.Context, vehicleID int64, customerID int64) error {
	return s.vehicleSalesRepository.AssignCustomer(ctx, s.db, vehicleID, customerID)
}

// RemoveCustomerFromVehicle removes the customer assignment from a vehicle sale
func (s *VehicleService) RemoveCustomerFromVehicle(ctx context.Context, vehicleID int64) error {
	return s.vehicleSalesRepository.RemoveCustomer(ctx, s.db, vehicleID)
}

// GetVehiclesByCustomer retrieves all vehicles associated with a specific customer
func (s *VehicleService) GetVehiclesByCustomer(ctx context.Context, customerID int64) ([]entity.VehicleComplete, error) {
	// Get vehicle IDs for this customer
	vehicleIDs, err := s.vehicleSalesRepository.GetVehiclesByCustomerID(ctx, s.db, customerID)
	if err != nil {
		return nil, err
	}

	if len(vehicleIDs) == 0 {
		return []entity.VehicleComplete{}, nil
	}

	// Fetch complete vehicle information for each ID
	var vehicles []entity.VehicleComplete
	for _, vehicleID := range vehicleIDs {
		vehicle, err := s.GetVehicleByID(ctx, vehicleID)
		if err != nil {
			return nil, err
		}
		vehicles = append(vehicles, *vehicle)
	}

	return vehicles, nil
}

// GetShippingHistory retrieves shipping status change history for a vehicle
func (s *VehicleService) GetShippingHistory(ctx context.Context, vehicleID int64) ([]entity.VehicleShippingHistoryWithDetails, error) {
	return s.vehicleShippingHistoryRepository.GetHistoryByVehicleID(ctx, s.db, vehicleID)
}

// GetRecentShippingHistory retrieves recent shipping changes across all vehicles
func (s *VehicleService) GetRecentShippingHistory(ctx context.Context, limit int) ([]entity.VehicleShippingHistoryWithDetails, error) {
	return s.vehicleShippingHistoryRepository.GetRecentHistory(ctx, s.db, limit)
}

//
//func (s *VehicleService) CreateVehicleMake(vehicleMake request.CreateVehicleMake) (*entity.VehicleMake, error) {
//	query := `
//        INSERT INTO vehicle_makes (make_name, country_origin, is_active)
//        VALUES ($1, $2, $3)
//        RETURNING id, make_name, country_origin, is_active, created_at
//    `
//
//	var make entity.VehicleMake
//	err := s.db.Db.QueryRow(query, vehicleMake.MakeName, vehicleMake.CountryOrigin, vehicleMake.IsActive).Scan(
//		&make.ID, &make.MakeName, &make.CountryOrigin, &make.IsActive, &make.CreatedAt,
//	)
//	if err != nil {
//		return nil, err
//	}
//
//	return &make, nil
//}
//
//func (s *VehicleService) GetAllVehicleMakes(activeOnly bool) ([]entity.VehicleMake, error) {
//	query := `SELECT id, make_name, country_origin, is_active, created_at FROM vehicle_makes`
//	if activeOnly {
//		query += ` WHERE is_active = true`
//	}
//	query += ` ORDER BY make_name`
//
//	rows, err := s.db.Db.Query(query)
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//
//	var makes []entity.VehicleMake
//	for rows.Next() {
//		var make entity.VehicleMake
//		err := rows.Scan(&make.ID, &make.MakeName, &make.CountryOrigin, &make.IsActive, &make.CreatedAt)
//		if err != nil {
//			return nil, err
//		}
//		makes = append(makes, make)
//	}
//
//	return makes, nil
//}
//
//func (s *VehicleService) UpdateVehicleMake(id int, makeName, countryOrigin *string, isActive *bool) error {
//	query := `
//        UPDATE vehicle_makes
//        SET make_name = COALESCE($2, make_name),
//            country_origin = COALESCE($3, country_origin),
//            is_active = COALESCE($4, is_active)
//        WHERE id = $1
//    `
//
//	_, err := s.db.Db.Exec(query, id, makeName, countryOrigin, isActive)
//	return err
//}
//
//// Vehicle Model Service Methods
//func (s *VehicleService) CreateVehicleModel(req request.CreateVehicleModel) (*entity.VehicleModel, error) {
//
//	query := `
//        INSERT INTO vehicle_models (make_id, model_name, body_type, fuel_type, transmission_type, engine_size_cc, is_active)
//        VALUES ($1, $2, $3, $4, $5, $6, $7)
//        RETURNING id, make_id, model_name, body_type, fuel_type, transmission_type, engine_size_cc, is_active, created_at
//    `
//
//	var model entity.VehicleModel
//	err := s.db.Db.QueryRow(query, req.MakeID, req.ModelName, req.BodyType, req.FuelType, req.TransmissionType, req.EngineSizeCC, req.IsActive).Scan(
//		&model.ID, &model.MakeID, &model.ModelName, &model.BodyType, &model.FuelType,
//		&model.TransmissionType, &model.EngineSizeCC, &model.IsActive, &model.CreatedAt,
//	)
//	if err != nil {
//		return nil, err
//	}
//
//	return &model, nil
//}
//
//func (s *VehicleService) GetVehicleModels(makeID *int, activeOnly bool) ([]entity.VehicleModelWithMake, error) {
//	query := `
//        SELECT vm.id, vm.make_id, vm.model_name, vm.body_type, vm.fuel_type,
//               vm.transmission_type, vm.engine_size_cc, vm.is_active, vm.created_at,
//               vma.make_name
//        FROM vehicle_models vm
//        JOIN vehicle_makes vma ON vm.make_id = vma.id
//    `
//
//	var args []interface{}
//	var conditions []string
//
//	if makeID != nil {
//		conditions = append(conditions, "vm.make_id = $1")
//		args = append(args, *makeID)
//	}
//
//	if activeOnly {
//		if len(args) > 0 {
//			conditions = append(conditions, "vm.is_active = $2")
//		} else {
//			conditions = append(conditions, "vm.is_active = $1")
//		}
//		args = append(args, true)
//	}
//
//	if len(conditions) > 0 {
//		query += " WHERE " + strings.Join(conditions, " AND ")
//	}
//
//	query += " ORDER BY vma.make_name, vm.model_name"
//
//	rows, err := s.db.Db.Query(query, args...)
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//
//	var models []entity.VehicleModelWithMake
//	for rows.Next() {
//		var model entity.VehicleModelWithMake
//		err := rows.Scan(&model.ID, &model.MakeID, &model.ModelName, &model.BodyType,
//			&model.FuelType, &model.TransmissionType, &model.EngineSizeCC, &model.IsActive,
//			&model.CreatedAt, &model.MakeName)
//		if err != nil {
//			return nil, err
//		}
//		models = append(models, model)
//	}
//
//	return models, nil
//}
//
//func (s *VehicleService) GetVehicleModelByID(id int) (*entity.VehicleModelWithMake, error) {
//	query := `
//        SELECT vm.id, vm.make_id, vm.model_name, vm.body_type, vm.fuel_type,
//               vm.transmission_type, vm.engine_size_cc, vm.is_active, vm.created_at,
//               vma.make_name
//        FROM vehicle_models vm
//        JOIN vehicle_makes vma ON vm.make_id = vma.id
//        WHERE vm.id = $1
//    `
//
//	var model entity.VehicleModelWithMake
//	err := s.db.Db.QueryRow(query, id).Scan(
//		&model.ID, &model.MakeID, &model.ModelName, &model.BodyType,
//		&model.FuelType, &model.TransmissionType, &model.EngineSizeCC,
//		&model.IsActive, &model.CreatedAt, &model.MakeName,
//	)
//	if err != nil {
//		return nil, err
//	}
//
//	return &model, nil
//}
//
//func (s *VehicleService) UpdateVehicleModel(id int, modelName, bodyType, fuelType,
//	transmissionType *string, engineSizeCC *int, isActive *bool) error {
//
//	query := `
//        UPDATE vehicle_models
//        SET model_name = COALESCE($2, model_name),
//            body_type = COALESCE($3, body_type),
//            fuel_type = COALESCE($4, fuel_type),
//            transmission_type = COALESCE($5, transmission_type),
//            engine_size_cc = COALESCE($6, engine_size_cc),
//            is_active = COALESCE($7, is_active)
//        WHERE id = $1
//    `
//
//	_, err := s.db.Db.Exec(query, id, modelName, bodyType, fuelType, transmissionType, engineSizeCC, isActive)
//	return err
//}
