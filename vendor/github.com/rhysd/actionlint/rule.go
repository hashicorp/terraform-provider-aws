package actionlint

import (
	"fmt"
	"io"
)

// RuleBase is a struct to be a base of rule structs. Embed this struct to define default methods
// automatically
type RuleBase struct {
	name string
	errs []*Error
	dbg  io.Writer
}

// VisitStep is callback when visiting Step node.
func (r *RuleBase) VisitStep(node *Step) error { return nil }

// VisitJobPre is callback when visiting Job node before visiting its children.
func (r *RuleBase) VisitJobPre(node *Job) error { return nil }

// VisitJobPost is callback when visiting Job node after visiting its children.
func (r *RuleBase) VisitJobPost(node *Job) error { return nil }

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (r *RuleBase) VisitWorkflowPre(node *Workflow) error { return nil }

// VisitWorkflowPost is callback when visiting Workflow node after visiting its children.
func (r *RuleBase) VisitWorkflowPost(node *Workflow) error { return nil }

func (r *RuleBase) error(pos *Pos, msg string) {
	err := errorAt(pos, r.name, msg)
	r.errs = append(r.errs, err)
}

func (r *RuleBase) errorf(pos *Pos, format string, args ...interface{}) {
	err := errorfAt(pos, r.name, format, args...)
	r.errs = append(r.errs, err)
}

func (r *RuleBase) debug(format string, args ...interface{}) {
	if r.dbg == nil {
		return
	}
	format = fmt.Sprintf("[%s] %s\n", r.name, format)
	fmt.Fprintf(r.dbg, format, args...)
}

// Errs returns errors found by the rule.
func (r *RuleBase) Errs() []*Error {
	return r.errs
}

// Name is name of the rule.
func (r *RuleBase) Name() string {
	return r.name
}

// EnableDebug enables debug output from the rule. Given io.Writer instance is used to print debug
// information to console
func (r *RuleBase) EnableDebug(out io.Writer) {
	r.dbg = out
}

// Rule is an interface which all rule structs must meet
type Rule interface {
	Pass
	Errs() []*Error
	Name() string
	EnableDebug(out io.Writer)
}
