package clipper

import (
	"fmt"
	"os"
	"time"
)

func (c *Clipper) clipSplitScreen(config *ClipperClipsConfig) (string, error) {
	editedVideoPath := fmt.Sprintf("clipper/temp/edited-%v.mp4", config.id)
	_, err := c.ffmpegClient.MergeTwoVidsWithCaptions(config.InputVideoPath, config.BottomVideoPath, config.InputAudioPath, editedVideoPath, config.CaptionsFilePath)
	if err != nil {
		return "", fmt.Errorf("error merging video with slime and captions: %v\n", err)
	}
	defer os.Remove(editedVideoPath)

	return editedVideoPath, nil
}

func (c *Clipper) clipSingleClipBlurredBackground(config *ClipperClipsConfig) (string, error) {
	editedVideoPath := fmt.Sprintf("clipper/temp/edited-%v.mp4", config.id)
	_, err := c.ffmpegClient.ClipWithBlurredBackground(config.InputVideoPath, editedVideoPath, config.InputAudioPath)
	if err != nil {
		return "", fmt.Errorf("error clipping with blurred background: %v\n", err)
	}
	defer os.Remove(editedVideoPath)

	return editedVideoPath, nil
}

// Helper function to format duration as HH:mm:ss
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func writeTempCaptionsFile(id string, cs string) (string, error) {
	fileName := fmt.Sprintf("clipper/temp/%v.ass", id)
	return fileName, os.WriteFile(fileName, []byte(cs), 0644)

}

func deleteTempCaptionsFile(filepath string) error {
	return os.Remove(filepath)
}
