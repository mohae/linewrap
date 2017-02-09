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
	"unicode"
	"unicode/utf8"
)

const (
	zeroNBSP      = '\ufeff' // Zero-width no break space
	cr            = '\r'
	lf            = '\n'
	CommentPrefix = "// "
)

var (
	LineLength = 80
	TabSize    = 5
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
	// If the wrapped string should be unwrappable. Unwrappable means all inserted
	// linebreaks can be removed and the unwrapped string will retain all of its
	// original formatting. If Unwrappable, the wrapped text will not be indented.
	// If there was a new line sequence substitution during line wrapping the
	// wrapped new line char(s) will be kept.
	Unwrappable bool
	indentLen   int // The number of chars in IndnetVal; this accounts for tabs.
	// The new line sequence to use. If this isn't created with New, which
	// sets it to '\n', it must be set by the user.
	NewLine string
	// this value may include a zero-width NBSP if Unwrappable. This one is used
	// for the actual insertion of newlines. The exported version is used as the
	// user settable one and also for replacing existing new lines. This
	// differentiation is necessary in the case of Unwrappable to help ensure
	// that the wrapped line is unwrapped properly.
	newLine string
	// lineComment, when set all lines are prefixed with "// ". IndentVal
	// should not be set then LineComment is enabled as it will be set during
	// the setting of lineComment (see LineComment()).
	lineComment bool

	r     bytes.Reader
	runes []rune // line buffer
	newNL bool   //if the last thing done was a nl: for whitespace elision
	buf   bytes.Buffer
	l     int // the length of the current line
}

// New returns a new Wrap with default Length and TabWidth.
func New() Wrap {
	return Wrap{Length: LineLength, TabSize: TabSize, NewLine: string(lf), newLine: string(lf)}
}

// Reset's the wrapper and sets its reader to the string to be wrapped. NewLine,
// Indent, IndentVal, TabSize, and lineComment settings are not affected.
func (w *Wrap) reset(v []byte) {
	w.r.Reset(v)
	w.buf.Reset()
	w.runes = w.runes[:0]
}

// LineComment set whether or not the text to be wrapped should be treated as
// line comments. If enabled, Indent will be set to true and IndentVal will be
// set to the CommentPreifx. If unset, Indent and IndentVal will not be affected
// and the text to be wrapped will not be treated as line comments.
//
// The difference between enabling line comments and setting Indent values is
// that only wrapped lines are indented, the first line is not, while line
// comments prefixes every line with the line comment prefix, e.g. IndentVal.
func (w *Wrap) LineComment(b bool) {
	if b {
		w.Indent = true
		w.IndentVal = CommentPrefix
	}
	w.lineComment = b
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
	b, err := w.Bytes([]byte(s))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Wrap bytes and return the wrapped bytes
func (w *Wrap) Bytes(s []byte) (b []byte, err error) {
	if len(s) == 0 { // if the string is empty, no comment
		return s, nil
	}

	// reset Wrap stuff
	w.reset(s)

	// set the new line chars to be inserted
	if w.Unwrappable {
		w.newLine = string(zeroNBSP) + w.NewLine
	} else {
		w.newLine = w.NewLine
	}

	if w.Indent { // If indenting lines; figure out the actual indent width.
		for _, v := range w.IndentVal {
			if v == '\t' {
				w.indentLen += w.TabSize
				continue
			}
			w.indentLen++
		}
	}

	// if this is being transformed to a line comment; the first line needs to
	// be prefixed with the CommentPrefix and the length counter needs to be
	// updated.
	if w.lineComment {
		n, err := w.buf.WriteString(CommentPrefix)
		if err != nil {
			return s, err
		}
		w.l = n
	}
	// Whether or not the chunk is unicode spaces. This starts as true because
	// the bool is negated at the top of the loop and we assume that it starts
	// with chars and not whitespaces.
	space := true
	for {
		w.runes = w.runes[:0]
		space = !space // flip what we are looking for
		var rerr error // error from reading the runes
		if space {
			// Process the space chunk; any new line sequences in this chunk will
			// be respected. The actual new line sequnce gets replaced with the
			// configured sequence.
			rerr = w.spaces()
			if rerr != nil && rerr != io.EOF {
				return s, rerr
			}

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
			continue
		}

		// process non-whitespace runs
		w.newNL = false // set to false because chars are being processed
		rerr = w.word()
		if rerr != nil && rerr != io.EOF {
			return s, rerr
		}

		if w.l+len(w.runes) >= w.Length { // if adding this chunk would exceed line length; emit a newline
			err := w.writeNewLine()
			if err != nil {
				return s, err
			}
			// since chars will be written next; this flag can be reset as it's for whitespace elision on indent
			w.newNL = false
		}

		_, err := w.buf.WriteString(string(w.runes))
		if err != nil {
			return s, err
		}
		// If the last chunk processed ended with an io.EOF, we're done.
		if rerr != nil && rerr == io.EOF {
			return w.buf.Bytes(), nil // If EOF was reached, return what we have.
		}
		w.l += len(w.runes)
	}
}

// Write out the new line + indent,
func (w *Wrap) writeNewLine() error {
	w.l = 0 // a newline results in reseting the line counter
	w.newNL = true
	_, err := w.buf.WriteString(w.newLine)
	if err != nil {
		return err
	}
	if w.Indent {
		_, err = w.buf.WriteString(w.IndentVal)
		if err != nil {
			return err
		}
		w.l = w.indentLen
	}
	return nil
}

// a word is a series of runes until a separator char is encountered
func (w *Wrap) word() error {
	for {
		r, _, err := w.r.ReadRune()
		if err != nil {
			return err
		}
		if isSpace(r) {
			// back up because we only return the word, not the separator
			err = w.r.UnreadRune()
			if err != nil {
				return err
			}
			return nil
		}
		w.runes = append(w.runes, r)
		// a hyphen is included with a chunk of non-whitespace chars but causes
		// a chunk boundary just in case there's a wrap after the hyphen but
		// before the end of the next chunk.
		if isHyphen(r) {
			return nil
		}
	}
}

// spaces returns all space characters between two words; unicode.IsSpace is used to
// evaluate if the rune is a space.
func (w *Wrap) spaces() error {
	for {
		r, _, err := w.r.ReadRune()
		if err != nil {
			return err
		}
		if !isSpace(r) && !isHyphen(r) {
			// back up because we only return the spaces, not non-Space chars
			err = w.r.UnreadRune()
			if err != nil {
				return err
			}
			return nil
		}
		// should actually be considered a space
		w.runes = append(w.runes, r)
	}
}

// process a chunk of runes that are a series of whitespace characters.
func (w *Wrap) processSpaceChunk() error {
	var (
		nl    bool
		runes []rune // tmp runes
		err   error
	)
	for _, r := range w.runes {
		// \r aren't copied into the temp run
		if r == '\r' {
			continue
		}
		// if it's not a new line, append the char to the current run
		if r != '\n' {
			runes = append(runes, r)
			continue
		}

		// handle new line
		nl = true
		if len(runes) > 0 { // when there is a new line in a run of whitespace, the line may exceed length
			_, err = w.buf.WriteString(string(runes))
			if err != nil {
				return err
			}
			runes = runes[:0]
		}
		if w.Unwrappable {
			_, err = w.buf.WriteString(w.NewLine)
			if err != nil {
				return err
			}
			w.l = 0
			continue
		}
		err = w.writeNewLine()
		if err != nil {
			return err
		}
		if w.Indent { // if indented, any spaces after the new line are ignored
			runes = runes[:0]
		}
	}
	if !nl { // if there wasn't a nl see if this chunk exceeds the line
		if w.l+len(runes) >= w.Length {
			err := w.writeNewLine()
			if err != nil {
				return err
			}
			// after a new line, if it is indented, remaining spaces are ignored
			if w.Indent {
				return nil
			}

		}
	}
	// If a newline was just written and new lines are indented; remaining
	// whitespaces are not written.
	if w.newNL && w.Indent {
		w.newNL = false
		return nil
	}
	w.newNL = false
	if len(runes) > 0 { // If there is something to write.
		_, err = w.buf.WriteString(string(runes))
		if err != nil {
			return err
		}
		w.l += len(runes)
	}
	return nil
}

// Unwrap unwraps a wrapped string: all new line chars inserted by Wrap.Line
// will be elided. The resulting string will be returned.
func Unwrap(s string) string {
	b := make([]byte, 0, len(s))
	r := make([]byte, 4) // for encoding rune
	var elide bool
	for _, v := range s {
		if v == zeroNBSP { // zero-width no-break space probably starts a seq
			elide = true
			// mark the start of the sequence. There is a chance that zero-width no-break
			// space doesn't mark the start of an inserted new line sequence.
			continue
		}
		if elide { // if this is part of a zero-width no-break space sequence
			if v == cr { // a carriage return is
				continue
			}
			if v == lf {
				// reset the info
				elide = false
				continue
			}
			// the zero-width no-break space didn't delimit an inserted sequence so
			// write it out to preserve original string.
			n := utf8.EncodeRune(r, zeroNBSP)
			b = append(b, r[:n]...)
			elide = false
		}
		n := utf8.EncodeRune(r, v)
		b = append(b, r[:n]...)
	}
	return string(b[:len(b)])
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
