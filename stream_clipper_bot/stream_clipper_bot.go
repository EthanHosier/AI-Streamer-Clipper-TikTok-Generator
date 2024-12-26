package stream_clipper_bot

import (
	"context"
	"fmt"

	"github.com/ethanhosier/clips/stream_watcher"
)

type StreamClipperBot struct {
	streamWatcher *stream_watcher.StreamWatcher
}

func NewStreamClipperBot(streamWatcher *stream_watcher.StreamWatcher) *StreamClipperBot {
	return &StreamClipperBot{streamWatcher: streamWatcher}
}

func (s *StreamClipperBot) Start(ctx context.Context, streamURL, streamerName string) error {
	createdClipsCh, doneCh, errCh := s.streamWatcher.Watch(ctx, streamURL, streamerName)

	for {
		select {
		case err := <-errCh:
			return err
		case createdClip := <-createdClipsCh:
			// Drain all clips before considering `doneCh`
			s.processCreatedClip(createdClip)
			for {
				select {
				case createdClip = <-createdClipsCh:
					s.processCreatedClip(createdClip)
				default:
					// Exit draining if no more clips are available
					break
				}
			}
		case <-doneCh:
			// Ensure all clips are processed before exiting
			for {
				select {
				case createdClip := <-createdClipsCh:
					s.processCreatedClip(createdClip)
				default:
					// Exit once all clips are handled
					return nil
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *StreamClipperBot) processCreatedClip(createdClip *stream_watcher.CreatedClipResult) error {
	fmt.Printf("Created clip: %+v\n", createdClip)
	return nil
}
