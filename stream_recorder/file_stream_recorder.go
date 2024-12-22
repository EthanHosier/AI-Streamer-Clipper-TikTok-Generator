package stream_recorder

import (
	"fmt"

	"github.com/ethanhosier/clips/ffmpeg"
)

type FileStreamRecorder struct {
	ffmpegClient ffmpeg.FfmpegHandler
}

func NewFileStreamRecorder(ffmpegClient ffmpeg.FfmpegHandler) *FileStreamRecorder {
	return &FileStreamRecorder{ffmpegClient: ffmpegClient}
}

func (s *FileStreamRecorder) Record(streamUrl, _ string, segmentTime int) (chan string, chan struct{}, chan error) {
	clipsCh := make(chan string)
	doneCh := make(chan struct{})
	errorCh := make(chan error)

	clips, err := s.ffmpegClient.SplitVideo(streamUrl, segmentTime)
	if err != nil {
		errorCh <- fmt.Errorf("failed to split video: %v", err)
		return nil, nil, nil
	}

	go func() {
		for _, clip := range clips {
			clipsCh <- clip
		}
		close(clipsCh)
		doneCh <- struct{}{}
	}()

	return clipsCh, doneCh, errorCh
}
