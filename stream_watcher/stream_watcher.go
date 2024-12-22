package stream_watcher

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ethanhosier/clips/ffmpeg"
	"github.com/ethanhosier/clips/gemini"
	"github.com/ethanhosier/clips/stream_recorder"
	"github.com/ethanhosier/clips/supabase"
)

const (
	segmentTime     = 120
	recordedVidsDir = "recorded-vids"
	defaultContext  = "[This is the first clip of the stream, so no context is available.]"
	defaultLast20s  = "[This is the first clip of the stream, so no last 20 seconds context is available.]"
)

type StreamWatcher struct {
	streamRecorder stream_recorder.StreamRecorder
	supabaseClient *supabase.Supabase
	geminiClient   *gemini.GeminiClient
	ffmpegClient   ffmpeg.FfmpegHandler
	streamID       int
}

func NewStreamWatcher(streamRecorder stream_recorder.StreamRecorder, supabaseClient *supabase.Supabase, geminiClient *gemini.GeminiClient, ffmpegClient ffmpeg.FfmpegHandler, streamID int) *StreamWatcher {
	return &StreamWatcher{streamRecorder: streamRecorder, supabaseClient: supabaseClient, geminiClient: geminiClient, ffmpegClient: ffmpegClient, streamID: streamID}
}

func (s *StreamWatcher) Watch(ctx context.Context, streamUrl string) error {
	outputDir := fmt.Sprintf("%s/%s", recordedVidsDir, streamUrl)
	clipsCh, doneCh, errorCh := s.streamRecorder.Record(streamUrl, outputDir, segmentTime)

	vidPositionSecs := 0.0
	vidContext := defaultContext
	last20secs := defaultLast20s

	receivedDone := false
	for {
		select {
		case <-ctx.Done():
			return nil
		case clip, ok := <-clipsCh:
			if !ok {
				slog.Info("clips Channel closed")
				// Channel is closed and no more clips
				if receivedDone {
					slog.Info("All clips processed, exiting")
					return nil
				}
				continue
			}
			slog.Info("Watching clip", "clip", clip)
			clipSummary, err := s.handleWatchClipAndStoreSummary(clip, vidContext, last20secs, vidPositionSecs)
			if err != nil {
				return err
			}
			vidContext = clipSummary.UpdatedContext
			last20secs = clipSummary.Last20Secs

			ffmpegDuration, err := s.ffmpegClient.VideoDuration(clip)
			if err != nil {
				return err
			}
			vidPositionSecs += ffmpegDuration
			slog.Info("Video position", "position", vidPositionSecs)

		case err := <-errorCh:
			slog.Error("Error while processing stream", "error", err)
			return err
		case <-doneCh:
			receivedDone = true
			// Don't exit immediately, wait for remaining clips

		default:
			fmt.Println("yeahhh")
		}
	}
}

func (s *StreamWatcher) handleWatchClipAndStoreSummary(clip string, vidContext string, last20secs string, streamPositionSecs float64) (*ClipSummary, error) {
	clipSummary, err := s.handleSummariseClip(clip, vidContext, last20secs, streamPositionSecs)
	if err != nil {
		return nil, err
	}

	if err := s.storeClipSummaryParts(clipSummary); err != nil {
		return nil, err
	}

	return clipSummary, nil
}

func (s *StreamWatcher) checkForViralClip(beforeTime time.Time) error {
	fmt.Println("Checking for viral clip before", beforeTime)
	return nil
}
