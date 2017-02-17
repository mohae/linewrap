// Copyright 2017 Joel Scoble
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package linewrap wraps text so that they are n characters, or less in
// length. Wrapped lines can be indented or turned into comments; c, c++, and
// shell style comments are supported.
//
// Any /r characters encountered will be elided during the wrapping process;
// only /n is supported for new lines.
//
// The size of tabs is configurable.
//
// With a few exceptions, lines can be wrapped at unicode dash and whitespace
// characters.
//
// The classification of unicode tokens is drawn from Jukka "Yucca" Korpela's
// unicode tables on: https://www.cs.tut.fi/~jkorpela/unicode/linebr.html,
// https://www.cs.tut.fi/~jkorpela/chars/spaces.html, and
// https://www.cs.tut.fi/~jkorpela/dashes.html. Additionally, information from
// http://www.unicode.org/reports/tr14/#Properties was used.
//
// The list of symbols handled is not exhaustive.
//
// Line breaks may be inserted before or after whitespace characters. Any
// trailing spaces on a line will be elided. With the exception of indentation,
// all leading whitespaces on a wrapped line will be elided.
//
//     space                      U+0020
//     ogham space mark           U+1680
//     mongolian vowel separator  U+180E
//     en quad                    U+2000
//     em quad                    U+2001
//     en space                   U+2002
//     em space                   U+2003
//     three per em space         U+2004
//     four per em space          U+2005
//     six per em space           U+2006
//     figure space               U+2007
//     punctuation space          U+2008
//     thin space                 U+2009
//     hair space                 U+200A
//     zero width space           U+200B
//     medium mathematical space  U+205F
//     ideographic space          U+3000
//
// Exceptions to whitespace characters (no break will occur):
//
//     no-break space             U+00A0
//     zero width no-break space  U+202F
//
// Line breaks may be inserted after a dash (hyphen) character. An em dash
// (U+2014) can have a break before or after its occurrence but linewrap will
// only break after its occurrence. A hyphen minus (U+002D) is not supposed to
// break on a numeric context but linewrap does not make that differentiation.
//
//     hyphen minus                            U+002D
//     soft hyphen                             U+00AD
//     armenian hyphen                         U+058A
//     hyphen                                  U+2010
//     figure dash                             U+2012
//     en dash                                 U+2013
//     em dash                                 U+2014
//     horizontal bar                          U+2015
//     swung dash                              U+2053
//     superscript mnus                        U+207B
//     subscript minus                         U+208B
//     two em dash                             U+2E3A
//     three em dash                           U+2E3B
//     presentation form for vertical em dash  U+FE31
//     presentation form for vertical en dash  U+FE32
//     small em dash                           U+FE58
//     small hyphen minus                      U+FE63
//     full width hyphen minus                 U+FF0D
//
// Exceptions to dash characters (no break will occur):
//
//      tilde                  U+007E
//      minus sign             U+2212
//      wavy dash              U+301C
//      wavy dash              U+3939
//      mongolian todo hyphen  U+1806
package linewrap

import (
	"fmt"
	"strings"
)

const (
	LineLength = 80 // default line length
	TabSize    = 8  // default tab size
)

var (
	cppComment    = []byte("// ")
	shellComment  = []byte("# ")
	cCommentBegin = []byte("/*\n") // the comment begin is on a separate line
	cCommentEnd   = []byte("*/\n") // the comment end
)

type CommentStyle int

const (
	NoComment    CommentStyle = iota
	CPPComment                // C++ style line comment: //
	ShellComment              // shell style line comment: #
	CComment                  // c style block comment: /* */
)

func (c CommentStyle) String() string {
	switch c {
	case NoComment:
		return "none"
	case CPPComment:
		return "c++ style comments"
	case ShellComment:
		return "shell style comments"
	case CComment:
		return "c style comments"
	default:
		return fmt.Sprintf("invalid: %d style comments", c)
	}
}

func ParseCommentStyle(s string) CommentStyle {
	s = strings.ToLower(s)
	switch s {
	case "c":
		return CComment
	case "cpp", "c++":
		return CPPComment
	case "shell", "perl":
		return ShellComment
	default:
		return NoComment
	}
}

// Wrapper wraps lines so that the output is lines of Length characters or less.
type Wrapper struct {
	Length       int    // Max length of the line.
	tabSize      int    // The size of a tab, in chars.
	indentText   []byte // The string used to indent wrapped lines; if empty no indent will be done.
	indentLen    int    // the length, in chars, of the indent text. tabs in the indentText count as tabSize cars.
	CommentStyle        // the type of comment,
	priorToken   token
	l            int // the length of the current line, in chars
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

// Reset resets the non-configuration fields so that it's usable for a new
// input. The Wrapper's configuration is not affected.
func (w *Wrapper) Reset() {
	w.lexer = nil
	w.b = w.b[:0]
	w.l = 0
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

	// if b hasn't already been allocated, do an initial allocation.
	if w.b == nil {
		w.b = make([]byte, 0, len(s))
	}

	// If there's a comment type; lead with that. If CommentType == none, nothing
	// will be done.
	w.commentBegin()

	var (
		skip bool
		tkn  token
	)

	w.lexer = lex(s)
	for {
		w.priorToken = tkn
		tkn = w.lexer.nextToken()
		if tkn.typ == tokenEOF { // if eof has been reached, stop processing
			break
		}
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
	w.commentEnd()

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
	// always reset the indent len
	w.indentLen = 0
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
	switch w.CommentStyle {
	case NoComment:
		return
	case CPPComment, ShellComment:
		w.lineComment()
	case CComment:
		w.b = append(w.b, cCommentBegin...)
	}
}

func (w *Wrapper) commentEnd() {
	if w.CommentStyle == CComment {
		w.b = append(w.b, cCommentEnd...)
	}
}

func (w *Wrapper) lineComment() bool {
	switch w.CommentStyle {
	case CPPComment:
		w.cppComment()
		return true
	case ShellComment:
		w.shellComment()
		return true
	}
	return false
}
func (w *Wrapper) shellComment() {
	w.b = append(w.b, shellComment...)
	w.l = 2
}

func (w *Wrapper) cppComment() {
	w.b = append(w.b, cppComment...)
	w.l = 3
}

func (w *Wrapper) nl() {
	// see if the priorToken was a tokenSpace; if so back up to elide
	// trailing spaces from the line prior to a nl
	if w.priorToken.typ == tokenSpace {
		w.b = w.b[:len(w.b)-len(w.priorToken.value)]
	}

	// If a line comment see if the current line is a blank comment line and elide
	// the trailing space if it is.
	w.cleanBlankCommentLine()

	// newline
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

// if the text is being wrapped as line comments and current line is a
// blank comment line, e.g. // with no text, make sure the trailing space
// is elided: "// " becomes "//" and "# " becomes "#"
func (w *Wrapper) cleanBlankCommentLine() {
	switch w.CommentStyle {
	case CPPComment:
		w.cleanBlankCPPCommentLine()
	case ShellComment:
		w.cleanBlankShellCommentLine()
	}
}

func (w *Wrapper) cleanBlankCPPCommentLine() {
	if w.b[len(w.b)-1] == 0x20 {
		if w.b[len(w.b)-2] == '/' && w.b[len(w.b)-3] == '/' {
			w.b = w.b[:len(w.b)-1]
		}
	}
}

func (w *Wrapper) cleanBlankShellCommentLine() {
	if w.b[len(w.b)-1] == 0x20 {
		if w.b[len(w.b)-2] == '#' {
			w.b = w.b[:len(w.b)-1]
		}
	}
}
