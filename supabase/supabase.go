package supabase

import (
	"fmt"

	"github.com/nedpals/supabase-go"
)

type Supabase struct {
	client *supabase.Client
}

func NewSupabase(url string, key string) *Supabase {
	return &Supabase{
		client: supabase.CreateClient(url, key),
	}
}

func (s *Supabase) GetStreamer(id int) (*Streamer, error) {
	var result []interface{}

	err := s.client.DB.From("streamers").Select("*").Eq("id", fmt.Sprintf("%d", id)).Execute(&result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no streamer found")
	}

	// Get the map from the result
	resultMap, ok := result[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert result to map")
	}

	retID := int(resultMap["id"].(float64))
	// Create a new Streamer and populate it from the map
	streamer := &Streamer{
		ID:   &retID,
		Name: resultMap["name"].(string),
		Info: resultMap["info"], // This can be nil
	}

	return streamer, nil
}

func (s *Supabase) CreateStreamer(streamer *Streamer) (int, error) {
	result := []interface{}{}
	err := s.client.DB.From("streamers").Insert(streamer).Execute(&result)
	if err != nil {
		return 0, err
	}

	return int(result[0].(map[string]interface{})["id"].(float64)), nil
}

func (s *Supabase) GetStreamerPlatforms(streamerID int) ([]StreamerPlatform, error) {
	result := []interface{}{}
	err := s.client.DB.From("streamer_platforms").Select("*").Eq("streamer_id", fmt.Sprintf("%d", streamerID)).Execute(&result)
	if err != nil {
		return nil, err
	}

	streamerPlatforms := []StreamerPlatform{}
	for _, item := range result {
		resultMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to convert result to map")
		}
		retID := int(resultMap["id"].(float64))
		streamerPlatforms = append(streamerPlatforms, StreamerPlatform{
			ID:         &retID,
			StreamerID: int(resultMap["streamer_id"].(float64)),
			Platform:   Platform(resultMap["platform"].(string)),
			Url:        resultMap["url"].(string),
		})
	}

	return streamerPlatforms, nil
}

func (s *Supabase) CreateStreamerPlatform(streamerPlatform *StreamerPlatform) (int, error) {
	result := []interface{}{}
	err := s.client.DB.From("streamer_platforms").Insert(streamerPlatform).Execute(&result)
	if err != nil {
		return 0, err
	}

	return int(result[0].(map[string]interface{})["id"].(float64)), nil
}

func (s *Supabase) CreateStream(stream *Stream) (int, error) {
	result := []interface{}{}
	err := s.client.DB.From("streams").Insert(stream).Execute(&result)
	if err != nil {
		return 0, err
	}

	fmt.Printf("Result: %+v\n", result)

	return int(result[0].(map[string]interface{})["id"].(float64)), nil
}

func (s *Supabase) GetStream(id int) (*Stream, error) {
	result := []interface{}{}
	err := s.client.DB.From("streams").Select("*").Eq("id", fmt.Sprintf("%d", id)).Execute(&result)
	if err != nil {
		return nil, err
	}

	stream := &Stream{}

	resultMap, ok := result[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert result to map")
	}

	retID := int(resultMap["id"].(float64))
	stream.ID = &retID
	stream.StreamerID = int(resultMap["streamer_id"].(float64))
	stream.Platform = resultMap["platform"].(string)
	stream.Info = resultMap["info"]

	return stream, nil
}

func (s *Supabase) GetStreamEvents(streamID int) ([]StreamEvent, error) {
	result := []interface{}{}
	err := s.client.DB.From("stream_events").Select("*").Eq("stream_id", fmt.Sprintf("%d", streamID)).Execute(&result)
	if err != nil {
		return nil, err
	}

	streamEvents := []StreamEvent{}
	for _, item := range result {
		resultMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to convert result to map")
		}
		retID := int(resultMap["id"].(float64))
		streamEvents = append(streamEvents, StreamEvent{
			ID:              &retID,
			StreamID:        int(resultMap["stream_id"].(float64)),
			StartSecs:       int(resultMap["start_secs"].(float64)),
			EndSecs:         int(resultMap["end_secs"].(float64)),
			Description:     resultMap["description"].(string),
			StreamContextID: int(resultMap["stream_context_id"].(float64)),
		})
	}

	return streamEvents, nil
}

func (s *Supabase) GetStreamEventsAfter(startSecs int, streamID int) ([]StreamEvent, error) {
	result := []interface{}{}
	err := s.client.DB.From("stream_events").Select("*").Eq("stream_id", fmt.Sprintf("%d", streamID)).Gte("start_secs", fmt.Sprintf("%d", startSecs)).Execute(&result)

	if err != nil {
		return nil, err
	}

	streamEvents := []StreamEvent{}
	for _, item := range result {
		resultMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to convert result to map")
		}
		retID := int(resultMap["id"].(float64))
		streamEvents = append(streamEvents, StreamEvent{
			ID:              &retID,
			StreamID:        int(resultMap["stream_id"].(float64)),
			StartSecs:       int(resultMap["start_secs"].(float64)),
			EndSecs:         int(resultMap["end_secs"].(float64)),
			Description:     resultMap["description"].(string),
			StreamContextID: int(resultMap["stream_context_id"].(float64)),
		})
	}

	return streamEvents, nil
}

func (s *Supabase) CreateStreamEvent(streamEvent *StreamEvent) (int, error) {
	result := []interface{}{}
	err := s.client.DB.From("stream_events").Insert(streamEvent).Execute(&result)
	if err != nil {
		return 0, err
	}

	return int(result[0].(map[string]interface{})["id"].(float64)), nil
}

func (s *Supabase) CreateStreamEvents(streamEvents []StreamEvent) ([]int, error) {
	result := []interface{}{}
	err := s.client.DB.From("stream_events").Insert(streamEvents).Execute(&result)
	if err != nil {
		return nil, err
	}

	ids := []int{}
	for _, item := range result {
		ids = append(ids, int(item.(map[string]interface{})["id"].(float64)))
	}

	return ids, nil
}

func (s *Supabase) GetStreamContexts(streamID int) ([]StreamContext, error) {
	result := []interface{}{}
	err := s.client.DB.From("stream_contexts").Select("*").OrderBy("created_at", "desc").Eq("stream_id", fmt.Sprintf("%d", streamID)).Execute(&result) // newest first
	if err != nil {
		return nil, err
	}

	streamContexts := []StreamContext{}
	for _, item := range result {
		resultMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to convert result to map")
		}
		retID := int(resultMap["id"].(float64))
		retCreatedAt := resultMap["created_at"].(string)
		streamContexts = append(streamContexts, StreamContext{
			ID:        &retID,
			StreamID:  int(resultMap["stream_id"].(float64)),
			Context:   resultMap["context"].(string),
			CreatedAt: &retCreatedAt,
		})
	}

	return streamContexts, nil
}

func (s *Supabase) CreateStreamContext(streamContext *StreamContext) (int, error) {
	result := []interface{}{}
	err := s.client.DB.From("stream_contexts").Insert(streamContext).Execute(&result)
	if err != nil {
		return 0, err
	}

	return int(result[0].(map[string]interface{})["id"].(float64)), nil
}

func (s *Supabase) CreateClip(clip *Clip) (int, error) {
	result := []interface{}{}
	err := s.client.DB.From("clips").Insert(clip).Execute(&result)
	if err != nil {
		return 0, err
	}

	return int(result[0].(map[string]interface{})["id"].(float64)), nil
}

func (s *Supabase) GetClips(streamID int) ([]Clip, error) {
	result := []interface{}{}
	err := s.client.DB.From("clips").Select("*").Eq("stream_id", fmt.Sprintf("%d", streamID)).Execute(&result)
	if err != nil {
		return nil, err
	}

	clips := []Clip{}
	for _, item := range result {
		resultMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to convert result to map")
		}
		retID := int(resultMap["id"].(float64))
		clips = append(clips, Clip{
			ID:              &retID,
			StreamID:        int(resultMap["stream_id"].(float64)),
			StartSecs:       int(resultMap["start_secs"].(float64)),
			EndSecs:         int(resultMap["end_secs"].(float64)),
			Caption:         resultMap["caption"].(string),
			Description:     resultMap["description"].(string),
			BufferStartSecs: int(resultMap["buffer_start_secs"].(float64)),
			BufferEndSecs:   int(resultMap["buffer_end_secs"].(float64)),
			URL:             resultMap["url"].(string),
		})
	}

	return clips, nil
}
