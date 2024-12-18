package main

import (
	"fmt"
	"log"

	"github.com/ethanhosier/clips/captions"
	"github.com/ethanhosier/clips/clipper"
	"github.com/ethanhosier/clips/ffmpeg"
	"github.com/ethanhosier/clips/openai"
	"github.com/ethanhosier/clips/youtube"
	"github.com/joho/godotenv"
)

const id = "vp5sSqyZ5Go"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	clipper := clipper.NewClipper(
		openai.NewOpenaiClient(),
		ffmpeg.NewFfmpegClient(),
		captions.NewCaptionsClient(),
		youtube.NewYoutubeClient(),
	)

	_, err = clipper.ClipEntireYtVideo(id, captions.CaptionsSingleWord)
	if err != nil {
		fmt.Printf(err.Error())
	}

}
