package actionlint

// RuleGlob is a rule to check glob syntax.
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet
type RuleGlob struct {
	RuleBase
}

// NewRuleGlob creates new RuleGlob instance.
func NewRuleGlob() *RuleGlob {
	return &RuleGlob{
		RuleBase: RuleBase{name: "glob"},
	}
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RuleGlob) VisitWorkflowPre(n *Workflow) error {
	for _, e := range n.On {
		if w, ok := e.(*WebhookEvent); ok {
			rule.checkGitRefGlobs(w.Branches)
			rule.checkGitRefGlobs(w.BranchesIgnore)
			rule.checkGitRefGlobs(w.Tags)
			rule.checkGitRefGlobs(w.TagsIgnore)
			rule.checkFilePathGlobs(w.Paths)
			rule.checkFilePathGlobs(w.PathsIgnore)
		}
	}
	return nil
}

func (rule *RuleGlob) checkGitRefGlobs(filter *WebhookEventFilter) {
	if filter == nil {
		return
	}
	for _, v := range filter.Values {
		// Empty value is already checked by parser. Avoid duplicate errors
		if v.Value != "" {
			rule.globErrors(ValidateRefGlob(v.Value), v.Pos, v.Quoted)
		}
	}
}

func (rule *RuleGlob) checkFilePathGlobs(filter *WebhookEventFilter) {
	if filter == nil {
		return
	}
	for _, v := range filter.Values {
		// Empty value is already checked by parser. Avoid duplicate errors
		if v.Value != "" {
			rule.globErrors(ValidatePathGlob(v.Value), v.Pos, v.Quoted)
		}
	}
}

func (rule *RuleGlob) globErrors(errs []InvalidGlobPattern, pos *Pos, quoted bool) {
	for i := range errs {
		err := &errs[i]
		p := *pos
		if quoted {
			p.Col++
		}
		if err.Column != 0 {
			p.Col += err.Column - 1
		}
		rule.errorf(&p, "%s. note: filter pattern syntax is explained at https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet", err.Message)
	}
}
