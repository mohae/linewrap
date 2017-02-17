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

package linewrap

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	cr                    = '\r'
	nl                    = '\n'
	tab                   = '\t'
	zeroWidthNoBreakSpace = "\uFEFF"
)

// Pos is a byte position in the original input text.
type Pos int

type token struct {
	typ   tokenType
	pos   Pos
	len   int // kength in chars (not bytes)
	value string
}

func (t token) String() string {
	switch {
	case t.typ == tokenEOF:
		return "EOF"
	case t.typ == tokenError:
		return t.value
	}
	return t.value
}

func (t token) Error() string {
	return fmt.Sprintf("lex error at %d: %s", int(t.pos), t.value)
}

type tokenType int

const (
	tokenNone tokenType = iota
	tokenError
	tokenEOF
	tokenText                  // anything that isn't one of the following
	tokenZeroWidthNoBreakSpace // U+FEFF used for unwrappable
	tokenNL                    // \n
	tokenCR                    // \r

	// unicode tokens we care about, mostly because of breaking rules. The whitespace
	// and dash tokens listed may be different than what Go uses in the relevant Go
	// unicode tables.

	// whitespace tokens from https://www.cs.tut.fi/~jkorpela/chars/spaces.html
	//
	// exceptions to the table:
	//   no-break space            U+00A0 is not considered whitespace for line break purposes
	//   narrow no-break space     U+202F is not considered whitespace for line break purposes
	//   zero width no-break space U+FEFF is not considered whitespace for line break purposes
	tokenTab                     // \t
	tokenSpace                   // U+0020
	tokenOghamSpaceMark          // U+1680
	tokenMongolianVowelSeparator // U+180E

	tokenEnQuad          // U+2000
	tokenEmQuad          // U+2001
	tokenEnSpace         // U+2002
	tokenEmSpace         // U+2003
	tokenThreePerEmSpace // U+2004

	tokenFourPerEmSpace   // U+2005
	tokenSixPerEmSpace    // U+2006
	tokenFigureSpace      // U+2007
	tokenPunctuationSpace // U+2008
	tokenThinSpace        // U+2009

	tokenHairSpace               // U+200A
	tokenZeroWidthSpace          // U+200B
	tokenMediumMathematicalSpace // U+205F
	tokenIdeographicSpace        // U+3000

	// dash tokens from https://www.cs.tut.fi/~jkorpela/dashes.html
	// hyphens and dashes in lines breaking rules sections
	//
	// exceptions to the table:
	//   tilde            U+007E does not cause a line break because of possibility of ~/dir, ~=, etc.
	//   hyphen minus     U+002D this is not supposed to break on a numeric context but no differentiation is done
	//   minus sign       U+2212 does not cause a line break
	//   wavy dash        U+301C does not cause a line break
	//   wavy dash        U+3939 does not cause a line break
	//   two em dash      U+2E3A is not in table but is here.
	//   three em dash    U+2E3B is not in table but is here.
	//   small em dash    U+FE58 is not in table but is here.
	//   small hyphen-minus       U+FE63 is not in table but is here.
	//   full width hyphen-minus  U+FF0D is not in table but is here.
	//   mongolian todo hyphen    U+1806  does not cause a line break becaues it is a break before char
	//   presentation form for vertical em dash U+FE31 is not in table but is here.
	//   presentation form for vertical en dash U+FE32 is not in table but is here.
	tokenHyphenMinus // U+002D

	tokenSoftHyphen     // U+00AD
	tokenArmenianHyphen // U+058A
	tokenHyphen         // U+2010
	tokenFigureDash     // U+2012

	tokenEnDash           // U+2013
	tokenEmDash           // U+2014 can be before or after but only after is supported here
	tokenHorizontalBar    // U+2015
	tokenSwungDash        // U+2053
	tokenSuperscriptMinus // U+207B

	tokenSubScriptMinus                    // U+208B
	tokenTwoEmDash                         // U+2E3A
	tokenThreeEmDash                       // U+2E3B
	tokenPresentationFormForVerticalEmDash // U+FE31
	tokenPresentationFormForVerticalEnDash // U+FE32

	tokenSmallEmDash          // U+FE58
	tokenSmallHyphenMinus     // U+FE63
	tokenFullWidthHyphenMinus // U+FF0D
)

var key = map[string]tokenType{
	"\r":     tokenCR,
	"\n":     tokenNL,
	"\t":     tokenTab,
	"\uFEFF": tokenZeroWidthNoBreakSpace,
	"\u0020": tokenSpace,
	"\u1680": tokenOghamSpaceMark,
	"\u180E": tokenMongolianVowelSeparator,
	"\u2000": tokenEnQuad,
	"\u2001": tokenEmQuad,
	"\u2002": tokenEnSpace,
	"\u2003": tokenEmSpace,
	"\u2004": tokenThreePerEmSpace,
	"\u2005": tokenFourPerEmSpace,
	"\u2006": tokenSixPerEmSpace,
	"\u2007": tokenFigureSpace,
	"\u2008": tokenPunctuationSpace,
	"\u2009": tokenThinSpace,
	"\u200A": tokenHairSpace,
	"\u200B": tokenZeroWidthSpace,
	"\u205F": tokenMediumMathematicalSpace,
	"\u3000": tokenIdeographicSpace,
	"\u002D": tokenHyphenMinus,
	"\u00AD": tokenSoftHyphen,
	"\u058A": tokenArmenianHyphen,
	"\u2010": tokenHyphen,
	"\u2012": tokenFigureDash,
	"\u2013": tokenEnDash,
	"\u2014": tokenEmDash,
	"\u2015": tokenHorizontalBar,
	"\u2053": tokenSwungDash,
	"\u207B": tokenSuperscriptMinus,
	"\u208B": tokenSubScriptMinus,
	"\u2E3A": tokenTwoEmDash,
	"\u2E3B": tokenThreeEmDash,
	"\uFE31": tokenPresentationFormForVerticalEmDash,
	"\uFE32": tokenPresentationFormForVerticalEnDash,
	"\uFE58": tokenSmallEmDash,
	"\uFE63": tokenSmallHyphenMinus,
	"\uFF0D": tokenFullWidthHyphenMinus,
}

var vals = map[tokenType]string{
	tokenNone:                              "none",
	tokenError:                             "error",
	tokenEOF:                               "eof",
	tokenText:                              "text",
	tokenZeroWidthNoBreakSpace:             "zero width no break space",
	tokenNL:                                "nl",
	tokenCR:                                "cr",
	tokenTab:                               "tab",
	tokenSpace:                             "space",
	tokenOghamSpaceMark:                    "ogham space mark",
	tokenMongolianVowelSeparator:           "mongolian vowel separator",
	tokenEnQuad:                            "en quad",
	tokenEmQuad:                            "em quad",
	tokenEnSpace:                           "en space",
	tokenEmSpace:                           "em space",
	tokenThreePerEmSpace:                   "three per em space",
	tokenFourPerEmSpace:                    "four per em space",
	tokenSixPerEmSpace:                     "siz per em space",
	tokenFigureSpace:                       "token figure space",
	tokenPunctuationSpace:                  "punctuation space",
	tokenThinSpace:                         "thin space",
	tokenHairSpace:                         "hair space",
	tokenZeroWidthSpace:                    "width space",
	tokenMediumMathematicalSpace:           "medium mathematical space",
	tokenIdeographicSpace:                  "ideographic space",
	tokenHyphenMinus:                       "hyphen minus",
	tokenSoftHyphen:                        "soft hyphen",
	tokenArmenianHyphen:                    "armenian hyphen",
	tokenHyphen:                            "hyphen",
	tokenFigureDash:                        "figure dash",
	tokenEnDash:                            "en dash",
	tokenEmDash:                            "em dash",
	tokenHorizontalBar:                     "horizontal bar",
	tokenSwungDash:                         "swung dash",
	tokenSuperscriptMinus:                  "superscript minus",
	tokenSubScriptMinus:                    "subscript minus",
	tokenTwoEmDash:                         "two em dash",
	tokenThreeEmDash:                       "three em dash",
	tokenPresentationFormForVerticalEmDash: "presentation form for vertical em dash",
	tokenPresentationFormForVerticalEnDash: "presentation form for vertical em dash",
	tokenSmallEmDash:                       "small em dash",
	tokenSmallHyphenMinus:                  "small hyphen minus",
	tokenFullWidthHyphenMinus:              "full width hyphen minus",
}

const eof = -1

const (
	classText tokenClass = iota
	classCR
	classNL
	classTab
	classSpace
	classHyphen
)

type tokenClass int

type stateFn func(*lexer) stateFn

type lexer struct {
	input   []byte     // the string being scanned
	state   stateFn    // the next lexing function to enter
	pos     Pos        // current position of this item
	start   Pos        // start position of this item
	width   Pos        // width of last rune read from input
	lastPos Pos        // position of most recent item returned by nextItem
	runeCnt int        // the number of runes in the current token sequence
	tokens  chan token // channel of scanned tokens
}

func lex(input []byte) *lexer {
	l := &lexer{
		input:  input,
		state:  lexText,
		tokens: make(chan token, 2),
	}
	go l.run()
	return l
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	l.runeCnt++
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRune(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
	l.runeCnt--
}

// emit passes an item back to the client.
func (l *lexer) emit(t tokenType) {
	l.tokens <- token{t, l.start, l.runeCnt, string(l.input[l.start:l.pos])}
	l.start = l.pos
	l.runeCnt = 0
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
	l.runeCnt = 0
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun cunsumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	if strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// error returns an error token and terminates the scan by passing back a nil
// pointer that will be the next state, terminating l.run.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token{tokenError, l.start, 0, fmt.Sprintf(format, args...)}
	return nil
}

// nextToken returns the next token from the input.
func (l *lexer) nextToken() token {
	token := <-l.tokens
	l.lastPos = token.pos
	return token
}

// drain the channel so the lex go routine will exit: called by caller.
func (l *lexer) drain() {
	for range l.tokens {
	}
}

// run lexes the input by executing state functions until the state is nil.
func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.tokens) // No more tokens will be delivered
}

// lexText scans non whitespace/hyphen chars.
func lexText(l *lexer) stateFn {
	for {
		is, class := l.atBreakPoint() // a breakpoint is any char after which a new line can be
		if is {
			if l.pos > l.start {
				l.emit(tokenText)
			}
			switch class {
			case classCR:
				return lexCR
			case classNL:
				return lexNL
			case classSpace:
				return lexSpace
			case classTab:
				return lexTab
			case classHyphen:
				return lexHyphen
			}
		}
		if l.next() == eof {
			l.runeCnt-- // eof doesn't count.
			break
		}
	}
	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(tokenText)
	}
	l.emit(tokenEOF) // Useful to make EOF a token
	return nil       // Stop the run loop.
}

// a breakpoint is any character afterwhich a wrap may occur. If it is a
// breakpoint char, the type of char is returned.
func (l *lexer) atBreakPoint() (breakpoint bool, class tokenClass) {
	r, _ := utf8.DecodeRune(l.input[l.pos:])
	t, ok := key[string(r)]
	if !ok || t <= tokenZeroWidthNoBreakSpace {
		return false, classText
	}
	switch t {
	case tokenCR:
		return true, classCR
	case tokenNL:
		return true, classNL
	case tokenTab:
		return true, classTab
	}
	if isSpace(t) {
		return true, classSpace
	}
	if isHyphen(t) {
		return true, classHyphen
	}
	// it really shouldn't get to here, but if it does, treat it like classText
	return false, classText
}

// lexCR handles a carriage return, `\r`; these are skipped. The prior token
// should already have been emitted and the next token should be a CR, which
// are skipped.  The next token is checked to ensure that it really is a CR.
func lexCR(l *lexer) stateFn {
	r := l.next()
	t := key[string(r)] // don't need to check ok, as the zero value won't match
	if t == tokenCR {
		l.ignore()
	}
	return lexText
}

// lexNL handles a new line, `\n`; the prior token should already have been
// emitted and the next token should be a NL. The next token is checked to
// ensure that it really is a NL
func lexNL(l *lexer) stateFn {
	r := l.next()
	t := key[string(r)] // don't need to check ok, as the zero value won't match
	if t == tokenNL {
		l.emit(tokenNL)
	}
	return lexText
}

// lexTab handles a tab, '\t'; the prior token should already have been emitted
// and the next token should be a tab. The next token is checked to ensure that
// it really is a tab.
func lexTab(l *lexer) stateFn {
	r := l.next()
	t := key[string(r)] // don't need to check ok, as the zero value won't match
	if t == tokenTab {
		l.emit(tokenTab)
	}
	return lexText
}

// This scans until end of the space sequence is encountered. If no spaces were
// found, nothing will be emitted. The prior token should already have been
// emitted before this function gets called.
func lexSpace(l *lexer) stateFn {
	var i int
	// scan until the spaces are consumed
	for {
		r := l.next()
		// ok doesn't need to be checked as the zeroo value won't be classified as a hyphen.
		tkn := key[string(r)]
		if !isSpace(tkn) {
			break
		}
		i++
	}
	if i == 0 { // if no spaces were processed; nothing to emit.
		return lexText
	}
	// otherwise backup to ensure only space tokens are emitted.
	l.backup()
	l.emit(tokenSpace)
	return lexText
}

// Scan until end of the hyphen sequence is encountered. If no hyphens were
// found, nothing will be emitted. The prior token should already have been
// emitted before this function gets called.
func lexHyphen(l *lexer) stateFn {
	var i int
	// scan until the spaces are consumed
	for {
		r := l.next()
		// ok doesn't need to be checked as the zero value won't be classified as a hyphen.
		tkn := key[string(r)]
		if !isHyphen(tkn) {
			break
		}
		i++
	}
	if i == 0 { // if no hyphens. nothing to emit.
		return lexText
	}
	l.backup()
	l.emit(tokenHyphen)
	return lexText
}

func isSpace(t tokenType) bool {
	if t >= tokenTab && t <= tokenIdeographicSpace {
		return true
	}
	return false
}

func isHyphen(t tokenType) bool {
	if t >= tokenHyphenMinus && t <= tokenFullWidthHyphenMinus {
		return true
	}
	return false
}
