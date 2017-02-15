# linewrap
[![GoDoc](https://godoc.org/github.com/mohae/linewrap?status.svg)](https://godoc.org/github.com/mohae/linewrap)[![Build Status](https://travis-ci.org/mohae/linewrap.png)](https://travis-ci.org/mohae/linewrap)  
Wraps either a string or a byte slice so that each line doesn't exceed the specified number of characters. A character is defined as a unicode code point, not a byte. Any `\r` in the input will be elided.

Trailing and leading spaces on wrapped lines are elided.

Linewrap can also indent wrapped lines or format the input as comments:

    #      line comment
	//     line comment
	/* */  block comment

## References used:
Jukka "Yucca" Korpela's unicode tables were used as the main references especially his line break page: https://www.cs.tut.fi/~jkorpela/unicode/linebr.html and his unicode spaces and dash pages listed in their respective sections in this document.

In addition, some of the symbols were pulled from http://www.unicode.org/reports/tr14/#Properties.

The list of symbols handled is not exhaustive.


## Hyphen and spaces
Linewrap will wrap lines on most unicode whitespace and dash characters, with some exceptions. Characters in the not considered list will not be considered points at which the input can be wrapped. If there are any characters that are unaccounted for, please file an issue or make a pull request. Before doing so, check the docs and/or the code to see if it has been already listed as an exception.

The `\n` and `\t` characters are handled separately.  The width used for tabs is set by `Wrap.TabSize(int)`, which defaults to 8 spaces.

### Spaces
Whitespace tokens are mostly from https://www.cs.tut.fi/~jkorpela/chars/spaces.html

#### Not considered whitespace characters:
code point|symbol name  
--|:--  
U+200B|zero width space  
U+00A0|no-break space  
U+202F|zero width no-break space  

#### Whitespace characters
code point|symbol name  
--|:--|:--  
U+0020|space  
U+1680|ogham space mark  
U+180E|mongolian vowel separator  
U+2000|en quad  
U+2001|em quad  
U+2002|en space  
U+2003|em space  
U+2004|three per em space  
U+2005|four per em space  
U+2006|six per em space  
U+2007|figure space  
U+2008|punctuation space  
U+2009|thin space  
U+200A|hair space  
U+200B|zero width space  
U+205F|medium mathematical space  
U+3000|ideographic space  

### Dashes (Hyphens)
Dash tokens are mostly from dash tokens from https://www.cs.tut.fi/~jkorpela/dashes.html

__Additional explanations to entries in the tables:__
The `em dash (U+2014)` symbol can have a break before or after its occurrence but linewrap only breaks after its occurrence.

The `hyphen minus (U+002D)` is not supposed to break on a numeric context but linewrap does not make such a differentiation.

#### Dash characters not considered dashes  
code point|symbol name  
--|:--:  
U+007E|tilde            U+007E  
U+2212|minus sign       U+2212  
U+301C|wavy dash        U+301C  
U+3939|wavy dash        U+3939  
U+2E3A|two em dash      U+2E3A  
U+2E3B|three em dash    U+2E3B  
U+FE58|small em dash    U+FE58  
U+FE63|small hyphen-minus       U+FE63  
U+FF0D|full width hyphen-minus  U+FF0D  
U+1806|mongolian todo hyphen    U+1806  
U+FE31|presentation form for vertical em dash  
U+FE32|presentation form for vertical en dash  

#### Dash characters
code point|symbol name  
--|:--:  
U+002D|hyphen minus  
U+00AD|soft hyphen  
U+058A|armenian hyphen  
U+2010|hyphen  
U+2012|figure dash  
U+2013|en dash  
U+2014|em dash  
U+2015|horizontal bar  
U+2053|swung dash  
U+207B|superscript mnus  
U+208B|subscript minus  
U+2E3A|two em dash  
U+2E3B|three em dash  
U+FE31|presentation form for vertical em dash  
U+FE32|presentation form for vertical en dash  
U+FE58|small em dash  
U+FE63|small hyphen minus  
U+FF0D|full width hyphen minus  
