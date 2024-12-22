package stream_recorder

import (
	"testing"

	"github.com/ethanhosier/clips/ffmpeg"
)

func TestFileStreamRecorder(t *testing.T) {
	ffmpegClient := ffmpeg.NewFfmpegClient()
	fileStreamRecorder := NewFileStreamRecorder(ffmpegClient, "kc-test")

	fileUrl := "/home/ethanh/Desktop/go/clips/stream_recorder/kc-10-mins.mp4"
	segmentTimeSecs := 120

	clipsCh, doneCh, errorCh := fileStreamRecorder.Record(fileUrl, "", segmentTimeSecs)

	for {
		select {
		case clip := <-clipsCh:
			t.Logf("Clip: %s", clip)
		case <-doneCh:
			return
		case err := <-errorCh:
			t.Fatalf("Error: %v", err)
		}
	}
}
