package actionlint

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

type shellcheckError struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Level   string `json:"level"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// RuleShellcheck is a rule to check shell scripts at 'run:' using shellcheck.
// https://github.com/koalaman/shellcheck
type RuleShellcheck struct {
	RuleBase
	cmd           *externalCommand
	workflowShell string
	jobShell      string
	mu            sync.Mutex
}

// NewRuleShellcheck craetes new RuleShellcheck instance. Parameter executable can be command name
// or relative/absolute file path. When the given executable is not found in system, it returns an
// error as 2nd return value.
func NewRuleShellcheck(executable string, proc *concurrentProcess) (*RuleShellcheck, error) {
	cmd, err := proc.newCommandRunner(executable)
	if err != nil {
		return nil, err
	}
	r := &RuleShellcheck{
		RuleBase:      RuleBase{name: "shellcheck"},
		cmd:           cmd,
		workflowShell: "",
		jobShell:      "",
	}
	return r, nil
}

// VisitStep is callback when visiting Step node.
func (rule *RuleShellcheck) VisitStep(n *Step) error {
	run, ok := n.Exec.(*ExecRun)
	if !ok || run.Run == nil {
		return nil
	}

	name := rule.getShellName(run)
	if name != "bash" && name != "sh" {
		return nil
	}

	rule.runShellcheck(run.Run.Value, name, run.RunPos)
	return nil
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleShellcheck) VisitJobPre(n *Job) error {
	if n.Defaults != nil && n.Defaults.Run != nil && n.Defaults.Run.Shell != nil {
		rule.jobShell = n.Defaults.Run.Shell.Value
		return nil
	}

	if n.RunsOn == nil {
		return nil
	}

	for _, label := range n.RunsOn.Labels {
		l := strings.ToLower(label.Value)
		// Default shell on Windows is PowerShell.
		// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#using-a-specific-shell
		if l == "windows" || strings.HasPrefix(l, "windows-") {
			return nil
		}
	}

	// TODO: When bash is not found, GitHub-hosted runner fallbacks to sh. What OSes require this behavior?
	rule.jobShell = "bash"

	return nil
}

// VisitJobPost is callback when visiting Job node after visiting its children.
func (rule *RuleShellcheck) VisitJobPost(n *Job) error {
	rule.jobShell = ""
	return nil
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RuleShellcheck) VisitWorkflowPre(n *Workflow) error {
	if n.Defaults != nil && n.Defaults.Run != nil && n.Defaults.Run.Shell != nil {
		rule.workflowShell = n.Defaults.Run.Shell.Value
	}
	return nil
}

// VisitWorkflowPost is callback when visiting Workflow node after visiting its children.
func (rule *RuleShellcheck) VisitWorkflowPost(n *Workflow) error {
	rule.workflowShell = ""
	return rule.cmd.wait() // Wait until all processes running for this rule
}

func (rule *RuleShellcheck) getShellName(exec *ExecRun) string {
	if exec.Shell != nil {
		return exec.Shell.Value
	}
	if rule.jobShell != "" {
		return rule.jobShell
	}
	return rule.workflowShell
}

// Replace ${{ ... }} with underscores like __________
// Note: replacing with spaces sometimes causes syntax error. For example,
//
//	if ${{ contains(xs, s) }}; then
//	  echo 'hello'
//	fi
func sanitizeExpressionsInScript(src string) string {
	b := strings.Builder{}
	for {
		s := strings.Index(src, "${{")
		if s == -1 {
			if b.Len() == 0 {
				return src
			}
			b.WriteString(src)
			return b.String()
		}

		e := strings.Index(src[s:], "}}")
		if e == -1 {
			if b.Len() == 0 {
				return src
			}
			b.WriteString(src)
			return b.String()
		}
		e += s + 2 // 2 is offset for len("}}")

		// Note: If ${{ ... }} includes newline, line and column reported by shellcheck will be
		// shifted.
		b.WriteString(src[:s])
		for i := 0; i < e-s; i++ {
			b.WriteByte('_')
		}

		src = src[e:]
	}
}

func (rule *RuleShellcheck) runShellcheck(src, sh string, pos *Pos) {
	src = sanitizeExpressionsInScript(src)
	rule.debug("%s: Run shellcheck for %s script:\n%s", pos, sh, src)

	// Reasons to exclude the rules:
	//
	// - SC1091: File not found. Scripts are for CI environment. Not suitable for checking this in current local
	//           environment
	// - SC2194: The word is constant. This sometimes happens at constants by replacing ${{ }} with underscores.
	//           For example, `if ${{ matrix.foo }}; then ...` -> `if _________________; then ...`
	// - SC2050: The expression is constant. This sometimes happens at `if` condition by replacing ${{ }} with
	//           underscores (#45). For example, `if [ "${{ matrix.foo }}" = "x" ]` -> `if [ "_________________" = "x" ]`
	// - SC2154: The var is referenced but not assigned. Script at `run:` can refer variables defined in `env:` section
	//           so this rule can cause false positives (#53).
	// - SC2157: Argument to -z is always false due to literal strings. When the argument of -z is replaced from ${{ }},
	//           this can happen. For example, `if [ -z ${{ env.FOO }} ]` -> `if [ -z ______________ ]` (#113).
	args := []string{"--norc", "-f", "json", "-x", "--shell", sh, "-e", "SC1091,SC2194,SC2050,SC2154,SC2157", "-"}
	rule.debug("%s: Running %s command with %s", pos, rule.cmd.exe, args)

	// Use same options to run shell process described at document
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#using-a-specific-shell
	setup := "set -e"
	if sh == "bash" {
		setup = "set -eo pipefail"
	}
	script := fmt.Sprintf("%s\n%s\n", setup, src)

	rule.cmd.run(args, script, func(stdout []byte, err error) error {
		if err != nil {
			rule.debug("Command %s %s failed: %v", rule.cmd.exe, args, err)
			return fmt.Errorf("`%s %s` did not run successfully while checking script at %s: %w", rule.cmd.exe, strings.Join(args, " "), pos, err)
		}

		errs := []shellcheckError{}
		if err := json.Unmarshal(stdout, &errs); err != nil {
			return fmt.Errorf("could not parse JSON output from shellcheck: %w: stdout=%q", err, stdout)
		}
		if len(errs) == 0 {
			return nil
		}

		rule.mu.Lock()
		defer rule.mu.Unlock()
		// It's better to show source location in the script as position of error, but it's not
		// possible easily. YAML has multiple block styles with '|', '>', '|+', '>+', '|-', '>-'. Some
		// of them remove indentation and/or blank lines. So restoring source position in block string
		// is not possible. Sourcemap is necessary to do it.
		// Instead, actionlint shows position of 'run:' as position of error. And separately show
		// location in script which is reported by shellcheck in error message.
		for _, err := range errs {
			// Consider the first line is setup for running shell which was implicitly added for better check
			line := err.Line - 1
			msg := strings.TrimSuffix(err.Message, ".") // Trim period aligning style of error message
			rule.errorf(pos, "shellcheck reported issue in this script: SC%d:%s:%d:%d: %s", err.Code, err.Level, line, err.Column, msg)
		}

		return nil
	})
}
