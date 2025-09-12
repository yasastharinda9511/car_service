package services

import (
	"car_service/dto/response"
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

type AnalyticsService struct {
	db                          *sql.DB
	vehicleRepository           *repository.VehicleRepository
	vehicleIMageRepository      *repository.VehicleImageRepository
	vehicleFinancialsRepository *repository.VehicleFinancialsRepository
	vehicleShippingRepository   *repository.VehicleShippingRepository
	vehiclePurchaseRepository   *repository.VehiclePurchaseRepository
	vehicleSalesRepository      *repository.VehicleSalesRepository
}

func NewAnalyticsService(db *sql.DB) *AnalyticsService {
	return &AnalyticsService{db: db,
		vehicleRepository:           repository.NewVehicleRepository(),
		vehicleIMageRepository:      repository.NewVehicleImageRepository(),
		vehicleFinancialsRepository: repository.NewVehicleFinancialsRepository(),
		vehiclePurchaseRepository:   repository.NewVehiclePurchaseRepository(),
		vehicleSalesRepository:      repository.NewVehicleSalesRepository(),
		vehicleShippingRepository:   repository.NewVehicleShippingRepository(),
	}
}

func (as *AnalyticsService) GetShippingStatusCount(ctx context.Context, filter filters.Filter) (map[string]int, error) {

	count, err := as.vehicleShippingRepository.GetShippingStustVehicleCount(ctx, as.db, filter)
	if err != nil {
		return nil, err
	}

	return count, nil

}

func (as *AnalyticsService) GetSalesStatusCount(ctx context.Context, filter filters.Filter) (map[string]int, error) {

	countMap, err := as.vehicleSalesRepository.GetSalesStustVehicleCount(ctx, as.db, filter)
	if err != nil {
		return nil, err
	}

	return countMap, nil

}

func (as *AnalyticsService) GetVehicleBrandStatusCount(ctx context.Context, filter filters.Filter) (map[string]int, error) {

	countMap, err := as.vehicleRepository.GetVehicleBrandCount(ctx, as.db, filter)
	if err != nil {
		return nil, err
	}

	return countMap, nil

}

func (as *AnalyticsService) GetTotalCostRevenue(ctx context.Context, filter filters.Filter) (map[string]int, error) {

	countMap, err := as.vehicleRepository.GetVehicleBrandCount(ctx, as.db, filter)
	if err != nil {
		return nil, err
	}

	return countMap, nil

}

func (as *AnalyticsService) GetFinancialSummary(ctx context.Context, filter filters.Filter) (*response.DetailedFinancialSummary, error) {

	fiancialSummary, err := as.vehicleFinancialsRepository.GetDetailedFinancialSummary(ctx, as.db, filter)
	if err != nil {
		return nil, err
	}

	return fiancialSummary, nil

}
