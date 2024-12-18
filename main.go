package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ethanhosier/clips/gemini"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	geminiClient, err := gemini.NewGeminiClient(context.Background(), os.Getenv("GEMINI_API_KEY"), "clips-439820-video-uploads")
	if err != nil {
		log.Fatalf("Error creating Gemini client: %v", err)
	}

	resp, err := geminiClient.GetChatCompletionWithVideo(context.Background(), "Give a detailed, specific analysis of the video", "clips/angryginge13/output002.mp4")
	if err != nil {
		log.Fatalf("Error getting chat completion: %v", err)
	}

	fmt.Println(*resp)
}
