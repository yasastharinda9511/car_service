package services

import (
	"bytes"
	"car_service/dto/response"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"time"

	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3Service struct {
	client     *s3.Client
	bucketName string
	region     string
}

type UploadResult struct {
	Key      string
	URL      string
	Filename string
	FileSize int64
}

func NewS3Service(bucketName, region string) (*S3Service, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	return &S3Service{
		client:     s3.NewFromConfig(cfg),
		bucketName: bucketName,
		region:     region,
	}, nil
}

// UploadFile uploads a file to S3 and returns the key and URL
func (s *S3Service) UploadFile(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, prefix string) (*UploadResult, error) {
	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)

	// Create S3 key with prefix (e.g., "vehicles/images/uuid_timestamp.jpg")
	key := path.Join(prefix, filename)

	// Upload to S3
	log.Printf("Uploading %s to %s", fileHeader.Filename, key)

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(fileContent),
		ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Generate the public URL
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucketName, s.region, key)

	return &UploadResult{
		Key:      key,
		URL:      url,
		Filename: filename,
		FileSize: fileHeader.Size,
	}, nil
}

// GetPresignedURL generates a presigned URL for downloading a file
func (s *S3Service) GetPresignedURL(ctx context.Context, key string, expirationMinutes int) (*response.PresignedURLResponse, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expirationMinutes) * time.Minute
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	var preSignedURL response.PresignedURLResponse
	preSignedURL.PresignedURL = request.URL
	return &preSignedURL, nil
}

// DeleteFile deletes a file from S3
func (s *S3Service) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// DownloadFile downloads a file from S3 and returns its content
func (s *S3Service) DownloadFile(ctx context.Context, key string) ([]byte, string, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to download from S3: %w", err)
	}
	defer result.Body.Close()

	content, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read S3 object: %w", err)
	}

	contentType := ""
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	return content, contentType, nil
}

// CheckIfFileExists checks if a file exists in S3
func (s *S3Service) CheckIfFileExists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}
