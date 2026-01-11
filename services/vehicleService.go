package services

import (
	"car_service/dto/request"
	"car_service/dto/response"
	"car_service/entity"
	"car_service/filters"
	"car_service/logger"
	"car_service/middleware"
	"car_service/notificationHandlers"
	"car_service/repository"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"path"
	"time"

	_ "github.com/lib/pq"
)

type VehicleService struct {
	db                               *sql.DB
	vehicleRepository                *repository.VehicleRepository
	vehicleIMageRepository           *repository.VehicleImageRepository
	vehicleDocumentRepository        *repository.VehicleDocumentRepository
	vehicleFinancialsRepository      *repository.VehicleFinancialsRepository
	vehicleShippingRepository        *repository.VehicleShippingRepository
	vehiclePurchaseRepository        *repository.VehiclePurchaseRepository
	vehicleSalesRepository           *repository.VehicleSalesRepository
	vehicleShippingHistoryRepository *repository.VehicleShippingHistoryRepository
	vehiclePurchaseHistoryRepository *repository.VehiclePurchaseHistoryRepository
	vehicleShareTokenRepository      *repository.VehicleShareTokenRepository
	customerRepository               *repository.CustomerRepository
	supplierRepository               *repository.SupplierRepository
	notificationService              *NotificationService
	S3Service                        *S3Service
}

func NewVehicleService(db *sql.DB, notificationService *NotificationService, service *S3Service) *VehicleService {
	return &VehicleService{db: db,
		vehicleRepository:                repository.NewVehicleRepository(),
		vehicleIMageRepository:           repository.NewVehicleImageRepository(),
		vehicleDocumentRepository:        repository.NewVehicleDocumentRepository(),
		vehicleFinancialsRepository:      repository.NewVehicleFinancialsRepository(),
		vehiclePurchaseRepository:        repository.NewVehiclePurchaseRepository(),
		vehicleSalesRepository:           repository.NewVehicleSalesRepository(),
		vehicleShippingRepository:        repository.NewVehicleShippingRepository(),
		vehicleShippingHistoryRepository: repository.NewVehicleShippingHistoryRepository(),
		vehiclePurchaseHistoryRepository: repository.NewVehiclePurchaseHistoryRepository(),
		vehicleShareTokenRepository:      repository.NewVehicleShareTokenRepository(),
		customerRepository:               repository.NewCustomerRepository(),
		supplierRepository:               repository.NewSupplierRepository(),
		notificationService:              notificationService,
		S3Service:                        service,
	}
}

func (s *VehicleService) GetAllVehicles(ctx context.Context, limit int, offset int, filter filters.Filter) (*response.VehiclesResponse, error) {
	logger.WithFields(map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}).Debug("Fetching all vehicles")

	vehicles, err := s.vehicleRepository.GetAllVehicles(ctx, s.db, limit, offset, filter)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"error":  err.Error(),
		}).Error("Failed to fetch vehicles")
		return nil, err
	}

	vehicleCount, err := s.vehicleRepository.GetAllVehicleCount(ctx, s.db, filter)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to get vehicle count")
		return nil, err
	}

	logger.WithFields(map[string]interface{}{
		"count":  len(vehicles),
		"total":  vehicleCount,
		"limit":  limit,
		"offset": offset,
	}).Info("Successfully fetched vehicles")

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
	logger.WithField("vehicle_id", id).Debug("Fetching vehicle by ID")

	vehicle, err := s.vehicleRepository.GetVehicleByID(ctx, s.db, id)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": id,
			"error":      err.Error(),
		}).Error("Failed to fetch vehicle")
		return nil, err
	}

	logger.WithFields(map[string]interface{}{
		"vehicle_id": id,
		"code":       vehicle.Code,
	}).Debug("Fetching vehicle related data")

	shipping, err := s.vehicleShippingRepository.GetByVehicleID(ctx, s.db, vehicle.ID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": id,
			"error":      err.Error(),
		}).Error("Failed to fetch shipping data")
		return nil, err
	}

	purchase, err := s.vehiclePurchaseRepository.GetByVehicleID(ctx, s.db, vehicle.ID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": id,
			"error":      err.Error(),
		}).Error("Failed to fetch purchase data")
		return nil, err
	}

	finance, err := s.vehicleFinancialsRepository.GetByVehicleID(ctx, s.db, vehicle.ID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": id,
			"error":      err.Error(),
		}).Error("Failed to fetch financial data")
		return nil, err
	}

	sales, err := s.vehicleSalesRepository.GetByVehicleID(ctx, s.db, vehicle.ID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": id,
			"error":      err.Error(),
		}).Error("Failed to fetch sales data")
		return nil, err
	}

	images, err := s.vehicleIMageRepository.GetByVehicleID(ctx, s.db, vehicle.ID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": id,
			"error":      err.Error(),
		}).Error("Failed to fetch vehicle images")
		return nil, err
	}

	documents, err := s.vehicleDocumentRepository.GetByVehicleID(ctx, s.db, vehicle.ID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": id,
			"error":      err.Error(),
		}).Error("Failed to fetch vehicle documents")
		return nil, err
	}

	logger.WithFields(map[string]interface{}{
		"vehicle_id":     id,
		"image_count":    len(images),
		"document_count": len(documents),
	}).Info("Successfully fetched complete vehicle data")

	var vehicleComplete entity.VehicleComplete

	vehicleComplete.Vehicle = *vehicle
	vehicleComplete.VehicleShipping = *shipping
	vehicleComplete.VehiclePurchase = *purchase
	vehicleComplete.VehicleFinancials = *finance
	vehicleComplete.VehicleSales = *sales
	vehicleComplete.VehicleImages = images
	vehicleComplete.VehicleDocuments = documents

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

func (s *VehicleService) CreateVehicle(ctx context.Context, req request.CreateVehicleRequest, authHeader string) (*entity.Vehicle, error) {
	logger.WithFields(map[string]interface{}{
		"code":  req.Code,
		"make":  req.Make,
		"model": req.Model,
	}).Info("Creating new vehicle")

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to begin transaction for vehicle creation")
		return nil, err
	}
	defer tx.Rollback() // Will be ignored if tx is committed

	logger.Debug("Inserting vehicle record")
	vehicleID, err := s.vehicleRepository.Insert(ctx, tx, req)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"code":  req.Code,
			"error": err.Error(),
		}).Error("Failed to insert vehicle")
		return nil, err
	}

	logger.WithField("vehicle_id", vehicleID).Debug("Inserting default shipping record")
	err = s.vehicleShippingRepository.InsertDefault(ctx, tx, vehicleID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": vehicleID,
			"error":      err.Error(),
		}).Error("Failed to insert default shipping record")
		return nil, err
	}

	logger.WithField("vehicle_id", vehicleID).Debug("Inserting default financial record")
	err = s.vehicleFinancialsRepository.InsertDefault(ctx, tx, vehicleID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": vehicleID,
			"error":      err.Error(),
		}).Error("Failed to insert default financial record")
		return nil, err
	}

	logger.WithField("vehicle_id", vehicleID).Debug("Inserting default sales record")
	err = s.vehicleSalesRepository.InsertDefault(ctx, tx, vehicleID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": vehicleID,
			"error":      err.Error(),
		}).Error("Failed to insert default sales record")
		return nil, err
	}

	logger.WithField("vehicle_id", vehicleID).Debug("Inserting default purchase record")
	err = s.vehiclePurchaseRepository.InsertDefault(ctx, tx, vehicleID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": vehicleID,
			"error":      err.Error(),
		}).Error("Failed to insert default purchase record")
		return nil, err
	}

	// Commit transaction
	logger.WithField("vehicle_id", vehicleID).Debug("Committing transaction")
	err = tx.Commit()
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": vehicleID,
			"error":      err.Error(),
		}).Error("Failed to commit transaction for vehicle creation")
		return nil, err
	}

	vehicle, err := s.vehicleRepository.GetVehicleByID(ctx, s.db, vehicleID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": vehicleID,
			"error":      err.Error(),
		}).Error("Failed to fetch created vehicle")
		return nil, err
	}

	logger.WithFields(map[string]interface{}{
		"vehicle_id": vehicleID,
		"code":       vehicle.Code,
	}).Info("Vehicle created successfully")

	// Extract user ID from context
	userID, _ := middleware.GetUserIDFromContext(ctx)

	// Create notification handler
	vehicleCreatedHandler := notificationHandlers.NewVehicleCreatedNotificationHandler(
		vehicle,
		userID,
	)

	// Send notification asynchronously
	go func() {
		if err := s.notificationService.SendNotification(vehicleCreatedHandler, authHeader); err != nil {
			logger.WithFields(map[string]interface{}{
				"vehicle_id": vehicleID,
				"code":       vehicle.Code,
				"error":      err.Error(),
			}).Error("Failed to send vehicle creation notification")
		}
	}()

	logger.WithFields(map[string]interface{}{
		"vehicle_id":        vehicleID,
		"code":              vehicle.Code,
		"notification_type": "vehicle_created",
	}).Info("Vehicle creation notification triggered")

	return vehicle, nil
}

func (s *VehicleService) UpdateShippingStatus(ctx context.Context, vehicleID int64, detailsRequest request.ShippingDetailsRequest, authHeader string) error {
	// Fetch old shipping status before update
	oldShipping, err := s.vehicleShippingRepository.GetByVehicleID(ctx, s.db, vehicleID)
	if err != nil {
		return err
	}

	var oldStatus string
	if oldShipping != nil {
		oldStatus = oldShipping.ShippingStatus
	}

	// Auto-set dates based on shipping status if not provided
	currentTime := time.Now().Format(time.RFC3339)

	// Auto-set shipment_date when status changes to SHIPPED or IN_TRANSIT
	if (detailsRequest.ShippingStatus == "SHIPPED" || detailsRequest.ShippingStatus == "IN_TRANSIT") && detailsRequest.ShipmentDate == nil {
		detailsRequest.ShipmentDate = &currentTime
	}

	// Auto-set arrival_date when status changes to ARRIVED
	if detailsRequest.ShippingStatus == "ARRIVED" && detailsRequest.ArrivalDate == nil {
		detailsRequest.ArrivalDate = &currentTime
	}

	// Auto-set clearing_date when status changes to CLEARED
	if detailsRequest.ShippingStatus == "CLEARED" && detailsRequest.ClearingDate == nil {
		detailsRequest.ClearingDate = &currentTime
	}

	// Update shipping status
	err = s.vehicleShippingRepository.UpdateShippingStatus(ctx, s.db, vehicleID, detailsRequest)
	if err != nil {
		return err
	}

	// Only send email if status actually changed
	newStatus := detailsRequest.ShippingStatus
	if oldStatus != newStatus {
		// Fetch vehicle details
		vehicle, err := s.vehicleRepository.GetVehicleByID(ctx, s.db, vehicleID)
		if err != nil {
			// Log error but don't fail the update
			return nil
		}

		// Fetch sales information to get customer
		sales, err := s.vehicleSalesRepository.GetByVehicleID(ctx, s.db, vehicleID)
		var customer *entity.Customer
		if sales != nil && sales.CustomerID != nil {
			customer, err = s.customerRepository.GetCustomerByID(ctx, s.db, *sales.CustomerID)
			if err != nil {
				return err
			}
		}

		userID, _ := middleware.GetUserIDFromContext(ctx)

		shippingStatusHandler := notificationHandlers.NewShippingStatusNotificationHandler(
			vehicle, customer, oldStatus, newStatus, userID,
		)

		// Send notification asynchronously with error handling
		go func() {
			if err := s.notificationService.SendNotification(shippingStatusHandler, authHeader); err != nil {
				logger.WithFields(map[string]interface{}{
					"vehicle_id": vehicleID,
					"old_status": oldStatus,
					"new_status": newStatus,
					"error":      err.Error(),
				}).Error("Failed to send shipping status notification")
			}
		}()
	}

	return nil
}

func (s *VehicleService) UpdatePurchaseDetails(ctx context.Context, id int64, purchaseRequest *request.PurchaseRequest, authHeader string) error {
	// Fetch old purchase status before update
	oldPurchase, err := s.vehiclePurchaseRepository.GetByVehicleID(ctx, s.db, id)
	if err != nil {
		return err
	}

	var oldStatus string
	if oldPurchase != nil {
		oldStatus = oldPurchase.PurchaseStatus
	}

	// Update purchase status
	err = s.vehiclePurchaseRepository.UpdateVehiclePurchase(ctx, s.db, id, purchaseRequest)
	if err != nil {
		return err
	}

	// If LC cost or exchange rate was provided, recalculate the financial total
	if purchaseRequest.LCCostJPY != nil || purchaseRequest.ExchangeRate != nil {
		// Get current financial details
		financial, err := s.vehicleFinancialsRepository.GetByVehicleID(ctx, s.db, id)
		if err == nil && financial != nil {
			// Create a financial update request with current values to trigger recalculation
			financialUpdate := &request.FinancialDetailsRequest{
				ChargesLKR:       financial.ChargesLKR,
				TTLKR:            financial.TTLKR,
				DutyLKR:          financial.DutyLKR,
				ClearingLKR:      financial.ClearingLKR,
				OtherExpensesLKR: financial.OtherExpensesLKR,
			}
			// This will recalculate total_cost_lkr including the new LC cost with exchange rate
			_ = s.vehicleFinancialsRepository.UpdateFinancialDetails(ctx, s.db, id, financialUpdate)
		}
	}

	// Only send email if status actually changed
	// Handle nil check for new status
	if purchaseRequest.PurchaseStatus == nil {
		return nil
	}

	newStatus := *purchaseRequest.PurchaseStatus
	if oldStatus != newStatus {
		// Fetch vehicle details
		vehicle, err := s.vehicleRepository.GetVehicleByID(ctx, s.db, id)
		if err != nil {
			// Log error but don't fail the update
			return nil
		}

		// Fetch sales information to get customer
		sales, err := s.vehicleSalesRepository.GetByVehicleID(ctx, s.db, id)

		// Fetch customer details
		var customer *entity.Customer
		if sales != nil && sales.CustomerID != nil {
			customer, err = s.customerRepository.GetCustomerByID(ctx, s.db, *sales.CustomerID)
			if err != nil {
				return err
			}
		}

		// Fetch updated purchase details to get all fields
		updatedPurchase, err := s.vehiclePurchaseRepository.GetByVehicleID(ctx, s.db, id)
		if err != nil {
			return nil
		}

		// Fetch supplier details if available
		var supplierName *string
		if updatedPurchase != nil && updatedPurchase.SupplierID != nil {
			supplier, err := s.supplierRepository.GetSupplierByID(ctx, s.db, *updatedPurchase.SupplierID)
			if err == nil && supplier != nil {
				supplierName = &supplier.SupplierName
			}
		}

		userID, _ := middleware.GetUserIDFromContext(ctx)

		purchasingStatusHandler := notificationHandlers.NewPurchasingStatusNotificationHandler(
			vehicle, customer, oldStatus, newStatus, supplierName, userID,
		)

		// Send notification asynchronously with error handling
		go func() {
			if err := s.notificationService.SendNotification(purchasingStatusHandler, authHeader); err != nil {
				logger.WithFields(map[string]interface{}{
					"vehicle_id": id,
					"old_status": oldStatus,
					"new_status": newStatus,
					"error":      err.Error(),
				}).Error("Failed to send purchasing status notification")
			}
		}()
	}

	return nil
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

// InsertVehicleDocument inserts vehicle documents into the database
func (s *VehicleService) InsertVehicleDocument(ctx context.Context, documents []entity.VehicleDocument) ([]entity.VehicleDocument, error) {
	var vehicleDocuments []entity.VehicleDocument
	for _, doc := range documents {
		document, err := s.vehicleDocumentRepository.InsertVehicleDocument(ctx, s.db, &doc)
		if err != nil {
			return nil, err
		}
		vehicleDocuments = append(vehicleDocuments, *document)
	}
	return vehicleDocuments, nil
}

// GetVehicleDocumentByID retrieves a single document by ID
func (s *VehicleService) GetVehicleDocumentByID(ctx context.Context, id int64) (*entity.VehicleDocument, error) {
	return s.vehicleDocumentRepository.GetByID(ctx, s.db, id)
}

// DeleteVehicle deletes a vehicle by ID
func (s *VehicleService) DeleteVehicle(ctx context.Context, vehicleID int64, authHeader string) error {
	logger.WithField("vehicle_id", vehicleID).Info("Deleting vehicle")

	// Fetch vehicle details before deletion for notification
	vehicle, err := s.vehicleRepository.GetVehicleByID(ctx, s.db, vehicleID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": vehicleID,
			"error":      err.Error(),
		}).Error("Failed to fetch vehicle before deletion")
		return err
	}

	// Delete the vehicle
	err = s.vehicleRepository.DeleteVehicle(ctx, s.db, vehicleID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": vehicleID,
			"error":      err.Error(),
		}).Error("Failed to delete vehicle")
		return err
	}

	logger.WithFields(map[string]interface{}{
		"vehicle_id": vehicleID,
		"code":       vehicle.Code,
	}).Info("Vehicle deleted successfully")

	// Extract user ID from context
	userID, _ := middleware.GetUserIDFromContext(ctx)

	// Create notification handler
	vehicleDeletedHandler := notificationHandlers.NewVehicleDeletedNotificationHandler(
		vehicle,
		userID,
	)

	// Send notification asynchronously
	go func() {
		if err := s.notificationService.SendNotification(vehicleDeletedHandler, authHeader); err != nil {
			logger.WithFields(map[string]interface{}{
				"vehicle_id": vehicleID,
				"code":       vehicle.Code,
				"error":      err.Error(),
			}).Error("Failed to send vehicle deletion notification")
		}
	}()

	logger.WithFields(map[string]interface{}{
		"vehicle_id":        vehicleID,
		"code":              vehicle.Code,
		"notification_type": "vehicle_deleted",
	}).Info("Vehicle deletion notification triggered")

	return nil
}

// GetPurchaseHistory retrieves purchase change history for a vehicle
func (s *VehicleService) GetPurchaseHistory(ctx context.Context, vehicleID int64) ([]entity.VehiclePurchaseHistoryWithDetails, error) {
	return s.vehiclePurchaseHistoryRepository.GetHistoryByVehicleID(ctx, s.db, vehicleID)
}

// GetRecentPurchaseHistory retrieves recent purchase changes across all vehicles
func (s *VehicleService) GetRecentPurchaseHistory(ctx context.Context, limit int) ([]entity.VehiclePurchaseHistoryWithDetails, error) {
	return s.vehiclePurchaseHistoryRepository.GetRecentHistory(ctx, s.db, limit)
}

// GetPurchaseHistoryByStatus retrieves purchase history for a specific status
func (s *VehicleService) GetPurchaseHistoryByStatus(ctx context.Context, status string) ([]entity.VehiclePurchaseHistoryWithDetails, error) {
	return s.vehiclePurchaseHistoryRepository.GetHistoryByStatus(ctx, s.db, status)
}

// GetPurchaseHistoryBySupplier retrieves purchase history for a specific supplier
func (s *VehicleService) GetPurchaseHistoryBySupplier(ctx context.Context, supplierID int64) ([]entity.VehiclePurchaseHistoryWithDetails, error) {
	return s.vehiclePurchaseHistoryRepository.GetHistoryBySupplier(ctx, s.db, supplierID)
}

// SetVehicleFeatured marks a vehicle as featured or unfeatured
func (s *VehicleService) SetVehicleFeatured(ctx context.Context, vehicleID int64, isFeatured bool, authHeader string) error {
	logger.WithFields(map[string]interface{}{
		"vehicle_id":  vehicleID,
		"is_featured": isFeatured,
	}).Info("Setting vehicle featured status")

	// Fetch vehicle details before update for notification
	vehicle, err := s.vehicleRepository.GetVehicleByID(ctx, s.db, vehicleID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": vehicleID,
			"error":      err.Error(),
		}).Error("Failed to fetch vehicle before updating featured status")
		return err
	}

	err = s.vehicleRepository.SetVehicleFeatured(ctx, s.db, vehicleID, isFeatured)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": vehicleID,
			"error":      err.Error(),
		}).Error("Failed to set vehicle featured status")
		return err
	}

	logger.WithField("vehicle_id", vehicleID).Info("Vehicle featured status updated successfully")

	// Extract user ID from context
	userID, _ := middleware.GetUserIDFromContext(ctx)

	// Create notification handler
	featuredVehicleHandler := notificationHandlers.NewFeaturedVehicleNotificationHandler(
		vehicle,
		isFeatured,
		userID,
	)

	// Send notification asynchronously
	go func() {
		if err := s.notificationService.SendNotification(featuredVehicleHandler, authHeader); err != nil {
			logger.WithFields(map[string]interface{}{
				"vehicle_id":  vehicleID,
				"code":        vehicle.Code,
				"is_featured": isFeatured,
				"error":       err.Error(),
			}).Error("Failed to send vehicle featured status notification")
		}
	}()

	logger.WithFields(map[string]interface{}{
		"vehicle_id":        vehicleID,
		"code":              vehicle.Code,
		"notification_type": "vehicle_featured_status_changed",
		"is_featured":       isFeatured,
	}).Info("Vehicle featured status notification triggered")

	return nil
}

// GetFeaturedVehicles retrieves all featured vehicles
func (s *VehicleService) GetFeaturedVehicles(ctx context.Context, limit int) ([]entity.VehicleComplete, error) {
	logger.WithField("limit", limit).Info("Fetching featured vehicles")

	// Get permissions from context
	permissions, ok := ctx.Value("permissions").([]string)
	if !ok {
		logger.Error("Permissions not found in context")
		return nil, fmt.Errorf("permissions not found in context")
	}

	vehicles, err := s.vehicleRepository.GetFeaturedVehicles(ctx, s.db, limit, permissions)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to fetch featured vehicles")
		return nil, err
	}

	logger.WithFields(map[string]interface{}{
		"count": len(vehicles),
		"limit": limit,
	}).Info("Featured vehicles fetched successfully")

	return vehicles, nil
}

// GenerateShareToken generates a shareable token for a vehicle
func (s *VehicleService) GenerateShareToken(ctx context.Context, vehicleID int64, req request.PublicTokenRequest) (*entity.VehicleShareToken, error) {
	logger.WithFields(map[string]interface{}{
		"vehicle_id":      vehicleID,
		"expire_in_days":  req.ExpireInDays,
		"include_details": req.IncludeDetails,
	}).Info("Generating share token for vehicle")

	// Verify vehicle exists
	_, err := s.vehicleRepository.GetVehicleByID(ctx, s.db, vehicleID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": vehicleID,
			"error":      err.Error(),
		}).Error("Vehicle not found")
		return nil, err
	}

	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		logger.WithField("error", err.Error()).Error("Failed to generate random token")
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes)

	// Calculate expiration date
	expiresAt := time.Now().AddDate(0, 0, req.ExpireInDays)

	// Get user ID from context
	userID, _ := middleware.GetUserIDFromContext(ctx)

	// Create share token entity
	shareToken := &entity.VehicleShareToken{
		VehicleID:      vehicleID,
		Token:          token,
		ExpiresAt:      expiresAt,
		IncludeDetails: req.IncludeDetails,
		CreatedBy:      userID,
		IsActive:       true,
	}

	// Insert into database
	id, err := s.vehicleShareTokenRepository.Insert(ctx, s.db, shareToken)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": vehicleID,
			"error":      err.Error(),
		}).Error("Failed to insert share token")
		return nil, err
	}

	shareToken.ID = id
	shareToken.CreatedAt = time.Now()

	logger.WithFields(map[string]interface{}{
		"vehicle_id": vehicleID,
		"token_id":   id,
		"expires_at": expiresAt,
	}).Info("Share token generated successfully")

	return shareToken, nil
}

// GetPublicVehicleData retrieves public vehicle data using a share token
func (s *VehicleService) GetPublicVehicleData(ctx context.Context, token string) (*response.PublicVehicleResponse, error) {
	logger.WithField("token", token[:8]+"...").Info("Fetching public vehicle data with share token")

	// Validate token and get token details
	shareToken, err := s.vehicleShareTokenRepository.GetByToken(ctx, s.db, token)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.WithField("token", token[:8]+"...").Warn("Invalid or expired share token")
			return nil, fmt.Errorf("invalid or expired share token")
		}
		logger.WithField("error", err.Error()).Error("Failed to retrieve share token")
		return nil, err
	}

	// Fetch vehicle details
	vehicle, err := s.vehicleRepository.GetVehicleByID(ctx, s.db, shareToken.VehicleID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"vehicle_id": shareToken.VehicleID,
			"error":      err.Error(),
		}).Error("Failed to fetch vehicle for public share")
		return nil, err
	}

	// Build public response with basic info (always included)
	publicResponse := &response.PublicVehicleResponse{
		Code:                vehicle.Code,
		Make:                vehicle.Make,
		Model:               vehicle.Model,
		YearOfManufacture:   vehicle.YearOfManufacture,
		Color:               vehicle.Color,
		ChassisID:           vehicle.ChassisID,
		ConditionStatus:     vehicle.ConditionStatus,
		TrimLevel:           vehicle.TrimLevel,
		MileageKm:           vehicle.MileageKm,
		YearOfRegistration:  vehicle.YearOfRegistration,
		AuctionGrade:        vehicle.AuctionGrade,
		AuctionPrice:        vehicle.AuctionPrice,
		Currency:            vehicle.Currency,
		ShareTokenExpiresAt: shareToken.ExpiresAt,
	}

	// Add optional details based on IncludeDetails
	for _, detail := range shareToken.IncludeDetails {
		switch detail {
		case "shipping":
			shipping, err := s.vehicleShippingRepository.GetByVehicleID(ctx, s.db, vehicle.ID)
			if err == nil && shipping != nil {
				publicResponse.ShippingStatus = &shipping.ShippingStatus
				publicResponse.VesselName = shipping.VesselName
				publicResponse.DepartureHarbour = shipping.DepartureHarbour
				publicResponse.ShipmentDate = shipping.ShipmentDate
				publicResponse.ArrivalDate = shipping.ArrivalDate
				publicResponse.ClearingDate = shipping.ClearingDate
			}

		case "financial":
			financial, err := s.vehicleFinancialsRepository.GetByVehicleID(ctx, s.db, vehicle.ID)
			if err == nil && financial != nil {
				publicResponse.TotalCostJPY = &financial.TotalCostLKR
			}

		case "purchase":
			purchase, err := s.vehiclePurchaseRepository.GetByVehicleID(ctx, s.db, vehicle.ID)
			if err == nil && purchase != nil {
				publicResponse.PurchaseStatus = &purchase.PurchaseStatus
				publicResponse.PurchaseDate = purchase.PurchaseDate
			}

		case "images":
			images, err := s.vehicleIMageRepository.GetByVehicleID(ctx, s.db, vehicle.ID)
			if err == nil && len(images) > 0 {
				var imageResponses []response.VehicleImageResponse
				for _, img := range images {
					presignedResponse, _ := s.S3Service.GetPresignedURL(ctx, path.Join((fmt.Sprintf("vehicles/%d/images", vehicle.ID)), img.Filename), 15)
					imageResponses = append(imageResponses, response.VehicleImageResponse{
						ID:        img.ID,
						ImageURL:  presignedResponse.PresignedURL,
						IsPrimary: img.IsPrimary,
					})
				}

				publicResponse.Images = imageResponses
			}
		}
	}

	logger.WithFields(map[string]interface{}{
		"vehicle_id":      vehicle.ID,
		"vehicle_code":    vehicle.Code,
		"include_details": shareToken.IncludeDetails,
	}).Info("Public vehicle data fetched successfully")

	return publicResponse, nil
}
