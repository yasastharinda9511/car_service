package services

import (
	"bytes"
	"car_service/dto/request"
	"car_service/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type EmailService struct {
	baseURL string
	client  *http.Client
}

func NewEmailService() *EmailService {
	return &EmailService{
		baseURL: os.Getenv("EMAIL_SERVICE_URL"),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendShippingStatusEmail sends an email notification when shipping status changes
func (e *EmailService) SendShippingStatusEmail(req request.ShippingStatusEmailRequest, authHeader string) error {
	logger.WithFields(map[string]interface{}{
		"to_email": req.ToEmail,
		"type":     "shipping_status",
	}).Info("Attempting to send shipping status email")

	// Skip if email service URL is not configured
	if e.baseURL == "" {
		logger.Warn("Email service URL not configured, skipping email send")
		return fmt.Errorf("email service URL not configured")
	}

	// Skip if customer email is not provided
	if req.ToEmail == "" {
		logger.Debug("No recipient email provided, skipping email send")
		return nil // Don't error, just skip silently
	}

	endpoint := fmt.Sprintf("%s/email-service/send-shipping-status", e.baseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"to_email": req.ToEmail,
			"error":    err.Error(),
		}).Error("Failed to marshal shipping status email request")
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"endpoint": endpoint,
			"error":    err.Error(),
		}).Error("Failed to create HTTP request for shipping status email")
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Add authorization header if provided
	if authHeader != "" {
		httpReq.Header.Set("Authorization", authHeader)
	}

	logger.WithFields(map[string]interface{}{
		"endpoint": endpoint,
		"to_email": req.ToEmail,
	}).Debug("Sending HTTP request to email service")

	resp, err := e.client.Do(httpReq)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"endpoint": endpoint,
			"to_email": req.ToEmail,
			"error":    err.Error(),
		}).Error("Failed to send shipping status email request")
		return fmt.Errorf("failed to send email request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		logger.WithFields(map[string]interface{}{
			"endpoint":    endpoint,
			"to_email":    req.ToEmail,
			"status_code": resp.StatusCode,
		}).Error("Email service returned non-success status code")
		return fmt.Errorf("email service returned status: %d", resp.StatusCode)
	}

	logger.WithField("to_email", req.ToEmail).Info("Shipping status email sent successfully")
	return nil
}

// SendPurchaseStatusEmail sends an email notification when purchase status changes
func (e *EmailService) SendPurchaseStatusEmail(req request.PurchaseStatusEmailRequest, authHeader string) error {
	logger.WithFields(map[string]interface{}{
		"to_email": req.ToEmail,
		"type":     "purchase_status",
	}).Info("Attempting to send purchase status email")

	// Skip if email service URL is not configured
	if e.baseURL == "" {
		logger.Warn("Email service URL not configured, skipping email send")
		return fmt.Errorf("email service URL not configured")
	}

	// Skip if customer email is not provided
	if req.ToEmail == "" {
		logger.Debug("No recipient email provided, skipping email send")
		return nil // Don't error, just skip silently
	}

	endpoint := fmt.Sprintf("%s/email-service/send-purchasing-status", e.baseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"to_email": req.ToEmail,
			"error":    err.Error(),
		}).Error("Failed to marshal purchase status email request")
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"endpoint": endpoint,
			"error":    err.Error(),
		}).Error("Failed to create HTTP request for purchase status email")
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Add authorization header if provided
	if authHeader != "" {
		httpReq.Header.Set("Authorization", authHeader)
	}

	logger.WithFields(map[string]interface{}{
		"endpoint": endpoint,
		"to_email": req.ToEmail,
	}).Debug("Sending HTTP request to email service")

	resp, err := e.client.Do(httpReq)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"endpoint": endpoint,
			"to_email": req.ToEmail,
			"error":    err.Error(),
		}).Error("Failed to send purchase status email request")
		return fmt.Errorf("failed to send email request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		logger.WithFields(map[string]interface{}{
			"endpoint":    endpoint,
			"to_email":    req.ToEmail,
			"status_code": resp.StatusCode,
		}).Error("Email service returned non-success status code")
		return fmt.Errorf("email service returned status: %d", resp.StatusCode)
	}

	logger.WithField("to_email", req.ToEmail).Info("Purchase status email sent successfully")
	return nil
}
