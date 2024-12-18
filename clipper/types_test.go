package clipper

import (
	"testing"

	"github.com/ethanhosier/clips/captions"
	"github.com/stretchr/testify/assert"
)

func TestClipperClipsConfigBuilder(t *testing.T) {
	t.Run("builds complete config with all fields", func(t *testing.T) {
		// Arrange & Act
		config := NewClipperClipsConfigBuilder().
			WithInputVideoPath("input.mp4").
			WithInputAudioPath("audio.mp3").
			WithBottomVideoPath("bottom.mp4").
			WithCaptionsFilePath("captions.srt").
			WithCaptionsPosition(ClipperCaptionsPositionCenter).
			WithCaptionsColor(captions.CaptionsWhite).
			WithCaptionsType(captions.CaptionsSingleWord).
			Build()

		// Assert
		assert.Equal(t, "input.mp4", config.InputVideoPath)
		assert.Equal(t, "audio.mp3", config.InputAudioPath)
		assert.Equal(t, "bottom.mp4", config.BottomVideoPath)
		assert.Equal(t, "captions.srt", config.CaptionsFilePath)
		assert.Equal(t, ClipperCaptionsPositionCenter, config.CaptionsPosition)
		assert.Equal(t, captions.CaptionsWhite, config.CaptionsColor)
		assert.Equal(t, captions.CaptionsSingleWord, config.CaptionsType)
	})

	t.Run("builds config with minimum required fields", func(t *testing.T) {
		// Arrange & Act
		config := NewClipperClipsConfigBuilder().
			WithInputVideoPath("input.mp4").
			WithInputAudioPath("audio.mp3").
			Build()

		// Assert
		assert.Equal(t, "input.mp4", config.InputVideoPath)
		assert.Equal(t, "audio.mp3", config.InputAudioPath)
		assert.Empty(t, config.BottomVideoPath)
		assert.Empty(t, config.CaptionsFilePath)
		assert.Empty(t, config.CaptionsPosition)
		assert.Empty(t, config.CaptionsColor)
		assert.Empty(t, config.CaptionsType)
	})

	t.Run("split screen builder initializes with correct type", func(t *testing.T) {
		// Arrange & Act
		builder := NewSplitScreenClipperClipsConfigBuilder()

		// Assert
		if builder.config.ClipType != ClipTypeSplitScreen {
			t.Errorf("expected ClipType to be %s, got %s", ClipTypeSplitScreen, builder.config.ClipType)
		}
	})

	t.Run("single clip blurred background builder initializes with correct type", func(t *testing.T) {
		// Arrange & Act
		config := NewSingleClipBlurredBackgroundClipperClipsConfigBuilder().config

		// Assert
		assert.Equal(t, ClipTypeSingleClipBlurredBackground, config.ClipType)
	})

	t.Run("split screen builder validates required fields", func(t *testing.T) {
		// Arrange & Act
		builder := NewSplitScreenClipperClipsConfigBuilder()

		// Assert
		assert.PanicsWithError(t, "bottom video path is required for split screen clips", func() {
			builder.Build()
		})

		builder.WithBottomVideoPath("bottom.mp4")
		assert.PanicsWithError(t, "captions file path is required for split screen clips", func() {
			builder.Build()
		})

		builder.WithCaptionsFilePath("captions.srt")
		assert.PanicsWithError(t, "captions position is required for split screen clips", func() {
			builder.Build()
		})

		builder.WithCaptionsPosition(ClipperCaptionsPositionCenter)
		assert.PanicsWithError(t, "captions color is required for split screen clips", func() {
			builder.Build()
		})

		builder.WithCaptionsColor(captions.CaptionsWhite)
		assert.PanicsWithError(t, "captions type is required for split screen clips", func() {
			builder.Build()
		})

		builder.WithCaptionsType(captions.CaptionsSingleWord)
		assert.PanicsWithError(t, "input video path is required for split screen clips", func() {
			builder.Build()
		})

		builder.WithInputVideoPath("input.mp4")
		assert.PanicsWithError(t, "input audio path is required for split screen clips", func() {
			builder.Build()
		})

		builder.WithInputAudioPath("audio.mp3")
		assert.NotPanics(t, func() {
			builder.Build()
		})
	})

	t.Run("single clip blurred background builder validates required fields", func(t *testing.T) {
		// Arrange & Act
		builder := NewSingleClipBlurredBackgroundClipperClipsConfigBuilder()

		// Assert
		assert.PanicsWithError(t, "input video path is required for single clip blurred background clips", func() {
			builder.Build()
		})

		builder.WithInputVideoPath("input.mp4")
		assert.PanicsWithError(t, "input audio path is required for single clip blurred background clips", func() {
			builder.Build()
		})

		builder.WithInputAudioPath("audio.mp3")
		assert.NotPanics(t, func() {
			builder.Build()
		})
	})
}
