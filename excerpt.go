package excerpt

import (
	"fmt"
	"index/suffixarray"
	"log"
	"sort"
	"strings"
	"time"
)

const (
	PADDING_WIDTH = 5
)

// type Match struct {
// 	Start int
// 	End   int
// }

type ExcerptWindow struct {
	Start      int
	ByteLength int
	CharLength int
	Score      float64
	Text       string
	Matches    []*Match
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
func FindExcerpts(searchterms map[string]float64, body string, length int, expand bool, findHighestScore bool) (excerptCandidates []*ExcerptWindow) {
	startTime := time.Now()
	b := []byte(strings.ToLower(body))
	var blength int = len(b)
	// log.Printf("body length (in bytes): %v\n", blength)
	var offsets []int
	var score, bytelength float64
	index := suffixarray.New(b)
	scores := make(map[int][]float64)
	for term, weight := range searchterms {
		termMatches := index.Lookup([]byte(strings.ToLower(term)), -1)
		// use the character(rune not byte) length multiplied with the weight
		// as score, we also need to know the byte length of the match in case
		// we need to extend the window if the match overlaps the window boundry
		score = float64(len([]rune(term))) * weight
		bytelength = float64(len([]byte(term)))
		for _, m := range termMatches {
			scores[m] = []float64{score, bytelength}
		}
		offsets = append(offsets, termMatches...)
	}
	log.Printf("finding matches took: %v. nr match positions: %v\n", time.Since(startTime), len(offsets))
	sort.Ints(offsets)

	// log.Println(offsets)
	// log.Println(scores)

	var nextMatchIdx int
	var sliceEnd int
	var HighestScore float64 = 0
	var HighestScoreIdx int = 0
	var ew *ExcerptWindow
	var r []rune

	// if we have no match we just send and excerpt that starts at the beginning
	if len(offsets) == 0 {
		scores[0] = []float64{0, 0}
		offsets = []int{0}
	}

	for i, offset := range offsets {
		// log.Printf("o")
		ew = &ExcerptWindow{
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
		// startTimeCnv := time.Now()
		r = []rune(body[offset:sliceEnd])
		// if the window would exceed the end of body we adjust the length
		if ew.CharLength >= len(r) {
			ew.CharLength = len(r)
		}
		r = r[0:ew.CharLength]
		ew.ByteLength = len([]byte(string(r)))
		// log.Printf("runtime []rune/[]byte conversions: %v\n", time.Since(startTimeCnv))

		nextMatchIdx = i + 1
		for {
			// log.Printf("n")
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
	log.Printf("runtime: %v\n", time.Since(startTime))
	return
}
