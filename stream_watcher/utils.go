package stream_watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ethanhosier/clips/ffmpeg"
	"github.com/ethanhosier/clips/supabase"
	"github.com/google/generative-ai-go/genai"
)

var (
	clipSummaryResponseSchema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"stream_events": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"start_time": {
							Type:        genai.TypeString,
							Description: "Should be in mm:ss format",
						},
						"end_time": {
							Type:        genai.TypeString,
							Description: "Should be in mm:ss format",
						},
						"description": {
							Type:        genai.TypeString,
							Description: "Make this detailed and specific. Detail on what the streamer is saying and interacting with other people. The people's reactions are important. Do not include any timestamps in the description.",
						},
					},
					Required: []string{"start_time", "end_time", "description"},
				},
			},
			"updated_context": {
				Type: genai.TypeString,
			},
			"last_20_secs": {
				Type:        genai.TypeString,
				Description: "Don't include any actual timestamps here. Just describe what is happening in the last ~20 seconds of the video, as this will be used in the next clip summary.",
			},
		},
		Required: []string{"stream_events", "updated_context", "last_20_secs"},
	}
)

func (s *StreamWatcher) handleSummariseClip(clipUrl string, vidContext string, last20secs string, streamPositionSecs float64, name string) (*ClipSummary, error) {
	prompt := `Here is a clip from a much longer live stream. The streamer is ` + name + `. Here is the context of what has happened up to this clip: ` + vidContext + `. More specifically, here is what happened just before this video was taken: ` + last20secs + `. 
	Give a detailed, specific analysis of this video. I will be passing these descriptions to another AI Agent which will determine which parts of the video clip to make viral tiktoks, so make this as detailed but unbiased as possible. You should specify each event in the video, using this format:
	
	0:16-0:25 Kai's friend makes fun of how Kai came to him about his new girl
	0:25-0:27 Kai's friend asks "Oh you telling them you got a girlfriend" to make fun of Kai
	0:27-0:37 Kai's friends make fun of him for acting like the news was a secret, when they all knew
	0:47-0:51 Kai's friend says "Put them on a chair" as a joke
	1:07-1:14 Kai's friends comment on how much he is smiling
	1:24-1:30 Kai's friend remembers the day Kai was at Target and looked stiff due to his new girlfriend
	1:38-1:47 Kai's friends start acting out how he was acting at Target
	1:58-2:02 Kai's friends look at him like they are annoyed and laughing at him
	2:06-2:13 Kai's friend mentions "Turks" to try and understand how he got the girl
	2:19-2:21 Kai's friends laughing after Kai states he has a girlfriend again
	
	For each event, you must explicitly state what is happening generally in the clip, any emotions, and then the exact words which are said in speech marks. If there are multiple people talking, you must state the exact words of each person.

	You must specify what happens in the last ~20 seconds of the video (as it cuts off). You should also update the context so that it is representative of the previous context and the events of this video.`

	resp, err := s.geminiClient.GetChatCompletionWithVideo(context.TODO(), prompt, clipUrl, clipSummaryResponseSchema)
	if err != nil {
		return nil, err
	}

	var clipSummary ClipSummaryResponse
	err = json.Unmarshal([]byte(*resp), &clipSummary)
	if err != nil {
		return nil, err
	}

	streamEvents := []supabase.StreamEvent{}
	for _, event := range clipSummary.StreamEvents {
		startSecs, err := ConvertMMSS(event.StartTime)
		if err != nil {
			return nil, err
		}
		endSecs, err := ConvertMMSS(event.EndTime)
		if err != nil {
			return nil, err
		}
		streamEvents = append(streamEvents, supabase.StreamEvent{
			StartSecs:   startSecs + int(streamPositionSecs),
			EndSecs:     endSecs + int(streamPositionSecs),
			Description: event.Description,
			StreamID:    s.streamID,
		})
	}
	return &ClipSummary{
		StreamEvents:   streamEvents,
		UpdatedContext: clipSummary.UpdatedContext,
		Last20Secs:     clipSummary.Last20Secs,
	}, nil
}

func (s *StreamWatcher) storeClipSummaryParts(clipSummary *ClipSummary) error {
	streamContext := supabase.StreamContext{
		StreamID:   s.streamID,
		Context:    clipSummary.UpdatedContext,
		Last20Secs: clipSummary.Last20Secs,
	}

	streamContextID, err := s.supabaseClient.CreateStreamContext(&streamContext)
	if err != nil {
		return fmt.Errorf("error creating stream context: %v", err)
	}

	streamEvents := []supabase.StreamEvent{}
	for _, event := range clipSummary.StreamEvents {
		streamEvents = append(streamEvents, supabase.StreamEvent{
			StreamContextID: streamContextID,
			StartSecs:       event.StartSecs,
			EndSecs:         event.EndSecs,
			Description:     event.Description,
			StreamID:        s.streamID,
		})
	}

	_, err = s.supabaseClient.CreateStreamEvents(streamEvents)
	if err != nil {
		return fmt.Errorf("error creating stream events: %v", err)
	}

	return nil
}

func ConvertMMSS(timeStr string) (int, error) {
	// Split the input string into minutes and seconds
	timeParts := strings.Split(timeStr, ":")
	if len(timeParts) != 2 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	// Convert minutes and seconds to integers
	minutes, err := strconv.Atoi(timeParts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes value: %s", timeParts[0])
	}

	seconds, err := strconv.Atoi(timeParts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid seconds value: %s", timeParts[1])
	}

	// Calculate total seconds
	totalSeconds := (minutes * 60) + seconds
	return totalSeconds, nil
}

func (s *StreamWatcher) findClips(windowStartSecs int, videoContext string, name string) ([]FoundClip, error) {
	streamEvents, err := s.supabaseClient.GetStreamEventsAfter(windowStartSecs, s.streamID)
	if err != nil {
		return nil, err
	}

	if len(streamEvents) == 0 {
		slog.Warn("No stream events found, skipping clip search")
		return []FoundClip{}, nil
	}

	prompt := `You are an expert clipper who takes a section of a live stream and finds the best clips from it.
	You can't see video, instead you are given a list of what happens in the stream at each timestamp.
	
	We are looking for clips which follow on of the following criteria. The clip must be:
	1. A funny moment
	2. A high energy moment (dancing, screaming, intense emotions etc)
	3. An interesting conversation, particularly a streamer "giving their thoughts on ..."
	4. A fail / blunder
	5. An interesting interaction with the chat
	6. Extremely relatable content
	7. The streamer telling a story about something that happened
	8. References to inside jokes / memes
	9. Mentions of other streamers
	10. Iconic quotes or phrases that represent the streamer’s brand or personality.
	11. Rage moments
	12. Spontaneous dancing, singing, or any musical moments from the streamer.
	13. Racy content
	14. Announcements
	
	Ensure that you focus on the STREAMER. Just winning a game or getting a kill is not a good clip, unless the streamer is doing something funny or interesting about it.

  Important: If two moments are about the same topic, and the start time of the second moment is less than 60 seconds after the end time of the first moment, they must be combined into a single clip. The combined clip’s start time is the first moment’s start, and its end time is the last moment’s end. Provide a single caption and single description for the combined clip.

	Clips must be at least 10 seconds long, and max 3 minutes. If given the choice, longer clips are better, but keep them relevant.

	Be strict about what is worth making a clip from. False positives are worse than false negatives. It is perfectly fine (and normal) to not find any clips.

	---
Here is the stream of events and timestamps to consider: ` + streamEventsStrFromStreamEvents(streamEvents) + `
	---

	---
Here is some general context about the stream up to this point. Use this as a general guide, but don't refer to anything in the context if it isnt specifically in the stream of events: ` + videoContext + `
  ---

	For each found clip (if there are any), you must include the start secs, the end secs, the caption and the description (of the clip, which category it false under 1-12 and why it is good). 

	The caption of the clip should be a short, clickbaity and engaging caption which references the streamer and something specificabout the clip. This should be whatever the main part of the clip is about. Make it informal and voiced as how a cool 16 year old texting their friend would sound. The caption must mention the streamer's name ("` + name + `").
	Examples include: 
	"` + name + ` reacts to the first article about him 🔥"
	"` + name + ` joins the Power Rangers 😭😭"
	"` + name + ` meets Breckie Hill for the first time 💀"
	"` + name + ` reveals his $500,000 production team 🤯"
	Don't include hashtags or punctuation in the caption.
`

	type FoundClipResponseFormat struct {
		FoundClips []FoundClip `json:"found_clips"`
	}

	response, err := s.openaiClient.CreateChatCompletionWithResponseFormatForO1Mini(context.TODO(), prompt, "If it is clear that there are no clips, return an empty array. Otherwise, return the clips in this format:", FoundClipResponseFormat{})
	if err != nil {
		return nil, err
	}

	var foundClips FoundClipResponseFormat
	err = json.Unmarshal([]byte(*response), &foundClips)
	if err != nil {
		return nil, err
	}

	return foundClips.FoundClips, nil
}

// TODO: Maybe clean this up?
func (s *StreamWatcher) getActualClipFrom(c *FoundClip, vidFiles []string, bufferStartSecs, bufferEndSecs int) (string, error) {
	// Find which files contain our clip by calculating cumulative duration
	var cumulativeDuration float64
	var startFileIndex, endFileIndex int
	var startFileOffset float64

	//get duration of all files:
	totalDuration := 0.0
	for _, file := range vidFiles {
		duration, err := s.ffmpegClient.VideoDuration(file)
		if err != nil {
			return "", fmt.Errorf("failed to get duration for file %s: %v", file, err)
		}
		totalDuration += duration
	}

	bufferedClip := &FoundClip{
		StartSecs: math.Max(c.StartSecs-float64(bufferStartSecs), 0),
		EndSecs:   math.Min(c.EndSecs+float64(bufferEndSecs), totalDuration),
	}

	// Scan through files to find start and end positions
	for i, file := range vidFiles {
		duration, err := s.ffmpegClient.VideoDuration(file)
		if err != nil {
			return "", fmt.Errorf("failed to get duration for file %s: %v", file, err)
		}

		nextDuration := cumulativeDuration + duration

		// Found start file
		if cumulativeDuration <= bufferedClip.StartSecs && bufferedClip.StartSecs < nextDuration {
			startFileIndex = i
			startFileOffset = bufferedClip.StartSecs - cumulativeDuration
		}

		// Found end file
		if cumulativeDuration <= bufferedClip.EndSecs && bufferedClip.EndSecs <= nextDuration {
			endFileIndex = i
			break
		}

		cumulativeDuration = nextDuration
	}

	if endFileIndex >= len(vidFiles) {
		return "", fmt.Errorf("clip end time (%f) exceeds available video duration", bufferedClip.EndSecs)
	}

	var inputFile string
	if startFileIndex == endFileIndex {
		// Clip is contained within a single file
		inputFile = vidFiles[startFileIndex]
	} else {
		// Create temp directory if it doesn't exist

		randomStr := strconv.Itoa(rand.Intn(1000000))
		tempDir := fmt.Sprintf("tmp-join-clips-%s/", randomStr)
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Need to combine multiple files
		tempFile := fmt.Sprintf("%s/temp_merged_%d_%d.mp4", tempDir, startFileIndex, endFileIndex)

		// Create a temporary concat file
		concatFile := fmt.Sprintf("%s/concat_%d_%d.txt", tempDir, startFileIndex, endFileIndex)
		f, err := os.Create(concatFile)
		if err != nil {
			return "", fmt.Errorf("failed to create concat file: %v", err)
		}
		defer os.Remove(concatFile)

		// Write the file list in the required format
		for i := startFileIndex; i <= endFileIndex; i++ {
			absPath, err := filepath.Abs(vidFiles[i])
			if err != nil {
				return "", fmt.Errorf("failed to get absolute path: %v", err)
			}
			_, err = f.WriteString(fmt.Sprintf("file '%s'\n", absPath))
			if err != nil {
				return "", fmt.Errorf("failed to write to concat file: %v", err)
			}
		}
		f.Close()

		// Merge the files using ffmpeg concat demuxer
		cmd := exec.Command("ffmpeg",
			"-f", "concat",
			"-safe", "0",
			"-i", concatFile,
			"-c", "copy",
			tempFile)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to merge video files: %v", err)
		}
		inputFile = tempFile
		defer os.Remove(tempFile)
	}

	// Extract the actual clip
	startTime := secondsToTimeString(startFileOffset)
	duration := secondsToTimeString(bufferedClip.EndSecs - bufferedClip.StartSecs)

	outputFile, err := s.ffmpegClient.ClipVideo(inputFile, ffmpeg.RandomOutputPathFor(ffmpeg.MP4, "tmp/", strconv.Itoa(int(c.StartSecs)), strconv.Itoa(int(c.EndSecs))), startTime, duration)
	if err != nil {
		return "", fmt.Errorf("failed to extract clip: %v", err)
	}

	return outputFile, nil
}

// Helper function to convert seconds to FFmpeg time format (HH:MM:SS.mmm)
func secondsToTimeString(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}

func streamEventsStrFromStreamEvents(streamEvents []supabase.StreamEvent) string {
	streamEventsStr := "Start Secs, End Secs, Description\n"
	for _, event := range streamEvents {
		streamEventsStr += fmt.Sprintf("%d, %d, %s\n", event.StartSecs, event.EndSecs, event.Description)
	}
	return streamEventsStr
}
