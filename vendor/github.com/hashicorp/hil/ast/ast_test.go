package ast

import (
	"strconv"
	"testing"
)

func TestPosString(t *testing.T) {
	cases := []struct {
		Input  Pos
		String string
	}{
		{
			Pos{Line: 1, Column: 1},
			"1:1",
		},
		{
			Pos{Line: 2, Column: 3},
			"2:3",
		},
		{
			Pos{Line: 3, Column: 2, Filename: "template.hil"},
			"template.hil:3:2",
		},
	}

	for i, tc := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := tc.Input.String()
			if want, got := tc.String, got; want != got {
				t.Errorf("%#v produced %q; want %q", tc.Input, got, want)
			}
		})
	}
}
