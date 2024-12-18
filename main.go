package main

import (
	"log"

	"github.com/ethanhosier/clips/stream_recorder"
)

const id = "JFh7vQEoqX4"

func main() {
	outputDir := "./clips"
	streamURL := "https://www.twitch.tv/freyzplayz"

	recorder := stream_recorder.NewStreamRecorder()

	clipsCh, doneCh, errorCh := recorder.Record(streamURL, outputDir)

	// Handle events
	go func() {
		for {
			select {
			case filename, ok := <-clipsCh:
				if !ok {
					return
				}
				log.Printf("New clip saved yeahhh: %s", filename)
			case err, ok := <-errorCh:
				if !ok {
					return
				}
				log.Printf("Error: %v", err)
			case <-doneCh:
				log.Println("Stream ended.")
				return
			}
		}
	}()

	<-doneCh
}
