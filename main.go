package main

import (
	"fmt"
	"log"
	"os"

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

	fmt.Println(vid.CaptionTrackURL)

	c := captions.NewCaptionsClient()
	cs, err := c.CaptionsFrom(vid.CaptionTrackURL, captions.CaptionsHormozi)
	if err != nil {
		fmt.Printf("error getting captions: %v\n", err)
		return
	}

	err = os.WriteFile("output.ass", []byte(cs), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Printf("cs: %s\n", cs)
}
