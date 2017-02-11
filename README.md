# linewrap
[![GoDoc](https://godoc.org/github.com/mohae/linewrap?status.svg)](https://godoc.org/github.com/mohae/linewrap)[![Build Status](https://travis-ci.org/mohae/linewrap.png)](https://travis-ci.org/mohae/linewrap)  
Wraps either a string or a byte slice so that each line doesn't exceed the specified number of characters. The wrapping can be done in an unwrappable manner, have wrapped lines be indented, or remove all white spaces around the wrapped lines. A character is defined as a unicode code point, not a byte.

If wrapping is done in an unwrappable manner, a zero-space no break space, `U+FEFF`, will be inserted prior to any inserted new line characters. This allows `Unwrap` to detect inserted new line characters and elide them while preserving any pre-existing new line characters. When a line is being wrapped in a unwrappable, all existing white spaces are preserved with the possible exception of new line characters. Since new line characters can be replaced during wrapping, e.g. `\r\n` replaced by `\n`, the original values cannot be restored as there is no way of determining what the original new line characters were.

If wrapping is not done in an unwrappable manner, lines can be optionally indented using the configured `Wrap.Indent` value. When indenting isn't enabled, linewrap will preserve all white spaces in the string and only insert new line characters.

Linewrap calculates line length on a per character, or rune, basis with the exception of '\t', which is counted as TabSize spaces, which is configurable using `Wrap.TabSize`, for the sake of calculating the current line length. While the defined number of spaces for a tab _depends_, this is the best approximation that can be made in this situation. Alternatively, the tab can be replaced with `Wrap.TabSize` spaces. The new line character to use is also configurable using `Wrap.NewLine`. Any existing new line characters in the string  will be replaced with the `Wrap.NewLine` new line character and any inserted new line will also use `Wrap.NewLine`. This can be set on the package level, which will result in any `Wrap` created by `New` having those settings.

If lines are indented the whitespace around the line-break is elided. The indent value is configurable using `Wrap.IndentVal`. Indenting is controlled by the `Wrap.Indent` bool.

Linewrap can also be used to create line-comment blocks out of text. If it is configured to do line comments the text will not be indented after a new line; all lines will be left justified.

The `Wrap` struct can be re-used.

## Hyphen and spaces
Linewrap will wrap lines on unicode whitespace and hyphens.

### Spaces
Go's `unicode.IsSpace()` function is used to determine if a char is a whitespace character. Linewrap classifies a few code points differently than `unicode.IsSpace` does. These differences are shown in the table below.

code point|name|IsSpace  
--|:--|:--  
U+200B|Zero Width Space|True  
U+00A0|No-break Space|False  
U+202F|Zero Width No-break Space|False

### Hyphens
Go's unicode.Dash range table, along with a few ASCII and extended ASCII code points, are used to determine if a char is a hyphen character. Linewrap classifies a few code points differently. These exceptions are shown in the table below.

code point|symbol|name|IsHyphen  
--|:--:|:--|:--  
U+2011|‑|No-break Hyphen|False  
U+207B|⁻|Superscript Minus|False  
U+208B|₋|Subscript Minus|False
