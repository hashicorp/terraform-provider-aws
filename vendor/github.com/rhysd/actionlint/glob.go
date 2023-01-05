package actionlint

import (
	"fmt"
	"strings"
	"text/scanner"
	"unicode"
)

// Note:
// - Broken pattern causes a syntax error
//   - '+' or '?' at top of pattern
//   - preceding character of '+' or '?' is special character like '+', '?', '*'
//   - Missing ] in [...] pattern like '[0-9'
//   - Missing end of range in [...] like '[0-]'
// - \ can escape special characters like '['. Otherwise \ is handled as normal character
// - invalid characters for Git ref names are not checked on GitHub Actions runtime
//   - `man git-check-ref-format` for more details
//   - \ is invalid character for ref names. it means that \ can be used only for escaping special chars

// InvalidGlobPattern is an error on invalid glob pattern.
type InvalidGlobPattern struct {
	// Message is a human readable error message.
	Message string
	// Column is a column number of the error in the glob pattern. This value is 1-based, but zero
	// is valid value. Zero means the error occurred before reading first character. This happens
	// when a given pattern is empty. When the given pattern include a newline and line number
	// increases (invalid pattern), the column number falls back into always 0.
	Column int
}

func (err *InvalidGlobPattern) Error() string {
	return fmt.Sprintf("%d: %s", err.Column, err.Message)
}

func (err *InvalidGlobPattern) String() string {
	return err.Error()
}

type globValidator struct {
	isRef bool
	prec  bool
	errs  []InvalidGlobPattern
	scan  scanner.Scanner
}

func (v *globValidator) error(msg string) {
	p := v.scan.Pos()
	// - 1 because character at the error position is already eaten from scanner
	c := p.Column - 1
	if p.Line > 1 {
		c = 0 // fallback to 0
	}
	v.errs = append(v.errs, InvalidGlobPattern{msg, c})
}

func (v *globValidator) unexpected(char rune, what, why string) {
	unexpected := "unexpected EOF"
	if char != scanner.EOF {
		unexpected = fmt.Sprintf("unexpected character %q", char)
	}

	while := ""
	if what != "" {
		while = fmt.Sprintf(" while checking %s", what)
	}

	v.error(fmt.Sprintf("invalid glob pattern. %s%s. %s", unexpected, while, why))
}

func (v *globValidator) invalidRefChar(c rune, why string) {
	cfmt := "%q"
	if unicode.IsPrint(c) {
		cfmt = "'%c'" // avoid '\\'
	}
	format := "character " + cfmt + " is invalid for branch and tag names. %s. see `man git-check-ref-format` for more details. note that regular expression is unavailable"
	msg := fmt.Sprintf(format, c, why)
	v.error(msg)
}

func (v *globValidator) init(pat string) {
	v.errs = []InvalidGlobPattern{}
	v.prec = false
	v.scan.Init(strings.NewReader(pat))
	v.scan.Error = func(s *scanner.Scanner, m string) {
		v.error(fmt.Sprintf("error while scanning glob pattern %q: %s", pat, m))
	}
}

func (v *globValidator) validateNext() bool {
	c := v.scan.Next()
	prec := true

	switch c {
	case '\\':
		switch v.scan.Peek() {
		case '[', '?', '*':
			c = v.scan.Next() // eat escaped character
			if v.isRef {
				v.invalidRefChar(v.scan.Peek(), "ref name cannot contain spaces, ~, ^, :, [, ?, *")
			}
		case '+', '\\', '!':
			c = v.scan.Next() // eat escaped character
		default:
			// file path can contain '\' (`mkdir 'foo\bar'` works)
			if v.isRef {
				v.invalidRefChar('\\', "only special characters [, ?, +, *, \\ ! can be escaped with \\")
				c = v.scan.Next()
			}
		}
	case '?':
		if !v.prec {
			v.unexpected('?', "special character ? (zero or one)", "the preceding character must not be special character")
		}
		prec = false
	case '+':
		if !v.prec {
			v.unexpected('+', "special character + (one or more)", "the preceding character must not be special character")
		}
		prec = false
	case '*':
		prec = false
	case '[':
		if v.scan.Peek() == ']' {
			c = v.scan.Next() // eat ]
			v.unexpected(']', "content of character match []", "character match must not be empty")
			break
		}

		chars := 0
	Loop:
		for {
			c = v.scan.Next()
			switch c {
			case ']':
				break Loop
			case scanner.EOF:
				v.unexpected(c, "end of character match []", "missing ]")
				return false
			default:
				if v.scan.Peek() != '-' {
					// in case of single character
					chars++
					continue Loop
				}
				// When match is range of character like 0-9

				chars += 2 // actually one or more. but this is ok since we only check chars > 1 later
				s := c
				//lint:ignore SA4006 c should always holds the current character even if it is unused
				c = v.scan.Next() // eat -
				switch v.scan.Peek() {
				case ']':
					c = v.scan.Next() // eat ]
					v.unexpected(c, "character range in []", "end of range is missing")
					break Loop
				case scanner.EOF:
					// do nothing
				default:
					c = v.scan.Next() // eat end of range
					if s > c {
						why := fmt.Sprintf("start of range %q (%d) is larger than end of range %q (%d)", s, s, c, c)
						v.unexpected(c, "character range in []", why)
					}
				}
			}
		}

		if chars == 1 {
			v.unexpected(c, "character match []", "character match with single character is useless. simply use x instead of [x]")
		}
	case '\r':
		if v.scan.Peek() == '\n' {
			c = v.scan.Next()
		}
		v.unexpected(c, "", "newline cannot be contained")
	case '\n':
		v.unexpected('\n', "", "newline cannot be contained")
	case ' ', '\t', '~', '^', ':':
		if v.isRef {
			v.invalidRefChar(c, "ref name cannot contain spaces, ~, ^, :, [, ?, *")
		}
	default:
	}
	v.prec = prec

	if v.scan.Peek() == scanner.EOF {
		if v.isRef && (c == '/' || c == '.') {
			v.invalidRefChar(c, "ref name must not end with / and .")
		}
		return false
	}

	return true
}

func (v *globValidator) validate(pat string) {
	v.init(pat)

	if pat == "" {
		v.error("glob pattern cannot be empty")
		return
	}

	// Handle first character if necessary
	switch v.scan.Peek() {
	case '/':
		if v.isRef {
			v.scan.Next()
			v.invalidRefChar('/', "ref name must not start with /")
			v.prec = true
		}
	case '!':
		v.scan.Next()
		if v.scan.Peek() == scanner.EOF {
			v.unexpected('!', "! at first character (negate pattern)", "at least one character must follow !")
			return
		}
		v.prec = false
	}

	for v.validateNext() {
	}
}

func validateGlob(pat string, isRef bool) []InvalidGlobPattern {
	v := globValidator{}
	v.isRef = isRef
	v.validate(pat)
	return v.errs
}

// ValidateRefGlob checks a given input as glob pattern for Git ref names. It returns list of
// errors found by the validation. See the following URL for more details of the syntax:
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet
func ValidateRefGlob(pat string) []InvalidGlobPattern {
	return validateGlob(pat, true)
}

// ValidatePathGlob checks a given input as glob pattern for file paths. It returns list of
// errors found by the validation. See the following URL for more details of the syntax:
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet
func ValidatePathGlob(pat string) []InvalidGlobPattern {
	if strings.HasPrefix(pat, " ") {
		return []InvalidGlobPattern{
			{"path value must not start with spaces", 0},
		}
	}
	if strings.HasSuffix(pat, " ") {
		return []InvalidGlobPattern{
			{"path value must not end with spaces", len(pat)},
		}
	}
	return validateGlob(pat, false)
}
