package excerpt

import (
	"github.com/fvbock/substr/src/substr"
	// "log"
	"strings"
	"sync"
	// "time"
)

const (
	END_PUNCT_CHAR = ":：.。;?？¿!！¡。…"
)

/*
FindExcerpts searches searchterms in body and returns the highest scoring
excerpt of a given length that contains the terms.
*/
func FindBestExcerptBM(searchterms map[string]float64, body string, eLength int,
	prependChars uint32, prependFullwords bool) (e *ExcerptWindowBM) {
	return FindExcerptsBM(searchterms, body, eLength, true, prependChars, prependFullwords)[0]
}

/*
FindExcerpts searches searchterms in body and returns excerpts of a given
length that contain the terms.

The terms can be weighted and each ExcerptWindow gets a cummulative score
of the searchterms it contains. setting findHighestScore the function will
only return the ExcerptWindow with the hightest score.

If a match is at the end of a window and overlaps its boundry the window
will be extended to include the full match.

If prepend option is set the excerpt will be prepended with a maximum of
the set number if characters. If prependFullwords option is set the prepended
text will go back to the last white space character when it hit the character
limit. this makes sense for languages that separate words with white space
characters and does make less sense for languages like chinese or japanese
that don't.
*/
func FindExcerptsBM(searchterms map[string]float64, body string, eLength int,
	findHighestScore bool, prependChars uint32, prependFullwords bool) (excerptCandidates []*ExcerptWindowBM) {
	// startTime := time.Now()
	var excerptLength uint32 = uint32(eLength)
	var offsets []uint32
	var channelkeys []string
	var sortBuffers = make(map[string][]uint32)
	var outSorted []uint32
	var flushAllItems bool
	var flush bool = false
	var maximumFlushOffset uint32 = 0
	var minFlushCount, minFlushCountDown int = eLength, eLength
	var scores = struct {
		sync.RWMutex
		s map[uint32]*TermScore
	}{s: make(map[uint32]*TermScore)}
	offsetChannels := make(map[string]<-chan substr.Result)
	finishedOffsetChannels := make(map[string]bool)
	termScores := make(map[string]*TermScore, len(searchterms))
	offsetSink := make(chan uint32)

	bodyReader := strings.NewReader(body)

	for term, weight := range searchterms {
		offsetChannels[term] = substr.IndexesWithinReaderStr(strings.NewReader(body), strings.ToLower(term))
		termScores[term] = &TermScore{
			Score:      float64(len([]rune(term))) * weight,
			ByteLength: uint32(len([]byte(term))),
		}
		finishedOffsetChannels[term] = false
		channelkeys = append(channelkeys, term)
	}
	go func() {
	fanin:
		for {
			for i, term := range channelkeys {
				if finishedOffsetChannels[term] == true && len(sortBuffers[term]) == 0 {
					continue
				}
				o, open := <-offsetChannels[term]
				if !open {
					finishedOffsetChannels[term] = true
					if len(sortBuffers[term]) == 0 {
						// log.Printf("Remove %s (%v) from channelkeys: %v", term, i, channelkeys)
						if len(channelkeys) > 1 {
							channelkeys = append(channelkeys[:i], channelkeys[i+1:]...)
						} else {
							channelkeys = []string{}
						}
						// log.Printf("OK. removed: %s (%v). channelkeys: %v", term, i, channelkeys)
						continue fanin
					} else {
						flush = true
					}
				} else {
					scores.Lock()
					// do only count the highest scored match at a position
					// this collision can happen if a word and it's stemmed
					// version both get passed into the algorithm
					if _, haveMatch := scores.s[o.Offset]; haveMatch {
						if scores.s[o.Offset].Score < termScores[term].Score {
							// we had a previous match with a lower score.
							// we have to remove the previous match from the
							// sortbuffer of the lower scored terms and replace
							// the termscore at the current offset
							for _, otherterm := range channelkeys {
								// can only happen if the other term is a substring of term
								if len(otherterm) < len(term) && otherterm == term[0:len(otherterm)] {
									for oi, otherOffset := range sortBuffers[otherterm] {
										if otherOffset == o.Offset {
											if len(sortBuffers[otherterm]) == oi+1 {
												sortBuffers[otherterm] = sortBuffers[otherterm][:oi]
											} else {
												sortBuffers[otherterm] = append(sortBuffers[otherterm][:oi], sortBuffers[otherterm][oi+1:]...)
											}
											break
										}
									}
								}
							}
							sortBuffers[term] = append(sortBuffers[term], o.Offset)
							// minFlushCountDown -= 1 // is this needed? check!
							scores.s[o.Offset] = termScores[term]
						}
					} else {
						sortBuffers[term] = append(sortBuffers[term], o.Offset)
						minFlushCountDown -= 1
						scores.s[o.Offset] = termScores[term]
					}
					scores.Unlock()

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
						var smallestFirst uint32 = 0
						for _, key := range channelkeys {
							if smallestFirst == 0 || sortBuffers[key][0] < smallestFirst {
								smallestFirst = sortBuffers[key][0]
							}
						}
						if smallestFirst > 0 {
							maximumFlushOffset = smallestFirst
						}
					}

					flush = false
					minFlushCountDown = minFlushCount
				}

			}
			if len(channelkeys) == 0 {
				// log.Printf("extraction runtime: %v\n", time.Since(startTime))
				close(offsetSink)
				break
			}
		}
	}()

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
		scores.RLock()
		m := &Match{
			Start:      matchOffset,
			Score:      scores.s[matchOffset].Score,
			ByteLength: scores.s[matchOffset].ByteLength,
		}
		scores.RUnlock()
		for {
			if currentWindow.AddMatch(m) == false {
				currentWindow.RemoveFirstMatch()
				continue
			}
			break
		}
		if currentWindow.Score > highestScoreWindow.Score {
			currentWindow.AdjustWindow(bodyReader, prependChars, prependFullwords)
			if currentWindow.Score < highestScoreWindow.Score {
				continue
			}
			highestScoreWindow.Start = currentWindow.Start
			highestScoreWindow.ByteLength = currentWindow.ByteLength
			highestScoreWindow.Score = currentWindow.Score
			highestScoreWindow.Matches = currentWindow.Matches
		}
		if findHighestScore == false {
			w := &ExcerptWindowBM{
				Start:      currentWindow.Start,
				CharLength: currentWindow.CharLength,
				ByteLength: currentWindow.ByteLength,
				Score:      currentWindow.Score,
				Matches:    currentWindow.Matches,
			}
			w.AdjustWindow(bodyReader, prependChars, prependFullwords)
			w.MaterializeWindow(bodyReader)
			excerptCandidates = append(excerptCandidates, w)
		}
	}
	if findHighestScore == true {
		// catch zero match case
		if len(offsets) == 0 {
			highestScoreWindow.AddMatch(&Match{
				Start:      0,
				Score:      0,
				ByteLength: highestScoreWindow.CharLength,
			})
			highestScoreWindow.AdjustWindow(bodyReader, 0, false)
		}

		highestScoreWindow.MaterializeWindow(bodyReader)
		excerptCandidates = append(excerptCandidates, highestScoreWindow)
	}
	// log.Printf("runtime: %v\n", time.Since(startTime))
	return
}
