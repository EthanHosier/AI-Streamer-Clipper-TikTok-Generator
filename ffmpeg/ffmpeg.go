package ffmpeg

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FfmpegHandler interface {
	RemoveAudio(inputFile, outputFile string) (string, error)
	ClipVideo(inputFile, outputPath, startTime, duration string) (string, error)
	VideoDuration(inputFile string) (float64, error)
	SplitVideo(inputFile string, segmentTime int, outputFolder string) ([]string, error)
}

type FfmpegClient struct {
}

func NewFfmpegClient() *FfmpegClient {
	return &FfmpegClient{}
}

func (ff *FfmpegClient) RemoveAudio(inputFile, outputFile string) (string, error) {
	cmd := exec.Command("ffmpeg", "-i", inputFile, "-an", "-c:v", "copy", outputFile)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return outputFile, nil
}

func (ff *FfmpegClient) ClipVideoRandomSecs(inputFile string, clipLength int, outputPath string) (string, error) {
	duration, err := ff.VideoDuration(inputFile)
	if err != nil {
		return "", fmt.Errorf("failed to get video duration: %v", err)
	}

	totalDurationSeconds := int(duration) // WHY IS THIS INT?

	if clipLength >= totalDurationSeconds {
		return "", errors.New("clip length must be shorter than video duration")
	}

	rand.Seed(time.Now().UnixNano())
	maxStart := totalDurationSeconds - clipLength
	randomStartSeconds := rand.Intn(maxStart)

	randomStartTime := secondsToTimeString(randomStartSeconds)
	clipDuration := secondsToTimeString(clipLength)

	return ff.ClipVideo(inputFile, outputPath, randomStartTime, clipDuration)
}

func (ff *FfmpegClient) ClipVideo(inputFile, outputPath, startTime, duration string) (string, error) {
	if !isValidTime(startTime) {
		return "", fmt.Errorf("invalid start time: %s", startTime)
	}

	if !isValidTime(duration) {
		return "", fmt.Errorf("invalid duration: %s", duration)
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	cmd := exec.Command("ffmpeg", "-ss", startTime, "-i", inputFile, "-t", duration, "-c", "copy", outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg error: %v\nOutput: %s", err, string(output))
	}
	return outputPath, nil
}

func (ff *FfmpegClient) SplitVideo(inputPath string, segmentTime int, outputFolder string) ([]string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get the base filename without path and extension
	baseFileName := filepath.Base(strings.TrimSuffix(inputPath, filepath.Ext(inputPath)))
	ext := filepath.Ext(inputPath)

	// Construct the output pattern for segmented files
	outputPattern := filepath.Join(outputFolder, baseFileName+"_%03d"+ext)

	// FFmpeg command
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-c", "copy",
		"-map", "0:v:0",
		"-map", "0:a:0",
		"-segment_time", strconv.Itoa(segmentTime),
		"-f", "segment",
		"-reset_timestamps", "1",
		outputPattern,
	)

	// Run the command and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running ffmpeg: %w\nOutput: %s", err, string(output))
	}

	// Generate the list of output file paths
	var outputPaths []string
	for i := 0; i < 1000; i++ { // Assuming no more than 1000 segments
		outputFile := filepath.Join(outputFolder, fmt.Sprintf("%s_%03d%s", baseFileName, i, ext))

		if !fileExists(outputFile) {
			break
		}
		outputPaths = append(outputPaths, outputFile)
	}

	return outputPaths, nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (ff *FfmpegClient) MergeTwoVidsWithCaptions(vid1, vid2, audioFile, output, captions string) (string, error) {
	cmd := fmt.Sprintf(`ffmpeg -i %s -i %s -i %s -filter_complex "\
[0:v]scale=720:-1,scale=720:640:force_original_aspect_ratio=increase,crop=720:640:(iw-720)/2:(ih-640)/2[top]; \
[1:v]scale=720:-1,scale=720:640:force_original_aspect_ratio=increase,crop=720:640:(iw-720)/2:(ih-640)/2[bottom]; \
[top][bottom]vstack=inputs=2[vstacked]; \
[vstacked]subtitles=%s[final]" \
-map "[final]" -map 2:a -r 30 -vsync 2 -c:v libx264 -c:a aac %s`, vid1, vid2, audioFile, captions, output)

	execCmd := exec.Command("bash", "-c", cmd)

	// Redirect stdout and stderr to the console
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	fmt.Println("Running FFmpeg command:")
	fmt.Println(cmd)

	// Run the command
	if err := execCmd.Run(); err != nil {
		return "", fmt.Errorf("error running ffmpeg: %w", err)
	}

	return output, nil
}

func (ff *FfmpegClient) ClipWithBlurredBackground(inputFile, outputFile, audioFile string) (string, error) {
	cmd := fmt.Sprintf(`ffmpeg -i %s -i %s -filter_complex "[0:v]scale=-1:1920,crop=1080:1920,boxblur=luma_radius=30:luma_power=2[blurred]; [0:v]scale=1080:-1[main]; [blurred][main]overlay=(W-w)/2:(H-h)/2[v]" -map "[v]" -map 1:a -c:v libx264 -c:a aac %s`,
		inputFile,
		audioFile,
		outputFile)

	execCmd := exec.Command("bash", "-c", cmd)

	// Redirect stdout and stderr to the console
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		return "", fmt.Errorf("error running ffmpeg: %w", err)
	}

	return outputFile, nil
}
