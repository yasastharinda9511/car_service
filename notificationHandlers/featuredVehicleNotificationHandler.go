package notificationHandlers

import (
	"car_service/dto/request"
	"car_service/entity"
	"fmt"
)

// FeaturedVehicleNotificationHandler handles building vehicle featured/unfeatured notification payloads
type FeaturedVehicleNotificationHandler struct {
	Vehicle    *entity.Vehicle
	IsFeatured bool
	UserID     string
}

// NewFeaturedVehicleNotificationHandler creates a new featured vehicle notification handler
func NewFeaturedVehicleNotificationHandler(
	vehicle *entity.Vehicle,
	isFeatured bool,
	userID string,
) NotificationHandler {
	return &FeaturedVehicleNotificationHandler{
		Vehicle:    vehicle,
		IsFeatured: isFeatured,
		UserID:     userID,
	}
}

// BuildNotificationRequest constructs the notification request for vehicle featured status change
func (h *FeaturedVehicleNotificationHandler) BuildNotificationRequest() *request.NotificationRequest {
	// Build the payload with vehicle information
	payload := map[string]interface{}{
		"vehicle_id":   h.Vehicle.ID,
		"vehicle_code": h.Vehicle.Code,
		"message":      h.buildMessage(),
		"make":         h.Vehicle.Make,
		"model":        h.Vehicle.Model,
		"year":         h.Vehicle.YearOfManufacture,
		"chassis_id":   h.Vehicle.ChassisID,
		"color":        h.Vehicle.Color,
		"mileage":      h.Vehicle.MileageKm,
		"condition":    h.Vehicle.ConditionStatus,
		"is_featured":  h.IsFeatured,
	}

	// Build metadata
	metadata := map[string]interface{}{
		"user_id":     h.UserID,
		"service":     "car-service",
		"event":       "vehicle_featured_status_changed",
		"is_featured": h.IsFeatured,
	}

	return &request.NotificationRequest{
		NotificationType: h.GetNotificationType(),
		Source:           "car-service",
		Payload:          payload,
		Priority:         h.determinePriority(),
		ReferenceID:      fmt.Sprintf("VEH-%d", h.Vehicle.ID),
		Metadata:         metadata,
	}
}

// GetNotificationType returns the notification type
func (h *FeaturedVehicleNotificationHandler) GetNotificationType() string {
	return "vehicle_featured_status_changed"
}

// buildMessage creates a human-readable message for the notification
func (h *FeaturedVehicleNotificationHandler) buildMessage() string {
	action := "unfeatured"
	if h.IsFeatured {
		action = "featured"
	}

	return fmt.Sprintf(
		"Vehicle %s (%s %s %d) has been %s",
		h.Vehicle.Code,
		h.Vehicle.Make,
		h.Vehicle.Model,
		h.Vehicle.YearOfManufacture,
		action,
	)
}

// determinePriority sets the priority based on featured status
func (h *FeaturedVehicleNotificationHandler) determinePriority() string {
	if h.IsFeatured {
		return "high"
	}
	return "normal"
}
