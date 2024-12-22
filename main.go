package main

import (
	"fmt"

	"github.com/ethanhosier/clips/stream_recorder"
)

func main() {
	streamlinkRecorder := stream_recorder.NewStreamlinkRecorder("kc-test")
	clipsCh, doneCh, errorCh := streamlinkRecorder.Record("https://www.twitch.tv/angryginge13", "stream_recorder/tmp/angryginge/", 10)

	for {
		select {
		case clip := <-clipsCh:
			fmt.Println(clip)
		case <-doneCh:
			return
		case err := <-errorCh:
			fmt.Println(err)
		}
	}
}
