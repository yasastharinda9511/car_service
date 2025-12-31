package notificationHandlers

import (
	"car_service/dto/request"
	"car_service/entity"
	"fmt"
)

// CustomerDeletedNotificationHandler handles building customer deletion notification payloads
type CustomerDeletedNotificationHandler struct {
	Customer *entity.Customer
	UserID   string
}

// NewCustomerDeletedNotificationHandler creates a new customer deleted notification handler
func NewCustomerDeletedNotificationHandler(
	customer *entity.Customer,
	userID string,
) NotificationHandler {
	return &CustomerDeletedNotificationHandler{
		Customer: customer,
		UserID:   userID,
	}
}

// BuildNotificationRequest constructs the notification request for customer deletion
func (h *CustomerDeletedNotificationHandler) BuildNotificationRequest() *request.NotificationRequest {
	// Build the payload with customer information
	payload := map[string]interface{}{
		"customer_id":   h.Customer.ID,
		"customer_name": h.Customer.CustomerName,
		"customer_type": h.Customer.CustomerType,
		"message":       h.buildMessage(),
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
		"event":   "customer_deleted",
	}

	return &request.NotificationRequest{
		NotificationType: h.GetNotificationType(),
		Source:           "car-service",
		Payload:          payload,
		Priority:         "high",
		ReferenceID:      fmt.Sprintf("CUST-%d", h.Customer.ID),
		Metadata:         metadata,
	}
}

// GetNotificationType returns the notification type
func (h *CustomerDeletedNotificationHandler) GetNotificationType() string {
	return "customer_deleted"
}

// buildMessage creates a human-readable message for the notification
func (h *CustomerDeletedNotificationHandler) buildMessage() string {
	return fmt.Sprintf(
		"Customer '%s' (%s) has been deleted from the system",
		h.Customer.CustomerName,
		h.Customer.CustomerType,
	)
}
