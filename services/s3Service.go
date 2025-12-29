package services

import (
	"bytes"
	"car_service/logger"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3Service struct {
	client     *s3.Client
	bucketName string
	endpoint   string // Store for URL generation
	region     string // e.g., "nyc3", "sfo3", "ams3"
}

type UploadResult struct {
	Key      string
	URL      string
	Filename string
	FileSize int64
}

// NewS3Service creates a new S3 service configured for Digital Ocean Spaces
func NewS3Service(region string, accessKey string, secretKey string, bucketName string) (*S3Service, error) {
	logger.WithFields(map[string]interface{}{
		"region": region,
		"bucket": bucketName,
	}).Info("Initializing S3 service for Digital Ocean Spaces")

	// Construct the endpoint URL
	// Format: https://<region>.digitaloceanspaces.com
	endpoint := fmt.Sprintf("https://%s.digitaloceanspaces.com", region)

	// Create custom endpoint resolver for Digital Ocean Spaces
	customResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, reg string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: endpoint,
			}, nil
		},
	)

	// Load AWS SDK config with custom settings for Spaces
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
		config.WithRegion("us-east-1"), // Required by SDK but ignored by Spaces
	)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"region": region,
			"bucket": bucketName,
			"error":  err.Error(),
		}).Error("Failed to load S3 config")
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = false // Spaces uses virtual-hosted style URLs
	})

	logger.WithFields(map[string]interface{}{
		"endpoint": endpoint,
		"bucket":   bucketName,
	}).Info("S3 service initialized successfully")

	return &S3Service{
		client:     client,
		bucketName: bucketName,
		endpoint:   endpoint,
		region:     region,
	}, nil
}

// UploadFile uploads a file to Digital Ocean Spaces
func (s *S3Service) UploadFile(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, prefix string) (*UploadResult, error) {
	logger.WithFields(map[string]interface{}{
		"original_filename": fileHeader.Filename,
		"size_bytes":        fileHeader.Size,
		"prefix":            prefix,
	}).Info("Starting file upload to S3")

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"filename": fileHeader.Filename,
			"error":    err.Error(),
		}).Error("Failed to read file content")
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)

	// Create key with prefix (e.g., "vehicles/images/uuid_timestamp.jpg")
	key := path.Join(prefix, filename)

	logger.WithFields(map[string]interface{}{
		"original_filename": fileHeader.Filename,
		"generated_key":     key,
		"bucket":            s.bucketName,
	}).Debug("Uploading file to S3")

	// Upload to Spaces
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(fileContent),
		ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
		ACL:         "private", // Keep files private, use presigned URLs for access
	})
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"filename": fileHeader.Filename,
			"key":      key,
			"bucket":   s.bucketName,
			"error":    err.Error(),
		}).Error("Failed to upload file to S3")
		return nil, fmt.Errorf("failed to upload to Spaces: %w", err)
	}

	// Generate the URL
	// Format: https://<bucket>.<region>.digitaloceanspaces.com/<key>
	// Or with CDN: https://<bucket>.<region>.cdn.digitaloceanspaces.com/<key>
	url := fmt.Sprintf("https://%s.%s.digitaloceanspaces.com/%s", s.bucketName, s.region, key)

	logger.WithFields(map[string]interface{}{
		"original_filename": fileHeader.Filename,
		"key":               key,
		"url":               url,
		"size_bytes":        fileHeader.Size,
	}).Info("File uploaded successfully to S3")

	return &UploadResult{
		Key:      key,
		URL:      url,
		Filename: filename,
		FileSize: fileHeader.Size,
	}, nil
}

// GetPresignedURL generates a presigned URL for downloading a file
func (s *S3Service) GetPresignedURL(ctx context.Context, key string, expirationMinutes int) (*PresignedURLResponse, error) {
	logger.WithFields(map[string]interface{}{
		"key":                key,
		"expiration_minutes": expirationMinutes,
	}).Debug("Generating presigned URL")

	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expirationMinutes) * time.Minute
	})

	if err != nil {
		logger.WithFields(map[string]interface{}{
			"key":   key,
			"error": err.Error(),
		}).Error("Failed to generate presigned URL")
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	logger.WithField("key", key).Info("Presigned URL generated successfully")

	return &PresignedURLResponse{
		PresignedURL: request.URL,
	}, nil
}

// CheckIfFileExists checks if a file exists in the Space
func (s *S3Service) CheckIfFileExists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		// Check if it's a "not found" error
		return false, nil
	}

	return true, nil
}

// DeleteFile deletes a file from the Space
func (s *S3Service) DeleteFile(ctx context.Context, key string) error {
	logger.WithField("key", key).Info("Deleting file from S3")

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		logger.WithFields(map[string]interface{}{
			"key":   key,
			"error": err.Error(),
		}).Error("Failed to delete file from S3")
		return fmt.Errorf("failed to delete file: %w", err)
	}

	logger.WithField("key", key).Info("File deleted successfully from S3")
	return nil
}

// GetCDNURL returns the CDN URL for a file (if CDN is enabled on your Space)
func (s *S3Service) GetCDNURL(key string) string {
	return fmt.Sprintf("https://%s.%s.cdn.digitaloceanspaces.com/%s", s.bucketName, s.region, key)
}

// PresignedURLResponse - adjust this to match your existing response struct
type PresignedURLResponse struct {
	PresignedURL string `json:"presigned_url"`
}
