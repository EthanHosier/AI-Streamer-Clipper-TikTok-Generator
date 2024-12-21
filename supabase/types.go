package supabase

type Platform string

const (
	PlatformTwitch  Platform = "twitch"
	PlatformYoutube Platform = "youtube"
	PlatformKick    Platform = "kick"
	PlatformTikTok  Platform = "tiktok"
	PlatformRumble  Platform = "rumble"
)

type Streamer struct {
	ID   *int        `json:"id,omitempty"`
	Name string      `json:"name"`
	Info interface{} `json:"info"`
}

type StreamerPlatform struct {
	ID         *int     `json:"id,omitempty"`
	StreamerID int      `json:"streamer_id"`
	Platform   Platform `json:"platform"`
	Url        string   `json:"url"`
}

type Stream struct {
	ID         *int        `json:"id,omitempty"`
	StreamerID int         `json:"streamer_id"`
	Platform   string      `json:"platform"`
	Info       interface{} `json:"info"`
}

type StreamContext struct {
	ID        *int    `json:"id,omitempty"`
	CreatedAt *string `json:"created_at,omitempty"`
	StreamID  int     `json:"stream_id"`
	Context   string  `json:"context"`
}
type StreamEvent struct {
	ID              *int   `json:"id,omitempty"`
	StartSecs       int    `json:"start_secs"`
	EndSecs         int    `json:"end_secs"`
	Description     string `json:"description"`
	StreamID        int    `json:"stream_id"`
	StreamContextID int    `json:"stream_context_id"`
}
