# go fmt
| format | description |
| --- | --- |
| %[+\|#]v | default (+ field names, # syntax) |
| %T | type |
| %% | escape |
| %t | bool |
| %b, %o, %d, %x, %X | integer base 2, 8, 10, 16 lower, 16 upper |
| %c | unicode codepoint |
| %U | unicode format (U+1234) |
| %b | decimal-less scinotation exp a power of two e.g. -123456p-78 |
| %e, %E |scientific notation, e.g. -1.234456e+78|
| %f, %F |decimal point but no exponent, e.g. 123.456|
| %g, %G |%e\|E for large exponents, %f\|F otherwise|
|%s | the uninterpreted bytes of the string or slice |
|%q | a double-quoted string safely escaped with Go syntax |
|%x, %X | base 16, lower\|upper-case, two characters per byte |
| %p | pointer|
