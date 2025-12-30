package notificationHandlers

import (
	"car_service/dto/request"
	"car_service/entity"
	"fmt"
)

// SupplierCreatedNotificationHandler handles building supplier creation notification payloads
type SupplierCreatedNotificationHandler struct {
	Supplier *entity.Supplier
	UserID   string
}

// NewSupplierCreatedNotificationHandler creates a new supplier created notification handler
func NewSupplierCreatedNotificationHandler(
	supplier *entity.Supplier,
	userID string,
) NotificationHandler {
	return &SupplierCreatedNotificationHandler{
		Supplier: supplier,
		UserID:   userID,
	}
}

// BuildNotificationRequest constructs the notification request for supplier creation
func (h *SupplierCreatedNotificationHandler) BuildNotificationRequest() *request.NotificationRequest {
	// Build the payload with supplier information
	payload := map[string]interface{}{
		"supplier_id":   h.Supplier.ID,
		"supplier_name": h.Supplier.SupplierName,
		"message":       h.buildMessage(),
		"supplier_type": h.Supplier.SupplierType,
		"country":       h.Supplier.Country,
	}

	// Add optional fields if they exist
	if h.Supplier.SupplierTitle != nil {
		payload["supplier_title"] = *h.Supplier.SupplierTitle
	}
	if h.Supplier.ContactNumber != nil {
		payload["contact_number"] = *h.Supplier.ContactNumber
	}
	if h.Supplier.Email != nil {
		payload["email"] = *h.Supplier.Email
	}
	if h.Supplier.Address != nil {
		payload["address"] = *h.Supplier.Address
	}

	// Build metadata
	metadata := map[string]interface{}{
		"user_id": h.UserID,
		"service": "car-service",
		"event":   "supplier_created",
	}

	return &request.NotificationRequest{
		NotificationType: h.GetNotificationType(),
		Source:           "car-service",
		Payload:          payload,
		Priority:         "normal",
		ReferenceID:      fmt.Sprintf("SUP-%d", h.Supplier.ID),
		Metadata:         metadata,
	}
}

// GetNotificationType returns the notification type
func (h *SupplierCreatedNotificationHandler) GetNotificationType() string {
	return "supplier_created"
}

// buildMessage creates a human-readable message for the notification
func (h *SupplierCreatedNotificationHandler) buildMessage() string {
	return fmt.Sprintf(
		"New supplier '%s' (%s) from %s has been created in the system",
		h.Supplier.SupplierName,
		h.Supplier.SupplierType,
		h.Supplier.Country,
	)
}
