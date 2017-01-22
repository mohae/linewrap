# linewrap
wraps a string so that each line the are <= Length. Lines can be optionally indented. When indenting isn't enabled, linewrap will preserve all whitespaces in the string and only insert new line characters.

Linewrap calculates line length on a per character basis with the exception of '\t', which is counted as TabSize spaces, which is configurable using `Wrap.TabSize`, for the sake of calculating the current line length. The new line character to use is also configurable using `Wrap.NewLine`. Any existing new line characters in the string  will be replaced with the `Wrap.NewLine` new line character and any inserted new line will also use `Wrap.NewLine`. This can be set on the package level.

If lines are indented the whitespace around the line-break is elided. The indent value is configurable using `Wrap.IndentVal`. Indenting is controlled by the `Wrap.Indent` bool.

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
