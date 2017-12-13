package getter

import (
	"path/filepath"
	"testing"
)

func TestZipDecompressor(t *testing.T) {
	cases := []TestDecompressCase{
		{
			"empty.zip",
			false,
			true,
			nil,
			"",
		},

		{
			"single.zip",
			false,
			false,
			nil,
			"d3b07384d113edec49eaa6238ad5ff00",
		},

		{
			"single.zip",
			true,
			false,
			[]string{"file"},
			"",
		},

		{
			"multiple.zip",
			true,
			false,
			[]string{"file1", "file2"},
			"",
		},

		{
			"multiple.zip",
			false,
			true,
			nil,
			"",
		},

		{
			"subdir.zip",
			true,
			false,
			[]string{"file1", "subdir/", "subdir/child"},
			"",
		},

		{
			"subdir_empty.zip",
			true,
			false,
			[]string{"file1", "subdir/"},
			"",
		},

		{
			"subdir_missing_dir.zip",
			true,
			false,
			[]string{"file1", "subdir/", "subdir/child"},
			"",
		},
	}

	for i, tc := range cases {
		cases[i].Input = filepath.Join("./test-fixtures", "decompress-zip", tc.Input)
	}

	TestDecompressor(t, new(ZipDecompressor), cases)
}
