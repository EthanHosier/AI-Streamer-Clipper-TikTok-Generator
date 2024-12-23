package openai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type OpenaiHandler interface{}

type OpenaiClient struct {
	client *openai.Client
}

func NewOpenaiClient(apiKey string) *OpenaiClient {
	client := openai.NewClient(apiKey)

	return &OpenaiClient{client: client}
}

func (oc *OpenaiClient) CreateChatCompletion(ctx context.Context, prompt string) (*string, error) {
	resp, err := oc.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("error creating chat completion: %w", err)
	}

	return &resp.Choices[0].Message.Content, nil
}

func (oc *OpenaiClient) CreateChatCompletionWithResponseFormatForO1Mini(ctx context.Context, o1MiniPrompt string, secondPrompt string, responseFormat interface{}) (*string, error) {
	resp, err := oc.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.O1Mini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: o1MiniPrompt,
				},
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("error creating chat completion: %w", err)
	}

	fmt.Printf("\n\nO1 mini output: \"%s\"\n\n", resp.Choices[0].Message.Content)

	newPrompt := fmt.Sprintf("Here is some O1Mini output: `%s`\n Now, from this output:`%s`", resp.Choices[0].Message.Content, secondPrompt)

	return oc.CreateChatCompletionWithResponseFormat(ctx, newPrompt, responseFormat)
}

func (oc *OpenaiClient) CreateChatCompletionWithResponseFormat(ctx context.Context, prompt string, responseFormat interface{}) (*string, error) {
	schema, err := jsonschema.GenerateSchemaForType(responseFormat)
	if err != nil {
		return nil, err
	}

	resp, err := oc.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:   "response_format",
					Schema: schema,
					Strict: true,
				},
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("error creating chat completion: %w", err)
	}

	return &resp.Choices[0].Message.Content, nil
}

func (oc *OpenaiClient) CreateTranscription(ctx context.Context, audioPath string, format OpenaiAudioResponseFormat) (*OpenaiAudioResponse, error) {
	resp, err := oc.client.CreateTranscription(ctx, openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: audioPath,
		Format:   openai.AudioResponseFormat(format),
	})

	if err != nil {
		return nil, err
	}

	return &OpenaiAudioResponse{
		Task:     resp.Task,
		Language: resp.Language,
		Duration: resp.Duration,
		Segments: resp.Segments,
		Words:    resp.Words,
		Text:     resp.Text,
	}, nil
}
