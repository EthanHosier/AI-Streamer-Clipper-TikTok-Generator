package ffmpeg

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type FfmpegExtension string

const (
	MP4 FfmpegExtension = "mp4"
)

func isValidTime(time string) bool {
	re := regexp.MustCompile(`^(?:[0-1]\d|2[0-3]):[0-5]\d:[0-5]\d$`)
	return re.MatchString(time)
}

func (ff *FfmpegClient) VideoDuration(inputFile string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", inputFile)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, err
	}
	return duration, nil
}

func timeStringToSeconds(timeStr string) (int, error) {
	re := regexp.MustCompile(`^(?:(\d+):)?(\d{1,2}):(\d{2})$`)
	matches := re.FindStringSubmatch(timeStr)
	if len(matches) != 4 {
		return 0, fmt.Errorf("invalid time format")
	}

	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.Atoi(matches[3])

	return hours*3600 + minutes*60 + seconds, nil
}

func secondsToTimeString(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	seconds = seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func outputPathFor(inputFile string, extension FfmpegExtension) string {
	bytes := make([]byte, 8)
	rand.Read(bytes)

	baseName := filepath.Base(inputFile)
	nameOnly := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	return fmt.Sprintf("/ffmpeg/temp/%s-%s.%s", nameOnly, hex.EncodeToString(bytes), string(extension))
}

func RandomOutputPathFor(extension FfmpegExtension, basePath string, params ...string) string {
	bytes := make([]byte, 8)
	rand.Read(bytes)

	fileName := ""
	for i, param := range params {
		fileName += fmt.Sprintf("%s", param)
		if i < len(params)-1 {
			fileName += "-"
		}
	}

	strippedPath := strings.TrimSuffix(basePath, "/")

	return fmt.Sprintf("%s/%s-%s.%s", strippedPath, fileName, hex.EncodeToString(bytes), string(extension))
}
