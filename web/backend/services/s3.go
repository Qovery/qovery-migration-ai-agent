package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

// UploadZipToS3 uploads a zip file to an S3 bucket and returns the object key
func UploadZipToS3(filename, bucketName, region, accessKeyID, secretAccessKey string) (string, error) {
	// Create a new AWS session with provided credentials
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}

	// Create an S3 service client
	svc := s3.New(sess)

	// Open the zip file
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Generate a unique object key
	timestamp := time.Now().Format("20060102-150405")
	uniqueID := uuid.New().String()
	baseFilename := filepath.Base(filename)
	objectKey := fmt.Sprintf("uploads/%s-%s-%s", timestamp, uniqueID, baseFilename)

	// Upload the file to S3
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %v", err)
	}

	fmt.Printf("Successfully uploaded %s to %s with key %s\n", filename, bucketName, objectKey)
	return objectKey, nil
}
