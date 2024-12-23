package stream_recorder

import (
	"fmt"
	"log/slog"

	"github.com/ethanhosier/clips/ffmpeg"
)

type FileStreamRecorder struct {
	ffmpegClient ffmpeg.FfmpegHandler
	name         string
}

func NewFileStreamRecorder(ffmpegClient ffmpeg.FfmpegHandler, name string) *FileStreamRecorder {
	return &FileStreamRecorder{ffmpegClient: ffmpegClient, name: name}
}

func (s *FileStreamRecorder) Record(streamUrl, _ string, segmentTime int) (chan string, chan struct{}, chan error) {
	clipsCh := make(chan string)
	doneCh := make(chan struct{})
	errorCh := make(chan error)

	slog.Info("splitting video", "streamUrl", streamUrl, "segmentTime", segmentTime, "outputDir", fmt.Sprintf("tmp/%s/", s.name))
	clips, err := s.ffmpegClient.SplitVideo(streamUrl, segmentTime, fmt.Sprintf("tmp/%s/", s.name))
	if err != nil {
		errorCh <- fmt.Errorf("failed to split video: %v", err)
		return nil, nil, nil
	}

	slog.Info("split video", "clips", clips)

	go func() {
		for _, clip := range clips {
			clipsCh <- clip
		}
		close(clipsCh)
		doneCh <- struct{}{}
	}()

	return clipsCh, doneCh, errorCh
}
