package srt

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	BOM      = "\ufeff"
	zeroNano = -6826986978871345152
)

// Subtitle represents a single SRT entry
type Subtitle struct {
	// The cardinality of this subtitle, starting at 1
	Number int
	// The time at which the subtitle appears on screen
	Start time.Duration
	// The time at which the subtitle leaves the screen
	End time.Duration
	// The content of the subtitle
	Text string
}

// Reader reads Subtitles from an SRT stream
type Reader struct {
	number  int
	scanner *bufio.Scanner
}

func NewReader(r io.Reader) *Reader {
	return &Reader{scanner: bufio.NewScanner(r)}
}

func (r *Reader) ReadSubtitle() (*Subtitle, error) {
	var subtitle Subtitle
	var index int
	for r.scanner.Scan() {
		text := r.scanner.Text()
		// Sometimes SRT files contain byte order marks to indicate UTF-8
		if len(text) >= 3 && text[:3] == BOM {
			text = text[3:]
		}
		switch index {
		case 0:
			// Be tolerant of extra newlines
			if len(text) == 0 {
				continue
			}
			n, err := strconv.Atoi(text)
			if err != nil {
				return nil, fmt.Errorf("srt: expected a sequence number at %q", text)
			}
			subtitle.Number = n
		case 1:

			times := strings.Split(text, " --> ")
			if len(times) != 2 {
				return nil, fmt.Errorf("srt: expected start and end time at %q", text)
			}
			// Ugh Go time parsing can't handle sub-seconds with commas
			t1, t2 := strings.Replace(times[0], ",", ".", -1), strings.Replace(times[1], ",", ".", -1)
			start, err := time.Parse("15:04:05.000", t1)
			if err != nil {
				return nil, fmt.Errorf("srt: invalid time format at %q", times[0])
			}
			end, err := time.Parse("15:04:05.000", t2)
			if err != nil {
				return nil, fmt.Errorf("srt: invalid time format at %q", times[1])
			}
			subtitle.Start = time.Duration(start.UnixNano() - zeroNano)
			subtitle.End = time.Duration(end.UnixNano() - zeroNano)
		case 2:
			subtitle.Text = text
			for r.scanner.Scan() {
				next := r.scanner.Text()
				if next == "" {
					return &subtitle, nil
				}
				subtitle.Text += "\n" + next
			}
			if err := r.scanner.Err(); err != nil {
				return nil, err
			}
			return &subtitle, nil
		}
		index++
	}
	if err := r.scanner.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}
