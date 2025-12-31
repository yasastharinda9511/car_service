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

type CustomerService struct {
	db                  *sql.DB
	customerRepository  *repository.CustomerRepository
	notificationService *NotificationService
}

func NewCustomerService(db *sql.DB, notificationService *NotificationService) *CustomerService {
	return &CustomerService{
		db:                  db,
		customerRepository:  repository.NewCustomerRepository(),
		notificationService: notificationService,
	}
}

// CreateCustomer creates a new customer and sends a notification
func (s *CustomerService) CreateCustomer(ctx context.Context, req request.CreateCustomerRequest, authHeader string) (*entity.Customer, error) {
	logger.WithFields(map[string]interface{}{
		"customer_name": req.CustomerName,
		"customer_type": req.CustomerType,
	}).Info("Creating new customer")

	// Validation
	if req.CustomerName == "" {
		logger.Warn("Create customer request missing customer name")
		return nil, fmt.Errorf("customer name is required")
	}

	if req.CustomerType == "" {
		logger.Warn("Create customer request missing customer type")
		return nil, fmt.Errorf("customer type is required")
	}

	// Validate customer type enum
	validTypes := map[string]bool{"INDIVIDUAL": true, "BUSINESS": true}
	if !validTypes[req.CustomerType] {
		logger.WithField("customer_type", req.CustomerType).Warn("Invalid customer type provided")
		return nil, fmt.Errorf("invalid customer type. Must be INDIVIDUAL or BUSINESS")
	}

	// Set default for is_active
	if req.IsActive == nil {
		defaultActive := true
		req.IsActive = &defaultActive
	}

	logger.WithFields(map[string]interface{}{
		"customer_name": req.CustomerName,
		"customer_type": req.CustomerType,
	}).Debug("Creating customer in repository")

	customer, err := s.customerRepository.CreateCustomer(ctx, s.db, req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			logger.WithFields(map[string]interface{}{
				"customer_name": req.CustomerName,
				"error":         err.Error(),
			}).Warn("Duplicate customer creation attempted")
			return nil, fmt.Errorf("customer with this information already exists")
		}
		logger.WithFields(map[string]interface{}{
			"customer_name": req.CustomerName,
			"error":         err.Error(),
		}).Error("Failed to create customer")
		return nil, err
	}

	logger.WithFields(map[string]interface{}{
		"customer_id":   customer.ID,
		"customer_name": customer.CustomerName,
	}).Info("Customer created successfully")

	// Extract user ID from context
	userID, _ := middleware.GetUserIDFromContext(ctx)

	// Create notification handler
	customerCreatedHandler := notificationHandlers.NewCustomerCreatedNotificationHandler(
		customer,
		userID,
	)

	// Send notification asynchronously
	go func() {
		if err := s.notificationService.SendNotification(customerCreatedHandler, authHeader); err != nil {
			logger.WithFields(map[string]interface{}{
				"customer_id":   customer.ID,
				"customer_name": customer.CustomerName,
				"error":         err.Error(),
			}).Error("Failed to send customer creation notification")
		}
	}()

	logger.WithFields(map[string]interface{}{
		"customer_id":       customer.ID,
		"customer_name":     customer.CustomerName,
		"notification_type": "customer_created",
	}).Info("Customer creation notification triggered")

	return customer, nil
}

// GetAllCustomers retrieves all customers with optional filtering, search, and pagination
func (s *CustomerService) GetAllCustomers(ctx context.Context, limit, offset int, customerType *string, activeOnly bool, searchTerm string) ([]entity.Customer, int64, error) {
	logger.WithFields(map[string]interface{}{
		"limit":         limit,
		"offset":        offset,
		"customer_type": customerType,
		"active_only":   activeOnly,
		"search_term":   searchTerm,
	}).Info("Fetching all customers")

	customers, err := s.customerRepository.GetAllCustomers(ctx, s.db, limit, offset, customerType, activeOnly, searchTerm)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to fetch customers")
		return nil, 0, err
	}

	count, err := s.customerRepository.GetAllCustomersCount(ctx, s.db, customerType, activeOnly, searchTerm)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to get customer count")
		return nil, 0, err
	}

	logger.WithFields(map[string]interface{}{
		"count":         len(customers),
		"total":         count,
		"customer_type": customerType,
		"active_only":   activeOnly,
		"search_term":   searchTerm,
	}).Info("Customers fetched successfully")

	return customers, count, nil
}

// GetCustomerByID retrieves a customer by ID
func (s *CustomerService) GetCustomerByID(ctx context.Context, id int64) (*entity.Customer, error) {
	logger.WithField("customer_id", id).Debug("Fetching customer by ID")

	customer, err := s.customerRepository.GetCustomerByID(ctx, s.db, id)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"customer_id": id,
			"error":       err.Error(),
		}).Error("Failed to fetch customer")
		return nil, err
	}

	logger.WithFields(map[string]interface{}{
		"customer_id":   id,
		"customer_name": customer.CustomerName,
	}).Info("Customer fetched successfully")

	return customer, nil
}

// UpdateCustomer updates a customer's information
func (s *CustomerService) UpdateCustomer(ctx context.Context, id int64, req request.UpdateCustomerRequest) error {
	logger.WithField("customer_id", id).Info("Updating customer")

	// Validate customer type if provided
	if req.CustomerType != nil && *req.CustomerType != "" {
		validTypes := map[string]bool{"INDIVIDUAL": true, "BUSINESS": true}
		if !validTypes[*req.CustomerType] {
			logger.WithField("customer_type", *req.CustomerType).Warn("Invalid customer type provided")
			return fmt.Errorf("invalid customer type. Must be INDIVIDUAL or BUSINESS")
		}
	}

	err := s.customerRepository.UpdateCustomer(ctx, s.db, id, req)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"customer_id": id,
			"error":       err.Error(),
		}).Error("Failed to update customer")
		return err
	}

	logger.WithField("customer_id", id).Info("Customer updated successfully")
	return nil
}

// DeleteCustomer soft deletes a customer
func (s *CustomerService) DeleteCustomer(ctx context.Context, id int64, authHeader string) error {
	logger.WithField("customer_id", id).Info("Deleting customer")

	// Fetch customer details before deletion for notification
	customer, err := s.customerRepository.GetCustomerByID(ctx, s.db, id)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"customer_id": id,
			"error":       err.Error(),
		}).Error("Failed to fetch customer before deletion")
		return err
	}

	// Delete the customer
	err = s.customerRepository.DeleteCustomer(ctx, s.db, id)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"customer_id": id,
			"error":       err.Error(),
		}).Error("Failed to delete customer")
		return err
	}

	logger.WithFields(map[string]interface{}{
		"customer_id":   id,
		"customer_name": customer.CustomerName,
	}).Info("Customer deleted successfully")

	// Extract user ID from context
	userID, _ := middleware.GetUserIDFromContext(ctx)

	// Create notification handler
	customerDeletedHandler := notificationHandlers.NewCustomerDeletedNotificationHandler(
		customer,
		userID,
	)

	// Send notification asynchronously
	go func() {
		if err := s.notificationService.SendNotification(customerDeletedHandler, authHeader); err != nil {
			logger.WithFields(map[string]interface{}{
				"customer_id":   id,
				"customer_name": customer.CustomerName,
				"error":         err.Error(),
			}).Error("Failed to send customer deletion notification")
		}
	}()

	logger.WithFields(map[string]interface{}{
		"customer_id":       id,
		"customer_name":     customer.CustomerName,
		"notification_type": "customer_deleted",
	}).Info("Customer deletion notification triggered")

	return nil
}

// SearchCustomers searches for customers by search term
func (s *CustomerService) SearchCustomers(ctx context.Context, searchTerm string) ([]entity.Customer, error) {
	logger.WithField("search_term", searchTerm).Info("Searching customers")

	if searchTerm == "" {
		return nil, fmt.Errorf("search term is required")
	}

	customers, err := s.customerRepository.SearchCustomers(ctx, s.db, searchTerm)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"search_term": searchTerm,
			"error":       err.Error(),
		}).Error("Failed to search customers")
		return nil, err
	}

	logger.WithFields(map[string]interface{}{
		"search_term": searchTerm,
		"count":       len(customers),
	}).Info("Customer search completed successfully")

	return customers, nil
}
