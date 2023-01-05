package actionlint

import (
	"regexp"
	"strings"
)

var jobIDPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)

// RuleID is a rule to check step IDs in workflow.
type RuleID struct {
	RuleBase
	seen map[string]*Pos
}

// NewRuleID creates a new RuleID instance.
func NewRuleID() *RuleID {
	return &RuleID{
		RuleBase: RuleBase{name: "id"},
	}
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleID) VisitJobPre(n *Job) error {
	rule.seen = map[string]*Pos{}

	rule.validateConvention(n.ID, "job")
	for _, j := range n.Needs {
		rule.validateConvention(j, "job")
	}

	return nil
}

// VisitJobPost is callback when visiting Job node after visiting its children.
func (rule *RuleID) VisitJobPost(n *Job) error {
	rule.seen = nil
	return nil
}

// VisitStep is callback when visiting Step node.
func (rule *RuleID) VisitStep(n *Step) error {
	if n.ID == nil {
		return nil
	}

	rule.validateConvention(n.ID, "step")

	id := strings.ToLower(n.ID.Value)
	if prev, ok := rule.seen[id]; ok {
		rule.errorf(n.ID.Pos, "step ID %q duplicates. previously defined at %s. step ID must be unique within a job. note that step ID is case insensitive", n.ID.Value, prev.String())
		return nil
	}
	rule.seen[id] = n.ID.Pos
	return nil
}

func (rule *RuleID) validateConvention(id *String, what string) {
	if id == nil || id.Value == "" || containsPlaceholder(id.Value) || jobIDPattern.MatchString(id.Value) {
		return
	}
	rule.errorf(id.Pos, "invalid %s ID %q. %s ID must start with a letter or _ and contain only alphanumeric characters, -, or _", what, id.Value, what)
}

func containsPlaceholder(s string) bool {
	i := strings.Index(s, "${{")
	j := strings.Index(s, "}}")
	return i >= 0 && j >= 0 && i < j
}
