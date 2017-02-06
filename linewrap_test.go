package linewrap

import (
	"strings"
	"testing"
)

func TestWrapLine(t *testing.T) {
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
		// 5
		{"This sentence is a\n meaningless one", 20, 4, true, "    ", "\n", "This sentence is a\n    meaningless one"},
		{"This sentence is a\n meaningless one", 20, 4, true, "\t", "\n", "This sentence is a\n\tmeaningless one"},
		{"This sentence is a\r\n meaningless one", 20, 4, true, "    ", "\n", "This sentence is a\n    meaningless one"},
		{"This sentence is a\r\n meaningless one", 20, 4, true, "\t", "\n", "This sentence is a\n\tmeaningless one"},
		{"This sentence isn't\r\n a meaningless one", 20, 4, true, "    ", "\n", "This sentence isn't\n    a meaningless \n    one"},
		// 10
		{"This sentence isn't\r\n a meaningless one", 20, 4, true, "\t", "\n", "This sentence isn't\n\ta meaningless \n\tone"},
		{"This sentence is a\n meaningless one", 20, 4, true, "\t", "\r\n", "This sentence is a\r\n\tmeaningless one"},
		{"This sentence is a\r\n meaningless one", 20, 4, true, "    ", "\r\n", "This sentence is a\r\n    meaningless one"},
		{"This sentence is a\r\n meaningless one", 20, 4, true, "\t", "\r\n", "This sentence is a\r\n\tmeaningless one"},
		{"This sentence isn't\r\n a meaningless one", 20, 4, true, "    ", "\r\n", "This sentence isn't\r\n    a meaningless \r\n    one"},
		// 15
		{"This sentence isn't\r\n a meaningless one", 20, 4, true, "\t", "\r\n", "This sentence isn't\r\n\ta meaningless \r\n\tone"},
		{"Reality is frequently inaccurate. One is never alone with a rubber duck.", 34, 4, false, "", "\n", "Reality is frequently inaccurate.\n One is never alone with a rubber\n duck."},
		{"Reality is frequently inaccurate. One is never alone with a rubber duck.", 35, 4, false, "", "\n", "Reality is frequently inaccurate. \nOne is never alone with a rubber \nduck."},
		{"Reality is frequently inaccurate.     One is never alone with a rubber duck.", 34, 4, false, "", "\n", "Reality is frequently inaccurate.\n     One is never alone with a \nrubber duck."},
		{"Reality is frequently inaccurate.     One is never alone with a rubber duck.", 35, 4, false, "", "\n", "Reality is frequently inaccurate.\n     One is never alone with a \nrubber duck."},
		// 20
		{"Reality is frequently inaccurate.     One is never alone with a rubber duck.", 40, 4, false, "", "\n", "Reality is frequently inaccurate.     \nOne is never alone with a rubber duck."},
		{"A common mistake\n that people make when trying to design something completely foolproof is to underestimate the ingenuity of complete fools.", 20, 4, false, "", "\n", "A common mistake\n that people make \nwhen trying to \ndesign something \ncompletely \nfoolproof is to \nunderestimate the \ningenuity of \ncomplete fools."},
		{"못\t알아\t듣겠어요\t전혀\t모르겠어요", 20, 4, false, "", "\n", "못\t알아\t듣겠어요\t\n전혀\t모르겠어요"},
		{"못\t알아\t듣겠어요\t전혀\t모르겠어요", 20, 4, true, "    ", "\n", "못\t알아\t듣겠어요\t\n    전혀\t모르겠어요"},
		{"못\t알아\t듣겠어요\t전혀\t모르겠어요", 20, 4, true, "\t", "\n", "못\t알아\t듣겠어요\t\n\t전혀\t모르겠어요"},
		// 25
		{"hello\nΧαίρετε\t\tЗдравствуйте", 20, 4, false, "", "\n", "hello\nΧαίρετε\t\t\nЗдравствуйте"},
		{"hello\nΧαίρετε\t\tЗдравствуйте", 20, 4, true, "    ", "\n", "hello\n    Χαίρετε\t\t\n    Здравствуйте"},
		{"hello\nΧαίρετε\t\tЗдравствуйте", 20, 4, true, "\t", "\n", "hello\n\tΧαίρετε\t\t\n\tЗдравствуйте"},
		{"hello\nΧαίρετε\t\tЗдравствуйте", 20, 4, false, "    ", "\n", "hello\nΧαίρετε\t\t\nЗдравствуйте"},
		{"hello\nΧαίρετε\t\tЗдравствуйте", 20, 4, false, "\t", "\n", "hello\nΧαίρετε\t\t\nЗдравствуйте"},

		// 30
		{"Reality is\u00A0frequently inaccurate.", 20, 4, false, "", "\n", "Reality \nis\u00A0frequently \ninaccurate."},
		{"Reality is\u00a0frequently inaccurate.", 20, 4, false, "", "\n", "Reality \nis\u00a0frequently \ninaccurate."},
		{"Reality is\u2005frequently inaccurate.", 20, 4, false, "", "\n", "Reality is\u2005\nfrequently \ninaccurate."},
		{"Reality is\u2001frequently inaccurate.", 20, 4, false, "", "\n", "Reality is\u2001\nfrequently \ninaccurate."},
		{"Reality is\u200Bfrequently inaccurate.", 20, 4, false, "", "\n", "Reality is\u200B\nfrequently \ninaccurate."},
		// 35
		{"Reality is\u200bfrequently inaccurate.", 20, 4, false, "", "\n", "Reality is\u200b\nfrequently \ninaccurate."},
		{"Reality is\u202Ffrequently inaccurate.", 20, 4, false, "", "\n", "Reality \nis\u202Ffrequently \ninaccurate."},
		{"Reality is\u202ffrequently inaccurate.", 20, 4, false, "", "\n", "Reality \nis\u202ffrequently \ninaccurate."},
		{"Reality is\uFEFFfrequently inaccurate.", 20, 4, false, "", "\n", "Reality \nis\uFEFFfrequently \ninaccurate."},
		{"Space is big. You just won't believe how vastly, hugely, mind-bogglingly big it is.", 30, 4, false, "", "\n", "Space is big. You just won't \nbelieve how vastly, hugely, \nmind-bogglingly big it is."},
		// 40
		{"Space is big. You just won't believe how vastly, hugely, mind\u00adbogglingly big it is.", 30, 4, false, "", "\n", "Space is big. You just won't \nbelieve how vastly, hugely, \nmind\u00adbogglingly big it is."},
		{"Space is big. You just won't believe how vastly, hugely, mind\u2011bogglingly big it is.", 30, 4, false, "", "\n", "Space is big. You just won't \nbelieve how vastly, hugely, \nmind\u2011bogglingly big it is."},
		{"Space is big. You just won't believe how vastly, hugely, mind\u207bbogglingly big it is.", 30, 4, false, "", "\n", "Space is big. You just won't \nbelieve how vastly, hugely, \nmind\u207bbogglingly big it is."},
		{"Space is big. You just won't believe how vastly, hugely, mind\u208bbogglingly big it is.", 30, 4, false, "", "\n", "Space is big. You just won't \nbelieve how vastly, hugely, \nmind\u208bbogglingly big it is."},
		{"Space is big. You just won't believe how vastly, hugely, mind\u2e3abogglingly big it is.", 30, 4, false, "", "\n", "Space is big. You just won't \nbelieve how vastly, hugely, \nmind\u2e3abogglingly big it is."},
		// 45
		{"Space is big. You just won't believe how vastly, hugely, mind-bogglingly big it is.", 34, 4, false, "", "\n", "Space is big. You just won't \nbelieve how vastly, hugely, mind-\nbogglingly big it is."},
		{"Space is big. You just won't believe how vastly, hugely, mind\u00adbogglingly big it is.", 34, 4, false, "", "\n", "Space is big. You just won't \nbelieve how vastly, hugely, mind\u00ad\nbogglingly big it is."},
		{"Space is big. You just won't believe how vastly, hugely, mind\u2011bogglingly big it is.", 34, 4, false, "", "\n", "Space is big. You just won't \nbelieve how vastly, hugely, \nmind\u2011bogglingly big it is."},
		{"Space is big. You just won't believe how vastly, hugely, mind\u207bbogglingly big it is.", 35, 4, false, "", "\n", "Space is big. You just won't \nbelieve how vastly, hugely, \nmind\u207bbogglingly big it is."},
		{"Space is big. You just won't believe how vastly, hugely, mind\u208bbogglingly big it is.", 35, 4, false, "", "\n", "Space is big. You just won't \nbelieve how vastly, hugely, \nmind\u208bbogglingly big it is."},
		// 50
		{"Space is big. You just won't believe how vastly, hugely, mind\u2e3abogglingly big it is.", 35, 4, false, "", "\n", "Space is big. You just won't \nbelieve how vastly, hugely, mind\u2e3a\nbogglingly big it is."},
	}
	w := New()
	for i, test := range tests {
		w.Length = test.length
		w.TabSize = test.tabSize
		w.Indent = test.indent
		w.IndentVal = test.indentVal
		w.NewLine = test.newLine
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

func TestUnwrappable(t *testing.T) {
	tests := []struct {
		s       string
		newLine string
		wrapped string
	}{
		{"This sentence is a meaningless one", "\n", "This sentence is a \ufeff\nmeaningless one"},
		{"This sentence is a meaningless one", "\r\n", "This sentence is a \ufeff\r\nmeaningless one"},
		{"This sentence isn't \na meaningless sentence", "\n", "This sentence isn't \na meaningless \ufeff\nsentence"},
		{"This sentence isn't \r\na meaningless sentence", "\r\n", "This sentence isn't \r\na meaningless \ufeff\r\nsentence"},
		{"A common mistake\n that people make \nwhen trying to \ndesign something\n completely foolproof is to underestimate the ingenuity of complete fools.", "\n", "A common mistake\n that people make \nwhen trying to \ndesign something\n completely \ufeff\nfoolproof is to \ufeff\nunderestimate the \ufeff\ningenuity of \ufeff\ncomplete fools."},

		{"못\n알아\t듣겠어요\t전혀\t모르겠어요", "\n", "못\n알아\t듣겠어요\t\ufeff\n전혀\t모르겠어요"},
		{"못\n알아\ufeff\t듣겠어요\t전혀\t모르겠어요", "\n", "못\n알아\ufeff\t듣겠어요\t\ufeff\n전혀\t모르겠어요"},
		{"hello\nΧαίρετε\t\tЗдравствуйте", "\n", "hello\nΧαίρετε\t\t\ufeff\nЗдравствуйте"},
	}
	w := New()
	w.Length = 20
	w.Unwrappable = true
	for i, test := range tests {
		w.NewLine = test.newLine
		s, err := w.Line(test.s)
		if err != nil {
			t.Errorf("%d: unexpected error: %q", i, err)
			continue
		}
		if s != test.wrapped {
			t.Errorf("wrap %d: got %q want %q", i, s, test.wrapped)
		}
		s = Unwrap(s)
		if s != test.s {
			t.Errorf("unwrap %d: got %q want %q", i, s, test.s)
		}
	}
}

func TestIsHyphen(t *testing.T) {
	tests := []struct {
		r rune
		b bool
	}{
		{'\u0020', false},
		{'\u2011', false},
		{'\u207b', false},
		{'\u208b', false},
		{'\u002d', true},
		{'\u007e', true},
		{'\u00ad', true},
		{'\u058a', true},
		{'\u05be', true},
		{'\u1400', true},
		{'\u1806', true},
		{'\u2010', true},
		{'\u2012', true},
		{'\u2013', true},
		{'\u2014', true},
		{'\u2015', true},
		{'\u2053', true},
		{'\u2212', true},
		{'\u2e17', true},
		{'\u2e3a', true},
		{'\u2e3b', true},
		{'\u301c', true},
		{'\u3030', true},
		{'\u30a0', true},
		{'\ufe31', true},
		{'\ufe32', true},
		{'\ufe58', true},
		{'\ufe63', true},
		{'\uff0d', true},
	}
	for _, test := range tests {
		b := isHyphen(test.r)
		if b != test.b {
			t.Errorf("%x %c: got %t; want %t", test.r, test.r, b, test.b)
		}
	}
}

func TestProcessSpaceChunk(t *testing.T) {
	suffix := "uvwxyz"
	tests := []struct {
		line     string
		length   int
		val      string
		expected string
	}{
		{"", 10, "", "uvwxyz"},
		{"Tallamore", 10, "    ", "Tallamore\n    uvwxyz"},
		{"Tallamore", 10, "\n    ", "Tallamore\n    uvwxyz"},
		{"Monaco", 10, "\t   ", "Monaco\n\t   uvwxyz"},
		{"Monaco", 10, "\t   \n", "Monaco\t   \nuvwxyz"},
		{"Firenza", 10, "\t\n\n", "Firenza\t\n\nuvwxyz"},
		{"Amsterdam", 10, "\n\n", "Amsterdam\n\nuvwxyz"},
		{"Modena", 10, "  \n \n  ", "Modena  \n \n  uvwxyz"},
	}
	for i, test := range tests {
		w := New()
		w.Length = test.length
		w.buf.WriteString(test.line)
		w.l = w.buf.Len()
		for _, r := range test.val {
			w.runes = append(w.runes, r)
		}
		err := w.processSpaceChunk()
		if err != nil {
			t.Errorf("%d: unexpected error: %s", i, err)
			continue
		}
		w.buf.WriteString(suffix)
		if w.buf.String() != test.expected {
			t.Errorf("%d: got %q want %q", i, w.buf.String(), test.expected)
		}
	}
}

// This is to validate that making a line comment block out of text works.
// This is different than regular indent behavior in that the first line is
// also indented.
func TestCommenting(t *testing.T) {
	gpl20 := `Copyright (C) yyyy name of author
This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; version 2.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.`

	expected := `// Copyright (C) yyyy name of author
// This program is free software; you can redistribute it and/or modify it 
// under the terms of the GNU General Public License as published by the Free 
// Software Foundation; version 2.
// 
// This program is distributed in the hope that it will be useful, but WITHOUT 
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or 
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for 
// more details.
// 
// You should have received a copy of the GNU General Public License along with
// this program; if not, write to the Free Software Foundation, Inc., 51 
// Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.`

	w := New()
	w.LineComment(true)
	cmt, err := w.Line(gpl20)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	gots := strings.Split(cmt, "\n")
	wants := strings.Split(expected, "\n")
	if len(gots) != len(wants) {
		t.Errorf("got %d lines; want %d", len(gots), len(wants))
		t.Errorf("got %q\nwant %q", cmt, expected)
		return
	}
	for i, got := range gots {
		if got != wants[i] {
			t.Errorf("%d: got %q want %q", i, got, wants[i])
		}
	}
}
