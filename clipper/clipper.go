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

// TODO: Make this more generic? (to make use of the same function for all clip types)
func (c *Clipper) ClipFromConfig(config *ClipperClipsConfig) (*ClipperClipResult, error) {
	var clipPath string
	var err error

	switch config.ClipType {
	case ClipTypeSplitScreen:
		clipPath, err = c.clipSplitScreen(config)
	case ClipTypeSingleClipBlurredBackground:
		clipPath, err = c.clipSingleClipBlurredBackground(config)
	default:
		return nil, fmt.Errorf("unknown clip type: %v", config.ClipType)
	}

	if err != nil {
		return nil, fmt.Errorf("error clipping: %v", err)
	}

	if !config.ShouldSubClip {
		return &ClipperClipResult{
			outputFilePaths: []string{clipPath},
		}, nil
	}

	outputPaths, err := c.ffmpegClient.SplitVideo(clipPath, config.SubClipLengthSecs)
	if err != nil {
		return nil, fmt.Errorf("error splitting video: %v", err)
	}
	return &ClipperClipResult{
		outputFilePaths: outputPaths,
	}, nil
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
	twoVidsConfig := NewSplitScreenClipperClipsConfigBuilder()
	twoVidsConfig.WithInputVideoPath(videoYtFile)
	twoVidsConfig.WithBottomVideoPath(slimeVideoPath)
	twoVidsConfig.WithInputAudioPath(audioYtFile)
	twoVidsConfig.WithCaptionsFilePath(captionsFilePath)
	twoVidsConfig.WithShouldSubClip(true)
	twoVidsConfig.WithSubClipLengthSecs(120)

	res, err := c.ClipFromConfig(twoVidsConfig.Build())
	if err != nil {
		return nil, fmt.Errorf("error clipping: %v", err)
	}

	return res.outputFilePaths, nil
}
