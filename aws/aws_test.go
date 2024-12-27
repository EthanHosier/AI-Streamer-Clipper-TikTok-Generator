package aws

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestUploadVideo(t *testing.T) {
	// Skip if running in CI or if explicitly requested
	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("Skipping integration test")
	}

	// Use the actual video file
	videoPath := "/home/ethanh/Desktop/go/clips/input.mp4"
	if _, err := os.Stat(videoPath); err != nil {
		t.Fatalf("Video file not found at %s: %v", videoPath, err)
	}

	// Initialize AWS client
	cloudfrontDomain := "d11afdhum95bkz.cloudfront.net"
	client, err := NewAwsClient(context.Background(), cloudfrontDomain)
	if err != nil {
		t.Fatalf("Failed to create AWS client: %v", err)
	}

	// Test uploading the file
	filePathToStore := "test/integration_test.mp4"
	url, err := client.UploadVideo(videoPath, filePathToStore)
	if err != nil {
		t.Fatalf("Failed to upload video: %v", err)
	}

	// Log the CloudFront URL
	t.Logf("Video uploaded successfully. CloudFront URL: %s", url)

	// Verify the returned URL
	expectedPrefix := "https://" + cloudfrontDomain
	if !strings.HasPrefix(url, expectedPrefix) {
		t.Errorf("Expected URL to start with %s, got %s", expectedPrefix, url)
	}

	expectedSuffix := filePathToStore
	if !strings.HasSuffix(url, expectedSuffix) {
		t.Errorf("Expected URL to end with %s, got %s", expectedSuffix, url)
	}
}
