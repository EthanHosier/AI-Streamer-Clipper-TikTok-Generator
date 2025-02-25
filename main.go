package main

import (
	"context"
	"fmt"

	"github.com/ethanhosier/clips/config"
	"github.com/ethanhosier/clips/stream_clipper_bot"
	"github.com/joho/godotenv"

	"github.com/ethanhosier/clips/stream_recorder"
	"github.com/ethanhosier/clips/stream_watcher"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		panic(fmt.Sprintf("Error loading .env file: %v", err))
	}

	var (
		conf         = config.MustNewDefaultConfig()
		streamerName = "dannyaarons"
		streamerUrl  = "https://www.twitch.tv/dannyaarons"
		streamID     = 3

		streamRecorder   = stream_recorder.NewStreamlinkRecorder(streamerName)
		streamWatcher    = stream_watcher.NewStreamWatcher(streamRecorder, conf.SupabaseClient, conf.GeminiClient, conf.OpenaiClient, conf.FfmpegClient, streamID)
		streamClipperBot = stream_clipper_bot.NewStreamClipperBot(streamWatcher)
	)

	panic(streamClipperBot.Start(context.Background(), streamerUrl, streamerName))
}
