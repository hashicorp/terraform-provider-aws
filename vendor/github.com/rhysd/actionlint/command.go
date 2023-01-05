package actionlint

import (
	"flag"
	"fmt"
	"io"
	"runtime"
	"runtime/debug"
)

// These variables might be modified by ldflags on building release binaries by GoReleaser. Do not modify manually
var (
	version       = ""
	installedFrom = "installed by building from source"
)

const (
	// ExitStatusSuccessNoProblem is the exit status when the command ran successfully with no problem found.
	ExitStatusSuccessNoProblem = 0
	// ExitStatusSuccessProblemFound is the exit status when the command ran successfully with some problem found.
	ExitStatusSuccessProblemFound = 1
	// ExitStatusInvalidCommandOption is the exit status when parsing command line options failed.
	ExitStatusInvalidCommandOption = 2
	// ExitStatusFailure is the exit status when the command stopped due to some fatal error while checking workflows.
	ExitStatusFailure = 3
)

const commandUsageHeader = `Usage: actionlint [FLAGS] [FILES...] [-]

  actionlint is a linter for GitHub Actions workflow files.

  To check all YAML files in current repository, just run actionlint without
  arguments. It automatically finds the nearest '.github/workflows' directory:

    $ actionlint

  To check specific files, pass the file paths as arguments:

    $ actionlint file1.yaml file2.yaml

  To check content which is not saved in file yet (e.g. output from some
  command), pass - argument. It reads stdin and checks it as workflow file:

    $ actionlint -

  To serialize errors into JSON, use -format option. It allows to format error
  messages flexibly with Go template syntax.

    $ actionlint -format '{{json .}}'

Documents:

  https://github.com/rhysd/actionlint/tree/main/docs

Flags:`

func getCommandVersion() string {
	if version != "" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown" // Reaches only when actionlint package is built outside module
	}

	return info.Main.Version
}

// Command represents entire actionlint command. Given stdin/stdout/stderr are used for input/output.
type Command struct {
	// Stdin is a reader to read input from stdin
	Stdin io.Reader
	// Stdout is a writer to write output to stdout
	Stdout io.Writer
	// Stderr is a writer to write output to stderr
	Stderr io.Writer
}

func (cmd *Command) runLinter(args []string, opts *LinterOptions, initConfig bool) ([]*Error, error) {
	l, err := NewLinter(cmd.Stdout, opts)
	if err != nil {
		return nil, err
	}

	if initConfig {
		return nil, l.GenerateDefaultConfig(".")
	}

	if len(args) == 0 {
		return l.LintRepository(".")
	}

	if len(args) == 1 && args[0] == "-" {
		b, err := io.ReadAll(cmd.Stdin)
		if err != nil {
			return nil, fmt.Errorf("could not read stdin: %w", err)
		}
		n := "<stdin>"
		if opts.StdinFileName != "" {
			n = opts.StdinFileName
		}
		return l.Lint(n, b, nil)
	}

	return l.LintFiles(args, nil)
}

type ignorePatternFlags []string

func (i *ignorePatternFlags) String() string {
	return "option for ignore patterns"
}
func (i *ignorePatternFlags) Set(v string) error {
	*i = append(*i, v)
	return nil
}

// Main is main function of actionlint. It takes command line arguments as string slice and returns
// exit status. The args should be entire arguments including the program name, usually given via
// os.Args.
func (cmd *Command) Main(args []string) int {
	var ver bool
	var opts LinterOptions
	var ignorePats ignorePatternFlags
	var initConfig bool
	var noColor bool
	var color bool

	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(cmd.Stderr)
	flags.Var(&ignorePats, "ignore", "Regular expression matching to error messages you want to ignore. This flag is repeatable")
	flags.StringVar(&opts.Shellcheck, "shellcheck", "shellcheck", "Command name or file path of \"shellcheck\" external command. If empty, shellcheck integration will be disabled")
	flags.StringVar(&opts.Pyflakes, "pyflakes", "pyflakes", "Command name or file path of \"pyflakes\" external command. If empty, pyflakes integration will be disabled")
	flags.BoolVar(&opts.Oneline, "oneline", false, "Use one line per one error. Useful for reading error messages from programs")
	flags.StringVar(&opts.Format, "format", "", "Custom template to format error messages in Go template syntax. See https://github.com/rhysd/actionlint/tree/main/docs/usage.md#format")
	flags.StringVar(&opts.ConfigFile, "config-file", "", "File path to config file")
	flags.BoolVar(&initConfig, "init-config", false, "Generate default config file at .github/actionlint.yaml in current project")
	flags.BoolVar(&noColor, "no-color", false, "Disable colorful output")
	flags.BoolVar(&color, "color", false, "Always enable colorful output. This is useful to force colorful outputs")
	flags.BoolVar(&opts.Verbose, "verbose", false, "Enable verbose output")
	flags.BoolVar(&opts.Debug, "debug", false, "Enable debug output (for development)")
	flags.BoolVar(&ver, "version", false, "Show version and how this binary was installed")
	flags.StringVar(&opts.StdinFileName, "stdin-filename", "", "File name when reading input from stdin")
	flags.Usage = func() {
		fmt.Fprintln(cmd.Stderr, commandUsageHeader)
		flags.PrintDefaults()
	}
	if err := flags.Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			// When -h or -help
			return ExitStatusSuccessNoProblem
		}
		return ExitStatusInvalidCommandOption
	}

	if ver {
		fmt.Fprintf(
			cmd.Stdout,
			"%s\n%s\nbuilt with %s compiler for %s/%s\n",
			getCommandVersion(),
			installedFrom,
			runtime.Version(),
			runtime.GOOS,
			runtime.GOARCH,
		)
		return ExitStatusSuccessNoProblem
	}

	opts.IgnorePatterns = ignorePats
	opts.LogWriter = cmd.Stderr

	if color {
		opts.Color = ColorOptionKindAlways
	}
	if noColor {
		opts.Color = ColorOptionKindNever
	}

	errs, err := cmd.runLinter(flags.Args(), &opts, initConfig)
	if err != nil {
		fmt.Fprintln(cmd.Stderr, err.Error())
		return ExitStatusFailure
	}
	if len(errs) > 0 {
		return ExitStatusSuccessProblemFound // Linter found some issues, yay!
	}

	return ExitStatusSuccessNoProblem
}
