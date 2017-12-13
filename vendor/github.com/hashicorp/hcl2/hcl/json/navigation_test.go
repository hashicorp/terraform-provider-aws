package json

import (
	"strconv"
	"testing"
)

func TestNavigationContextString(t *testing.T) {
	src := `
{
  "version": 1,
  "resource": {
    "null_resource": {
      "baz": {
        "id": "foo"
      }
    }
  }
}
`
	file, diags := Parse([]byte(src), "test.json")
	if len(diags) != 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
	if file == nil {
		t.Fatalf("Got nil file")
	}
	nav := file.Nav.(navigation)

	tests := []struct {
		Offset int
		Want   string
	}{
		{0, ``},
		{8, ``},
		{36, `resource`},
		{60, `resource.null_resource`},
		{89, `resource.null_resource.baz`},
	}

	for _, test := range tests {
		t.Run(strconv.Itoa(test.Offset), func(t *testing.T) {
			got := nav.ContextString(test.Offset)

			if got != test.Want {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", got, test.Want)
			}
		})
	}
}
