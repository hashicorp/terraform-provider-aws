package complete

import (
	"bytes"
	"os"
	"sort"
	"strings"
	"testing"
)

func TestCompleter_Complete(t *testing.T) {
	t.Parallel()
	initTests()

	c := Command{
		Sub: Commands{
			"sub1": {
				Flags: Flags{
					"-flag1": PredictAnything,
					"-flag2": PredictNothing,
				},
			},
			"sub2": {
				Flags: Flags{
					"-flag2": PredictNothing,
					"-flag3": PredictSet("opt1", "opt2", "opt12"),
				},
				Args: PredictFiles("*.md"),
			},
		},
		Flags: Flags{
			"-o": PredictFiles("*.txt"),
		},
		GlobalFlags: Flags{
			"-h":       PredictNothing,
			"-global1": PredictAnything,
		},
	}
	cmp := New("cmd", c)

	tests := []struct {
		args string
		want []string
	}{
		{
			args: "",
			want: []string{"sub1", "sub2"},
		},
		{
			args: "-",
			want: []string{"-h", "-global1", "-o"},
		},
		{
			args: "-h ",
			want: []string{"sub1", "sub2"},
		},
		{
			args: "-global1 ", // global1 is known follow flag
			want: []string{},
		},
		{
			args: "sub",
			want: []string{"sub1", "sub2"},
		},
		{
			args: "sub1",
			want: []string{"sub1"},
		},
		{
			args: "sub2",
			want: []string{"sub2"},
		},
		{
			args: "sub1 ",
			want: []string{},
		},
		{
			args: "sub1 -",
			want: []string{"-flag1", "-flag2", "-h", "-global1"},
		},
		{
			args: "sub2 ",
			want: []string{"./", "dir/", "outer/", "readme.md"},
		},
		{
			args: "sub2 ./",
			want: []string{"./", "./readme.md", "./dir/", "./outer/"},
		},
		{
			args: "sub2 re",
			want: []string{"readme.md"},
		},
		{
			args: "sub2 ./re",
			want: []string{"./readme.md"},
		},
		{
			args: "sub2 -flag2 ",
			want: []string{"./", "dir/", "outer/", "readme.md"},
		},
		{
			args: "sub1 -fl",
			want: []string{"-flag1", "-flag2"},
		},
		{
			args: "sub1 -flag1",
			want: []string{"-flag1"},
		},
		{
			args: "sub1 -flag1 ",
			want: []string{}, // flag1 is unknown follow flag
		},
		{
			args: "sub1 -flag2 -",
			want: []string{"-flag1", "-flag2", "-h", "-global1"},
		},
		{
			args: "-no-such-flag",
			want: []string{},
		},
		{
			args: "-no-such-flag ",
			want: []string{"sub1", "sub2"},
		},
		{
			args: "-no-such-flag -",
			want: []string{"-h", "-global1", "-o"},
		},
		{
			args: "no-such-command",
			want: []string{},
		},
		{
			args: "no-such-command ",
			want: []string{"sub1", "sub2"},
		},
		{
			args: "-o ",
			want: []string{"a.txt", "b.txt", "c.txt", ".dot.txt", "./", "dir/", "outer/"},
		},
		{
			args: "-o ./no-su",
			want: []string{},
		},
		{
			args: "-o ./",
			want: []string{"./a.txt", "./b.txt", "./c.txt", "./.dot.txt", "./", "./dir/", "./outer/"},
		},
		{
			args: "-o=./",
			want: []string{"./a.txt", "./b.txt", "./c.txt", "./.dot.txt", "./", "./dir/", "./outer/"},
		},
		{
			args: "-o .",
			want: []string{"./a.txt", "./b.txt", "./c.txt", "./.dot.txt", "./", "./dir/", "./outer/"},
		},
		{
			args: "-o ./b",
			want: []string{"./b.txt"},
		},
		{
			args: "-o=./b",
			want: []string{"./b.txt"},
		},
		{
			args: "-o ./read",
			want: []string{},
		},
		{
			args: "-o=./read",
			want: []string{},
		},
		{
			args: "-o ./readme.md",
			want: []string{},
		},
		{
			args: "-o ./readme.md ",
			want: []string{"sub1", "sub2"},
		},
		{
			args: "-o=./readme.md ",
			want: []string{"sub1", "sub2"},
		},
		{
			args: "-o sub2 -flag3 ",
			want: []string{"opt1", "opt2", "opt12"},
		},
		{
			args: "-o sub2 -flag3 opt1",
			want: []string{"opt1", "opt12"},
		},
		{
			args: "-o sub2 -flag3 opt",
			want: []string{"opt1", "opt2", "opt12"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.args, func(t *testing.T) {
			got := runComplete(cmp, tt.args)

			sort.Strings(tt.want)
			sort.Strings(got)

			if !equalSlices(got, tt.want) {
				t.Errorf("failed '%s'\ngot: %s\nwant: %s", t.Name(), got, tt.want)
			}
		})
	}
}

func TestCompleter_Complete_SharedPrefix(t *testing.T) {
	t.Parallel()
	initTests()

	c := Command{
		Sub: Commands{
			"status": {
				Flags: Flags{
					"-f3": PredictNothing,
				},
			},
			"job": {
				Sub: Commands{
					"status": {
						Flags: Flags{
							"-f4": PredictNothing,
						},
					},
				},
			},
		},
		Flags: Flags{
			"-o": PredictFiles("*.txt"),
		},
		GlobalFlags: Flags{
			"-h":       PredictNothing,
			"-global1": PredictAnything,
		},
	}

	cmp := New("cmd", c)

	tests := []struct {
		args string
		want []string
	}{
		{
			args: "",
			want: []string{"status", "job"},
		},
		{
			args: "-",
			want: []string{"-h", "-global1", "-o"},
		},
		{
			args: "j",
			want: []string{"job"},
		},
		{
			args: "job ",
			want: []string{"status"},
		},
		{
			args: "job -",
			want: []string{"-h", "-global1"},
		},
		{
			args: "job status ",
			want: []string{},
		},
		{
			args: "job status -",
			want: []string{"-f4", "-h", "-global1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.args, func(t *testing.T) {
			got := runComplete(cmp, tt.args)

			sort.Strings(tt.want)
			sort.Strings(got)

			if !equalSlices(got, tt.want) {
				t.Errorf("failed '%s'\ngot = %s\nwant: %s", t.Name(), got, tt.want)
			}
		})
	}
}

// runComplete runs the complete login for test purposes
// it gets the complete struct and command line arguments and returns
// the complete options
func runComplete(c *Complete, args string) (completions []string) {
	os.Setenv(envComplete, "cmd "+args)
	b := bytes.NewBuffer(nil)
	c.Out = b
	c.Complete()
	completions = parseOutput(b.String())
	return
}

func parseOutput(output string) []string {
	lines := strings.Split(output, "\n")
	options := []string{}
	for _, l := range lines {
		if l != "" {
			options = append(options, l)
		}
	}
	return options
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
