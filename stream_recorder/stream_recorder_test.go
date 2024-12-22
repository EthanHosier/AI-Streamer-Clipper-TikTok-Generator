package stream_recorder

import (
	"testing"
)

func TestStreamRecorder(t *testing.T) {
	recorder := NewStreamlinkRecorder("angryginge13")
	clipsCh, doneCh, errorCh := recorder.Record("https://www.twitch.tv/angryginge13", "/tmp/angryginge13", 3)

	select {
	case clip := <-clipsCh:
		t.Logf("Clip: %s", clip)
	case err := <-errorCh:
		t.Fatalf("Error: %v", err)
	case <-doneCh:
		t.Log("Done")
	}
}
