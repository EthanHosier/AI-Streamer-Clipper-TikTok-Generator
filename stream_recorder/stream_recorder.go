package stream_recorder

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type StreamRecorder interface {
	Record(streamUrl, outputDir string, segmentTime int) (chan string, chan struct{}, chan error)
}

type StreamlinkRecorder struct {
	name string
}

func NewStreamlinkRecorder(name string) *StreamlinkRecorder {
	return &StreamlinkRecorder{name: name}
}

func (s *StreamlinkRecorder) Record(streamUrl, outputDir string, segmentTime int) (chan string, chan struct{}, chan error) {
	clipsCh := make(chan string)
	doneCh := make(chan struct{})
	errorCh := make(chan error)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		errorCh <- fmt.Errorf("failed to create output directory: %v", err)
		return nil, nil, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-doneCh
		cancel()
	}()

	go s.recordStream(ctx, streamUrl, outputDir, clipsCh, doneCh, errorCh, segmentTime)

	return clipsCh, doneCh, errorCh
}

func (s *StreamlinkRecorder) recordStream(ctx context.Context, streamUrl, outputDir string, clipsCh chan string, doneCh chan struct{}, errorCh chan error, segmentTime int) {
	defer close(clipsCh)
	defer close(doneCh)
	defer close(errorCh)

	streamlinkCmd := exec.CommandContext(ctx, "streamlink", "--stdout", streamUrl, "best")
	ffmpegCmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", "pipe:0",
		"-f", "segment",
		"-segment_time", fmt.Sprintf("%d", segmentTime),
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

	cleanup := func() {
		fmt.Println("Cleaning up")
		if streamlinkCmd.Process != nil {
			streamlinkCmd.Process.Kill()
		}
		if ffmpegCmd.Process != nil {
			ffmpegCmd.Process.Kill()
		}
	}

	defer cleanup()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		os.Interrupt,    // SIGINT (Ctrl+C)
		syscall.SIGTERM, // SIGTERM
		syscall.SIGTSTP, // Ctrl+Z
		syscall.SIGQUIT, // Ctrl+\
		syscall.SIGHUP,  // Terminal closed
	)
	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal: %v\n", sig)
		cleanup()

		// If it's SIGTSTP, we need to exit explicitly since it's a stop signal
		if sig == syscall.SIGTSTP {
			os.Exit(0)
		}

		// For other signals
		os.Exit(1)
	}()

	if err := streamlinkCmd.Start(); err != nil {
		errorCh <- fmt.Errorf("failed to start streamlink: %v", err)
		return
	}

	if err := ffmpegCmd.Start(); err != nil {
		errorCh <- fmt.Errorf("failed to start ffmpeg: %v", err)
		return
	}

	slog.Info("Started recording", "streamUrl", streamUrl, "outputDir", outputDir, "segmentTime", segmentTime)

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

					if processedFiles[fullPath] {
						continue
					}

					if fullPath > currentFile {
						if currentFile != "" && !processedFiles[currentFile] {
							slog.Info("Sending segment clip", "clip", currentFile)
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

	errStreamlink := streamlinkCmd.Wait()
	if errStreamlink != nil {
		errorCh <- fmt.Errorf("streamlink exited with error: %v", errStreamlink)
	}

	errFFmpeg := ffmpegCmd.Wait()
	if errFFmpeg != nil {
		errorCh <- fmt.Errorf("ffmpeg exited with error: %v", errFFmpeg)
	}

	doneCh <- struct{}{}
}
