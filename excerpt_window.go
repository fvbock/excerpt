package excerpt

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

type TermScore struct {
	Score      float64
	ByteLength uint32
}

func (t *TermScore) String() string {
	return fmt.Sprintf("<TermScore: %v Length: %v>", t.Score, t.ByteLength)
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

func (e *ExcerptWindowBM) AdjustWindow(body *strings.Reader, prependChars uint32,
	prependFullwords bool) {
	var bufSize int
	body.Seek(int64(e.Matches[0].Start), 0)
	// i am not sure why but i get positions that were not rune starts in
	// multibyte character texts. until the reason is understood i will have to
	// adjust the start position here...
	var moveOffsets uint32 = 0
	for {
		b, err := body.ReadByte()
		if err == io.EOF {
			break
		}
		if utf8.RuneStart(b) {
			break
		}
		moveOffsets += 1
		body.Seek(-2, 1)
	}
	if moveOffsets > 0 {
		e.Start -= moveOffsets
		e.ByteLength += moveOffsets
		for n, _ := range e.Matches {
			e.Matches[n].Start -= moveOffsets
		}
	}
	body.Seek(int64(e.Matches[0].Start), 0)
	var rc uint32 = 0
	for rc < e.CharLength {
		_, size, err := body.ReadRune()
		if err == io.EOF {
			// bufSize -= 1
			break
		}
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
	if prependChars == 0 {
		return
	}

	// extend window to show more context at the beginning of the excerpt
	if e.Start == 0 {
		return
	}
	if e.Start < prependChars {
		e.ByteLength += e.Start
		e.Start = 0
		return
	}

	var moveStart uint32 = 0
	var lastWhiteSpace uint32 = 0
	var lastWhiteSpaceLength int = 0

	body.Seek(int64(e.Start)-1, 0)
	for {
		b, err := body.ReadByte()
		if err == io.EOF {
			break
		}
		if !utf8.RuneStart(b) {
			body.Seek(-2, 1)
			continue
		}

		body.Seek(-1, 1)
		r, rs, err := body.ReadRune()
		if err == io.EOF {
			break
		}
		moveStart += uint32(rs)
		// log.Println("read rune:", string(r), "rs:", rs, "moveStart", moveStart)
		if unicode.IsSpace(r) {
			lastWhiteSpace = moveStart
			lastWhiteSpaceLength = rs
		}

		if strings.Index(END_PUNCT_CHAR, string(r)) != -1 {
			// log.Println("found END_PUNCT_CHAR")
			moveStart -= uint32(rs)
			if moveStart == lastWhiteSpace {
				moveStart -= uint32(lastWhiteSpaceLength)
			}
			break
		}

		if moveStart >= prependChars {
			// log.Println("moveStart >= prependChars")
			if !unicode.IsSpace(r) && prependFullwords {
				moveStart = lastWhiteSpace - uint32(lastWhiteSpaceLength)
			}
			break
		}
		spos := int64(-(rs + 1))
		body.Seek(spos, 1)
	}
	e.Start -= moveStart
	e.ByteLength += moveStart
}

func (e *ExcerptWindowBM) MaterializeWindow(body *strings.Reader) {
	var buffer bytes.Buffer
	var bc int = 0

	// check that the last rune is not cut off
	body.Seek(int64(e.Start+e.ByteLength-1), 0)
	var lastRune int = 1
	for {
		b, err := body.ReadByte()
		if err == io.EOF {
			break
		}
		if !utf8.RuneStart(b) {
			body.Seek(-2, 1)
			lastRune += 1
			continue
		} else {
			break
		}
	}

	body.Seek(-1, 1)
	_, rs, _ := body.ReadRune()
	// log.Println("lastRune", lastRune, "rs", rs)
	if lastRune != rs {
		// log.Println("lastRune != rs", lastRune, rs, "add", rs-lastRune, "bytes")
		e.ByteLength += uint32(rs - lastRune)
	}

	// body.Seek(int64(e.Matches[0].Start), 0)
	body.Seek(int64(e.Start), 0)
	for bc < int(e.ByteLength) {
		b, err := body.ReadByte()
		if err == io.EOF {
			break
		}
		buffer.WriteByte(b)
		bc += 1
	}

	e.Text = strings.TrimSpace(buffer.String())
	e.ByteLength = uint32(len(e.Text))
	// log.Println(e.Text)
}
