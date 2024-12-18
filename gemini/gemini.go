package gemini

import (
	"context"
	"fmt"
	"log"
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
	// Upload the video file to Gemini API
	file, err := gc.uploadFileToGemini(ctx, videoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to upload video: %v", err)
	}
	defer func() {
		// Clean up the uploaded file after processing
		if err := gc.client.DeleteFile(ctx, file.Name); err != nil {
			log.Printf("failed to delete file: %v", err)
		}
	}()

	// Ensure the file is fully processed before proceeding
	for file.State == genai.FileStateProcessing {
		log.Printf("Processing file: %s", file.Name)
		time.Sleep(5 * time.Second)
		if file, err = gc.client.GetFile(ctx, file.Name); err != nil {
			return nil, fmt.Errorf("failed to check file status: %v", err)
		}
	}
	if file.State != genai.FileStateActive {
		return nil, fmt.Errorf("file is not active; state: %s", file.State)
	}

	// Use the uploaded and processed file URI in the prompt
	model := gc.client.GenerativeModel("gemini-2.0-flash-exp")
	resp, err := model.GenerateContent(ctx,
		genai.FileData{URI: file.URI, MIMEType: "video/mp4"},
		genai.Text(prompt),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, fmt.Errorf("no response candidates received")
	}

	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("failed to convert content to text")
	}

	strText := string(text)
	return &strText, nil
}

func (gc *GeminiClient) uploadFileToGemini(ctx context.Context, videoPath string) (*genai.File, error) {
	// Use the Gemini API's UploadFileFromPath method to upload the video
	file, err := gc.client.UploadFileFromPath(ctx, videoPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to Gemini: %v", err)
	}
	return file, nil
}
