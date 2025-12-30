package services

import (
	"car_service/dto/request"
	"car_service/entity"
	"car_service/logger"
	"car_service/middleware"
	"car_service/notificationHandlers"
	"car_service/repository"
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type SupplierService struct {
	db                  *sql.DB
	supplierRepository  *repository.SupplierRepository
	notificationService *NotificationService
}

func NewSupplierService(db *sql.DB, notificationService *NotificationService) *SupplierService {
	return &SupplierService{
		db:                  db,
		supplierRepository:  repository.NewSupplierRepository(),
		notificationService: notificationService,
	}
}

// CreateSupplier creates a new supplier with validation and sends a notification
func (s *SupplierService) CreateSupplier(ctx context.Context, req request.CreateSupplierRequest, authHeader string) (*entity.Supplier, error) {
	logger.WithFields(map[string]interface{}{
		"supplier_name": req.SupplierName,
		"supplier_type": req.SupplierType,
	}).Info("Creating new supplier")

	// Validation
	if req.SupplierName == "" {
		logger.Warn("Create supplier request missing supplier name")
		return nil, fmt.Errorf("supplier name is required")
	}

	if req.SupplierType == "" {
		logger.Warn("Create supplier request missing supplier type")
		return nil, fmt.Errorf("supplier type is required")
	}

	// Validate supplier type enum
	validTypes := map[string]bool{"AUCTION": true, "DEALER": true, "INDIVIDUAL": true}
	if !validTypes[req.SupplierType] {
		logger.WithField("supplier_type", req.SupplierType).Warn("Invalid supplier type provided")
		return nil, fmt.Errorf("invalid supplier type. Must be AUCTION, DEALER, or INDIVIDUAL")
	}

	// Set default for is_active
	if req.IsActive == nil {
		defaultActive := true
		req.IsActive = &defaultActive
	}

	logger.WithFields(map[string]interface{}{
		"supplier_name": req.SupplierName,
		"supplier_type": req.SupplierType,
	}).Debug("Creating supplier in repository")

	supplier, err := s.supplierRepository.CreateSupplier(ctx, s.db, req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			logger.WithFields(map[string]interface{}{
				"supplier_name": req.SupplierName,
				"error":         err.Error(),
			}).Warn("Duplicate supplier creation attempted")
			return nil, fmt.Errorf("supplier with this information already exists")
		}
		logger.WithFields(map[string]interface{}{
			"supplier_name": req.SupplierName,
			"error":         err.Error(),
		}).Error("Failed to create supplier")
		return nil, err
	}

	logger.WithFields(map[string]interface{}{
		"supplier_id":   supplier.ID,
		"supplier_name": supplier.SupplierName,
	}).Info("Supplier created successfully")

	// Extract user ID from context
	userID, _ := middleware.GetUserIDFromContext(ctx)

	// Create notification handler
	supplierCreatedHandler := notificationHandlers.NewSupplierCreatedNotificationHandler(
		supplier,
		userID,
	)

	// Send notification asynchronously
	go func() {
		if err := s.notificationService.SendNotification(supplierCreatedHandler, authHeader); err != nil {
			logger.WithFields(map[string]interface{}{
				"supplier_id":   supplier.ID,
				"supplier_name": supplier.SupplierName,
				"error":         err.Error(),
			}).Error("Failed to send supplier creation notification")
		}
	}()

	logger.WithFields(map[string]interface{}{
		"supplier_id":       supplier.ID,
		"supplier_name":     supplier.SupplierName,
		"notification_type": "supplier_created",
	}).Info("Supplier creation notification triggered")

	return supplier, nil
}

// GetAllSuppliers retrieves all suppliers with optional filtering, search, and pagination
func (s *SupplierService) GetAllSuppliers(ctx context.Context, limit, offset int, supplierType *string, activeOnly bool, searchTerm string) ([]entity.Supplier, int64, error) {
	logger.WithFields(map[string]interface{}{
		"limit":         limit,
		"offset":        offset,
		"supplier_type": supplierType,
		"active_only":   activeOnly,
		"search_term":   searchTerm,
	}).Info("Fetching all suppliers")

	suppliers, err := s.supplierRepository.GetAllSuppliers(ctx, s.db, limit, offset, supplierType, activeOnly, searchTerm)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to fetch suppliers")
		return nil, 0, err
	}

	count, err := s.supplierRepository.GetAllSuppliersCount(ctx, s.db, supplierType, activeOnly, searchTerm)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to get supplier count")
		return nil, 0, err
	}

	logger.WithFields(map[string]interface{}{
		"count":         len(suppliers),
		"total":         count,
		"supplier_type": supplierType,
		"active_only":   activeOnly,
		"search_term":   searchTerm,
	}).Info("Suppliers fetched successfully")

	return suppliers, count, nil
}

// GetSupplierByID retrieves a supplier by ID
func (s *SupplierService) GetSupplierByID(ctx context.Context, id int64) (*entity.Supplier, error) {
	logger.WithField("supplier_id", id).Debug("Fetching supplier by ID")

	supplier, err := s.supplierRepository.GetSupplierByID(ctx, s.db, id)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"supplier_id": id,
			"error":       err.Error(),
		}).Error("Failed to fetch supplier")
		return nil, err
	}

	logger.WithFields(map[string]interface{}{
		"supplier_id":   id,
		"supplier_name": supplier.SupplierName,
	}).Info("Supplier fetched successfully")

	return supplier, nil
}

// UpdateSupplier updates a supplier's information
func (s *SupplierService) UpdateSupplier(ctx context.Context, id int64, req request.UpdateSupplierRequest) error {
	logger.WithField("supplier_id", id).Info("Updating supplier")

	// Validate supplier type if provided
	if req.SupplierType != nil && *req.SupplierType != "" {
		validTypes := map[string]bool{"AUCTION": true, "DEALER": true, "INDIVIDUAL": true}
		if !validTypes[*req.SupplierType] {
			logger.WithField("supplier_type", *req.SupplierType).Warn("Invalid supplier type provided")
			return fmt.Errorf("invalid supplier type. Must be AUCTION, DEALER, or INDIVIDUAL")
		}
	}

	err := s.supplierRepository.UpdateSupplier(ctx, s.db, id, req)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"supplier_id": id,
			"error":       err.Error(),
		}).Error("Failed to update supplier")
		return err
	}

	logger.WithField("supplier_id", id).Info("Supplier updated successfully")
	return nil
}

// DeleteSupplier soft deletes a supplier
func (s *SupplierService) DeleteSupplier(ctx context.Context, id int64) error {
	logger.WithField("supplier_id", id).Info("Deleting supplier")

	err := s.supplierRepository.DeleteSupplier(ctx, s.db, id)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"supplier_id": id,
			"error":       err.Error(),
		}).Error("Failed to delete supplier")
		return err
	}

	logger.WithField("supplier_id", id).Info("Supplier deleted successfully")
	return nil
}

// SearchSuppliers searches for suppliers by search term
func (s *SupplierService) SearchSuppliers(ctx context.Context, searchTerm string) ([]entity.Supplier, error) {
	logger.WithField("search_term", searchTerm).Info("Searching suppliers")

	if searchTerm == "" {
		return nil, fmt.Errorf("search term is required")
	}

	suppliers, err := s.supplierRepository.SearchSuppliers(ctx, s.db, searchTerm)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"search_term": searchTerm,
			"error":       err.Error(),
		}).Error("Failed to search suppliers")
		return nil, err
	}

	logger.WithFields(map[string]interface{}{
		"search_term": searchTerm,
		"count":       len(suppliers),
	}).Info("Supplier search completed successfully")

	return suppliers, nil
}
