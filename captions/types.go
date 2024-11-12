package captions

type TimedText struct {
	Body Body `xml:"body"`
}

type Body struct {
	P []P `xml:"p"`
}

type P struct {
	T int `xml:"t,attr"`
	D int `xml:"d,attr"`
	S []S `xml:"s"`
}

type S struct {
	T    int    `xml:"t,attr"`
	Ac   int    `xml:"ac,attr"`
	Text string `xml:",chardata"`
}

type Captions struct {
	WireMagic string  `json:"wireMagic"`
	Events    []Event `json:"events"`
}

type Event struct {
	TStartMs     int   `json:"tStartMs"`
	DDurationMs  int   `json:"dDurationMs"`
	ID           int   `json:"id"`
	WpWinPosId   int   `json:"wpWinPosId"`
	WsWinStyleId int   `json:"wsWinStyleId"`
	WWinId       int   `json:"wWinId"`
	Segs         []Seg `json:"segs"`
}

type Seg struct {
	UTF8      string `json:"utf8"`
	AcAsrConf int    `json:"acAsrConf"`
	TOffsetMs int    `json:"tOffsetMs"`
}

type CaptionsType string

const (
	CaptionsSingleWord      CaptionsType = "single-word"
	CaptionsHormozi         CaptionsType = "hormozi"
	CaptionsBackgroundColor CaptionsType = "background-color"
)
