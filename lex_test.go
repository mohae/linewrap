package linewrap

import "testing"

type lexTest struct {
	input  string
	tokens []token
}

// token{tokenEOF, 0, ""}

var lexTests = []lexTest{
	{"", []token{token{tokenEOF, 0, 0, ""}}},
	{"hello world", []token{{tokenText, 0, 5, "hello"}, {tokenSpace, 5, 1, " "}, {tokenText, 6, 5, "world"}, token{tokenEOF, 11, 0, ""}}},
	{"Time is an illusion. Lunchtime doubly so.",
		[]token{
			{tokenText, 0, 4, "Time"}, {tokenSpace, 4, 1, " "}, {tokenText, 5, 2, "is"}, {tokenSpace, 7, 1, " "},
			{tokenText, 8, 2, "an"}, {tokenSpace, 10, 1, " "}, {tokenText, 11, 9, "illusion."}, {tokenSpace, 20, 1, " "},
			{tokenText, 21, 9, "Lunchtime"}, {tokenSpace, 30, 1, " "}, {tokenText, 31, 6, "doubly"}, {tokenSpace, 37, 1, " "},
			{tokenText, 38, 3, "so."}, token{tokenEOF, 41, 0, ""},
		},
	},
	{"Time is an illusion.\u2001Lunchtime doubly so.",
		[]token{
			{tokenText, 0, 4, "Time"}, {tokenSpace, 4, 1, " "}, {tokenText, 5, 2, "is"}, {tokenSpace, 7, 1, " "},
			{tokenText, 8, 2, "an"}, {tokenSpace, 10, 1, " "}, {tokenText, 11, 9, "illusion."}, {tokenSpace, 20, 1, "\u2001"},
			{tokenText, 23, 9, "Lunchtime"}, {tokenSpace, 32, 1, " "}, {tokenText, 33, 6, "doubly"}, {tokenSpace, 39, 1, " "},
			{tokenText, 40, 3, "so."}, token{tokenEOF, 43, 0, ""},
		},
	},
	{"Time is an illusion.\u2014Lunchtime doubly so.",
		[]token{
			{tokenText, 0, 4, "Time"}, {tokenSpace, 4, 1, " "}, {tokenText, 5, 2, "is"}, {tokenSpace, 7, 1, " "},
			{tokenText, 8, 2, "an"}, {tokenSpace, 10, 1, " "}, {tokenText, 11, 9, "illusion."}, {tokenHyphen, 20, 1, "\u2014"},
			{tokenText, 23, 9, "Lunchtime"}, {tokenSpace, 32, 1, " "}, {tokenText, 33, 6, "doubly"}, {tokenSpace, 39, 1, " "},
			{tokenText, 40, 3, "so."}, token{tokenEOF, 43, 0, ""},
		},
	},
	{"Time is an illusion.-Lunchtime doubly so.",
		[]token{
			{tokenText, 0, 4, "Time"}, {tokenSpace, 4, 1, " "}, {tokenText, 5, 2, "is"}, {tokenSpace, 7, 1, " "},
			{tokenText, 8, 2, "an"}, {tokenSpace, 10, 1, " "}, {tokenText, 11, 9, "illusion."}, {tokenHyphen, 20, 1, "-"},
			{tokenText, 21, 9, "Lunchtime"}, {tokenSpace, 30, 1, " "}, {tokenText, 31, 6, "doubly"}, {tokenSpace, 37, 1, " "},
			{tokenText, 38, 3, "so."}, token{tokenEOF, 41, 0, ""},
		},
	},
	{"Time is an illusion.\tLunchtime doubly so.",
		[]token{
			{tokenText, 0, 4, "Time"}, {tokenSpace, 4, 1, " "}, {tokenText, 5, 2, "is"}, {tokenSpace, 7, 1, " "},
			{tokenText, 8, 2, "an"}, {tokenSpace, 10, 1, " "}, {tokenText, 11, 9, "illusion."}, {tokenTab, 20, 1, "\t"},
			{tokenText, 21, 9, "Lunchtime"}, {tokenSpace, 30, 1, " "}, {tokenText, 31, 6, "doubly"}, {tokenSpace, 37, 1, " "},
			{tokenText, 38, 3, "so."}, token{tokenEOF, 41, 0, ""},
		},
	},
	{"Time is an illusion.\nLunchtime doubly so.",
		[]token{
			{tokenText, 0, 4, "Time"}, {tokenSpace, 4, 1, " "}, {tokenText, 5, 2, "is"}, {tokenSpace, 7, 1, " "},
			{tokenText, 8, 2, "an"}, {tokenSpace, 10, 1, " "}, {tokenText, 11, 9, "illusion."}, {tokenNL, 20, 1, "\n"},
			{tokenText, 21, 9, "Lunchtime"}, {tokenSpace, 30, 1, " "}, {tokenText, 31, 6, "doubly"}, {tokenSpace, 37, 1, " "},
			{tokenText, 38, 3, "so."}, token{tokenEOF, 41, 0, ""},
		},
	},
	{"Time is an illusion.\r\nLunchtime doubly so.",
		[]token{
			{tokenText, 0, 4, "Time"}, {tokenSpace, 4, 1, " "}, {tokenText, 5, 2, "is"}, {tokenSpace, 7, 1, " "},
			{tokenText, 8, 2, "an"}, {tokenSpace, 10, 1, " "}, {tokenText, 11, 9, "illusion."}, {tokenNL, 21, 1, "\n"},
			{tokenText, 22, 9, "Lunchtime"}, {tokenSpace, 31, 1, " "}, {tokenText, 32, 6, "doubly"}, {tokenSpace, 38, 1, " "},
			{tokenText, 39, 3, "so."}, token{tokenEOF, 42, 0, ""},
		},
	},
}

// collect gathers the emitted items into a slice.
func collect(t *lexTest, left, right string) (tokens []token) {
	l := lex([]byte(t.input))
	for {
		token := l.nextToken()
		tokens = append(tokens, token)
		if token.typ == tokenEOF || token.typ == tokenError {
			break
		}
	}
	return tokens
}

func equal(t *testing.T, i int, i1, i2 []token) {
	if len(i1) != len(i2) {
		t.Errorf("%d: got %d tokens want %d", i, len(i1), len(i2))
		return
	}
	for k := range i1 {
		if i1[k].typ != i2[k].typ {
			t.Errorf("%d:%d:typ: got %v want %v\ttoken: %#v", i, k, i1[k].typ, i2[k].typ, i1[k])
			continue
		}
		if i1[k].value != i2[k].value {
			t.Errorf("%d:%d:value: got %v want %v\ttoken: %#v", i, k, i1[k].value, i2[k].value, i1[k])
			continue
		}
		if i1[k].len != i2[k].len {
			t.Errorf("%d:%d:len: got %v want %v\ttoken: %#v", i, k, i1[k].len, i2[k].len, i1[k])
			continue
		}
		if i1[k].pos != i2[k].pos {
			t.Errorf("%d:%d:pos: got %v want %v\ttoken: %#v", i, k, i1[k].pos, i2[k].pos, i1[k])
			continue
		}
	}
}

// test comment lexing
func TestLex(t *testing.T) {
	for i, test := range lexTests {
		tokens := collect(&test, "", "")
		equal(t, i, tokens, test.tokens)
	}
}

func TestIsHyphen(t *testing.T) {
	tests := []struct {
		r rune
		b bool
	}{
		{'\u0020', false},
		{'\u2011', false},
		{'\u207b', true},
		{'\u208b', true},
		{'\u002d', true},

		{'\u007e', false},
		{'\u00ad', true},
		{'\u058a', true},
		{'\u05be', false},
		{'\u1400', false},

		{'\u1806', false}, // is a hyphen but not considered one because break before hyphen not supported
		{'\u2010', true},
		{'\u2012', true},
		{'\u2013', true},
		{'\u2014', true},

		{'\u2015', true},
		{'\u2053', true},
		{'\u2212', false},
		{'\u2e17', false},
		{'\u2e3a', true},

		{'\u2e3b', true},
		{'\u301c', false},
		{'\u3030', false},
		{'\u30a0', false},
		{'\ufe31', true},

		{'\ufe32', true},
		{'\ufe58', true},
		{'\ufe63', true},
		{'\uff0d', true},
	}
	for i, test := range tests {
		tkn, ok := key[string(test.r)]
		if !ok { // anything not in the key map should be false
			if test.b != ok {
				t.Errorf("%d:%q: got %t want %t", i, string(test.r), ok, test.b)
			}
			continue
		}
		b := isHyphen(tkn)
		if b != test.b {
			t.Errorf("%x %c: got %t; want %t", test.r, test.r, b, test.b)
		}
	}
}
