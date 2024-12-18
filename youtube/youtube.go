package youtube

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
)

type YoutubeHandler interface {
	CaptionsForId(id string)
}

type YoutubeClient struct {
	client *youtube.Client
}

type Video struct {
	ID              string
	Title           string
	Description     string
	Author          string
	ChannelID       string
	ChannelHandle   string
	Views           int
	Duration        time.Duration
	PublishDate     time.Time
	DASHManifestURL string
	HLSManifestURL  string
	CaptionTrackURL string
	AudioURL        string
	VideoURL        string
}

func NewYoutubeClient() *YoutubeClient {
	return &YoutubeClient{
		client: &youtube.Client{},
	}
}

func (y *YoutubeClient) VideoForId(id string) (*Video, error) {
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", id)
	video, err := y.client.GetVideo(url)
	fmt.Println(video.Duration)
	if err != nil {
		return nil, fmt.Errorf("error getting video: %v\n\n", err)
	}

	hdFormat := video.Formats[0]
	audioFormat := video.Formats.Itag(140)[0]

	json3Url, err := replaceWithJson3Url(video.CaptionTracks[0].BaseURL)

	return &Video{
		ID:              video.ID,
		Title:           video.Title,
		Description:     video.Description,
		Author:          video.Author,
		ChannelID:       video.ChannelID,
		ChannelHandle:   video.ChannelHandle,
		Views:           video.Views,
		Duration:        video.Duration,
		PublishDate:     video.PublishDate,
		DASHManifestURL: video.DASHManifestURL,
		HLSManifestURL:  video.HLSManifestURL,
		CaptionTrackURL: json3Url,
		AudioURL:        audioFormat.URL,
		VideoURL:        hdFormat.URL,
	}, nil
}

func replaceWithJson3Url(originalURL string) (string, error) {
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return "", fmt.Errorf("error parsing URL: %v", err)
	}

	query := parsedURL.Query()
	query.Set("fmt", "json3")

	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}

func (y *YoutubeClient) retryDownload(video *youtube.Video, format *youtube.Format, filePath string) error {
	for i := 0; i < 3; i++ { // Retry up to 3 times
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}
		defer file.Close()

		stream, _, err := y.client.GetStream(video, format)
		if err != nil {
			fmt.Printf("Attempt %d: error getting stream: %v\n", i+1, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		_, err = io.Copy(file, stream)
		if err == nil {
			return nil // Success
		}
		fmt.Printf("Attempt %d: error saving file: %v\n", i+1, err)
		time.Sleep(2 * time.Second) // Wait before retrying
	}
	return errors.New("failed to download after multiple attempts")
}

func extractBeforeSubstring(input, substring string) (string, error) {
	index := strings.Index(input, substring)
	if index == -1 {
		return "", fmt.Errorf("substring not found in input")
	}
	return input[:index], nil
}

func (y *YoutubeClient) DownloadVideoAndAudio(videoID, videoPath, audioPath string) error {

	return y.downloadVideoAndAudioPython(videoID, videoPath, audioPath)
	// client := youtube.Client{}
	// videoURL := "https://www.youtube.com/watch?v=" + videoID
	// video, err := client.GetVideo(videoURL)
	// if err != nil {
	// 	return fmt.Errorf("error fetching video details: %w", err)
	// }

	// videoFormat := video.Formats[1]
	// audioFormat := video.Formats.WithAudioChannels()[0]

	// if err := y.retryDownload(video, &videoFormat, videoPath); err != nil {
	// 	return fmt.Errorf("error downloading video: %w", err)
	// }
	// fmt.Println("Video downloaded successfully!")

	// if err := y.retryDownload(video, &audioFormat, audioPath); err != nil {
	// 	return fmt.Errorf("error downloading audio: %w", err)
	// }
	// fmt.Println("Audio downloaded successfully!")
	// return nil
}

func (y *YoutubeClient) downloadVideoAndAudioPython(videoID, videoPath, audioPath string) error {
	// Construct the Python script command
	pythonScript := "youtube/download.py" // Replace with the actual script file name
	videoURL := "https://www.youtube.com/watch?v=" + videoID

	// Call the Python script as a subprocess
	cmd := exec.Command("python3", pythonScript, videoURL, videoPath, audioPath)

	// Connect subprocess stdout and stderr to os.Stdout and os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the Python script
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running Python script: %w", err)
	}

	fmt.Println("Python script completed successfully.")
	return nil
}
