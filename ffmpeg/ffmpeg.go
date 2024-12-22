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
	ClipVideo(inputFile, startTime, duration string) (string, error)
	VideoDuration(inputFile string) (float64, error)
	SplitVideo(inputFile string, segmentTime int) ([]string, error)
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

func (ff *FfmpegClient) ClipVideoRandomSecs(inputFile string, clipLength int) (string, error) {
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

	return ff.ClipVideo(inputFile, randomStartTime, clipDuration)
}

func (ff *FfmpegClient) ClipVideo(inputFile string, startTime, duration string) (string, error) {
	if !isValidTime(startTime) {
		return "", fmt.Errorf("invalid start time: %s", startTime)
	}

	if !isValidTime(duration) {
		return "", fmt.Errorf("invalid duration: %s", duration)
	}

	outputPath := outputPathFor(inputFile, MP4)
	fmt.Println("outputPath: ", outputPath)

	cmd := exec.Command("ffmpeg", "-ss", startTime, "-i", inputFile, "-t", duration, "-c", "copy", outputPath)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return outputPath, nil
}

func (ff *FfmpegClient) SplitVideo(inputPath string, segmentTime int) ([]string, error) {
	// Construct the output pattern for segmented files
	outputPattern := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + "_%03d" + filepath.Ext(inputPath)

	// Convert segment time to string
	segmentTimeStr := strconv.Itoa(segmentTime)

	// FFmpeg command
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-c", "copy",
		"-map", "0",
		"-segment_time", segmentTimeStr,
		"-f", "segment",
		"-reset_timestamps", "1",
		outputPattern,
	)

	// Run the command and capture any errors
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error running ffmpeg: %w", err)
	}

	// Generate the list of output file paths based on the expected naming pattern
	var outputPaths []string
	base := strings.TrimSuffix(inputPath, filepath.Ext(inputPath))
	ext := filepath.Ext(inputPath)

	// Check for a reasonable number of expected output files
	for i := 0; i < 1000; i++ { // Assuming no more than 1000 segments
		outputFile := fmt.Sprintf("%s_%03d%s", base, i, ext)

		// Break early if the output file doesn't exist
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
