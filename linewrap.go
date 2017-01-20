package linewrap

import (
	"bytes"
	"fmt"
	"unicode"
)

var (
	LineLength = 80
	TabWidth   = 5
)

// Wrap processes strings into wrapped lines. If the wrapped lines are indented,
// the whitespace character at the wrap point is elided and replaced with the
// IndentVal.
type Wrap struct {
	Length int // Max length of the line.
	// Number of chars that should be added for each tab encountered.
	TabWidth  int
	Indent    bool   // Indent wrapped lines
	IndentVal string // The string used to indent wrapped lines
	buf       bytes.Buffer
}

// New returns a new Wrap with default Length and TabWidth.
func New() Wrap {
	return Wrap{Length: LineLength, TabWidth: TabWidth}
}

// Line inserts a new line at Length. If the position is a non-Unicode space
// character, the new line is inserted at the position of the last space
// character. The resulting string is returned. If an error occurs, both the
// original string and the error are returned.
func (w Wrap) Line(s string) (string, error) {
	if s == "" { // if the string is empty, no comment
		return s, nil
	}

	var (
		r []rune // line buffer
		l int    // current line length in characters
		j int    // index of the last space in
	)

	w.buf.Reset()

	// range through the runes in s
	for i, v := range s {
		_ = i
		if l > w.Length {
			// only do if space was encountered
			if j != 0 {
				var isNL, isRN bool
				// if the last space char was a new line we might need this info in
				// the future
				if r[j-1] == '\n' {
					isNL = true
				} else if r[j-1] == '\r' { // detect /r/n where /r is on the split boundary
					fmt.Printf("%x\n", r)
					j++ // skip to \n
					fmt.Printf("%x\n", r[j-1])
					isRN = true
				}
				fmt.Printf("%q\t%q\n", r[j-1], string(r[:j]))

				var err error
				// if indenting; don't use the last whitespace, unless it was a \n
				_, err = w.buf.WriteString(string(r[:j]))
				if err != nil {
					return s, err
				}
				if w.Indent && (isNL || isRN) && j < len(r) && unicode.IsSpace(r[j]) {
					j++
				}
				if !isNL && !isRN {
					err = w.buf.WriteByte('\n')
					if err != nil {
						return s, err
					}
				}
				if w.Indent {
					_, err = w.buf.WriteString(w.IndentVal)
					if err != nil {
						return s, err
					}
				}
				if j <= len(r) {
					r = r[j:]

				} else {
					r = r[:0] // if the line happened on a space boundary, the next line starts out empty
				}
				l = len(r) // keep track of number of chars already in the next line
				j = 0      // reset space tracker

				// Skip the current char being processed if the space was at the end of the line
				// and we are indenting and the current character being processed is a space.
				if l == 0 && w.Indent && unicode.IsSpace(v) {
					continue
				}
			}
		}
		if unicode.IsSpace(v) {
			j = l // set space index to current
		}
		r = append(r, v) // add the rune to the current line.
		l++              // increment the character count for the current line
	}
	_, err := w.buf.WriteString(string(r))
	if err != nil {
		return s, err
	}
	return w.buf.String(), nil
}
