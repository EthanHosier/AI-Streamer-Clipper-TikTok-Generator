package config

import (
	"context"
	"fmt"
	"os"

	"github.com/ethanhosier/clips/ffmpeg"
	"github.com/ethanhosier/clips/gemini"
	"github.com/ethanhosier/clips/openai"
	"github.com/ethanhosier/clips/supabase"
)

type Config struct {
	FfmpegClient   ffmpeg.FfmpegHandler
	GeminiClient   *gemini.GeminiClient
	OpenaiClient   *openai.OpenaiClient
	SupabaseClient *supabase.Supabase
}

func MustNewDefaultConfig() *Config {

	geminiClient, err := gemini.NewGeminiClient(context.Background(), os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		panic(fmt.Sprintf("Error creating Gemini client: %v", err))
	}

	openaiClient := openai.NewOpenaiClient(os.Getenv("OPENAI_KEY"))
	supabaseClient := supabase.NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
	ffmpegClient := ffmpeg.NewFfmpegClient()

	return &Config{
		FfmpegClient:   ffmpegClient,
		GeminiClient:   geminiClient,
		OpenaiClient:   openaiClient,
		SupabaseClient: supabaseClient,
	}
}
