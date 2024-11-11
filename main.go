package main

import (
	"fmt"
	"log"

	"github.com/ethanhosier/clips/captions"
	"github.com/ethanhosier/clips/youtube"
	"github.com/joho/godotenv"
)

const id = "5GZeLzMe8NE"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	yt := youtube.NewYoutubeClient()
	vid, err := yt.VideoForId(id)
	if err != nil {
		fmt.Printf("error getting video: %v\n", err)
		return
	}

	// fmt.Printf("vid: %+v\n", vid)

	c := captions.NewCaptionsClient()
	cs, err := c.CaptionsFrom(vid.CaptionTrackURL, captions.CaptionsASS)
	if err != nil {
		fmt.Printf("error getting captions: %v\n", err)
		return
	}

	fmt.Printf("cs: %s\n", cs)
}
