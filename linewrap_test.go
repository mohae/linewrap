package linewrap

import "testing"

func TestLine(t *testing.T) {
	tests := []struct {
		s         string
		length    int
		tabSize   int
		indent    bool
		indentVal string
		newLine   string
		expected  string
	}{
		{"", 20, 4, false, "", "\n", ""},
		{"Hello", 20, 4, false, "", "\n", "Hello"},
		{"Hello World", 20, 4, false, "", "\n", "Hello World"},
		{"This sentence is a\n meaningless one", 20, 4, false, "", "\n", "This sentence is a\n meaningless one"},
		{"This sentence is a \nmeaningless one", 20, 4, false, "", "\n", "This sentence is a \nmeaningless one"},

		{"This sentence is a\n meaningless one", 20, 4, true, "    ", "\n", "This sentence is a\n    meaningless one"},
		{"This sentence is a\n meaningless one", 20, 4, true, "\t", "\n", "This sentence is a\n\tmeaningless one"},
		{"This sentence is a\r\n meaningless one", 20, 4, true, "    ", "\n", "This sentence is a\n    meaningless one"},
		{"This sentence is a\r\n meaningless one", 20, 4, true, "\t", "\n", "This sentence is a\n\tmeaningless one"},
		{"This sentence isn't\r\n a meaningless one", 20, 4, true, "    ", "\n", "This sentence isn't\n    a meaningless \n    one"},

		{"This sentence isn't\r\n a meaningless one", 20, 4, true, "\t", "\n", "This sentence isn't\n\ta meaningless \n\tone"},
		{"This sentence is a\n meaningless one", 20, 4, true, "\t", "\r\n", "This sentence is a\r\n\tmeaningless one"},
		{"This sentence is a\r\n meaningless one", 20, 4, true, "    ", "\r\n", "This sentence is a\r\n    meaningless one"},
		{"This sentence is a\r\n meaningless one", 20, 4, true, "\t", "\r\n", "This sentence is a\r\n\tmeaningless one"},
		{"This sentence isn't\r\n a meaningless one", 20, 4, true, "    ", "\r\n", "This sentence isn't\r\n    a meaningless \r\n    one"},

		{"This sentence isn't\r\n a meaningless one", 20, 4, true, "\t", "\r\n", "This sentence isn't\r\n\ta meaningless \r\n\tone"},
		{"Reality is frequently inaccurate. One is never alone with a rubber duck.", 34, 4, false, "", "\n", "Reality is frequently inaccurate.\n One is never alone with a rubber\n duck."},
		{"Reality is frequently inaccurate. One is never alone with a rubber duck.", 35, 4, false, "", "\n", "Reality is frequently inaccurate. \nOne is never alone with a rubber \nduck."},
		{"Reality is frequently inaccurate.     One is never alone with a rubber duck.", 34, 4, false, "", "\n", "Reality is frequently inaccurate.\n     One is never alone with a \nrubber duck."},
		{"Reality is frequently inaccurate.     One is never alone with a rubber duck.", 35, 4, false, "", "\n", "Reality is frequently inaccurate.\n     One is never alone with a \nrubber duck."},

		{"Reality is frequently inaccurate.     One is never alone with a rubber duck.", 40, 4, false, "", "\n", "Reality is frequently inaccurate.     \nOne is never alone with a rubber duck."},
		{"A common mistake\n that people make when trying to design something completely foolproof is to underestimate the ingenuity of complete fools.", 20, 4, false, "", "\n", "A common mistake\n that people make \nwhen trying to \ndesign something \ncompletely \nfoolproof is to \nunderestimate the \ningenuity of \ncomplete fools."},
		{"못\t알아\t듣겠어요\t전혀\t모르겠어요", 20, 4, false, "", "\n", "못\t알아\t듣겠어요\t\n전혀\t모르겠어요"},
		{"못\t알아\t듣겠어요\t전혀\t모르겠어요", 20, 4, true, "    ", "\n", "못\t알아\t듣겠어요\t\n    전혀\t모르겠어요"},
		{"못\t알아\t듣겠어요\t전혀\t모르겠어요", 20, 4, true, "\t", "\n", "못\t알아\t듣겠어요\t\n\t전혀\t모르겠어요"},

		{"hello\nΧαίρετε\t\tЗдравствуйте", 20, 4, false, "", "\n", "hello\nΧαίρετε\t\t\nЗдравствуйте"},
		{"hello\nΧαίρετε\t\tЗдравствуйте", 20, 4, true, "    ", "\n", "hello\n    Χαίρετε\t\t\n    Здравствуйте"},
		{"hello\nΧαίρετε\t\tЗдравствуйте", 20, 4, true, "\t", "\n", "hello\n\tΧαίρετε\t\t\n\tЗдравствуйте"},
		{"hello\nΧαίρετε\t\tЗдравствуйте", 20, 4, false, "    ", "\n", "hello\nΧαίρετε\t\t\nЗдравствуйте"},
		{"hello\nΧαίρετε\t\tЗдравствуйте", 20, 4, false, "\t", "\n", "hello\nΧαίρετε\t\t\nЗдравствуйте"},
		// altspaces
		{"Reality is\u00A0frequently inaccurate.", 20, 4, false, "", "\n", "Reality \nis\u00A0frequently \ninaccurate."},
		{"Reality is\u00a0frequently inaccurate.", 20, 4, false, "", "\n", "Reality \nis\u00a0frequently \ninaccurate."},
		{"Reality is\u2005frequently inaccurate.", 20, 4, false, "", "\n", "Reality is\u2005\nfrequently \ninaccurate."},
		{"Reality is\u2001frequently inaccurate.", 20, 4, false, "", "\n", "Reality is\u2001\nfrequently \ninaccurate."},
		{"Reality is\u200Bfrequently inaccurate.", 20, 4, false, "", "\n", "Reality is\u200B\nfrequently \ninaccurate."},

		{"Reality is\u200bfrequently inaccurate.", 20, 4, false, "", "\n", "Reality is\u200b\nfrequently \ninaccurate."},
		{"Reality is\u202Ffrequently inaccurate.", 20, 4, false, "", "\n", "Reality \nis\u202Ffrequently \ninaccurate."},
		{"Reality is\u202ffrequently inaccurate.", 20, 4, false, "", "\n", "Reality \nis\u202ffrequently \ninaccurate."},
		{"Reality is\uFEFFfrequently inaccurate.", 20, 4, false, "", "\n", "Reality \nis\uFEFFfrequently \ninaccurate."},
		// dashes
		// nbdash
	}
	w := New()
	for i, test := range tests {
		w.Length = test.length
		w.TabSize = test.tabSize
		w.Indent = test.indent
		w.IndentVal = test.indentVal
		w.NewLine = []byte(test.newLine)
		s, err := w.Line(test.s)
		if err != nil {
			t.Errorf("%d: unexpected error: %q", i, err)
			continue
		}
		if s != test.expected {
			t.Errorf("%d: got %q want %q", i, s, test.expected)
		}
	}
}

/*
func TestStringInComments(t *testing.T) {
	tests := []struct {
		line  string
		l     int
		lines []string
	}{
		{"", 10, nil},
		{"Hello", 10, []string{"// Hello"}},
		{"Hello World", 10, []string{"// Hello", "// World"}},
		{"This sentence is a meaningless one", 0, []string{"// This sentence is a meaningless one"}},
		{"This sentence is a meaningless one", 20, []string{"// This sentence is", "// a meaningless one"}},
		{"못 알아 듣겠어요 전혀 모르겠어요", 10, []string{"// 못 알아", "// 듣겠어요 전혀", "// 모르겠어요"}},
		// outlier, but if a word > l then use the whole word anyways
		{"hello Χαίρετε Здравствуйте", 10, []string{"// hello", "// Χαίρετε", "// Здравствуйте"}},
	}
	for i, test := range tests {
		lines := StringToComments(test.line, test.l)
		if len(lines) != len(test.lines) {
			t.Errorf("%d: got %d lines; want %d", i, len(lines), len(test.lines))
			t.Errorf("%s", lines)
			continue
		}
		for j, v := range lines {
			if v != test.lines[j] {
				t.Errorf("%d:%d: got %q; want %q", i, j, v, test.lines[j])
				continue
			}
		}
	}
}
*/
