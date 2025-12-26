package services

import (
	"bytes"
	"car_service/dto/request"
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
	// Skip if email service URL is not configured
	if e.baseURL == "" {
		return fmt.Errorf("email service URL not configured")
	}

	// Skip if customer email is not provided
	if req.ToEmail == "" {
		return nil // Don't error, just skip silently
	}

	endpoint := fmt.Sprintf("%s/email-service/send-shipping-status", e.baseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Add authorization header if provided
	if authHeader != "" {
		httpReq.Header.Set("Authorization", authHeader)
	}

	resp, err := e.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send email request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("email service returned status: %d", resp.StatusCode)
	}

	return nil
}

// SendPurchaseStatusEmail sends an email notification when purchase status changes
func (e *EmailService) SendPurchaseStatusEmail(req request.PurchaseStatusEmailRequest, authHeader string) error {
	// Skip if email service URL is not configured
	if e.baseURL == "" {
		return fmt.Errorf("email service URL not configured")
	}

	// Skip if customer email is not provided
	if req.ToEmail == "" {
		return nil // Don't error, just skip silently
	}

	endpoint := fmt.Sprintf("%s/email-service/send-purchasing-status", e.baseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Add authorization header if provided
	if authHeader != "" {
		httpReq.Header.Set("Authorization", authHeader)
	}

	resp, err := e.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send email request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("email service returned status: %d", resp.StatusCode)
	}

	return nil
}
