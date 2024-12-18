package gemini

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiClient struct {
	client        *genai.Client
	bucket        string
	storageClient *storage.Client
}

func NewGeminiClient(ctx context.Context, apiKey string, bucket string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	return &GeminiClient{
		client:        client,
		bucket:        bucket,
		storageClient: storageClient,
	}, nil
}

func (gc *GeminiClient) GetChatCompletion(ctx context.Context, prompt string) (*string, error) {
	model := gc.client.GenerativeModel("gemini-2.0-flash-exp")

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response candidates received")
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response parts received")
	}

	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("failed to convert to text")
	}

	strText := string(text)

	return &strText, nil
}

func (gc *GeminiClient) GetChatCompletionWithVideo(ctx context.Context, prompt string, videoPath string) (*string, error) {
	// Generate a unique filename (you might want to adjust this logic)
	fileName := fmt.Sprintf("uploads/%d.mp4", time.Now().UnixNano())

	// Create the bucket handle and object handle
	bkt := gc.storageClient.Bucket(gc.bucket)
	obj := bkt.Object(fileName)

	// Create the writer
	writer := obj.NewWriter(ctx)

	// Open the local file
	file, err := os.Open(videoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open video file: %v", err)
	}
	defer file.Close()

	// Copy the file to GCS
	if _, err := io.Copy(writer, file); err != nil {
		return nil, fmt.Errorf("failed to copy file to GCS: %v", err)
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %v", err)
	}

	// Change the GCS URL to a public HTTPS URL
	videoURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", gc.bucket, fileName)

	model := gc.client.GenerativeModel("gemini-2.0-flash-exp")

	// Create prompt parts combining text and video URL
	parts := []genai.Part{
		genai.FileData{URI: videoURL, MIMEType: "video/mp4"},
		genai.Text(prompt),
	}

	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response parts received")
	}

	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("failed to convert to text")
	}

	strText := string(text)

	return &strText, nil
}
