package textseg

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestScanGraphemeClusters(t *testing.T) {
	tests := unicodeGraphemeTests

	for i, test := range tests {
		t.Run(fmt.Sprintf("%03d-%x", i, test.input), func(t *testing.T) {
			got, err := AllTokens(test.input, ScanGraphemeClusters)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !reflect.DeepEqual(got, test.output) {
				// Also get the rune values resulting from decoding utf8,
				// since they are generally easier to look up to figure out
				// what's failing.
				runes := make([]string, 0, len(test.input))
				seqs := make([][]byte, 0, len(test.input))
				categories := make([]string, 0, len(test.input))
				buf := test.input
				for len(buf) > 0 {
					r, size := utf8.DecodeRune(buf)
					runes = append(runes, fmt.Sprintf("0x%04x", r))
					seqs = append(seqs, buf[:size])
					categories = append(categories, _GraphemeRuneType(r).String())
					buf = buf[size:]
				}

				t.Errorf(
					"wrong result\ninput: %s\nutf8s: %s\nrunes: %s\ncats:  %s\ngot:   %s\nwant:  %s",
					formatBytes(test.input),
					formatByteRanges(seqs),
					strings.Join(runes, " "),
					strings.Join(categories, " "),
					formatByteRanges(got),
					formatByteRanges(test.output),
				)
			}
		})
	}
}

func formatBytes(buf []byte) string {
	strs := make([]string, len(buf))
	for i, b := range buf {
		strs[i] = fmt.Sprintf("0x%02x", b)
	}
	return strings.Join(strs, "   ")
}

func formatByteRanges(bufs [][]byte) string {
	strs := make([]string, len(bufs))
	for i, b := range bufs {
		strs[i] = formatBytes(b)
	}
	return strings.Join(strs, " | ")
}
