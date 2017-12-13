package match

import (
	"fmt"
	"os"
	"testing"
)

func TestMatch(t *testing.T) {
	t.Parallel()

	// Change to tests directory for testing completion of
	// files and directories
	err := os.Chdir("../tests")
	if err != nil {
		panic(err)
	}

	type matcherTest struct {
		prefix string
		want   bool
	}

	tests := []struct {
		m     Match
		long  string
		tests []matcherTest
	}{
		{
			m:    Prefix,
			long: "abcd",
			tests: []matcherTest{
				{prefix: "", want: true},
				{prefix: "ab", want: true},
				{prefix: "ac", want: false},
			},
		},
		{
			m:    Prefix,
			long: "",
			tests: []matcherTest{
				{prefix: "ac", want: false},
				{prefix: "", want: true},
			},
		},
		{
			m:    File,
			long: "file.txt",
			tests: []matcherTest{
				{prefix: "", want: true},
				{prefix: "f", want: true},
				{prefix: "./f", want: true},
				{prefix: "./.", want: false},
				{prefix: "file.", want: true},
				{prefix: "./file.", want: true},
				{prefix: "file.txt", want: true},
				{prefix: "./file.txt", want: true},
				{prefix: "other.txt", want: false},
				{prefix: "/other.txt", want: false},
				{prefix: "/file.txt", want: false},
				{prefix: "/fil", want: false},
				{prefix: "/file.txt2", want: false},
				{prefix: "/.", want: false},
			},
		},
		{
			m:    File,
			long: "./file.txt",
			tests: []matcherTest{
				{prefix: "", want: true},
				{prefix: "f", want: true},
				{prefix: "./f", want: true},
				{prefix: "./.", want: false},
				{prefix: "file.", want: true},
				{prefix: "./file.", want: true},
				{prefix: "file.txt", want: true},
				{prefix: "./file.txt", want: true},
				{prefix: "other.txt", want: false},
				{prefix: "/other.txt", want: false},
				{prefix: "/file.txt", want: false},
				{prefix: "/fil", want: false},
				{prefix: "/file.txt2", want: false},
				{prefix: "/.", want: false},
			},
		},
		{
			m:    File,
			long: "/file.txt",
			tests: []matcherTest{
				{prefix: "", want: true},
				{prefix: "f", want: false},
				{prefix: "./f", want: false},
				{prefix: "./.", want: false},
				{prefix: "file.", want: false},
				{prefix: "./file.", want: false},
				{prefix: "file.txt", want: false},
				{prefix: "./file.txt", want: false},
				{prefix: "other.txt", want: false},
				{prefix: "/other.txt", want: false},
				{prefix: "/file.txt", want: true},
				{prefix: "/fil", want: true},
				{prefix: "/file.txt2", want: false},
				{prefix: "/.", want: false},
			},
		},
		{
			m:    File,
			long: "./",
			tests: []matcherTest{
				{prefix: "", want: true},
				{prefix: ".", want: true},
				{prefix: "./", want: true},
				{prefix: "./.", want: false},
			},
		},
	}

	for _, tt := range tests {
		for _, ttt := range tt.tests {
			name := fmt.Sprintf("matcher=%T&long='%s'&prefix='%s'", tt.m, tt.long, ttt.prefix)
			t.Run(name, func(t *testing.T) {
				got := tt.m(tt.long, ttt.prefix)
				if got != ttt.want {
					t.Errorf("Failed %s: got = %t, want: %t", name, got, ttt.want)
				}
			})
		}
	}
}
