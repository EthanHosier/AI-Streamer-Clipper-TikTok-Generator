package stream_clipper_bot

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/ethanhosier/clips/config"
	"github.com/ethanhosier/clips/stream_recorder"
	"github.com/ethanhosier/clips/stream_watcher"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	os.Exit(m.Run())
}

func TestStreamClipperBot(t *testing.T) {
	var (
		conf         = config.MustNewDefaultConfig()
		streamerName = "dannyaarons"
		streamerUrl  = "https://www.twitch.tv/dannyaarons"
		streamID     = 3

		streamRecorder   = stream_recorder.NewStreamlinkRecorder(streamerName)
		streamWatcher    = stream_watcher.NewStreamWatcher(streamRecorder, conf.SupabaseClient, conf.GeminiClient, conf.OpenaiClient, conf.FfmpegClient, streamID)
		streamClipperBot = NewStreamClipperBot(streamWatcher)
	)

	t.Fatal(streamClipperBot.Start(context.Background(), streamerUrl, streamerName))
}
