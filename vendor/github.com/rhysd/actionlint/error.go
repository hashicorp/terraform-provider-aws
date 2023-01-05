package actionlint

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
)

var (
	bold   = color.New(color.Bold)
	green  = color.New(color.FgGreen)
	yellow = color.New(color.FgYellow)
	gray   = color.New(color.FgHiBlack)
)

// Error represents an error detected by actionlint rules
type Error struct {
	// Message is an error message.
	Message string
	// Filepath is a file path where the error occurred.
	Filepath string
	// Line is a line number where the error occurred. This value is 1-based.
	Line int
	// Column is a column number where the error occurred. This value is 1-based.
	Column int
	// Kind is a string to represent kind of the error. Usually rule name which found the error.
	Kind string
}

// Error returns summary of the error as string.
func (e *Error) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s [%s]", e.Filepath, e.Line, e.Column, e.Message, e.Kind)
}

func (e *Error) String() string {
	return e.Error()
}

func errorAt(pos *Pos, kind string, msg string) *Error {
	return &Error{
		Message: msg,
		Line:    pos.Line,
		Column:  pos.Col,
		Kind:    kind,
	}
}

func errorfAt(pos *Pos, kind string, format string, args ...interface{}) *Error {
	return &Error{
		Message: fmt.Sprintf(format, args...),
		Line:    pos.Line,
		Column:  pos.Col,
		Kind:    kind,
	}
}

// GetTemplateFields fields for formatting this error with Go template.
func (e *Error) GetTemplateFields(source []byte) *ErrorTemplateFields {
	snippet := ""
	end := e.Column
	if len(source) > 0 && e.Line > 0 {
		if l, ok := e.getLine(source); ok {
			snippet = l
			if len(l) >= e.Column-1 {
				if i := e.getIndicator(l); i != "" {
					snippet += "\n" + i
					end = len(i) // Byte length can be used here because this line only contains ASCII
				}
			}
		}
	}

	return &ErrorTemplateFields{
		Message:   e.Message,
		Filepath:  e.Filepath,
		Line:      e.Line,
		Column:    e.Column,
		Kind:      e.Kind,
		Snippet:   snippet,
		EndColumn: end,
	}
}

// PrettyPrint prints the error with user-friendly way. It prints file name, source position, error
// message with colorful output and source snippet with indicator. When nil is set to source, no
// source snippet is not printed. To disable colorful output, set true to fatih/color.NoColor.
func (e *Error) PrettyPrint(w io.Writer, source []byte) {
	yellow.Fprint(w, e.Filepath)
	gray.Fprint(w, ":")
	fmt.Fprint(w, e.Line)
	gray.Fprint(w, ":")
	fmt.Fprint(w, e.Column)
	gray.Fprint(w, ": ")
	bold.Fprint(w, e.Message)
	gray.Fprintf(w, " [%s]\n", e.Kind)

	if len(source) == 0 || e.Line <= 0 {
		return
	}
	line, ok := e.getLine(source)
	if !ok || len(line) < e.Column-1 {
		return
	}

	lnum := fmt.Sprintf("%d | ", e.Line)
	indent := strings.Repeat(" ", len(lnum)-2)
	gray.Fprintf(w, "%s|\n", indent)
	gray.Fprint(w, lnum)
	fmt.Fprintln(w, line)
	gray.Fprintf(w, "%s| ", indent)
	green.Fprintln(w, e.getIndicator(line))
}

func (e *Error) getLine(source []byte) (string, bool) {
	s := bufio.NewScanner(bytes.NewReader(source))
	l := 0
	for s.Scan() {
		l++
		if l == e.Line {
			return s.Text(), true
		}
	}
	return "", false
}

func (e *Error) getIndicator(line string) string {
	if e.Column <= 0 {
		return ""
	}

	start := e.Column - 1 // Column is 1-based

	// Count width of non-space characters after '^' for underline
	uw := 0
	r := strings.NewReader(line[start:])
	for {
		c, s, err := r.ReadRune()
		if err != nil || s == 0 || c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			break
		}
		uw += runewidth.RuneWidth(c)
	}
	if uw > 0 {
		uw-- // Decrement for place for '^'
	}

	// Count width of spaces before '^'
	sw := runewidth.StringWidth(line[:start])
	return fmt.Sprintf("%s^%s", strings.Repeat(" ", sw), strings.Repeat("~", uw))
}

// ByErrorPosition is predicate for sort.Interface. It sorts errors slice by file path, line, and
// column.
type ByErrorPosition []*Error

func (by ByErrorPosition) Len() int {
	return len(by)
}

func (by ByErrorPosition) Less(i, j int) bool {
	if c := strings.Compare(by[i].Filepath, by[j].Filepath); c != 0 {
		return c < 0
	}
	if by[i].Line == by[j].Line {
		return by[i].Column < by[j].Column
	}
	return by[i].Line < by[j].Line
}

func (by ByErrorPosition) Swap(i, j int) {
	by[i], by[j] = by[j], by[i]
}

// ErrorTemplateFields holds all fields to format one error message.
type ErrorTemplateFields struct {
	// Message is error message body.
	Message string `json:"message"`
	// Filepath is a canonical relative file path. This is empty when input was read from stdin.
	// When encoding into JSON, this field may be omitted when the file path is empty.
	Filepath string `json:"filepath,omitempty"`
	// Line is a line number of error position.
	Line int `json:"line"`
	// Column is a column number of error position.
	Column int `json:"column"`
	// Kind is a rule name the error belongs to.
	Kind string `json:"kind"`
	// Snippet is a code snippet and indicator to indicate where the error occurred.
	// When encoding into JSON, this field may be omitted when the snippet is empty.
	Snippet string `json:"snippet,omitempty"`
	// EndColumn is a column number where the error indicator (^~~~~~~) ends. When no indicator
	// can be shown, EndColumn is equal to Column.
	EndColumn int `json:"end_column"`
}

func unescapeBackslash(s string) string {
	// https://golang.org/ref/spec#Rune_literals
	r := strings.NewReplacer(
		`\a`, "\a",
		`\b`, "\b",
		`\f`, "\f",
		`\n`, "\n",
		`\r`, "\r",
		`\t`, "\t",
		`\v`, "\v",
		`\\`, "\\",
	)
	return r.Replace(s)
}

// ErrorFormatter is a formatter to format a slice of ErrorTemplateFields. It is used for
// formatting error messages with -format option.
type ErrorFormatter struct {
	temp *template.Template
}

// NewErrorFormatter creates new ErrorFormatter instance. Given format must contain at least one
// {{ }} placeholder. Escaped characters like \n in the format string are unescaped.
func NewErrorFormatter(format string) (*ErrorFormatter, error) {
	if !strings.Contains(format, "{{") {
		return nil, fmt.Errorf("template to format error messages must contain at least one {{ }} placeholder: %s", format)
	}
	funcs := template.FuncMap(map[string]interface{}{
		"json": func(data interface{}) (string, error) {
			var b strings.Builder
			enc := json.NewEncoder(&b)
			if err := enc.Encode(data); err != nil {
				return "", fmt.Errorf("could not encode template value into JSON: %w", err)
			}
			return b.String(), nil
		},
		"replace": func(s string, oldnew ...string) string {
			return strings.NewReplacer(oldnew...).Replace(s)
		},
	})
	t, err := template.New("error formatter").Funcs(funcs).Parse(unescapeBackslash(format))
	if err != nil {
		return nil, fmt.Errorf("template %q to format error messages could not be parsed: %w", format, err)
	}
	return &ErrorFormatter{t}, nil
}

// Print formats the slice of template fields and prints it with given writer.
func (f *ErrorFormatter) Print(out io.Writer, t []*ErrorTemplateFields) error {
	if err := f.temp.Execute(out, t); err != nil {
		return fmt.Errorf("could not format error messages: %w", err)
	}
	return nil
}

// PrintErrors prints the errors after formatting them with template.
func (f *ErrorFormatter) PrintErrors(out io.Writer, errs []*Error, src []byte) error {
	t := make([]*ErrorTemplateFields, 0, len(errs))
	for _, err := range errs {
		t = append(t, err.GetTemplateFields(src))
	}
	return f.Print(out, t)
}
