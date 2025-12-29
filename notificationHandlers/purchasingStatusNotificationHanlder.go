package notificationHandlers

import (
	"car_service/dto/request"
	"car_service/entity"
	"fmt"
)

// PurchasingStatusNotificationHandler handles building purchasing status notification payloads
type PurchasingStatusNotificationHandler struct {
	Vehicle         *entity.Vehicle
	Customer        *entity.Customer
	OldStatus       string
	NewStatus       string
	PurchaseDetails map[string]interface{}
	SupplierName    *string
	UserID          string
}

// NewPurchasingStatusNotificationHandler creates a new purchasing status notification handler
func NewPurchasingStatusNotificationHandler(
	vehicle *entity.Vehicle,
	customer *entity.Customer,
	oldStatus string,
	newStatus string,
	supplierName *string,
	userID string,
) NotificationHandler {
	return &PurchasingStatusNotificationHandler{
		Vehicle:      vehicle,
		Customer:     customer,
		OldStatus:    oldStatus,
		NewStatus:    newStatus,
		SupplierName: supplierName,
		UserID:       userID,
	}
}

// BuildNotificationRequest constructs the notification request for purchasing status updates
func (h *PurchasingStatusNotificationHandler) BuildNotificationRequest() *request.NotificationRequest {
	// Build the payload with purchasing status information
	payload := map[string]interface{}{
		"vehicle_code":  h.Vehicle.Code,
		"old_status":    h.OldStatus,
		"new_status":    h.NewStatus,
		"supplier_name": h.SupplierName,
		"message":       h.buildMessage(),
	}

	// Add customer information if available
	if h.Customer != nil && h.Customer.Email != nil && *h.Customer.Email != "" {
		payload["email"] = *h.Customer.Email
		payload["customer_name"] = h.Customer.CustomerName
	}

	// Add vehicle details
	payload["vehicle_details"] = map[string]interface{}{
		"make":       h.Vehicle.Make,
		"model":      h.Vehicle.Model,
		"year":       h.Vehicle.YearOfManufacture,
		"chassis_id": h.Vehicle.ChassisID,
		"color":      h.Vehicle.Color,
		"mileage":    h.Vehicle.MileageKm,
	}

	// Add purchase details if provided
	if h.PurchaseDetails != nil && len(h.PurchaseDetails) > 0 {
		payload["purchase_details"] = h.PurchaseDetails
	}

	// Build metadata
	metadata := map[string]interface{}{
		"user_id":     h.UserID,
		"service":     "car-service",
		"event":       "purchase_status_update",
		"customer_id": h.Customer.ID,
	}

	// Determine priority based on status
	priority := h.determinePriority()

	return &request.NotificationRequest{
		NotificationType: h.GetNotificationType(),
		Source:           "car-service",
		Payload:          payload,
		Priority:         priority,
		ReferenceID:      fmt.Sprintf("VEH-%d", h.Vehicle.ID),
		Metadata:         metadata,
	}
}

// GetNotificationType returns the notification type
func (h *PurchasingStatusNotificationHandler) GetNotificationType() string {
	return "purchase_status"
}

// buildMessage creates a human-readable message for the notification
func (h *PurchasingStatusNotificationHandler) buildMessage() string {
	return fmt.Sprintf(
		"Purchase status for vehicle %s changed from %s to %s",
		h.Vehicle.Code,
		h.OldStatus,
		h.NewStatus,
	)
}

// determinePriority determines notification priority based on purchasing status
func (h *PurchasingStatusNotificationHandler) determinePriority() string {
	switch h.NewStatus {
	case "CONFIRMED", "LC_ISSUED":
		return "high" // Important milestones
	case "PAYMENT_COMPLETED":
		return "urgent" // Payment is urgent
	case "PENDING", "PROCESSING":
		return "normal" // Initial/intermediate statuses
	case "CANCELLED", "REJECTED":
		return "urgent" // Critical status changes
	default:
		return "normal"
	}
}
