package gemini

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	code := m.Run()
	os.Exit(code)
}

func TestGeminiGetChatCompletion(t *testing.T) {
	geminiClient, err := NewGeminiClient(context.Background(), os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		t.Fatalf("Error creating Gemini client: %v", err)
	}

	resp, err := geminiClient.GetChatCompletion(context.Background(), "Say something cool")
	if err != nil {
		t.Fatalf("Error getting chat completion: %v", err)
	}

	fmt.Println(*resp)
}

func TestGeminiGetChatCompletionWithVideo(t *testing.T) {
	geminiClient, err := NewGeminiClient(context.Background(), os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		t.Fatalf("Error creating Gemini client: %v", err)
	}

	resp, err := geminiClient.GetChatCompletionWithVideo(context.Background(), "Give a detailed, specific analysis of the video", "/home/ethanh/Desktop/go/clips/clips/angryginge13/output001.mp4", nil)
	if err != nil {
		t.Fatalf("Error getting chat completion with video: %v", err)
	}

	fmt.Println(*resp)
}

func TestGeminiGetChatCompletionWithVideoWithResponseSchema(t *testing.T) {
	geminiClient, err := NewGeminiClient(context.Background(), os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		t.Fatalf("Error creating Gemini client: %v", err)
	}

	vidContext := "The streamer is using the gingerbread skin and is located near Retail Row. They're actively participating in a match with two other teammates, as shown by the team display in the top left. A quest titled \"Outlast Players\" has already been completed, giving the player XP."

	last20secs := "The streamer has just helped kill another player with his sniper rifle."

	responseSchema := &genai.Schema{
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
							Description: "Make this detailed and specific. Detail on what the streamer is saying and interacting with other people. The people's reactions are important.",
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

	resp, err := geminiClient.GetChatCompletionWithVideo(context.Background(), prompt, "/home/ethanh/Desktop/go/clips/clips/angryginge13/output001.mp4", responseSchema)
	if err != nil {
		t.Fatalf("Error getting chat completion with video: %v", err)
	}

	fmt.Println(*resp)
}
