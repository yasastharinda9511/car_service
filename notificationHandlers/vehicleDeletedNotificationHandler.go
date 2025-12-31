package notificationHandlers

import (
	"car_service/dto/request"
	"car_service/entity"
	"fmt"
)

// VehicleDeletedNotificationHandler handles building vehicle deletion notification payloads
type VehicleDeletedNotificationHandler struct {
	Vehicle *entity.Vehicle
	UserID  string
}

// NewVehicleDeletedNotificationHandler creates a new vehicle deleted notification handler
func NewVehicleDeletedNotificationHandler(
	vehicle *entity.Vehicle,
	userID string,
) NotificationHandler {
	return &VehicleDeletedNotificationHandler{
		Vehicle: vehicle,
		UserID:  userID,
	}
}

// BuildNotificationRequest constructs the notification request for vehicle deletion
func (h *VehicleDeletedNotificationHandler) BuildNotificationRequest() *request.NotificationRequest {
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
	}

	// Build metadata
	metadata := map[string]interface{}{
		"user_id": h.UserID,
		"service": "car-service",
		"event":   "vehicle_deleted",
	}

	return &request.NotificationRequest{
		NotificationType: h.GetNotificationType(),
		Source:           "car-service",
		Payload:          payload,
		Priority:         "high",
		ReferenceID:      fmt.Sprintf("VEH-%d", h.Vehicle.ID),
		Metadata:         metadata,
	}
}

// GetNotificationType returns the notification type
func (h *VehicleDeletedNotificationHandler) GetNotificationType() string {
	return "vehicle_deleted"
}

// buildMessage creates a human-readable message for the notification
func (h *VehicleDeletedNotificationHandler) buildMessage() string {
	return fmt.Sprintf(
		"Vehicle %s (%s %s %d) has been deleted from the system",
		h.Vehicle.Code,
		h.Vehicle.Make,
		h.Vehicle.Model,
		h.Vehicle.YearOfManufacture,
	)
}
