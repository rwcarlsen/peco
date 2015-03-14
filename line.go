package peco

import (
	"regexp"
	"strings"

	"github.com/nsf/termbox-go"
)

// Global var used to strips ansi sequences
var reANSIEscapeChars = regexp.MustCompile("\x1B\\[(?:[0-9]{1,2}(?:;[0-9]{1,2})?)*[a-zA-Z]")

// Function who strips ansi sequences
func stripANSISequence(s string) (clean string, fg, bg []termbox.Attribute) {
	bg = make([]termbox.Attribute, len(clean))
	fg = make([]termbox.Attribute, len(clean))
	clean = s
	for {
		start := strings.Index(clean, string(0x1b)+"[")
		if start == -1 {
			break
		}
		end := strings.Index(clean[start:], "m")
		if end == -1 {
			panic("unterminated escape sequence") // TODO: how to handle this?
		}
		end += start + 1
		for i := end; i < len(clean); i++ {
			fg[i], bg[i] = parseAttrib(fg[i], bg[i], clean[start+2:end-1])
		}
		clean = clean[:start] + clean[end:]
		fg = append(append([]termbox.Attribute{}, fg[:start]...), fg[end:]...)
		bg = append(append([]termbox.Attribute{}, bg[:start]...), bg[end:]...)
	}
	return clean, fg, bg
}

// Line defines the interface for each of the line that peco uses to display
// and match against queries. Note that to make drawing easier,
// we have a RawLine and MatchedLine types
type Line interface {
	Buffer() string                        // Raw buffer, may contain null
	DisplayString() string                 // Line to be displayed
	Output() string                        // Output string to be displayed after peco is done
	Indices() [][]int                      // If the type allows, indices into matched portions of the string
	Attribs() (fg, bg []termbox.Attribute) // If the type allows, indices into matched portions of the string
}

// baseLine is the common implementation between RawLine and MatchedLine
type baseLine struct {
	buf           string
	sepLoc        int
	displayString string
	fg, bg        []termbox.Attribute
}

func newBaseLine(v string, enableSep bool) *baseLine {
	m := &baseLine{
		v,
		-1,
		"",
		nil, nil,
	}
	if !enableSep {
		return m
	}

	// XXX This may be silly, but we're avoiding using strings.IndexByte()
	// here because it doesn't exist on go1.1. Let's remove support for
	// 1.1 when 1.4 comes out (or something)
	for i := 0; i < len(m.buf); i++ {
		if m.buf[i] == '\000' {
			m.sepLoc = i
		}
	}
	return m
}

func (m baseLine) Buffer() string {
	return m.buf
}

func (m baseLine) Attribs() (fg, bg []termbox.Attribute) {
	return m.fg, m.bg
}

func (m baseLine) DisplayString() string {
	if m.displayString != "" {
		return m.displayString
	}

	if i := m.sepLoc; i > -1 {
		m.displayString, m.fg, m.bg = stripANSISequence(m.buf[:i])
	} else {
		m.displayString, m.fg, m.bg = stripANSISequence(m.buf)
	}
	return m.displayString
}

func (m baseLine) Output() string {
	if i := m.sepLoc; i > -1 {
		return m.buf[i+1:]
	}
	return m.buf
}

// RawLine implements the Line interface. It represents a line with no matches,
// which means that it can only be used in the initial unfiltered view
type RawLine struct {
	*baseLine
}

// NewRawLine creates a RawLine struct
func NewRawLine(v string, enableSep bool) *RawLine {
	return &RawLine{newBaseLine(v, enableSep)}
}

// Indices always returns nil
func (m RawLine) Indices() [][]int {
	return nil
}

// MatchedLine contains the actual match, and the indices to the matches
// in the line
type MatchedLine struct {
	*baseLine
	matches [][]int
}

// NewMatchedLine creates a new MatchedLine struct
func NewMatchedLine(v string, enableSep bool, m [][]int) *MatchedLine {
	return &MatchedLine{newBaseLine(v, enableSep), m}
}

// Indices returns the indices in the buffer that matched
func (d MatchedLine) Indices() [][]int {
	return d.matches
}

func parseAttrib(currfg, currbg termbox.Attribute, esc string) (fg, bg termbox.Attribute) {
	fg, bg = currfg, currbg

	codes := strings.Split(esc, ";")
	for _, code := range codes {
		switch code {
		case "30":
			fg &= 16 // zero color bits
			fg |= termbox.ColorBlack
		case "31":
			fg &= 16 // zero color bits
			fg |= termbox.ColorRed
		case "32":
			fg &= 16 // zero color bits
			fg |= termbox.ColorGreen
		case "33":
			fg &= 16 // zero color bits
			fg |= termbox.ColorYellow
		case "34":
			fg &= 16 // zero color bits
			fg |= termbox.ColorBlue
		case "35":
			fg &= 16 // zero color bits
			fg |= termbox.ColorMagenta
		case "36":
			fg &= 16 // zero color bits
			fg |= termbox.ColorCyan
		case "37":
			fg &= 16 // zero color bits
			fg |= termbox.ColorWhite
		case "39":
			fg &= 16 // zero color bits
			fg |= termbox.ColorDefault
		case "40":
			bg &= 16 // zero color bits
			bg |= termbox.ColorBlack
		case "41":
			bg &= 16 // zero color bits
			bg |= termbox.ColorRed
		case "42":
			bg &= 16 // zero color bits
			bg |= termbox.ColorGreen
		case "43":
			bg &= 16 // zero color bits
			bg |= termbox.ColorYellow
		case "44":
			bg &= 16 // zero color bits
			bg |= termbox.ColorBlue
		case "45":
			bg &= 16 // zero color bits
			bg |= termbox.ColorMagenta
		case "46":
			bg &= 16 // zero color bits
			bg |= termbox.ColorCyan
		case "47":
			bg &= 16 // zero color bits
			bg |= termbox.ColorWhite
		case "49":
			bg &= 16 // zero color bits
			bg |= termbox.ColorDefault
		case "0", "": // reset all attribs
			bg = termbox.ColorDefault
			fg = termbox.ColorDefault
		}
	}

	return fg, bg
}
