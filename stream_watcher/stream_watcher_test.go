package stream_watcher

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/ethanhosier/clips/ffmpeg"
	"github.com/ethanhosier/clips/gemini"
	"github.com/ethanhosier/clips/openai"
	"github.com/ethanhosier/clips/stream_recorder"
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

	streamWatcher := NewStreamWatcher(nil, nil, geminiClient, nil, nil, 1)

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

	streamWatcher := NewStreamWatcher(nil, supabaseClient, geminiClient, nil, nil, 1)

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

	streamWatcher := NewStreamWatcher(nil, supabaseClient, geminiClient, nil, ffmpegClient, 3)

	clipSummary, err := streamWatcher.handleWatchClipAndStoreSummary("/home/ethanh/Desktop/go/clips/clips/angryginge13/output001.mp4", vidContext, last20secs, 32.5)
	if err != nil {
		t.Fatalf("Error watching clip: %v", err)
	}

	t.Logf("Clip summary: %+v", *clipSummary)
}

func TestStreamWatcherWatch(t *testing.T) {
	var (
		fileStreamRecorder = stream_recorder.NewFileStreamRecorder(ffmpeg.NewFfmpegClient())
		supabaseClient     = supabase.NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
		geminiClient, err  = gemini.NewGeminiClient(context.Background(), os.Getenv("GEMINI_API_KEY"))
		ffmpegClient       = ffmpeg.NewFfmpegClient()
		streamID           = 3
	)

	if err != nil {
		t.Fatalf("Error creating Gemini client: %v", err)
	}

	streamWatcher := NewStreamWatcher(fileStreamRecorder, supabaseClient, geminiClient, nil, ffmpegClient, streamID)
	err = streamWatcher.Watch(context.Background(), "/home/ethanh/Desktop/go/clips/stream_watcher/kc-10-mins.mp4")
	if err != nil {
		t.Fatalf("Error watching stream: %v", err)
	}
}

func TestStreamWatcherCheckForViralClip(t *testing.T) {
	openaiClient := openai.NewOpenaiClient()
	supabaseClient := supabase.NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
	streamWatcher := NewStreamWatcher(nil, supabaseClient, nil, openaiClient, nil, 3)

	vidContext := "Kai Cenat dances and jokes around, and becomes ecstatic as he gains new subscribers. He expresses his excitement by stating that he is glowing. Kai states that he missed his fans and that they have a long night ahead. He also mentions he needs to get his live stream habits back. He then has a moment where he gets confused about a detail on his stream before stating that he is going to be serious."

	foundClips, err := streamWatcher.findClips(0, vidContext)
	if err != nil {
		t.Fatalf("Error checking for viral clip: %v", err)
	}

	t.Logf("Found clips: %+v", foundClips)
}

func TestStreamWatcherGetActualClipFrom(t *testing.T) {
	ffmpegClient := ffmpeg.NewFfmpegClient()
	streamWatcher := NewStreamWatcher(nil, nil, nil, nil, ffmpegClient, 3)

	vidFiles := []string{
		"/home/ethanh/Desktop/go/clips/stream_recorder/kc-10-mins_000.mp4",
		"/home/ethanh/Desktop/go/clips/stream_recorder/kc-10-mins_001.mp4",
		"/home/ethanh/Desktop/go/clips/stream_recorder/kc-10-mins_002.mp4",
		"/home/ethanh/Desktop/go/clips/stream_recorder/kc-10-mins_003.mp4",
		"/home/ethanh/Desktop/go/clips/stream_recorder/kc-10-mins_004.mp4",
	}

	clip := &FoundClip{
		StartSecs: 142,
		EndSecs:   160,
	}

	bufferStartSecs := 10
	bufferEndSecs := 10

	c, err := streamWatcher.getActualClipFrom(clip, vidFiles, bufferStartSecs, bufferEndSecs)
	if err != nil {
		t.Fatalf("Error getting actual clip: %v", err)
	}

	t.Logf("Clip: %+v", c)
}
