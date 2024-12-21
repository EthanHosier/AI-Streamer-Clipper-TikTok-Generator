package stream_watcher

import "github.com/ethanhosier/clips/supabase"

type ClipSummary struct {
	StreamEvents   []supabase.StreamEvent `json:"stream_events"`
	UpdatedContext string                 `json:"updated_context"`
	Last20Secs     string                 `json:"last_20_secs"`
}

type ClipSummaryResponse struct {
	Last20Secs     string `json:"last_20_secs"`
	UpdatedContext string `json:"updated_context"`
	StreamEvents   []struct {
		Description string `json:"description"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
	} `json:"stream_events"`
}
