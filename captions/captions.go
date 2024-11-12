package captions

import (
	"fmt"

	"strings"
)

type CaptionsHandler interface {
	CaptionsFromXml(filepath string, typ CaptionsType)
}

type CaptionsClient struct {
}

func NewCaptionsClient() *CaptionsClient {
	return &CaptionsClient{}
}

func (c *CaptionsClient) CaptionsFrom(url string, CaptionsType CaptionsType) (string, error) {

	captions, err := captionsFromUrl(url)
	if err != nil {
		return "", err
	}

	singleWordCaptions := toSingleWordCaptions(*captions)

	header := `[Script Info]
; Script generated by XML to ASS converter
Title: Converted Subtitles
ScriptType: v4.00+
Collisions: Normal
PlayDepth: 0

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default,The Bold Font,14,&H00FFFFFF,&H0033FF33,&H00000000,&H64000000,-1,0,0,0,100,100,0,0,1,3,5,5,10,10,30,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
`
	switch CaptionsType {
	case CaptionsSingleWord:
		return header + "\n" + singleWordCaptionsFrom(singleWordCaptions), nil
	case CaptionsHormozi:
		return hormoziCaptionsFrom2(groupCaptionWords(singleWordCaptions)), nil
		// case CaptionsBackgroundColor:
		// 	return header + "\n" + backgroundColorCaptionsFrom(captions), nil
	}

	return "", fmt.Errorf("invalid caption type")
}

func singleWordCaptionsFrom(captions []CaptionWord) string {
	ret := ""

	for i, c := range captions {
		if i == len(captions)-1 {
			ret += fmt.Sprintf("Dialogue: 0,%s,%s,Default,,0,0,0,,%s\n", msToASSTime(c.StartTimeMs), msToASSTime(c.StartTimeMs+1000), c.Word)
		} else {
			ret += fmt.Sprintf("Dialogue: 0,%s,%s,Default,,0,0,0,,%s\n", msToASSTime(c.StartTimeMs), msToASSTime(captions[i+1].StartTimeMs-1), c.Word)
		}
	}

	return ret
}

func hormoziCaptionsFrom2(captionGroups [][]CaptionWord) string {
	header := `[Script Info]
	; Script generated by XML to ASS converter
	Title: Converted Subtitles
	ScriptType: v4.00+
	Collisions: Normal
	PlayDepth: 0
	
	[V4+ Styles]
	Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
	Style: Default,The Bold Font,14,&H00FFFFFF,&H0033FF33,&H00000000,&H64000000,-1,0,0,0,100,100,0,0,1,3,5,5,10,10,30,1
	
	[Events]
	Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text`

	ret := header + "\n"

	for i, cg := range captionGroups {
		for j, c := range cg {
			endTime := -1
			if j == len(cg)-1 {
				if i == len(captionGroups)-1 {
					endTime = c.StartTimeMs + 1000
				} else {
					endTime = captionGroups[i+1][0].StartTimeMs - 1
				}
			} else {
				endTime = cg[j+1].StartTimeMs - 1
			}

			dialogue := fmt.Sprintf("Dialogue: 0,%s,%s,Default,,0,0,0,,", msToASSTime(c.StartTimeMs), msToASSTime(endTime))
			for k, c := range cg {
				w := strings.TrimSpace(c.Word)
				if len(w) > charLimit {
					w = "\n" + w + "\n"
				}

				if k == j {
					dialogue += " " + "{\\c&H0033FF33}" + strings.TrimSpace(c.Word) + "{\\c&HFFFFFF&}"
				} else {
					dialogue += " " + strings.TrimSpace(c.Word)
				}
			}

			ret += dialogue + "\n"
		}
	}

	return header + "\n" + ret
}

func hormoziCaptionsFrom(captions *Captions) string {
	header := `[Script Info]
; Script generated by XML to ASS converter
Title: Converted Subtitles
ScriptType: v4.00+
Collisions: Normal
PlayDepth: 0

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default,The Bold Font,14,&H00FFFFFF,&H0033FF33,&H00000000,&H64000000,-1,0,0,0,100,100,0,0,1,3,5,5,10,10,30,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text`

	ret := ""

	for i, event := range captions.Events {
		if event.Segs == nil || len(event.Segs) == 1 && event.Segs[0].UTF8 == "\n" {
			continue
		}

		for j, s := range event.Segs {
			var endTime int

			if j == len(event.Segs)-1 {
				if i == len(captions.Events)-1 {
					endTime = event.TStartMs + event.DDurationMs
				} else {
					endTime = captions.Events[i+1].TStartMs - 1
				}
			} else {
				endTime = event.TStartMs + event.Segs[j+1].TOffsetMs - 1
			}

			dialogue := fmt.Sprintf("Dialogue: 0,%s,%s,Default,,0,0,0,,", msToASSTime(event.TStartMs+s.TOffsetMs), msToASSTime(endTime))
			for k, seg := range event.Segs {
				if k == j {
					dialogue += " " + "{\\c&H0033FF33}" + strings.TrimSpace(seg.UTF8) + "{\\c&HFFFFFF&}"
				} else {
					dialogue += " " + strings.TrimSpace(seg.UTF8)
				}
			}

			ret += dialogue + "\n"
		}
	}

	return header + "\n" + ret
}
