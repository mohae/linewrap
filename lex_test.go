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
			{tokenText, 8, 2, "an"}, {tokenSpace, 10, 1, " "}, {tokenText, 11, 9, "illusion."}, {tokenEmQuad, 20, 1, "\u2001"},
			{tokenText, 23, 9, "Lunchtime"}, {tokenSpace, 32, 1, " "}, {tokenText, 33, 6, "doubly"}, {tokenSpace, 39, 1, " "},
			{tokenText, 40, 3, "so."}, token{tokenEOF, 43, 0, ""},
		},
	},
	{"Time is an illusion.\u2014Lunchtime doubly so.",
		[]token{
			{tokenText, 0, 4, "Time"}, {tokenSpace, 4, 1, " "}, {tokenText, 5, 2, "is"}, {tokenSpace, 7, 1, " "},
			{tokenText, 8, 2, "an"}, {tokenSpace, 10, 1, " "}, {tokenText, 11, 9, "illusion."}, {tokenEmDash, 20, 1, "\u2014"},
			{tokenText, 23, 9, "Lunchtime"}, {tokenSpace, 32, 1, " "}, {tokenText, 33, 6, "doubly"}, {tokenSpace, 39, 1, " "},
			{tokenText, 40, 3, "so."}, token{tokenEOF, 43, 0, ""},
		},
	},

	{"Time is an illusion.-Lunchtime doubly so.",
		[]token{
			{tokenText, 0, 4, "Time"}, {tokenSpace, 4, 1, " "}, {tokenText, 5, 2, "is"}, {tokenSpace, 7, 1, " "},
			{tokenText, 8, 2, "an"}, {tokenSpace, 10, 1, " "}, {tokenText, 11, 9, "illusion."}, {tokenHyphenMinus, 20, 1, "-"},
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
	l := newLexer([]byte(t.input))
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
			t.Errorf("%d:%d:typ: got %v want %v", i, k, i1[k].typ, i2[k].typ)
			continue
		}
		if i1[k].value != i2[k].value {
			t.Errorf("%d:%d:value: got %v want %v", i, k, i1[k].value, i2[k].value)
			continue
		}
		if i1[k].len != i2[k].len {
			t.Errorf("%d:%d:len: got %v want %v", i, k, i1[k].len, i2[k].len)
			continue
		}
		if i1[k].pos != i2[k].pos {
			t.Errorf("%d:%d:pos: got %v want %v", i, k, i1[k].pos, i2[k].pos)
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
