package notificationHandlers

import (
	"car_service/dto/request"
	"car_service/entity"
	"fmt"
)

// VehicleCreatedNotificationHandler handles building vehicle creation notification payloads
type VehicleCreatedNotificationHandler struct {
	Vehicle *entity.Vehicle
	UserID  string
}

// NewVehicleCreatedNotificationHandler creates a new vehicle created notification handler
func NewVehicleCreatedNotificationHandler(
	vehicle *entity.Vehicle,
	userID string,
) NotificationHandler {
	return &VehicleCreatedNotificationHandler{
		Vehicle: vehicle,
		UserID:  userID,
	}
}

// BuildNotificationRequest constructs the notification request for vehicle creation
func (h *VehicleCreatedNotificationHandler) BuildNotificationRequest() *request.NotificationRequest {
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
		"event":   "vehicle_created",
	}

	return &request.NotificationRequest{
		NotificationType: h.GetNotificationType(),
		Source:           "car-service",
		Payload:          payload,
		Priority:         "normal",
		ReferenceID:      fmt.Sprintf("VEH-%d", h.Vehicle.ID),
		Metadata:         metadata,
	}
}

// GetNotificationType returns the notification type
func (h *VehicleCreatedNotificationHandler) GetNotificationType() string {
	return "vehicle_created"
}

// buildMessage creates a human-readable message for the notification
func (h *VehicleCreatedNotificationHandler) buildMessage() string {
	return fmt.Sprintf(
		"New vehicle %s (%s %s %d) has been created in the system",
		h.Vehicle.Code,
		h.Vehicle.Make,
		h.Vehicle.Model,
		h.Vehicle.YearOfManufacture,
	)
}
