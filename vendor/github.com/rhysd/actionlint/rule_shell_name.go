package actionlint

import (
	"strings"
)

type platformKind int

const (
	platformKindAny platformKind = iota
	platformKindMacOrLinux
	platformKindWindows
)

// RuleShellName is a rule to check 'shell' field. For more details, see
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#using-a-specific-shell
type RuleShellName struct {
	RuleBase
	platform platformKind
}

// NewRuleShellName creates new RuleShellName instance.
func NewRuleShellName() *RuleShellName {
	return &RuleShellName{
		RuleBase: RuleBase{name: "shell-name"},
		platform: platformKindAny,
	}
}

// VisitStep is callback when visiting Step node.
func (rule *RuleShellName) VisitStep(n *Step) error {
	if run, ok := n.Exec.(*ExecRun); ok {
		rule.checkShellName(run.Shell)
	}
	return nil
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleShellName) VisitJobPre(n *Job) error {
	if n.RunsOn == nil {
		return nil
	}
	rule.platform = rule.getPlatformFromRunner(n.RunsOn)
	if n.Defaults != nil && n.Defaults.Run != nil {
		rule.checkShellName(n.Defaults.Run.Shell)
	}
	return nil
}

// VisitJobPost is callback when visiting Job node after visiting its children.
func (rule *RuleShellName) VisitJobPost(n *Job) error {
	rule.platform = platformKindAny // Clear
	return nil
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RuleShellName) VisitWorkflowPre(n *Workflow) error {
	if n.Defaults != nil && n.Defaults.Run != nil {
		rule.checkShellName(n.Defaults.Run.Shell)
	}
	return nil
}

func (rule *RuleShellName) checkShellName(node *String) {
	if node == nil {
		return
	}

	// Ignore custom shell
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#custom-shell
	if strings.Contains(node.Value, "{0}") {
		return
	}

	// Ignore dynamic shell name
	if strings.Contains(node.Value, "${{") {
		return
	}

	name := strings.ToLower(node.Value)
	available := getAvailableShellNames(rule.platform)

	for _, s := range available {
		if name == s {
			return // ok
		}
	}

	onPlatform := ""
	switch rule.platform {
	case platformKindWindows:
		for _, p := range getAvailableShellNames(platformKindAny) {
			if name == p {
				onPlatform = " on Windows" // only when the shell is unavailable on Windows
			}
		}
	case platformKindMacOrLinux:
		for _, p := range getAvailableShellNames(platformKindAny) {
			if name == p {
				onPlatform = " on macOS or Linux" // only when the shell is unavailable on macOS or Linux
			}
		}
	}

	rule.errorf(
		node.Pos,
		"shell name %q is invalid%s. available names are %s",
		node.Value,
		onPlatform,
		sortedQuotes(available),
	)
}

func getAvailableShellNames(kind platformKind) []string {
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#using-a-specific-shell
	switch kind {
	case platformKindAny:
		return []string{
			"bash",
			"pwsh",
			"python",
			"sh",
			"cmd",
			"powershell",
		}
	case platformKindWindows:
		return []string{
			"bash",
			"pwsh",
			"python",
			"cmd",
			"powershell",
		}
	case platformKindMacOrLinux:
		return []string{
			"bash",
			"pwsh",
			"python",
			"sh",
		}
	default:
		panic("unreachable")
	}
}

func (rule *RuleShellName) getPlatformFromRunner(runner *Runner) platformKind {
	if runner == nil {
		return platformKindAny
	}

	// Note: Labels for self-hosted runners:
	// https://docs.github.com/en/actions/hosting-your-own-runners/using-labels-with-self-hosted-runners

	ret := platformKindAny
	for _, label := range runner.Labels {
		k := platformKindAny
		l := strings.ToLower(label.Value)
		if strings.HasPrefix(l, "windows-") || l == "windows" {
			k = platformKindWindows
		} else if strings.HasPrefix(l, "macos-") || strings.HasPrefix(l, "ubuntu-") || l == "macos" || l == "linux" {
			k = platformKindMacOrLinux
		}

		if k == platformKindAny {
			continue
		}
		if ret != platformKindAny && ret != k {
			// Conflicts are reported by runner-label rule so simply ignore here
			return platformKindAny
		}
		ret = k
	}

	return ret
}
