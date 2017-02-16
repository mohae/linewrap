package linewrap

import (
	"strings"
	"testing"
)

func TestWrapLine(t *testing.T) {
	tests := []struct {
		s          string
		length     int
		tabSize    int
		indentText string
		expected   string
	}{
		{"", 20, 4, "", ""},
		{"Hello", 20, 4, "", "Hello"},
		{"Hello World", 20, 4, "", "Hello World"},
		{"This sentence is a\n meaningless one", 20, 4, "", "This sentence is a\nmeaningless one"},
		{"This sentence is a \nmeaningless one", 20, 4, "", "This sentence is a\nmeaningless one"},
		// 5
		{"This sentence is a\n meaningless one", 20, 4, "    ", "This sentence is a\n    meaningless one"},
		{"This sentence is a\n meaningless one", 20, 4, "\t", "This sentence is a\n\tmeaningless one"},
		{"This sentence is a\r\n meaningless one", 20, 4, "    ", "This sentence is a\n    meaningless one"},
		{"This sentence is a\r\n meaningless one", 20, 4, "\t", "This sentence is a\n\tmeaningless one"},
		{"This sentence isn't\r\n a meaningless one", 20, 4, "    ", "This sentence isn't\n    a meaningless\n    one"},
		// 10
		{"This sentence isn't\r\n a meaningless one", 20, 4, "\t", "This sentence isn't\n\ta meaningless\n\tone"},
		{"Reality is frequently inaccurate. One is never alone with a rubber duck.", 34, 4, "", "Reality is frequently inaccurate.\nOne is never alone with a rubber\nduck."},
		{"Reality is frequently inaccurate. One is never alone with a rubber duck.", 35, 4, "", "Reality is frequently inaccurate.\nOne is never alone with a rubber\nduck."},
		{"Reality is frequently inaccurate.     One is never alone with a rubber duck.", 34, 4, "", "Reality is frequently inaccurate.\nOne is never alone with a rubber\nduck."},
		{"Reality is frequently inaccurate.     One is never alone with a rubber duck.", 35, 4, "", "Reality is frequently inaccurate.\nOne is never alone with a rubber\nduck."},
		// 15
		{"Reality is frequently inaccurate.     One is never alone with a rubber duck.", 40, 4, "", "Reality is frequently inaccurate.\nOne is never alone with a rubber duck."},
		{"A common mistake\n that people make when trying to design something completely foolproof is to underestimate the ingenuity of complete fools.", 20, 4, "", "A common mistake\nthat people make\nwhen trying to\ndesign something\ncompletely\nfoolproof is to\nunderestimate the\ningenuity of\ncomplete fools."},
		{"못\t알아\t듣겠어요\t전혀\t모르겠어요", 20, 4, "", "못\t알아\t듣겠어요\t\n전혀\t모르겠어요"},
		{"못\t알아\t듣겠어요\t전혀\t모르겠어요", 20, 4, "    ", "못\t알아\t듣겠어요\t\n    전혀\t모르겠어요"},
		{"못\t알아\t듣겠어요\t전혀\t모르겠어요", 20, 4, "\t", "못\t알아\t듣겠어요\t\n\t전혀\t모르겠어요"},
		// 20
		{"hello\nΧαίρετε\t\tЗдравствуйте", 20, 4, "", "hello\nΧαίρετε\t\t\nЗдравствуйте"},
		{"hello\nΧαίρετε\t\tЗдравствуйте", 20, 4, "    ", "hello\n    Χαίρετε\t\t\n    Здравствуйте"},
		{"hello\nΧαίρετε\t\tЗдравствуйте", 20, 4, "\t", "hello\n\tΧαίρετε\t\t\n\tЗдравствуйте"},
		{"Reality is\u00A0frequently inaccurate.", 20, 4, "", "Reality\nis\u00A0frequently\ninaccurate."},
		{"Reality is\u00a0frequently inaccurate.", 20, 4, "", "Reality\nis\u00a0frequently\ninaccurate."},
		// 25
		{"Reality is\u2005frequently inaccurate.", 20, 4, "", "Reality is\nfrequently\ninaccurate."},
		{"Reality is\u2001frequently inaccurate.", 20, 4, "", "Reality is\nfrequently\ninaccurate."},
		{"Reality is\u200bfrequently inaccurate.", 20, 4, "", "Reality is\nfrequently\ninaccurate."},
		{"Reality is\u202Ffrequently inaccurate.", 20, 4, "", "Reality\nis\u202Ffrequently\ninaccurate."},
		{"Reality is\uFEFFfrequently inaccurate.", 20, 4, "", "Reality\nis\uFEFFfrequently\ninaccurate."},
		// 30
		{"Space is big. You just won't believe how vastly, hugely, mind-bogglingly big it is.", 30, 4, "", "Space is big. You just won't\nbelieve how vastly, hugely,\nmind-bogglingly big it is."},
		{"Space is big. You just won't believe how vastly, hugely, mind\u00adbogglingly big it is.", 30, 4, "", "Space is big. You just won't\nbelieve how vastly, hugely,\nmind\u00adbogglingly big it is."},
		{"Space is big. You just won't believe how vastly, hugely, mind\u2011bogglingly big it is.", 30, 4, "", "Space is big. You just won't\nbelieve how vastly, hugely,\nmind\u2011bogglingly big it is."},
		{"Space is big. You just won't believe how vastly, hugely, mind\u207bbogglingly big it is.", 30, 4, "", "Space is big. You just won't\nbelieve how vastly, hugely,\nmind\u207bbogglingly big it is."},
		{"Space is big. You just won't believe how vastly, hugely, mind\u208bbogglingly big it is.", 30, 4, "", "Space is big. You just won't\nbelieve how vastly, hugely,\nmind\u208bbogglingly big it is."},
		// 35
		{"Space is big. You just won't believe how vastly, hugely, mind\u2e3abogglingly big it is.", 30, 4, "", "Space is big. You just won't\nbelieve how vastly, hugely,\nmind\u2e3abogglingly big it is."},
		{"Space is big. You just won't believe how vastly, hugely, mind\u00adbogglingly big it is.", 34, 4, "", "Space is big. You just won't\nbelieve how vastly, hugely, mind\u00ad\nbogglingly big it is."},
		{"Space is big. You just won't believe how vastly, hugely, mind\u2011bogglingly big it is.", 34, 4, "", "Space is big. You just won't\nbelieve how vastly, hugely,\nmind\u2011bogglingly big it is."},
		{"Space is big. You just won't believe how vastly, hugely, mind\u207bbogglingly big it is.", 35, 4, "", "Space is big. You just won't\nbelieve how vastly, hugely, mind\u207b\nbogglingly big it is."},
	}

	w := New()
	for i, test := range tests {
		w.Reset()
		w.Length = test.length
		w.TabSize(test.tabSize)
		w.IndentText(test.indentText)
		s, err := w.String(test.s)
		if err != nil {
			t.Errorf("%d: unexpected error: %q", i, err)
			continue
		}
		if s != test.expected {
			t.Errorf("%d: got %q want %q", i, s, test.expected)
		}
	}
}

var gpl20 = `Copyright (C) yyyy name of author
This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; version 2.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.`

// This is to validate that making a line comment block out of text works.
// This is different than regular indent behavior in that the first line is
// also indented.
func TestLineCommentSlashes(t *testing.T) {
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
	w.CommentStyle = CPPComment
	cmt, err := w.String(gpl20)
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

func TestLineCommentHashes(t *testing.T) {
	expected := `# Copyright (C) yyyy name of author
# This program is free software; you can redistribute it and/or modify it under
# the terms of the GNU General Public License as published by the Free Software
# Foundation; version 2.
#
# This program is distributed in the hope that it will be useful, but WITHOUT
# ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
# FOR A PARTICULAR PURPOSE. See the GNU General Public License for more
# details.
#
# You should have received a copy of the GNU General Public License along with
# this program; if not, write to the Free Software Foundation, Inc., 51
# Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.`

	w := New()
	w.CommentStyle = ShellComment
	cmt, err := w.String(gpl20)
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

var mit = `MIT License
Copyright (c) <year> <copyright holders>

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
`

func TestLineWrapLeftJustify(t *testing.T) {

	expected := `/*
MIT License
Copyright (c) <year> <copyright holders>

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
`
	// use the globala
	w := New()
	w.CommentStyle = CComment
	cmt, err := w.String(mit)
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

func TestCommentStyleStringer(t *testing.T) {
	tests := []struct {
		name     string
		style    CommentStyle
		expected string
	}{
		{"invalid", CommentStyle(-1), "invalid: -1 style comments"},
		{"none", NoComment, "none"},
		{"c++", CPPComment, "c++ style comments"},
		{"shell", ShellComment, "shell style comments"},
		{"c", CComment, "c style comments"},
	}

	for _, test := range tests {
		s := test.style.String()
		if s != test.expected {
			t.Errorf("%s: got %q want %q", test.name, s, test.expected)
		}
	}
}
