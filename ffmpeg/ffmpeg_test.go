package ffmpeg

import "testing"

func TestIsValidTime(t *testing.T) {
	testCases := []struct {
		time     string
		expected bool
	}{
		// Valid cases
		{"00:01:57", true},
		{"12:34:56", true},
		{"23:59:59", true},
		{"00:00:00", true},

		// Invalid cases
		{"25:00:00", false},
		{"01:60:00", false},
		{"00:00:60", false},
		{"123:00:00", false},
		{"00:123:00", false},
		{"00:00:123", false},
		{"invalid", false},
		{"00:00", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := isValidTime(tc.time)
		if result != tc.expected {
			t.Errorf("isValidTime(%q) = %v; expected %v", tc.time, result, tc.expected)
		}
	}
}

func TestVideoDuration(t *testing.T) {
	ffmpeg := NewFfmpegClient()
	duration, err := ffmpeg.VideoDuration("/home/ethanh/Desktop/go/clips/clips/angryginge13/output001.mp4")
	if err != nil {
		t.Errorf("VideoDuration() error = %v", err)
	}
	t.Logf("Duration: %v", duration)
}

func TestClipVideo(t *testing.T) {
	ffmpeg := NewFfmpegClient()
	clips, err := ffmpeg.ClipVideo("/home/ethanh/Desktop/go/clips/stream_watcher/kc-10-mins.mp4", "tmp/kc-stream.mp4", "00:02:00", "00:05:00")
	if err != nil {
		t.Errorf("ClipVideo() error = %v", err)
	}
	t.Logf("Clips: %v", clips)
}
