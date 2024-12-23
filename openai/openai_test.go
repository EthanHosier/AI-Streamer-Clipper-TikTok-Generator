package openai

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	os.Exit(m.Run())
}

func TestOpenaiClient_CreateChatCompletion(t *testing.T) {
	oc := NewOpenaiClient(os.Getenv("OPENAI_API_KEY"))

	prompt := "What is the capital of France?"

	response, err := oc.CreateChatCompletion(context.TODO(), prompt)
	if err != nil {
		t.Fatalf("Error creating chat completion: %v", err)
	}

	t.Logf("Response: %+v", *response)
}

func TestOpenaiClient_CreateChatCompletion_WithResponseFormat(t *testing.T) {
	oc := NewOpenaiClient(os.Getenv("OPENAI_API_KEY"))

	type CapitalResponse struct {
		Capital string   `json:"capital"`
		Colors  []string `json:"colors"`
	}

	prompt := "What is the capital of Germany and what colors in the flag?"
	responseFormat := CapitalResponse{}

	response, err := oc.CreateChatCompletionWithResponseFormat(context.TODO(), prompt, responseFormat)
	if err != nil {
		t.Fatalf("Error creating chat completion: %v", err)
	}

	t.Logf("Response: %+v", *response)
}

func TestOpenaiClient_CreateChatCompletion_WithResponseFormatForO1Mini(t *testing.T) {
	oc := NewOpenaiClient(os.Getenv("OPENAI_KEY"))

	prompt := "What is the capital of France?"

	type CapitalResponse struct {
		Capital string `json:"capital"`
	}

	responseFormat := CapitalResponse{}

	response, err := oc.CreateChatCompletionWithResponseFormat(context.TODO(), prompt, responseFormat)
	if err != nil {
		t.Fatalf("Error creating chat completion: %v", err)
	}

	t.Logf("Response: %+v", *response)
}
