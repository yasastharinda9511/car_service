package services

import (
	"bytes"
	"car_service/logger"
	"car_service/notificationHandlers"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type NotificationService struct {
	baseURL string
	client  *http.Client
}

func NewNotificationService(baseURL string) *NotificationService {
	return &NotificationService{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendNotification sends a generic notification request to the notification service
func (s *NotificationService) SendNotification(handler notificationHandlers.NotificationHandler, authHeader string) error {

	var req = handler.BuildNotificationRequest()

	logger.WithFields(map[string]interface{}{
		"notification_type": req.NotificationType,
		"source":            req.Source,
		"priority":          req.Priority,
		"reference_id":      req.ReferenceID,
	}).Info("Sending generic notification")

	// Skip if notification service URL is not configured
	if s.baseURL == "" {
		logger.Warn("Notification service URL not configured, skipping notification")
		return fmt.Errorf("notification service URL not configured")
	}

	// Set default source if not provided
	if req.Source == "" {
		req.Source = "car-service"
	}

	// Set default priority if not provided
	if req.Priority == "" {
		req.Priority = "normal"
	}

	endpoint := fmt.Sprintf("%s/notifications", s.baseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to marshal notification request")
		return fmt.Errorf("failed to marshal notification request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to create HTTP request for notification")
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Add authorization header if provided
	if authHeader != "" {
		httpReq.Header.Set("Authorization", authHeader)
	}

	logger.WithFields(map[string]interface{}{
		"endpoint":          endpoint,
		"notification_type": req.NotificationType,
	}).Debug("Sending HTTP request to notification service")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"endpoint":          endpoint,
			"notification_type": req.NotificationType,
			"error":             err.Error(),
		}).Error("Failed to send notification request")
		return fmt.Errorf("failed to send notification request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		logger.WithFields(map[string]interface{}{
			"endpoint":          endpoint,
			"notification_type": req.NotificationType,
			"status_code":       resp.StatusCode,
		}).Error("Notification service returned non-success status code")
		return fmt.Errorf("notification service returned status: %d", resp.StatusCode)
	}

	logger.WithFields(map[string]interface{}{
		"notification_type": req.NotificationType,
		"reference_id":      req.ReferenceID,
	}).Info("Notification sent successfully")

	return nil
}
