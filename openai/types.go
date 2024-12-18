package openai

type OpenaiAudioResponseFormat string

const (
	OpenaiAudioResponseFormatJSON        OpenaiAudioResponseFormat = "json"
	OpenaiAudioResponseFormatText        OpenaiAudioResponseFormat = "text"
	OpenaiAudioResponseFormatSRT         OpenaiAudioResponseFormat = "srt"
	OpenaiAudioResponseFormatVerboseJSON OpenaiAudioResponseFormat = "verbose_json"
	OpenaiAudioResponseFormatVTT         OpenaiAudioResponseFormat = "vtt"
)

type OpenaiAudioResponse struct {
	Task     string  `json:"task"`
	Language string  `json:"language"`
	Duration float64 `json:"duration"`
	Segments []struct {
		ID               int     `json:"id"`
		Seek             int     `json:"seek"`  // some internal thing
		Start            float64 `json:"start"` // how long into the video (or audio) the transcribed text begins.
		End              float64 `json:"end"`
		Text             string  `json:"text"`
		Tokens           []int   `json:"tokens"`
		Temperature      float64 `json:"temperature"`
		AvgLogprob       float64 `json:"avg_logprob"`
		CompressionRatio float64 `json:"compression_ratio"`
		NoSpeechProb     float64 `json:"no_speech_prob"`
		Transient        bool    `json:"transient"`
	} `json:"segments"`
	Words []struct {
		Word  string  `json:"word"`
		Start float64 `json:"start"`
		End   float64 `json:"end"`
	} `json:"words"`
	Text string `json:"text"`
}
