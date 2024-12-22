package main

import (
	"context"
	"log"
	"os"

	"github.com/ethanhosier/clips/ffmpeg"
	"github.com/ethanhosier/clips/gemini"
	"github.com/ethanhosier/clips/openai"
	"github.com/ethanhosier/clips/stream_recorder"
	"github.com/ethanhosier/clips/stream_watcher"
	"github.com/ethanhosier/clips/supabase"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	var (
		streamlinkRecorder = stream_recorder.NewStreamlinkRecorder("angryginge")
		supabaseClient     = supabase.NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
		geminiClient, err  = gemini.NewGeminiClient(context.Background(), os.Getenv("GEMINI_API_KEY"))
		ffmpegClient       = ffmpeg.NewFfmpegClient()
		openaiClient       = openai.NewOpenaiClient(os.Getenv("OPENAI_KEY"))
		streamID           = 3
	)

	if err != nil {
		log.Fatalf("Error creating Gemini client: %v", err)
	}

	streamWatcher := stream_watcher.NewStreamWatcher(streamlinkRecorder, supabaseClient, geminiClient, openaiClient, ffmpegClient, streamID)
	streamWatcher.Watch(context.Background(), "https://www.twitch.tv/angryginge13")

}
