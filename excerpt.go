package excerpt

import (
	"fmt"
	"index/suffixarray"
	"sort"
	"strings"
)

type ExcerptWindow struct {
	Start      int
	ByteLength int
	CharLength int
	Score      float64
	Text       string
}

func (e *ExcerptWindow) String() string {
	return fmt.Sprintf("<ExcerptWindow(charlength: %v|bytes: %v): starts at: %v score: %v>:\n%s\n", e.CharLength, e.ByteLength, e.Start, e.Score, e.Text)
}

/*
FindExcerpts searches searchterms in body and returns excerpts of a given
length that contain the terms.

The terms can be weighted and each ExcerptWindow gets a cummulative score
of the searchterms it contains. setting findHighestScore the function will
only return the ExcerptWindow with the hightest score.

If a match is at the end of a window and overlaps its boundry the window
will be extended to include the full match.

An ExcerptWindow always starts with a match. In the future an option might
be added to position/center the window around the matches.
*/
func FindExcerpts(searchterms map[string]float64, body string, length int,
	findHighestScore bool) (excerptCandidates []*ExcerptWindow) {
	b := []byte(strings.ToLower(body))
	var blength int = len(b)
	index := suffixarray.New(b)
	var offsets []int
	scores := make(map[int][]float64)
	for term, weight := range searchterms {
		termMatches := index.Lookup([]byte(strings.ToLower(term)), -1)
		// use the character(rune not byte) length multiplied with the weight
		// as score, we also need to know the byte length of the match in case
		// we need to extend the window if the match overlaps the window boundry
		for _, m := range termMatches {
			scores[m] = []float64{float64(len([]rune(term))) * weight, float64(len([]byte(term)))}
		}
		offsets = append(offsets, termMatches...)
	}

	sort.Ints(offsets)

	var nextMatchIdx int
	var sliceEnd int
	var HighestScore float64 = 0
	var HighestScoreIdx int
	for i, offset := range offsets {
		ew := &ExcerptWindow{
			Start:      offset,
			CharLength: length,
			Score:      scores[offset][0],
		}

		// runes use 4 bytes max - but 1~4 depending on the character
		// we want a string of character_length length so we look ahead
		// 4 * length bytes, check the character (rune) length and adjust
		sliceEnd = (offset + length) * 4
		if sliceEnd > blength {
			sliceEnd = blength
		}
		r := []rune(body[offset:sliceEnd])
		// if the window would exceed the end of body we adjust the length
		if ew.CharLength >= len(r) {
			ew.CharLength = len(r)
		}
		r = r[0:ew.CharLength]
		ew.ByteLength = len([]byte(string(r)))

		nextMatchIdx = i + 1
		for {
			if nextMatchIdx >= len(offsets) || offsets[nextMatchIdx] > offset+ew.ByteLength {
				break
			}
			ew.Score += scores[offsets[nextMatchIdx]][0]
			// if the match would be cut off we extend the window
			if offsets[nextMatchIdx]+int(scores[offsets[nextMatchIdx]][1]) >= ew.Start+ew.ByteLength {
				ew.ByteLength = offsets[nextMatchIdx] + int(scores[offsets[nextMatchIdx]][1]) - ew.Start
				break
			}
			nextMatchIdx += 1
		}

		ew.Text = strings.TrimSpace(body[ew.Start : ew.Start+ew.ByteLength])
		ew.CharLength = len([]rune(ew.Text))
		if ew.Score > HighestScore {
			HighestScore = ew.Score
			HighestScoreIdx = i
		}
		excerptCandidates = append(excerptCandidates, ew)
	}

	if findHighestScore {
		excerptCandidates = []*ExcerptWindow{excerptCandidates[HighestScoreIdx]}
	}
	return
}
