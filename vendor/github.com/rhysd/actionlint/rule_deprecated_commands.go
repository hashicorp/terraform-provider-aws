package actionlint

import "regexp"

var deprecatedCommandsPattern = regexp.MustCompile(`(?:::(save-state|set-output|set-env)\s+name=[a-zA-Z][a-zA-Z_-]*::\S+|::(add-path)::\S+)`)

// RuleDeprecatedCommands is a rule checker to detect deprecated workflow commands. Currently
// 'set-state', 'set-output', `set-env' and 'add-path' are detected as deprecated.
//
// - https://github.blog/changelog/2020-10-01-github-actions-deprecating-set-env-and-add-path-commands/
// - https://github.blog/changelog/2022-10-11-github-actions-deprecating-save-state-and-set-output-commands/
type RuleDeprecatedCommands struct {
	RuleBase
}

// NewRuleDeprecatedCommands creates a new RuleDeprecatedCommands instance.
func NewRuleDeprecatedCommands() *RuleDeprecatedCommands {
	return &RuleDeprecatedCommands{
		RuleBase: RuleBase{name: "deprecated-commands"},
	}
}

// VisitStep is callback when visiting Step node.
func (rule *RuleDeprecatedCommands) VisitStep(n *Step) error {
	if r, ok := n.Exec.(*ExecRun); ok && r.Run != nil {
		for _, m := range deprecatedCommandsPattern.FindAllStringSubmatch(r.Run.Value, -1) {
			c := m[1]
			if len(c) == 0 {
				c = m[2]
			}

			var a string
			switch c {
			case "set-output":
				a = `echo "{name}={value}" >> $GITHUB_OUTPUT`
			case "save-state":
				a = `echo "{name}={value}" >> $GITHUB_STATE`
			case "set-env":
				a = `echo "{name}={value}" >> $GITHUB_ENV`
			case "add-path":
				a = `echo "{path}" >> $GITHUB_PATH`
			default:
				panic("unreachable")
			}

			rule.errorf(
				r.Run.Pos,
				"workflow command %q was deprecated. use `%s` instead: https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions",
				c,
				a,
			)
		}
	}
	return nil
}
