package excerpt

import (
	"bytes"
	"fmt"
	"github.com/fvbock/substr/src/substr"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

type TermScore struct {
	Score      float64
	ByteLength uint32
}

type Match struct {
	Start      uint32
	ByteLength uint32
	Score      float64
}

func (m *Match) String() string {
	return fmt.Sprintf("<Match at: %v, Score: %v>", m.Start, m.Score)
}

type ExcerptWindowBM struct {
	Start      uint32
	ByteLength uint32
	CharLength uint32
	Score      float64
	Text       string
	Matches    []*Match
}

func (e *ExcerptWindowBM) String() string {
	return fmt.Sprintf("<ExcerptWindowBM(charlength: %v|bytes: %v): starts at: %v score: %v>:\n%s\n", e.CharLength, e.ByteLength, e.Start, e.Score, e.Text)
}

func (e *ExcerptWindowBM) AddMatch(m *Match) bool {
	if len(e.Matches) == 0 {
		e.Start = m.Start
	} else {
		if m.Start > e.Start+e.ByteLength {
			return false
		}
	}
	e.Matches = append(e.Matches, m)
	e.Score += m.Score
	return true
}

func (e *ExcerptWindowBM) RemoveFirstMatch() {
	if len(e.Matches) > 1 {
		e.Score -= e.Matches[0].Score
		e.Matches = e.Matches[1:]
		e.Start = e.Matches[0].Start
	} else {
		e.Matches = []*Match{}
		e.Start = 0
		e.Score = 0
	}
	return
}

func (e *ExcerptWindowBM) AdjustWindow(body *strings.Reader) {
	var bufSize int
	body.Seek(int64(e.Matches[0].Start), 0)
	var rc uint32 = 0
	for rc < e.CharLength {
		_, size, _ := body.ReadRune()
		bufSize += size
		rc += 1
	}
	e.ByteLength = uint32(bufSize)
	for i := len(e.Matches) - 1; i > 1; i-- {
		if e.Matches[i].Start > e.Start+e.ByteLength {
			e.Score -= e.Matches[i].Score
			e.Matches = e.Matches[:i]
		} else {
			break
		}
	}
	if e.Matches[len(e.Matches)-1].Start+e.Matches[len(e.Matches)-1].ByteLength > e.Start+e.ByteLength {
		e.ByteLength = (e.Matches[len(e.Matches)-1].Start + e.Matches[len(e.Matches)-1].ByteLength) - e.Start
	}
}

func (e *ExcerptWindowBM) MaterializeWindow(body *strings.Reader) {
	var buffer bytes.Buffer
	var bc int = 0
	body.Seek(int64(e.Matches[0].Start), 0)
	for bc < int(e.ByteLength) {
		b, _ := body.ReadByte()
		buffer.WriteByte(b)
		bc += 1
	}
	e.Text = buffer.String()
}

type Uint32Slice []uint32

func (p Uint32Slice) Len() int           { return len(p) }
func (p Uint32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Uint32s sorts a slice of uint32s in increasing order.
func SortUint32s(a []uint32) { sort.Sort(Uint32Slice(a)) }

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
func FindExcerptsBM(searchterms map[string]float64, body string, eLength int, expand bool, findHighestScore bool) (excerptCandidates []*ExcerptWindowBM) {
	startTime := time.Now()
	var excerptLength uint32 = uint32(eLength)
	var offsets []uint32
	var channelkeys []string
	var sortBuffers = make(map[string][]uint32)
	var outSorted []uint32
	var flushAllItems bool
	var flush bool = false
	var maximumFlushOffset uint32 = 0
	var minFlushCount, minFlushCountDown int = eLength, eLength
	scores := make(map[uint32]*TermScore)
	offsetChannels := make(map[string]<-chan substr.Result)
	finishedOffsetChannels := make(map[string]bool)
	termScores := make(map[string]*TermScore, len(searchterms))
	offsetSink := make(chan uint32)
	bodyReader := strings.NewReader(body)

	for term, weight := range searchterms {
		offsetChannels[term] = substr.IndexesWithinReaderStr(strings.NewReader(body), term)
		termScores[term] = &TermScore{
			Score:      float64(len([]rune(term))) * weight,
			ByteLength: uint32(len([]byte(term))),
		}
		finishedOffsetChannels[term] = false
		channelkeys = append(channelkeys, term)
	}

	go func() {
		for {
			for i, term := range channelkeys {
				if finishedOffsetChannels[term] == true && len(sortBuffers[term]) == 0 {
					continue
				}
				o, open := <-offsetChannels[term]
				if !open {
					finishedOffsetChannels[term] = true
					if len(sortBuffers[term]) == 0 {
						if len(channelkeys) > 1 {
							channelkeys = append(channelkeys[:i], channelkeys[i+1:]...)
						} else {
							channelkeys = []string{}
						}
						continue
					} else {
						flush = true
					}
				} else {
					sortBuffers[term] = append(sortBuffers[term], o.Offset)
					minFlushCountDown -= 1

					scores[o.Offset] = termScores[term]

					if minFlushCountDown > 0 && flush == false {
						continue
					}

					var smallestFirst uint32 = 0
					for _, key := range channelkeys {
						if len(sortBuffers[key]) == 0 {
							flush = false
							continue
						}
						if smallestFirst == 0 || sortBuffers[key][0] < smallestFirst {
							smallestFirst = sortBuffers[key][0]
						}
					}
					if smallestFirst > 0 {
						maximumFlushOffset = smallestFirst
					}
				}

				if flush == true {
					outSorted = []uint32{}

					for c, cKey := range channelkeys {
						flushAllItems = true
						for n, offs := range sortBuffers[cKey] {
							if offs > maximumFlushOffset {
								outSorted = append(outSorted, sortBuffers[cKey][0:n]...)
								sortBuffers[cKey] = sortBuffers[cKey][n:]
								flushAllItems = false
								break
							}
						}
						if flushAllItems == true {
							outSorted = append(outSorted, sortBuffers[cKey]...)
							sortBuffers[cKey] = []uint32{}
						}
						if len(sortBuffers[cKey]) == 0 && finishedOffsetChannels[cKey] == true {
							channelkeys = append(channelkeys[:c], channelkeys[c+1:]...)
						}
					}

					SortUint32s(outSorted)
					for _, soffset := range outSorted {
						offsetSink <- soffset
					}

					// reset the flags
					var setMax bool = true
					for _, cKey := range channelkeys {
						if len(sortBuffers[cKey]) == 0 && finishedOffsetChannels[cKey] == false {
							setMax = false
							break
						}
					}
					if setMax == true {
						// smallestFirstTimeStart := time.Now()
						var smallestFirst uint32 = 0
						for _, key := range channelkeys {
							if smallestFirst == 0 || sortBuffers[key][0] < smallestFirst {
								smallestFirst = sortBuffers[key][0]
							}
						}
						if smallestFirst > 0 {
							maximumFlushOffset = smallestFirst
						}
						// SmallestFirstTimeFlush += time.Since(smallestFirstTimeStart)
					}

					flush = false
					minFlushCountDown = minFlushCount
				}

			}
			if len(channelkeys) == 0 {
				close(offsetSink)
				log.Printf("finding matches took: %v.\n", time.Since(startTime))
				break
			}
		}
	}()

	if len(scores) == 0 {
		scores[0] = &TermScore{
			Score:      0,
			ByteLength: 0,
		}
	}

	// now find the highest scoring window
	var highestScoreWindow *ExcerptWindowBM
	var currentWindow *ExcerptWindowBM

	currentWindow = &ExcerptWindowBM{
		Start:      0,
		CharLength: excerptLength,
		ByteLength: excerptLength * 4,
		Score:      0,
		Matches:    []*Match{},
	}
	highestScoreWindow = &ExcerptWindowBM{
		Start:      0,
		CharLength: excerptLength,
		ByteLength: excerptLength * 4,
		Score:      0,
		Matches:    []*Match{},
	}

	for matchOffset := range offsetSink {
		offsets = append(offsets, matchOffset)
		if matchOffset < currentWindow.Start {
			log.Println("FRAKK! got", matchOffset, "and currentWindow.Start is ", currentWindow.Start)
			log.Println(offsets[len(offsets)-10:])
			os.Exit(0)
		}
		m := &Match{
			Start:      matchOffset,
			Score:      scores[matchOffset].Score,
			ByteLength: scores[matchOffset].ByteLength,
		}
		for {
			if currentWindow.AddMatch(m) == false {
				currentWindow.RemoveFirstMatch()
				continue
			}
			break
		}
		if currentWindow.Score > highestScoreWindow.Score {
			currentWindow.AdjustWindow(bodyReader)
			if currentWindow.Score < highestScoreWindow.Score {
				continue
			}
			highestScoreWindow.Start = currentWindow.Start
			highestScoreWindow.ByteLength = currentWindow.ByteLength
			highestScoreWindow.Score = currentWindow.Score
			highestScoreWindow.Matches = currentWindow.Matches
		}
	}

	// if we have no match we just send and excerpt that starts at the beginning
	if len(offsets) == 0 {
		highestScoreWindow.Text = "no match"
	} else {
		highestScoreWindow.MaterializeWindow(bodyReader)
	}

	log.Printf("total time took: %v. nr match positions: %v\n", time.Since(startTime), len(scores))
	// log.Println("BestMatch at:", highestScoreWindow.Start, "Score", highestScoreWindow.Score)
	// log.Println(highestScoreWindow.Text)

	excerptCandidates = append(excerptCandidates, highestScoreWindow)
	return
}
