package textseg

import (
	"bufio"
	"reflect"
	"testing"
)

func TestAllTokens(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{
			``,
			[]string{},
		},
		{
			`hello`,
			[]string{
				`hello`,
			},
		},
		{
			`hello world`,
			[]string{
				`hello`,
				`world`,
			},
		},
		{
			`hello worldly world`,
			[]string{
				`hello`,
				`worldly`,
				`world`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			gotBytes, err := AllTokens([]byte(test.input), bufio.ScanWords)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			got := make([]string, len(gotBytes))
			for i, buf := range gotBytes {
				got[i] = string(buf)
			}

			if !reflect.DeepEqual(got, test.want) {
				wantBytes := make([][]byte, len(test.want))
				for i, str := range test.want {
					wantBytes[i] = []byte(str)
				}

				t.Errorf(
					"wrong result\ninput: %s\ngot:   %s\nwant:  %s",
					formatBytes([]byte(test.input)),
					formatByteRanges(gotBytes),
					formatByteRanges(wantBytes),
				)
			}
		})
	}
}
