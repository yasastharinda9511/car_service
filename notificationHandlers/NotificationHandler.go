package notificationHandlers

import "car_service/dto/request"

// NotificationHandler defines the interface for building notification payloads
type NotificationHandler interface {
	// BuildNotificationRequest constructs and returns a complete notification request
	BuildNotificationRequest() *request.NotificationRequest

	// GetNotificationType returns the type of notification this handler manages
	GetNotificationType() string
}
