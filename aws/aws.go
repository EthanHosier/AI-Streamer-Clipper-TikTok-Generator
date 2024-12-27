package aws

import (
	"context"
	"fmt"
	"os"

	go_aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AwsClient struct {
	s3Client         *s3.Client
	cloudfrontDomain string
}

// NewAwsClient creates a new AWS client instance
func NewAwsClient(ctx context.Context, cloudfrontDomain string) (*AwsClient, error) {
	// Load AWS configuration with explicit region
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"), // Explicitly set the region
	)
	if err != nil {
		return nil, err
	}

	return &AwsClient{
		s3Client:         s3.NewFromConfig(cfg),
		cloudfrontDomain: cloudfrontDomain,
	}, nil
}

// UploadVideo uploads a video file to S3 and returns the CloudFront URL
func (a *AwsClient) UploadVideo(filePath, filePathToStore string) (string, error) {
	// Open the video file
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Upload to S3
	_, err = a.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: go_aws.String("stream-clips-bucket"),
		Key:    go_aws.String(filePathToStore),
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	// Return the CloudFront URL for the uploaded file
	return fmt.Sprintf("https://%s/%s", a.cloudfrontDomain, filePathToStore), nil
}
