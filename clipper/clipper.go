package clipper

import (
	"fmt"
	"os"
	"time"

	attentionclips "github.com/ethanhosier/clips/attention-clips"
	"github.com/ethanhosier/clips/captions"
	"github.com/ethanhosier/clips/ffmpeg"
	"github.com/ethanhosier/clips/openai"
	"github.com/ethanhosier/clips/youtube"
)

type Clipper struct {
	openaiClient   *openai.OpenaiClient
	ffmpegClient   *ffmpeg.FfmpegClient
	captionsClient *captions.CaptionsClient
	youtubeClient  *youtube.YoutubeClient
}

func NewClipper(openaiClient *openai.OpenaiClient, ffmpegClient *ffmpeg.FfmpegClient, captionsClient *captions.CaptionsClient, youtubeClient *youtube.YoutubeClient) *Clipper {
	return &Clipper{
		openaiClient:   openaiClient,
		ffmpegClient:   ffmpegClient,
		captionsClient: captionsClient,
		youtubeClient:  youtubeClient,
	}
}

func (c *Clipper) ClipEntireYtVideo(id string, captionsType captions.CaptionsType) ([]string, error) {
	vid, err := c.youtubeClient.VideoForId(id)
	if err != nil {
		return nil, fmt.Errorf("error getting vid: %v", err)
	}

	// Get captions for the video
	cs, err := c.captionsClient.CaptionsFrom(vid.CaptionTrackURL, captionsType)
	if err != nil {
		return nil, fmt.Errorf("error getting captions: %v\n", err)
	}

	captionsFilePath, err := writeTempCaptionsFile(id, cs)
	if err != nil {
		return nil, fmt.Errorf("error writing temp captions file: %v\n", err)
	}
	defer os.Remove(captionsFilePath)

	audioYtFile := fmt.Sprintf("clipper/temp/audio-%v.mp4", id)
	videoYtFile := fmt.Sprintf("clipper/temp/video-%v.mp4", id)

	if err = c.youtubeClient.DownloadVideoAndAudio(id, videoYtFile, audioYtFile); err != nil {
		return nil, fmt.Errorf("error downloading video and audio: %v", err)
	}
	audioYtFile += ".mp4"
	defer os.Remove(audioYtFile)
	defer os.Remove(videoYtFile)

	// Generate a "slime video" for merging
	slimeVideoPath, err := c.ffmpegClient.ClipVideoRandomSecs(string(attentionclips.Slime), int(vid.Duration.Seconds()))
	if err != nil {
		return nil, fmt.Errorf("error generating slime video: %v\n", err)
	}
	defer os.Remove(slimeVideoPath)

	// Merge the main video, slime video, and captions
	editedVideoPath := fmt.Sprintf("clipper/temp/edited-%v.mp4", id)
	_, err = c.ffmpegClient.MergeTwoVidsWithCaptions(videoYtFile, slimeVideoPath, audioYtFile, editedVideoPath, captionsFilePath)
	if err != nil {
		return nil, fmt.Errorf("error merging video with slime and captions: %v\n", err)
	}
	defer os.Remove(editedVideoPath)

	outputPaths, err := c.ffmpegClient.SplitVideo(editedVideoPath, 120)
	return outputPaths, err
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
