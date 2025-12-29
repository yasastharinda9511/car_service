package notificationHandlers

import (
	"car_service/dto/request"
	"car_service/entity"
	"fmt"
)

// ShippingStatusNotificationHandler handles building shipping status notification payloads
type ShippingStatusNotificationHandler struct {
	Vehicle   *entity.Vehicle
	Customer  *entity.Customer
	OldStatus string
	NewStatus string
	UserID    string
}

// NewShippingStatusNotificationHandler creates a new shipping status notification handler
func NewShippingStatusNotificationHandler(
	vehicle *entity.Vehicle,
	customer *entity.Customer,
	oldStatus string,
	newStatus string,
	userID string,
) NotificationHandler {
	return &ShippingStatusNotificationHandler{
		Vehicle:   vehicle,
		Customer:  customer,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		UserID:    userID,
	}
}

// BuildNotificationRequest constructs the notification request for shipping status updates
func (h *ShippingStatusNotificationHandler) BuildNotificationRequest() *request.NotificationRequest {
	// Build the payload with shipping status information
	payload := map[string]interface{}{
		"vehicle_code": h.Vehicle.Code,
		"old_status":   h.OldStatus,
		"new_status":   h.NewStatus,
		"message":      h.buildMessage(),
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

	// Build metadata
	metadata := map[string]interface{}{
		"user_id":     h.UserID,
		"service":     "car-service",
		"event":       "shipping_status_update",
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
func (h *ShippingStatusNotificationHandler) GetNotificationType() string {
	return "shipping_status"
}

// buildMessage creates a human-readable message for the notification
func (h *ShippingStatusNotificationHandler) buildMessage() string {
	return fmt.Sprintf(
		"Shipping status for vehicle %s changed from %s to %s",
		h.Vehicle.Code,
		h.OldStatus,
		h.NewStatus,
	)
}

// determinePriority determines notification priority based on shipping status
func (h *ShippingStatusNotificationHandler) determinePriority() string {
	switch h.NewStatus {
	case "SHIPPED", "ARRIVED", "CLEARED":
		return "high" // Important milestones
	case "DELIVERED":
		return "urgent" // Final delivery is urgent
	case "IN_TRANSIT":
		return "normal" // Regular update
	case "ORDERED":
		return "normal" // Initial status
	default:
		return "normal"
	}
}
