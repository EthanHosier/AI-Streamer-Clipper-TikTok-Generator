package clipper

import (
	"errors"

	"github.com/ethanhosier/clips/captions"
)

type ClipperClipResult struct {
	outputFilePaths []string
}

type ClipperClipsConfig struct {
	id       string
	ClipType ClipType

	InputVideoPath string
	InputAudioPath string

	BottomVideoPath string

	CaptionsFilePath string
	CaptionsPosition ClipperCaptionsPosition
	CaptionsColor    captions.CaptionsColor
	CaptionsType     captions.CaptionsType

	ShouldSubClip     bool
	SubClipLengthSecs int
}

type ClipperCaptionsPosition string

const (
	ClipperCaptionsPositionCenter     ClipperCaptionsPosition = "center"
	ClipperCaptionsPositionUnderneath ClipperCaptionsPosition = "underneath"
)

type ClipType string

const (
	ClipTypeSplitScreen                 ClipType = "split-screen"
	ClipTypeSingleClipBlurredBackground ClipType = "single-clip-background-blurred"
)

// ClipperClipsConfigBuilder provides a builder pattern for ClipperClipsConfig
type ClipperClipsConfigBuilder struct {
	config ClipperClipsConfig
}

// NewClipperClipsConfigBuilder creates a new builder instance
func NewClipperClipsConfigBuilder() *ClipperClipsConfigBuilder {
	return &ClipperClipsConfigBuilder{}
}

func NewSplitScreenClipperClipsConfigBuilder() *ClipperClipsConfigBuilder {
	return &ClipperClipsConfigBuilder{
		config: ClipperClipsConfig{
			ClipType: ClipTypeSplitScreen,
		},
	}
}

func NewSingleClipBlurredBackgroundClipperClipsConfigBuilder() *ClipperClipsConfigBuilder {
	return &ClipperClipsConfigBuilder{
		config: ClipperClipsConfig{
			ClipType: ClipTypeSingleClipBlurredBackground,
		},
	}
}

// WithInputVideoPath sets the input video path
func (b *ClipperClipsConfigBuilder) WithInputVideoPath(path string) *ClipperClipsConfigBuilder {
	b.config.InputVideoPath = path
	return b
}

// WithInputAudioPath sets the input audio path
func (b *ClipperClipsConfigBuilder) WithInputAudioPath(path string) *ClipperClipsConfigBuilder {
	b.config.InputAudioPath = path
	return b
}

// WithBottomVideoPath sets the bottom video path
func (b *ClipperClipsConfigBuilder) WithBottomVideoPath(path string) *ClipperClipsConfigBuilder {
	b.config.BottomVideoPath = path
	return b
}

// WithCaptionsFilePath sets the captions file path
func (b *ClipperClipsConfigBuilder) WithCaptionsFilePath(path string) *ClipperClipsConfigBuilder {
	b.config.CaptionsFilePath = path
	return b
}

// WithCaptionsPosition sets the captions position
func (b *ClipperClipsConfigBuilder) WithCaptionsPosition(position ClipperCaptionsPosition) *ClipperClipsConfigBuilder {
	b.config.CaptionsPosition = position
	return b
}

// WithCaptionsColor sets the captions color
func (b *ClipperClipsConfigBuilder) WithCaptionsColor(color captions.CaptionsColor) *ClipperClipsConfigBuilder {
	b.config.CaptionsColor = color
	return b
}

// WithCaptionsType sets the captions type
func (b *ClipperClipsConfigBuilder) WithCaptionsType(captionsType captions.CaptionsType) *ClipperClipsConfigBuilder {
	b.config.CaptionsType = captionsType
	return b
}

// WithShouldSubClip sets whether to subclip the video
func (b *ClipperClipsConfigBuilder) WithShouldSubClip(shouldSubClip bool) *ClipperClipsConfigBuilder {
	b.config.ShouldSubClip = shouldSubClip
	return b
}

// WithSubClipLengthSecs sets the subclip length in seconds
func (b *ClipperClipsConfigBuilder) WithSubClipLengthSecs(lengthSecs int) *ClipperClipsConfigBuilder {
	b.config.SubClipLengthSecs = lengthSecs
	return b
}

// Build creates the final ClipperClipsConfig
func (b *ClipperClipsConfigBuilder) Build() *ClipperClipsConfig {
	if b.config.ClipType == ClipTypeSplitScreen {
		if err := b.validateSplitScreenConfig(); err != nil {
			panic(err)
		}
	}

	if b.config.ClipType == ClipTypeSingleClipBlurredBackground {
		if err := b.validateSingleClipBlurredBackgroundConfig(); err != nil {
			panic(err)
		}
	}

	return &b.config
}

func (b *ClipperClipsConfigBuilder) validateSplitScreenConfig() error {
	if b.config.BottomVideoPath == "" {
		return errors.New("bottom video path is required for split screen clips")
	}

	if b.config.CaptionsFilePath == "" {
		return errors.New("captions file path is required for split screen clips")
	}

	if b.config.CaptionsPosition == "" {
		return errors.New("captions position is required for split screen clips")
	}

	if b.config.CaptionsColor == "" {
		return errors.New("captions color is required for split screen clips")
	}

	if b.config.CaptionsType == "" {
		return errors.New("captions type is required for split screen clips")
	}

	if b.config.InputVideoPath == "" {
		return errors.New("input video path is required for split screen clips")
	}

	if b.config.InputAudioPath == "" {
		return errors.New("input audio path is required for split screen clips")
	}

	return nil
}

func (b *ClipperClipsConfigBuilder) validateSingleClipBlurredBackgroundConfig() error {
	if b.config.InputVideoPath == "" {
		return errors.New("input video path is required for single clip blurred background clips")
	}

	if b.config.InputAudioPath == "" {
		return errors.New("input audio path is required for single clip blurred background clips")
	}

	return nil
}
