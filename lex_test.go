package linewrap

import "testing"

type lexTest struct {
	input  string
	tokens []token
}

// token{tokenEOF, 0, ""}

var lexTests = []lexTest{
	{"", []token{token{tokenEOF, 0, ""}}},
	{"hello world", []token{{tokenText, 0, "hello"}, {tokenSpace, 5, " "}, {tokenText, 6, "world"}, token{tokenEOF, 11, ""}}},
	{"Time is an illusion. Lunchtime doubly so.",
		[]token{
			{tokenText, 0, "Time"}, {tokenSpace, 4, " "}, {tokenText, 5, "is"}, {tokenSpace, 7, " "},
			{tokenText, 8, "an"}, {tokenSpace, 10, " "}, {tokenText, 11, "illusion."}, {tokenSpace, 20, " "},
			{tokenText, 21, "Lunchtime"}, {tokenSpace, 30, " "}, {tokenText, 31, "doubly"}, {tokenSpace, 37, " "},
			{tokenText, 38, "so."}, token{tokenEOF, 41, ""},
		},
	},
	{"Time is an illusion.\u2001Lunchtime doubly so.",
		[]token{
			{tokenText, 0, "Time"}, {tokenSpace, 4, " "}, {tokenText, 5, "is"}, {tokenSpace, 7, " "},
			{tokenText, 8, "an"}, {tokenSpace, 10, " "}, {tokenText, 11, "illusion."}, {tokenEmQuad, 20, "\u2001"},
			{tokenText, 23, "Lunchtime"}, {tokenSpace, 32, " "}, {tokenText, 33, "doubly"}, {tokenSpace, 39, " "},
			{tokenText, 40, "so."}, token{tokenEOF, 43, ""},
		},
	},
	{"Time is an illusion.\u2014Lunchtime doubly so.",
		[]token{
			{tokenText, 0, "Time"}, {tokenSpace, 4, " "}, {tokenText, 5, "is"}, {tokenSpace, 7, " "},
			{tokenText, 8, "an"}, {tokenSpace, 10, " "}, {tokenText, 11, "illusion."}, {tokenEmDash, 20, "\u2014"},
			{tokenText, 23, "Lunchtime"}, {tokenSpace, 32, " "}, {tokenText, 33, "doubly"}, {tokenSpace, 39, " "},
			{tokenText, 40, "so."}, token{tokenEOF, 43, ""},
		},
	},

	{"Time is an illusion.-Lunchtime doubly so.",
		[]token{
			{tokenText, 0, "Time"}, {tokenSpace, 4, " "}, {tokenText, 5, "is"}, {tokenSpace, 7, " "},
			{tokenText, 8, "an"}, {tokenSpace, 10, " "}, {tokenText, 11, "illusion."}, {tokenHyphenMinus, 20, "-"},
			{tokenText, 21, "Lunchtime"}, {tokenSpace, 30, " "}, {tokenText, 31, "doubly"}, {tokenSpace, 37, " "},
			{tokenText, 38, "so."}, token{tokenEOF, 41, ""},
		},
	},
	{"Time is an illusion.\nLunchtime doubly so.",
		[]token{
			{tokenText, 0, "Time"}, {tokenSpace, 4, " "}, {tokenText, 5, "is"}, {tokenSpace, 7, " "},
			{tokenText, 8, "an"}, {tokenSpace, 10, " "}, {tokenText, 11, "illusion."}, {tokenNL, 20, "\n"},
			{tokenText, 21, "Lunchtime"}, {tokenSpace, 30, " "}, {tokenText, 31, "doubly"}, {tokenSpace, 37, " "},
			{tokenText, 38, "so."}, token{tokenEOF, 41, ""},
		},
	},
	{"Time is an illusion.\r\nLunchtime doubly so.",
		[]token{
			{tokenText, 0, "Time"}, {tokenSpace, 4, " "}, {tokenText, 5, "is"}, {tokenSpace, 7, " "},
			{tokenText, 8, "an"}, {tokenSpace, 10, " "}, {tokenText, 11, "illusion."}, {tokenNL, 21, "\n"},
			{tokenText, 22, "Lunchtime"}, {tokenSpace, 31, " "}, {tokenText, 32, "doubly"}, {tokenSpace, 38, " "},
			{tokenText, 39, "so."}, token{tokenEOF, 42, ""},
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
