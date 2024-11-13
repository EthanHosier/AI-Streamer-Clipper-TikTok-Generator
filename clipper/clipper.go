package clipper

import (
	"fmt"
	"os"

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

func (c *Clipper) ClipEntireYtVideo(id string) ([]string, error) {
	vid, err := c.youtubeClient.VideoForId(id)
	if err != nil {
		return nil, fmt.Errorf("error getting vid: %v", err)
	}

	cs, err := c.captionsClient.CaptionsFrom(vid.CaptionTrackURL, captions.CaptionsHormozi)
	if err != nil {
		return nil, fmt.Errorf("error getting captions: %v\n", err)
	}

	filepath, err := writeTempCaptionsFile(id, cs)
	if err != nil {
		return nil, fmt.Errorf("error writing temp captions file: %v\n", err)
	}
	// defer deleteTempCaptionsFile(fileName)

	v2, err := c.ffmpegClient.ClipVideoRandomSecs(attentionclips.Slime, 40)

	_, err = c.ffmpegClient.MergeTwoVidsWithCaptions("v1.mp4", v2, "outputyeah.mp4", filepath)
	if err != nil {
		return nil, fmt.Errorf("error merging two vids with captions: %v\n", err)
	}

	return nil, err
}

func writeTempCaptionsFile(id string, cs string) (string, error) {
	fileName := fmt.Sprintf("clipper/temp/%v.ass", id)
	return fileName, os.WriteFile(fileName, []byte(cs), 0644)

}

func deleteTempCaptionsFile(filepath string) error {
	return os.Remove(filepath)
}
