package stream_watcher

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/ethanhosier/clips/ffmpeg"
	"github.com/ethanhosier/clips/gemini"
	"github.com/ethanhosier/clips/supabase"
	"github.com/joho/godotenv"
)

const (
	vidContext = "The streamer is using the gingerbread skin and is located near Retail Row. They're actively participating in a match with two other teammates, as shown by the team display in the top left. A quest titled \"Outlast Players\" has already been completed, giving the player XP."
	last20secs = "The streamer has just helped kill another player with his sniper rifle."
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	os.Exit(m.Run())
}

func TestStreamWatcherHandleSummariseClip(t *testing.T) {

	geminiClient, err := gemini.NewGeminiClient(context.Background(), os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		t.Fatalf("Error creating Gemini client: %v", err)
	}

	streamWatcher := NewStreamWatcher(nil, nil, geminiClient, nil, 1)

	clipSummary, err := streamWatcher.handleSummariseClip("/home/ethanh/Desktop/go/clips/clips/angryginge13/output001.mp4", vidContext, last20secs, 0.0)
	if err != nil {
		t.Fatalf("Error summarising clip: %v", err)
	}

	t.Logf("Clip summary: %+v", *clipSummary)
}

func TestStreamWatcherHandleWatchClip(t *testing.T) {
	geminiClient, err := gemini.NewGeminiClient(context.Background(), os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		t.Fatalf("Error creating Gemini client: %v", err)
	}

	supabaseClient := supabase.NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))

	streamWatcher := NewStreamWatcher(nil, supabaseClient, geminiClient, nil, 3)

	clipSummary, err := streamWatcher.handleWatchClipAndStoreSummary("/home/ethanh/Desktop/go/clips/clips/angryginge13/output001.mp4", vidContext, last20secs, 0.0)
	if err != nil {
		t.Fatalf("Error watching clip: %v", err)
	}

	t.Logf("Clip summary: %+v", *clipSummary)
}

func TestStreamWatcherHandleWatchClipWithOffset(t *testing.T) {
	geminiClient, err := gemini.NewGeminiClient(context.Background(), os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		t.Fatalf("Error creating Gemini client: %v", err)
	}

	supabaseClient := supabase.NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
	ffmpegClient := ffmpeg.NewFfmpegClient()

	streamWatcher := NewStreamWatcher(nil, supabaseClient, geminiClient, ffmpegClient, 3)

	clipSummary, err := streamWatcher.handleWatchClipAndStoreSummary("/home/ethanh/Desktop/go/clips/clips/angryginge13/output001.mp4", vidContext, last20secs, 32.5)
	if err != nil {
		t.Fatalf("Error watching clip: %v", err)
	}

	t.Logf("Clip summary: %+v", *clipSummary)
}
