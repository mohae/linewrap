package linewrap

import (
	"fmt"
	"strings"
	"unicode/utf8"
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
	tokenError tokenType = iota
	tokenEOF
	tokenText // anything that isn't one of the following
	tokenNL   // \n
	tokenCR   // \r

	tokenTab                   // \t
	tokenZeroWidthNoBreakSpace // U+FEFF used for unwrappable

	// unicode tokens we care about, mostly because of breaking rules. The whitespace
	// and dash tokens listed may be different than what Go uses in the relevant Go
	// unicode tables.

	// whitespace tokens from https://www.cs.tut.fi/~jkorpela/chars/spaces.html
	//
	// exceptions to the table:
	//   no-break space            U+00A0 is not considered whitespace for line break purposes
	//   narrow no-break space     U+202F is not considered whitespace for line break purposes
	//   zero width no-break space U+FEFF is not considered whitespace for line break purposes
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
	//   tilde        U+007E does not cause a line break because of possibility of ~/dir, ~=, etc.
	//   hyphen minus U+002D this is not supposed to break on a numeric context but no differentiation is done
	//   minus sign   U+2212 does not cause a line break
	//   wavy dash    U+301C does not cause a line break
	//   wavy dash    U+3939 does not cause a line break
	//   two em dash  U+2E3A is not in table but is here.
	//   three em dash  U+2E3B is not in table but is here.
	//   presentation form for vertical em dash U+FE31 is not in table but is here.
	//   presentation form for vertical en dash U+FE32 is not in table but is here.
	//   small em dash U+FE58 is not in table but is here.
	//   small hyphen-minus is not in table but is here.
	//   full width hyphen-minus is not in table but is here.
	tokenHyphenMinus // U+002D

	tokenSoftHyphen          // U+00AD
	tokenArmenianHyphen      // U+058A
	tokenMongolianTodoHyphen // U+1806 break before
	tokenHyphen              // U+2010
	tokenFigureDash          // U+2012

	tokenEnDash           // U+2013
	tokenEmDash           // U+2014 can be before or after
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
	"\n":     tokenNL,
	"\r":     tokenCR,
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
	"\u1806": tokenMongolianTodoHyphen,
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

const eof = -1

type stateFn func(*lexer) stateFn

type lexer struct {
	input      []byte     // the string being scanned
	state      stateFn    // the next lexing function to enter
	pos        Pos        // current position of this item
	start      Pos        // start position of this item
	width      Pos        // width of last rune read from input
	lastPos    Pos        // position of most recent item returned by nextItem
	runeCnt    int        // the number of runes in the current token sequence
	tokens     chan token // channel of scanned tokens
	commentTyp CommentType
}

func newLexer(input []byte) *lexer {
	return &lexer{
		input:  input,
		state:  lexText,
		tokens: make(chan token, 2),
	}
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
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

// error returns an error token and terminates the scan by passing back a nil
// pointer that will be the next state, terminating l.run.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token{tokenError, l.start, 0, fmt.Sprintf(format, args...)}
	return nil
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
	l.runeCnt = 0
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

// nextToken returns the next token from the input.
func (l *lexer) nextToken() token {
	for {
		select {
		case token := <-l.tokens:
			return token
		default:
			l.state = l.state(l)
		}
	}
}

// peek returns but does not consume the next rune in the input
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// run lexes the input by executing state functions until the state is nil.
func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.tokens) // No more tokens will be delivered
}

// scan until end of the space sequence is encountered
func lexSpace(l *lexer) stateFn {
	if l.pos > l.start {
		l.emit(tokenText)
	}
	// scan until the spaces are consumed
	for {
		r := l.next()
		tkn, ok := key[string(r)]
		if !ok {
			break
		}
		if !isSpace(tkn) {
			break
		}
		//		l.emit(tkn)
	}
	l.backup()
	l.emit(tokenSpace)
	return lexText
}

// scan until end of the hyphen sequence is encountered
func lexHyphen(l *lexer) stateFn {
	if l.pos > l.start {
		l.emit(tokenText)
	}
	// scan until the spaces are consumed
	for {
		r := l.next()
		tkn, ok := key[string(r)]
		if !ok {
			break
		}
		if !isHyphen(tkn) {
			break
		}
		l.emit(tkn)
	}
	l.backup()
	return lexText
}

// lexReturn handles a carriage return, `\r`; these are skipped.
func lexReturn(l *lexer) stateFn {
	if l.pos > l.start {
		l.emit(tokenText)
	}
	l.pos += Pos(len(string(cr)))
	l.ignore()
	return lexText
}

// lexNewLine handles a new line, `\n`
func lexNewLine(l *lexer) stateFn {
	if l.pos > l.start {
		l.emit(tokenText)
	}
	l.pos += Pos(len(string(nl)))
	l.runeCnt++
	l.emit(tokenNL)
	return lexText
}

// lexNewLine handles a tab line, `\t`
func lexTab(l *lexer) stateFn {
	if l.pos > l.start {
		l.emit(tokenText)
	}
	l.pos += Pos(len(string(nl)))
	l.runeCnt++
	l.emit(tokenTab)
	return lexText
}

// stateFn to process input and tokenize things
func lexText(l *lexer) stateFn {
	for {
		r := l.peek()
		if r == eof {
			break
		}
		tkn, ok := key[string(r)]
		if !ok {
			l.next()
			continue
		}
		switch tkn {
		case tokenCR:
			return lexReturn
		case tokenNL:
			return lexNewLine
		case tokenTab:
			return lexTab
		}
		if isSpace(tkn) { // this is the most likely so it's explicitly checked here
			return lexSpace
		}
		if isHyphen(tkn) {
			return lexHyphen
		}
		l.next()
	}

	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(tokenText)
	}
	l.emit(tokenEOF) // Useful to make EOF a token
	return nil       // Stop the run loop.
}

func isSpace(t tokenType) bool {
	if t >= tokenSpace && t <= tokenIdeographicSpace {
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
