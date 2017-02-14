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

const (
	LineLength            = 80 // default line length
	TabSize               = 8  // default tab size
	cr                    = '\r'
	nl                    = '\n'
	tab                   = '\t'
	zeroWidthNoBreakSpace = "\uFEFF"
)

var (
	lineCommentSlash  = []byte("// ")
	lineCommentHash   = []byte("# ")
	blockCommentBegin = []byte("/*")
	blockCommentEnd   = []byte("*/")
)

type CommentType int

const (
	None         CommentType = iota
	CommentSlash             // Line comment starting with //
	CommentHash              // Line comment starting with #
	CommentBlock             // Block comment delimited by /* and */
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
	priorToken  token
	input       []byte
	l           int // the length of the current line, in chars
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

	// If there's a comment type; lead with that. If CommentType == none, nothing
	// will be done.
	w.commentBegin()

	var (
		skip bool
		tkn  token
	)

	w.lexer = newLexer(s)
	go w.lexer.run()
	for {
		w.priorToken = tkn
		tkn = w.lexer.nextToken()
		switch tkn.typ {
		case tokenSpace:
			if w.priorToken.typ == tokenNL {
				continue
			}
		case tokenNL:
			w.nl()
			continue
		case tokenEOF:
			goto done
		case tokenError:
			return w.b, tkn
		}
		skip = w.wrap(&tkn)
		if skip {
			continue
		}
		w.b = append(w.b, tkn.String()...)
		w.l += tkn.len
	}

done:
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

func (w *Wrapper) commentBegin() {
	switch w.CommentType {
	case None:
		return
	case CommentSlash, CommentHash:
		w.lineComment()
	}
}

func (w *Wrapper) lineComment() bool {
	switch w.CommentType {
	case CommentSlash:
		w.slashComment()
		return true
	case CommentHash:
		w.hashComment()
		return true
	}
	return false
}
func (w *Wrapper) hashComment() {
	w.b = append(w.b, lineCommentHash...)
	w.l = 2
}

func (w *Wrapper) slashComment() {
	w.b = append(w.b, lineCommentSlash...)
	w.l = 3
}

func (w *Wrapper) nl() {
	// see if the priorToken was a tokenSpace; if so back up to elide
	// trailing spaces from the line prior to a nl
	if w.priorToken.typ == tokenSpace {
		w.b = w.b[:len(w.b)-w.priorToken.len]
	}

	w.b = append(w.b, nl)
	w.l = 0
	b := w.lineComment() // add a new line if applicable
	if b {               // if this is a line comment no indent is done
		return
	}
	if w.indentLen > 0 {
		w.b = append(w.b, w.indentText...)
		w.l += w.indentLen
	}

}
