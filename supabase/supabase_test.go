package supabase

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	err := godotenv.Load("../.env")
	if err != nil {
		panic(fmt.Sprintf("Warning: Error loading .env file: %v", err))
	}

	os.Exit(m.Run())
}

func TestGetStreamer(t *testing.T) {
	supabase := NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
	streamer, err := supabase.GetStreamer(1)
	if err != nil {
		t.Fatalf("Error getting streamer: %v", err)
	}

	fmt.Printf("%+v\n", *streamer)
}

func TestCreateStreamer(t *testing.T) {
	supabase := NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
	streamer := &Streamer{
		Name: "Test Streamer",
	}
	id, err := supabase.CreateStreamer(streamer)
	if err != nil {
		t.Fatalf("Error creating streamer: %v", err)
	}
	t.Logf("Streamer created with ID: %d", id)
}

func TestCreateStreamerPlatform(t *testing.T) {
	supabase := NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
	streamerPlatform := &StreamerPlatform{
		StreamerID: 1,
		Platform:   Platform("twitch"),
		Url:        "https://www.twitch.tv/test",
	}
	id, err := supabase.CreateStreamerPlatform(streamerPlatform)
	if err != nil {
		t.Fatalf("Error creating streamer platform: %v", err)
	}
	t.Logf("Streamer platform created with ID: %d", id)
}

func TestGetStreamerPlatforms(t *testing.T) {
	supabase := NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
	streamerPlatforms, err := supabase.GetStreamerPlatforms(1)
	if err != nil {
		t.Fatalf("Error getting streamer platforms: %v", err)
	}
	fmt.Printf("%+v\n", streamerPlatforms)
}

func TestCreateStream(t *testing.T) {
	supabase := NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))

	// Create test data
	stream := &Stream{
		StreamerID: 1,
		Platform:   "twitch",
		Info: map[string]interface{}{
			"url":   "https://www.twitch.tv/test",
			"title": "Test Stream", // Add more required fields if needed
		},
	}

	// Print what we're trying to insert
	t.Logf("Attempting to create stream: %+v", stream)

	id, err := supabase.CreateStream(stream)
	if err != nil {
		t.Fatalf("Error creating stream: %v", err)
	}

	t.Logf("Stream created with ID: %d", id)
}

func TestGetStream(t *testing.T) {
	supabase := NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
	stream, err := supabase.GetStream(1)
	if err != nil {
		t.Fatalf("Error getting stream: %v", err)
	}
	fmt.Printf("%+v\n", stream)
}

func TestGetStreamEvents(t *testing.T) {
	supabase := NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
	streamEvents, err := supabase.GetStreamEvents(0)
	if err != nil {
		t.Fatalf("Error getting stream events: %v", err)
	}
	fmt.Printf("%+v\n", streamEvents)
}

func TestCreateStreamEvent(t *testing.T) {
	supabase := NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
	streamEvent := &StreamEvent{
		StreamID:        2,
		StartSecs:       100,
		EndSecs:         200,
		Description:     "Test Stream Event",
		StreamContextID: 1,
	}
	id, err := supabase.CreateStreamEvent(streamEvent)
	if err != nil {
		t.Fatalf("Error creating stream event: %v", err)
	}
	t.Logf("Stream event created with ID: %d", id)
}

func TestCreateStreamContext(t *testing.T) {
	supabase := NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
	streamContext := &StreamContext{
		StreamID: 2,
		Context:  "Test Stream Context",
	}
	id, err := supabase.CreateStreamContext(streamContext)
	if err != nil {
		t.Fatalf("Error creating stream context: %v", err)
	}
	t.Logf("Stream context created with ID: %d", id)
}

func TestGetStreamContexts(t *testing.T) {
	supabase := NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
	streamContexts, err := supabase.GetStreamContexts(2)
	if err != nil {
		t.Fatalf("Error getting stream contexts: %v", err)
	}
	fmt.Printf("%+v\n", streamContexts)
}

func TestGetStreamEventsAfter(t *testing.T) {
	supabase := NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
	streamEvents, err := supabase.GetStreamEventsAfter(109, 3)
	if err != nil {
		t.Fatalf("Error getting stream events: %v", err)
	}
	fmt.Printf("%+v\n", streamEvents)
}

func TestGetClips(t *testing.T) {
	supabase := NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
	clips, err := supabase.GetClips(3)
	if err != nil {
		t.Fatalf("Error getting clips: %v", err)
	}
	fmt.Printf("%+v\n", clips)
}

func TestCreateClip(t *testing.T) {
	supabase := NewSupabase(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"))
	clip := &Clip{
		StreamID:        3,
		StartSecs:       100,
		EndSecs:         200,
		Caption:         "Test Clip",
		Description:     "Test Clip Description",
		BufferStartSecs: 90,
		BufferEndSecs:   210,
		URL:             "https://www.youtube.com/watch?v=123",
	}
	id, err := supabase.CreateClip(clip)
	if err != nil {
		t.Fatalf("Error creating clip: %v", err)
	}
	t.Logf("Clip created with ID: %d", id)
}
