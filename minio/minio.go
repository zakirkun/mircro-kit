package minio

import (
	"context"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioContext struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
}

func (m *MinioContext) Open() (*minio.Client, error) {
	// Initialize minio client object.
	minioClient, err := minio.New(m.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(m.AccessKeyID, m.SecretAccessKey, ""),
		Secure: m.UseSSL,
	})

	if err != nil {
		return nil, err
	}

	return minioClient, nil
}

func (m *MinioContext) Upload(bucketName, objectName, filePath string) {
	client, _ := m.Open()

	go func() {
		retryCount := 0
		maxRetries := 5
		for {
			err := uploadFile(client, bucketName, objectName, filePath)
			if err == nil {
				log.Printf("File %s successfully uploaded to bucket %s\n", objectName, bucketName)
				break
			}

			if retryCount >= maxRetries {
				log.Printf("Failed to upload %s after %d retries: %v\n", objectName, retryCount, err)
				break
			}

			retryCount++
			backoff := time.Duration(2<<retryCount) * time.Second
			log.Printf("Retrying upload of %s in %s...\n", objectName, backoff)
			time.Sleep(backoff)
		}
	}()
}

func (m *MinioContext) GetFile(bucketName, objectName string, expiry time.Duration) (string, error) {
	client, _ := m.Open()
	return getFileURL(client, bucketName, objectName, expiry)
}

func (m *MinioContext) Delete(bucketName, objectName string) {
	client, _ := m.Open()
	deleteFile(client, bucketName, objectName)
}

func uploadFile(minioClient *minio.Client, bucketName, objectName, filePath string) error {
	// Check if file exists before attempting upload.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return err
	}

	// Make bucket if it doesn't exist.
	err := minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(context.Background(), bucketName)
		if errBucketExists == nil && exists {
			log.Printf("Bucket %s already exists\n", bucketName)
		} else {
			return err
		}
	}

	// Upload the file.
	_, err = minioClient.FPutObject(context.Background(), bucketName, objectName, filePath, minio.PutObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}

func getFileURL(minioClient *minio.Client, bucketName, objectName string, expiry time.Duration) (string, error) {
	// Generate a pre-signed URL for the object.
	reqParams := url.Values{} // Optionally, add custom request parameters here.
	presignedURL, err := minioClient.PresignedGetObject(context.Background(), bucketName, objectName, expiry, reqParams)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

func deleteFile(minioClient *minio.Client, bucketName, objectName string) {
	// Delete the object.
	err := minioClient.RemoveObject(context.Background(), bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		log.Println(err)
	}

	log.Printf("Successfully deleted %s from bucket %s\n", objectName, bucketName)
}
