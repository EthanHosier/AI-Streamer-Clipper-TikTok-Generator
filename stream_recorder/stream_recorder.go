package stream_recorder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type StreamRecorder struct {
}

func NewStreamRecorder() *StreamRecorder {
	return &StreamRecorder{}
}

func (s *StreamRecorder) Record(streamUrl, outputDir string) (chan string, chan struct{}, chan error) {
	clipsCh := make(chan string)
	doneCh := make(chan struct{})
	errorCh := make(chan error)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		errorCh <- fmt.Errorf("failed to create output directory: %v", err)
		return nil, nil, nil
	}

	go s.recordStream(streamUrl, outputDir, clipsCh, doneCh, errorCh)

	return clipsCh, doneCh, errorCh
}

func (s *StreamRecorder) recordStream(streamUrl, outputDir string, clipsCh chan string, doneCh chan struct{}, errorCh chan error) {
	defer close(clipsCh)
	defer close(doneCh)
	defer close(errorCh)

	streamlinkCmd := exec.Command("streamlink", "--stdout", streamUrl, "best")
	ffmpegCmd := exec.Command("ffmpeg",
		"-i", "pipe:0",
		"-f", "segment",
		"-segment_time", "12",
		"-reset_timestamps", "1",
		"-c", "copy",
		filepath.Join(outputDir, "output%03d.mp4"),
	)

	stdout, err := streamlinkCmd.StdoutPipe()
	if err != nil {
		errorCh <- fmt.Errorf("failed to set up stdout pipe: %v", err)
		return
	}
	ffmpegCmd.Stdin = stdout

	// Start streamlink
	if err := streamlinkCmd.Start(); err != nil {
		errorCh <- fmt.Errorf("failed to start streamlink: %v", err)
		return
	}

	// Start ffmpeg
	if err := ffmpegCmd.Start(); err != nil {
		errorCh <- fmt.Errorf("failed to start ffmpeg: %v", err)
		return
	}

	// Monitor output directory for completed files
	go func() {
		processedFiles := make(map[string]bool)
		currentFile := ""

		for {
			files, err := os.ReadDir(outputDir)
			if err != nil {
				errorCh <- fmt.Errorf("error reading output directory: %v", err)
				return
			}

			for _, file := range files {
				if filepath.Ext(file.Name()) == ".mp4" {
					fullPath := filepath.Join(outputDir, file.Name())

					// Skip if we've already processed this file
					if processedFiles[fullPath] {
						continue
					}

					// If this is a new current file, update tracking
					if fullPath > currentFile {
						// Previous file is complete, send it if it exists
						if currentFile != "" && !processedFiles[currentFile] {
							processedFiles[currentFile] = true
							clipsCh <- currentFile
						}
						currentFile = fullPath
					}
				}
			}

			time.Sleep(time.Second)
		}
	}()

	// Wait for commands to complete
	errStreamlink := streamlinkCmd.Wait()
	if errStreamlink != nil {
		errorCh <- fmt.Errorf("streamlink exited with error: %v", errStreamlink)
	}

	errFFmpeg := ffmpegCmd.Wait()
	if errFFmpeg != nil {
		errorCh <- fmt.Errorf("ffmpeg exited with error: %v", errFFmpeg)
	}

	// Signal completion
	doneCh <- struct{}{}
}
