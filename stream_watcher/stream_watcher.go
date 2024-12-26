package stream_watcher

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math"

	"github.com/ethanhosier/clips/ffmpeg"
	"github.com/ethanhosier/clips/gemini"
	"github.com/ethanhosier/clips/openai"
	"github.com/ethanhosier/clips/stream_recorder"
	"github.com/ethanhosier/clips/supabase"
)

const (
	segmentTime     = 120
	recordedVidsDir = "recorded-vids"
	defaultContext  = "[This is the first clip of the stream, so no context is available.]"
	defaultLast20s  = "[This is the first clip of the stream, so no last 20 seconds context is available.]"

	bufferStartSecs = 0
	bufferEndSecs   = 0 // TODO: MAKE THESE CONFIGURABLE
)

type StreamWatcher struct {
	streamRecorder stream_recorder.StreamRecorder
	supabaseClient *supabase.Supabase
	geminiClient   *gemini.GeminiClient
	openaiClient   *openai.OpenaiClient
	ffmpegClient   ffmpeg.FfmpegHandler
	streamID       int
}

func NewStreamWatcher(streamRecorder stream_recorder.StreamRecorder, supabaseClient *supabase.Supabase, geminiClient *gemini.GeminiClient, openaiClient *openai.OpenaiClient, ffmpegClient ffmpeg.FfmpegHandler, streamID int) *StreamWatcher {
	return &StreamWatcher{streamRecorder: streamRecorder, supabaseClient: supabaseClient, geminiClient: geminiClient, openaiClient: openaiClient, ffmpegClient: ffmpegClient, streamID: streamID}
}

func (s *StreamWatcher) Watch(ctx context.Context, streamUrl string, name string) (chan *CreatedClipResult, chan bool, chan error) {
	createdClipsCh := make(chan *CreatedClipResult, 99)
	errorCh := make(chan error)
	doneCh := make(chan bool)

	go func() {
		err := s.watchLoop(ctx, streamUrl, name, createdClipsCh)
		if err != nil {
			errorCh <- err
		}
		doneCh <- true
	}()

	return createdClipsCh, doneCh, errorCh
}

func (s *StreamWatcher) watchLoop(ctx context.Context, streamUrl string, name string, createdClipsCh chan *CreatedClipResult) error {
	outputDir := fmt.Sprintf("%s/%d", recordedVidsDir, s.streamID)
	clipsCh, doneCh, errorCh := s.streamRecorder.Record(streamUrl, outputDir, segmentTime)

	vidPositionSecs := 0.0
	vidContext := defaultContext
	last20secs := defaultLast20s

	receivedDone := false

	allVideoFiles := []string{}

	clipWindowStartSecs := 0
	// so that when want to add buffer to clip, we can use some of the pending clip
	var pendingClip string

	createdClips := []CreatedClipResult{}

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errorCh:
			slog.Error("Error while processing stream", "error", err)
			return err
		case <-doneCh:
			receivedDone = true
			// Don't exit immediately, wait for remaining clips
			continue

		case clip, ok := <-clipsCh:
			if !ok { // Clips channel closed
				if pendingClip != "" {
					clipSummary, newStreamPositionSecs, err := s.processClip(pendingClip, vidContext, last20secs, vidPositionSecs, name)
					if err != nil {
						return err
					}

					vidPositionSecs = newStreamPositionSecs
					vidContext = clipSummary.UpdatedContext
					last20secs = clipSummary.Last20Secs
					pendingClip = ""

					newClips, cws, err := s.createClips(allVideoFiles, vidContext, int(clipWindowStartSecs), name)
					if err != nil {
						return err
					}

					createdClips = append(createdClips, newClips...)

					for _, c := range newClips {
						createdClipsCh <- &c
					}

					clipWindowStartSecs = cws

					if len(newClips) == 0 {
						log.Printf("No new clips found, exiting")
					} else {
						log.Printf("New clips found: %+v", newClips)
					}
				}

				// Channel is closed and no more clips
				if receivedDone {
					slog.Info("All clips processed, exiting")
					log.Println("Created clips: ")
					for _, c := range createdClips {
						log.Printf("Url: %s\n", c.Url)
						log.Printf("FoundClip: %+v\n", c.FoundClip)
						log.Printf("BufferStartSecs: %d\n", c.BufferStartSecs)
						log.Printf("BufferEndSecs: %d\n\n", c.BufferEndSecs)
					}
					return nil
				}
				continue
			}
			allVideoFiles = append(allVideoFiles, clip)

			if pendingClip == "" {
				slog.Info("Setting pending clip", "clip", clip)
				pendingClip = clip
				continue
			}

			slog.Info("Processing pending clip", "clip", pendingClip)
			clipSummary, newStreamPositionSecs, err := s.processClip(pendingClip, vidContext, last20secs, vidPositionSecs, name)
			if err != nil {
				return err
			}

			vidPositionSecs = newStreamPositionSecs
			vidContext = clipSummary.UpdatedContext
			last20secs = clipSummary.Last20Secs

			pendingClip = clip

			newClips, cws, err := s.createClips(allVideoFiles, vidContext, int(clipWindowStartSecs), name)
			if err != nil {
				return err
			}

			createdClips = append(createdClips, newClips...)

			for _, c := range newClips {
				createdClipsCh <- &c
			}

			clipWindowStartSecs = cws

			if len(newClips) == 0 {
				log.Printf("No new clips found, exiting")
			} else {
				log.Printf("New clips found: %+v", newClips)
			}
		}
	}
}

func (s *StreamWatcher) createClips(videoFiles []string, vidContext string, startWindowSecs int, name string) ([]CreatedClipResult, int, error) {
	slog.Info("Creating clips", "startWindowSecs", startWindowSecs, "videoFiles", videoFiles)
	clips, err := s.findClips(startWindowSecs, vidContext, name)
	if err != nil {
		return nil, 0, err
	}

	if len(clips) == 0 {
		return []CreatedClipResult{}, startWindowSecs, nil
	}

	var createdClips []CreatedClipResult

	for _, clip := range clips {
		slog.Info("Creating clip", "clip start", clip.StartSecs, "clip end", clip.EndSecs, "clip start buffer", bufferStartSecs, "clip end buffer", bufferEndSecs, "clip caption", clip.Caption)
		actualClip, err := s.getActualClipFrom(&clip, videoFiles, bufferStartSecs, bufferEndSecs)
		if err != nil {
			return nil, 0, err
		}

		createdClips = append(createdClips, CreatedClipResult{
			Url:             actualClip,
			FoundClip:       &clip,
			BufferStartSecs: bufferStartSecs,
			BufferEndSecs:   bufferEndSecs,
		})
	}

	maxEndTime := startWindowSecs
	for _, c := range clips {
		maxEndTime = int(math.Max(float64(maxEndTime), float64(int(c.EndSecs))))
	}

	return createdClips, maxEndTime, nil
}

func (s *StreamWatcher) processClip(clip string, vidContext string, last20secs string, streamPositionSecs float64, name string) (*ClipSummary, float64, error) {
	clipSummary, err := s.handleWatchClipAndStoreSummary(clip, vidContext, last20secs, streamPositionSecs, name)
	if err != nil {
		return nil, 0, err
	}

	videoDuration, err := s.ffmpegClient.VideoDuration(clip)
	if err != nil {
		return nil, 0, err
	}

	return clipSummary, streamPositionSecs + videoDuration, nil
}

func (s *StreamWatcher) handleWatchClipAndStoreSummary(clip string, vidContext string, last20secs string, streamPositionSecs float64, name string) (*ClipSummary, error) {
	clipSummary, err := s.handleSummariseClip(clip, vidContext, last20secs, streamPositionSecs, name)
	if err != nil {
		return nil, err
	}

	if err := s.storeClipSummaryParts(clipSummary); err != nil {
		return nil, err
	}

	return clipSummary, nil
}
