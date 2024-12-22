package stream_watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethanhosier/clips/supabase"
	"github.com/google/generative-ai-go/genai"
)

var responseSchema = &genai.Schema{
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
			Type: genai.TypeString,
		},
	},
	Required: []string{"stream_events", "updated_context", "last_20_secs"},
}

func (s *StreamWatcher) handleSummariseClip(clipUrl string, vidContext string, last20secs string, streamPositionSecs float64) (*ClipSummary, error) {
	prompt := `Here is a clip from a much longer live stream. Here is the context of what has happened up to this clip: ` + vidContext + `. More specifically, here is what happened just before this video was taken: ` + last20secs + `. 
	Give a detailed, specific analysis of this video. I will be passing these descriptions to another AI Agent which will determine which parts of the video clip to make viral tiktoks, so make this as detailed as possible. You should specify each event in the video, using this format:
	
	0:16-0:25 Kai's friend makes fun of how Kai came to him about his new girl
	0:25-0:27 Kai's friend asks "Oh you telling them you got a girlfriend" to make fun of Kai
	0:27-0:37 Kai's friends make fun of him for acting like the news was a secret, when they all knew
	0:47-0:51 Kai's friend says "Put them on a chair" as a joke
	1:07-1:14 Kai's friends comment on how much he is smiling
	1:24-1:30 Kai's friend remembers the day Kai was at Target and looked stiff due to his new girlfriend
	1:38-1:47 Kai's friends start acting out how he was acting at Target
	1:58-2:02 Kai's friends look at him like they are annoyed and laughing at him
	2:06-2:13 Kai's friend mentions Turks to try and understand how he got the girl
	2:19-2:21 Kai's friends laughing after Kai states he has a girlfriend again
	
	You must specify what happens in the last ~20 seconds of the video (as it cuts off). You should also update the context so that it is representative of the previous context and the events of this video.`

	resp, err := s.geminiClient.GetChatCompletionWithVideo(context.TODO(), prompt, clipUrl, responseSchema)
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
		StreamID: s.streamID,
		Context:  clipSummary.UpdatedContext,
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
