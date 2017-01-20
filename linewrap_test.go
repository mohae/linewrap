package linewrap

import "testing"

func TestLine(t *testing.T) {
	tests := []struct {
		s         string
		length    int
		tabWidth  int
		indent    bool
		indentVal string
		expected  string
	}{
		{"", 20, 4, false, "", ""},
		{"Hello", 20, 4, false, "", "Hello"},
		{"Hello World", 20, 4, false, "", "Hello World"},
		{"This sentence is a\n meaningless one", 20, 4, false, "", "This sentence is a\n meaningless one"},
		{"This sentence is a\n meaningless one", 20, 4, true, "    ", "This sentence is a\n    meaningless one"},

		{"This sentence is a\n meaningless one", 20, 4, true, "\t", "This sentence is a\n\tmeaningless one"},
		{"This sentence is a\r\n meaningless one", 20, 4, true, "    ", "This sentence is a\r\n    meaningless one"},
		{"This sentence is a\r\n meaningless one", 20, 4, true, "\t", "This sentence is a\r\n\tmeaningless one"},
		{"This sentence isn't\r\n a meaningless one", 20, 4, true, "    ", "This sentence isn't\r\n    a meaningless one"},
		{"This sentence isn't\r\n a meaningless one", 20, 4, true, "\t", "This sentence isn't\r\n\ta meaningless one"},

		{"Reality is frequently inaccurate. One is never alone with a rubber duck.", 34, 4, false, "", "Reality is frequently inaccurate.\n One is never alone with a rubber\n duck."},
		{"Reality is frequently inaccurate. One is never alone with a rubber duck.", 35, 4, false, "", "Reality is frequently inaccurate. \nOne is never alone with a rubber \nduck."},
		{"Reality is frequently inaccurate.     One is never alone with a rubber duck.", 34, 4, false, "", "Reality is frequently inaccurate.\n     One is never alone with a \nrubber duck."},
		{"Reality is frequently inaccurate.     One is never alone with a rubber duck.", 35, 4, false, "", "Reality is frequently inaccurate.\n     One is never alone with a \nrubber duck."},
		{"Reality is frequently inaccurate.     One is never alone with a rubber duck.", 40, 4, false, "", "Reality is frequently inaccurate.     \nOne is never alone with a rubber \nduck."},
		//{"A common mistake\n that people make when trying to design something completely foolproof is to underestimate the ingenuity of complete fools.", 20, 4, false, "", "A common mistake\n that people make \nwhen trying to \ndesign something \ncompletely foolproof\n is to underestimate\n the ingenuity of\n complete fools."},
		//		{"못\t알아\t듣겠어요\t전혀\t모르겠어요", 20, 4, false, "", "못\t알아\t듣겠어요\t\n전혀\t모르겠어요"},
		//		{"못\t알아\t듣겠어요\t전혀\t모르겠어요", 20, 4, true, "    ", "못\t알아\t듣겠어요\t\n    전혀\t모르겠어요"},
		//		{"못\t알아\t듣겠어요\t전혀\t모르겠어요", 20, 4, true, "\t", "못\t알아\t듣겠어요\t\n\t전혀\t모르겠어요"},
		//		{"hello\nΧαίρετε\t\tЗдравствуйте", 20, 4, false, "", "hello\nΧαίρετε\t\t\nЗдравствуйте"},
		//		{"hello\nΧαίρετε\t\tЗдравствуйте", 20, 4, false, "    ", "hello\nΧαίρετε\t\t\n    Здравствуйте"},
		//		{"hello\nΧαίρετε\t\tЗдравствуйте", 20, 4, false, "\t", "hello\nΧαίρετε\t\t\n\tЗдравствуйте"},
		// nbsp
		// alt spaces
		// dashes
		// nbdash
	}
	var w Wrap
	for i, test := range tests {
		w.Length = test.length
		w.TabWidth = test.tabWidth
		w.Indent = test.indent
		w.IndentVal = test.indentVal
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
