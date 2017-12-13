package getter

import (
	"path/filepath"
	"testing"
)

func TestTar(t *testing.T) {
	cases := []TestDecompressCase{
		{
			"extended_header.tar",
			true,
			false,
			[]string{"directory/", "directory/a", "directory/b"},
			"",
		},
		{
			"implied_dir.tar",
			true,
			false,
			[]string{"directory/", "directory/sub/", "directory/sub/a", "directory/sub/b"},
			"",
		},
	}

	for i, tc := range cases {
		cases[i].Input = filepath.Join("./test-fixtures", "decompress-tar", tc.Input)
	}

	TestDecompressor(t, new(tarDecompressor), cases)
}
