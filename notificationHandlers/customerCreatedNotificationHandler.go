package notificationHandlers

import (
	"car_service/dto/request"
	"car_service/entity"
	"fmt"
)

// CustomerCreatedNotificationHandler handles building customer creation notification payloads
type CustomerCreatedNotificationHandler struct {
	Customer *entity.Customer
	UserID   string
}

// NewCustomerCreatedNotificationHandler creates a new customer created notification handler
func NewCustomerCreatedNotificationHandler(
	customer *entity.Customer,
	userID string,
) NotificationHandler {
	return &CustomerCreatedNotificationHandler{
		Customer: customer,
		UserID:   userID,
	}
}

// BuildNotificationRequest constructs the notification request for customer creation
func (h *CustomerCreatedNotificationHandler) BuildNotificationRequest() *request.NotificationRequest {
	// Build the payload with customer information
	payload := map[string]interface{}{
		"customer_id":   h.Customer.ID,
		"customer_name": h.Customer.CustomerName,
		"message":       h.buildMessage(),
		"customer_type": h.Customer.CustomerType,
	}

	// Add optional fields if they exist
	if h.Customer.CustomerTitle != nil {
		payload["customer_title"] = *h.Customer.CustomerTitle
	}
	if h.Customer.ContactNumber != nil {
		payload["contact_number"] = *h.Customer.ContactNumber
	}
	if h.Customer.Email != nil {
		payload["email"] = *h.Customer.Email
	}
	if h.Customer.Address != nil {
		payload["address"] = *h.Customer.Address
	}

	// Build metadata
	metadata := map[string]interface{}{
		"user_id": h.UserID,
		"service": "car-service",
		"event":   "customer_created",
	}

	return &request.NotificationRequest{
		NotificationType: h.GetNotificationType(),
		Source:           "car-service",
		Payload:          payload,
		Priority:         "normal",
		ReferenceID:      fmt.Sprintf("CUST-%d", h.Customer.ID),
		Metadata:         metadata,
	}
}

// GetNotificationType returns the notification type
func (h *CustomerCreatedNotificationHandler) GetNotificationType() string {
	return "customer_created"
}

// buildMessage creates a human-readable message for the notification
func (h *CustomerCreatedNotificationHandler) buildMessage() string {
	return fmt.Sprintf(
		"New customer '%s' (%s) has been created in the system",
		h.Customer.CustomerName,
		h.Customer.CustomerType,
	)
}
