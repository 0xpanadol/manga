package uploader

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioUploader handles file uploads to a MinIO server.
type MinioUploader struct {
	client     *minio.Client
	bucketName string
	endpoint   string
	useSSL     bool
}

// NewMinioUploader creates and initializes a new MinioUploader.
func NewMinioUploader(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*MinioUploader, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	uploader := &MinioUploader{
		client:     client,
		bucketName: bucketName,
		endpoint:   endpoint,
		useSSL:     useSSL,
	}

	// Ensure the bucket exists.
	err = uploader.ensureBucket(context.Background())
	if err != nil {
		return nil, fmt.Errorf("bucket initialization failed: %w", err)
	}

	return uploader, nil
}

// ensureBucket creates the bucket if it doesn't already exist.
func (u *MinioUploader) ensureBucket(ctx context.Context) error {
	exists, err := u.client.BucketExists(ctx, u.bucketName)
	if err != nil {
		return err
	}
	if !exists {
		return u.client.MakeBucket(ctx, u.bucketName, minio.MakeBucketOptions{})
	}
	return nil
}

// UploadFile uploads a file to MinIO and returns its public URL.
func (u *MinioUploader) UploadFile(ctx context.Context, file io.Reader, fileSize int64, originalFilename string) (string, error) {
	// Generate a unique object name to prevent collisions
	// Format: <uuid>.<original_extension>
	ext := filepath.Ext(originalFilename)
	objectName := uuid.New().String() + ext

	// Upload the file
	_, err := u.client.PutObject(ctx, u.bucketName, objectName, file, fileSize, minio.PutObjectOptions{
		// Set content type based on extension if needed, or let MinIO detect
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to minio: %w", err)
	}

	// Construct the public URL
	scheme := "http"
	if u.useSSL {
		scheme = "https"
	}
	// Note: For production S3, the URL format would be different.
	// This format works for MinIO's local setup.
	url := fmt.Sprintf("%s://%s/%s/%s", scheme, u.endpoint, u.bucketName, objectName)

	return url, nil
}
