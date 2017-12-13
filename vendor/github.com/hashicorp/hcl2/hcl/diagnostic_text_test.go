package hcl

import (
	"bytes"
	"fmt"
	"testing"
)

func TestDiagnosticTextWriter(t *testing.T) {
	tests := []struct {
		Input *Diagnostic
		Want  string
	}{
		{
			&Diagnostic{
				Severity: DiagError,
				Summary:  "Splines not reticulated",
				Detail:   "All splines must be pre-reticulated.",
				Subject: &Range{
					Start: Pos{
						Byte:   0,
						Column: 1,
						Line:   1,
					},
					End: Pos{
						Byte:   3,
						Column: 4,
						Line:   1,
					},
				},
			},
			`Error: Splines not reticulated

  on  line 1, in hardcoded-context:
   1: foo = 1

All splines must be pre-reticulated.

`,
		},
		{
			&Diagnostic{
				Severity: DiagError,
				Summary:  "Unsupported attribute",
				Detail:   `"baz" is not a supported top-level attribute. Did you mean "bam"?`,
				Subject: &Range{
					Start: Pos{
						Column: 1,
						Line:   3,
					},
					End: Pos{
						Column: 4,
						Line:   3,
					},
				},
			},
			`Error: Unsupported attribute

  on  line 3, in hardcoded-context:
   3: baz = 3

"baz" is not a supported top-level
attribute. Did you mean "bam"?

`,
		},
		{
			&Diagnostic{
				Severity: DiagError,
				Summary:  "Unsupported attribute",
				Detail:   `"pizza" is not a supported attribute. Did you mean "pizzetta"?`,
				Subject: &Range{
					Start: Pos{
						Column: 3,
						Line:   5,
					},
					End: Pos{
						Column: 8,
						Line:   5,
					},
				},
				// This is actually not a great example of a context, but is here to test
				// whether we're able to show a multi-line context when needed.
				Context: &Range{
					Start: Pos{
						Column: 1,
						Line:   4,
					},
					End: Pos{
						Column: 2,
						Line:   6,
					},
				},
			},
			`Error: Unsupported attribute

  on  line 5, in hardcoded-context:
   4: block "party" {
   5:   pizza = "cheese"
   6: }

"pizza" is not a supported attribute.
Did you mean "pizzetta"?

`,
		},
	}

	files := map[string]*File{
		"": &File{
			Bytes: []byte(testDiagnosticTextWriterSource),
			Nav:   &diagnosticTestNav{},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			bwr := &bytes.Buffer{}
			dwr := NewDiagnosticTextWriter(bwr, files, 40, false)
			err := dwr.WriteDiagnostic(test.Input)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			got := bwr.String()
			if got != test.Want {
				t.Errorf("wrong result\n\ngot:\n%swant:\n%s", got, test.Want)
			}
		})
	}
}

const testDiagnosticTextWriterSource = `foo = 1
bar = 2
baz = 3
block "party" {
  pizza = "cheese"
}
`

type diagnosticTestNav struct {
}

func (tn *diagnosticTestNav) ContextString(offset int) string {
	return "hardcoded-context"
}
