// Copyright 2017 Joel Scoble
// Licensed under the MIT License;
// you may not use this file except in compliance with the License.
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This file includes code from Go's unicode package. Information about which
// file from which the code is copied is local to the actual code in the file.
// This is the original copyright notice for the relevant code:
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linewrap

import (
	"bytes"
	"io"
	"strings"
	"unicode"
)

var (
	LineLength = 80
	TabSize    = 5
	NewLine    = []byte("\n")
)

// Wrap processes strings into wrapped lines. If the wrapped lines are indented,
// the whitespace character at the wrap point is elided and replaced with the
// IndentVal.
type Wrap struct {
	Length int // Max length of the line.
	// Number of chars that should be added for each tab encountered.
	TabSize   int
	Indent    bool   // Indent wrapped lines
	IndentVal string // The string used to indent wrapped lines
	indentLen int    // The number of chars in IndnetVal; this accounts for tabs.
	// The new line sequence to use. If this isn't created with New, which
	// sets it to '\n', it must be set by the user.
	NewLine []byte
	r       strings.Reader
	runes   []rune // line buffer
	buf     bytes.Buffer
	l       int // the length of the current line
}

// New returns a new Wrap with default Length and TabWidth.
func New() Wrap {
	return Wrap{Length: LineLength, TabSize: TabSize, NewLine: NewLine}
}

// Reset's the wrapper and sets its reader to the string to be wrapped.
func (w *Wrap) reset(s string) {
	w.r.Reset(s)
	w.buf.Reset()
	w.runes = w.runes[:0]
}

// Line inserts a new line at Length. If the position is a non-Unicode space
// character, the new line is inserted at the position of the last space
// character. New line sequences in the text will be replaced with the
// Wrap.NewLine sequence.
//
// If the line length boundary occurs within a sequence of white space chars,
// there is a new line sequence within the whitespace sequence, and the
// sequence of whitespaces preceeding the new line would exceed the desired
// line length, those whitespace chars are allowed to spill over the line
// length to prevent a new line, a sequence of spaces, and another new line
// from occurring. If wrapped lines are to be indented any whitespace chars
// after the newline and prior to non-whitespace chars are elided and replaced
// with the IndentVal. This will result in loss of some white space chars. If
// this is undesirable behavior, set Indent to false.
//
// The resulting string is returned. If an error occurs, both the original
// string and the error are returned.
func (w Wrap) Line(s string) (string, error) {
	if s == "" { // if the string is empty, no comment
		return s, nil
	}
	w.reset(s)

	if w.Indent { // If indenting lines; figure out the actual indent width.
		for _, v := range w.IndentVal {
			if v == '\t' {
				w.indentLen += w.TabSize
				continue
			}
			w.indentLen++
		}
	}

	// Whether or not the chunk is unicode spaces. This starts as true because
	// the bool is negated at the top of the loop and we assume that it starts
	// with chars and not whitespaces.
	space := true
	for {
		space = !space // flip what we are looking for
		cerr := w.chunk(space)
		if cerr != nil && cerr != io.EOF {
			return s, cerr
		}
		if !space {
			if w.l+len(w.runes) >= w.Length { // if adding this chunk would exceed line length; emit a newline
				err := w.newLine()
				if err != nil {
					return s, err
				}
			}
		} else {
			// Process the space chunk; any new line sequences in this chunk will
			// be respected. The actual new line sequnce gets replaced with the
			// configured sequence.
			err := w.processSpaceChunk()
			if err != nil {
				return s, err
			}

			// account for tabs
			if len(w.runes) > 0 {
				for _, v := range w.runes {
					if v == '\t' {
						w.l += w.TabSize
					}
				}
			}
		}

		_, err := w.buf.WriteString(string(w.runes))
		if err != nil {
			return s, err
		}
		// If the last chunk processed ended with an io.EOF, we're done.
		if cerr != nil && cerr == io.EOF {
			return w.buf.String(), nil // If EOF was reached, return what we have.
		}
		w.l += len(w.runes)
	}
}

// Write out the new line + indent, if applicable.
func (w *Wrap) newLine() error {
	w.l = 0 // a newline results in reseting the line counter
	_, err := w.buf.Write(w.NewLine)
	if err != nil {
		return err
	}
	if w.Indent {
		w.buf.WriteString(w.IndentVal)
		if err != nil {
			return err
		}
		w.l = w.indentLen
	}
	return nil
}

// chunk gets a chunk of chars; either whitespace or non-whitespace.
func (w *Wrap) chunk(space bool) error {
	// reset the rune cache
	w.runes = w.runes[:0]
	if space {
		return w.spaces()
	}
	return w.word()
}

// a word is a series of runes until a separator char is encountered
func (w *Wrap) word() error {
	for {
		ch, _, err := w.r.ReadRune()
		if err != nil {
			return err
		}
		if isSpace(ch) {
			// back up because we only return the word, not the separator
			err = w.r.UnreadRune()
			if err != nil {
				return err
			}
			return nil
		}
		w.runes = append(w.runes, ch)
	}
}

// spaces returns all space characters between two words; unicode.IsSpace is used to
// evaluate if the rune is a space.
func (w *Wrap) spaces() error {
	for {
		ch, _, err := w.r.ReadRune()
		if err != nil {
			return err
		}
		if !isSpace(ch) {
			// back up because we only return the spaces, not non-Space chars
			err = w.r.UnreadRune()
			if err != nil {
				return err
			}
			return nil
		}
		// should actually be considered a space
		w.runes = append(w.runes, ch)
	}
}

// process a chunk of runes that are a series of whitespace characters.
func (w *Wrap) processSpaceChunk() error {
	var nl bool
	for i, r := range w.runes {
		if r == '\n' {
			nl = true
			var winLine bool
			// write the preceeding chars
			if i > 0 && w.runes[i-1] == '\r' { // back up index to elide the \r
				winLine = true
				i--
			}
			_, err := w.buf.WriteString(string(w.runes[:i]))
			if err != nil {
				return err
			}

			err = w.newLine()
			if err != nil {
				return err
			}
			if !w.Indent {
				if winLine {
					i++ // move forward the index to account for the prior back up.
				}
				if i == len(w.runes) {
					w.runes = w.runes[:0]
				} else {
					w.runes = w.runes[i+1:]
				}
			} else {
				w.runes = w.runes[:0]
			}
			//return w.processSpaceChunk()
		}
	}

	if !nl { // if there wasn't a nl see if this chunk exceeds the line
		if w.l+len(w.runes) >= w.Length {
			err := w.newLine()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// isSpace corrects (from the perspective of this package) some invalid
// evaluations of unicode.IsSpace.
func isSpace(r rune) bool {
	// check exceptions. some unicode spaces don't evaluate to true
	if r == '\u200b' {
		return true
	}
	// check exceptions, no break spaces are spaces but we don't
	// treat them as spaces since one shouldn't line break on them
	if r == '\u00A0' || r == '\u202F' {
		return false
	}
	return unicode.IsSpace(r)
}

// isHyphen checks to see if the rune is a unicode hyphen. The following hypens
// evaluate to false for the purposes of linewrap:
//   * U+2011 non-breaking hyphen
//   * U+207B superscript minus
//   * U+208B subscript minus
//
// This code is based on
//   https://golang.org/src/unicode/letter.go?h=isExcludingLatin#L170
// See the Go Authors copyright notice found at the top of this file.
func isHyphen(r rune) bool {
	// this handles ASCII
	if uint32(r) <= unicode.MaxLatin1 {
		switch r {
		case '\u002d', '\u007e', '\u00ad':
			return true
		default:
			return false
		}
	}
	// Dashes that we don't consider dash
	switch r {
	case '\u2011', '\u207B', '\u208B':
		return false
	}

	dashTab := unicode.Dash.R16
	if off := unicode.Dash.LatinOffset; len(dashTab) > off && r <= rune(dashTab[len(dashTab)-1].Hi) {
		return is16(dashTab[off:], uint16(r))
	}
	return false
}

// This code is based on
//   https://golang.org/src/unicode/letter.go?h=isExcludingLatin#L170
// See the Go Authors copyright notice found at the top of this file.

// linearMax is the maximum size table for linear search for non-Latin1 rune.
// Derived by running 'go test -calibrate'.
const linearMax = 18

// is16 reports whether r is in the sorted slice of 16-bit ranges.
func is16(ranges []unicode.Range16, r uint16) bool {
	if len(ranges) <= linearMax || r <= unicode.MaxLatin1 {
		for i := range ranges {
			range_ := &ranges[i]
			if r < range_.Lo {
				return false
			}
			if r <= range_.Hi {
				return (r-range_.Lo)%range_.Stride == 0
			}
		}
		return false
	}

	// binary search over ranges
	lo := 0
	hi := len(ranges)
	for lo < hi {
		m := lo + (hi-lo)/2
		range_ := &ranges[m]
		if range_.Lo <= r && r <= range_.Hi {
			return (r-range_.Lo)%range_.Stride == 0
		}
		if r < range_.Lo {
			hi = m
		} else {
			lo = m + 1
		}
	}
	return false
}
