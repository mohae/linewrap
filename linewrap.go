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

import "sync"

const (
	LineLength            = 80 // default line length
	TabSize               = 8  // default tab size
	lineCommentSlash      = "//"
	lineCommentHash       = "#"
	blockCommentBegin     = "/*"
	blockCommentEnd       = "*/"
	cr                    = '\r'
	nl                    = '\n'
	tab                   = '\t'
	zeroWidthNoBreakSpace = "\uFEFF"
)

type CommentType int

const (
	None         CommentType = iota
	CommentSlash             // Line comment starting with //
	CommentHash              // Line comment starting with #
	CommentBlock             // Block comment delimited by /* and */
)

var (
	// package global
	stdWrap = New()
	mu      sync.Mutex
)

// Wrapper wraps lines so that the output is lines of Length characters or less.
type Wrapper struct {
	Length     int    // Max length of the line.
	tabSize    int    // The size of a tab, in chars.
	indentText []byte // The string used to indent wrapped lines; if empty no indent will be done.
	indentLen  int    // the length, in chars, of the indent text. tabs in the indentText count as tabSize cars.
	// If the wrapped string should be unwrappable. Unwrappable means all inserted
	// linebreaks can be removed and the unwrapped string will retain all of its
	// original formatting. If Unwrappable, the wrapped text will not be indented.
	// If there was a new line sequence substitution during line wrapping the
	// wrapped new line char(s) will be kept.
	Unwrappable bool
	CommentType // the type of comment,

	input []byte
	l     int // the length of the current line, in chars
	*lexer
	b []byte
}

// New returns a new Wrap with default Length and TabWidth.
func New() *Wrapper {
	return &Wrapper{
		Length:  LineLength,
		tabSize: TabSize,
	}
}

// String returns a wrapped string. The resulting string will be consistent
// with Wrap's configuration.
func (w *Wrapper) String(s string) (string, error) {
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
func (w *Wrapper) Bytes(s []byte) (b []byte, err error) {
	if len(s) == 0 { // if the string is empty, no comment
		return s, nil
	}

	w.input = s
	// odds are, it'll be at least the length of the input. This minimizes
	// re-allocs.
	w.b = make([]byte, 0, len(s))
	w.lexer = newLexer(s)
	go w.lexer.run()
	for {
		token := w.lexer.nextToken()
		if token.typ == tokenEOF {
			break
		}
		if token.typ == tokenError {
			return b, token
		}
		skip := w.wrap(&token)
		if skip {
			continue
		}
		w.b = append(w.b, token.String()...)
		w.l += token.len
	}
	return w.b, nil
}

// Sets the tabsize for line length calculations, when a tab is encountered.
// Actual tabsize may vary.  See TabSize for the default value.
func (w *Wrapper) TabSize(i int) {
	w.tabSize = i
	w.setIndentLen() // the indent len may need to be updated
}

// IndentText sets the value that should be used to indent wrapped lines.
func (w *Wrapper) IndentText(s string) {
	if s == "" { // no indent
		w.indentText = nil
		return
	}
	w.indentText = []byte(s)
	w.setIndentLen()
}

// sets the indentLen based on indentText and tabsize.
func (w *Wrapper) setIndentLen() {
	// calculate the indentLen
	for _, v := range w.indentText {
		if v == tab {
			w.indentLen += w.tabSize
			continue
		}
		w.indentLen++
	}
}

// wrap figures out wrapping of line stuff
func (w *Wrapper) wrap(t *token) (skip bool) {
	if t.typ == tokenNL { // if this is a newline token, reset the length
		// make the line length == Length so that the next token will trigger a
		// newline; this handles trailing spaces.
		w.l = w.Length
		return true
	}

	if t.typ == tokenTab {
		t.len = w.tabSize
	}
	if w.l+t.len < w.Length { // if a new line isn't going to be emitted, return
		return
	}
	w.nl()
	if isSpace(t.typ) { // if this token is a space or spaces, it should be skipped
		return true
	}
	return false

}

func (w *Wrapper) nl() {
	w.b = append(w.b, nl)
	w.l = 0
	if w.indentLen > 0 {
		w.b = append(w.b, w.indentText...)
		w.l += w.indentLen
	}

}
