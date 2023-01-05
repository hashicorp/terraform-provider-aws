package actionlint

var allPermissionScopes = map[string]struct{}{
	"actions":             {},
	"checks":              {},
	"contents":            {},
	"deployments":         {},
	"id-token":            {},
	"issues":              {},
	"discussions":         {},
	"packages":            {},
	"pages":               {},
	"pull-requests":       {},
	"repository-projects": {},
	"security-events":     {},
	"statuses":            {},
}

// RulePermissions is a rule checker to check permission configurations in a workflow.
// https://docs.github.com/en/actions/security-guides/automatic-token-authentication#permissions-for-the-github_token
type RulePermissions struct {
	RuleBase
}

// NewRulePermissions creates new RulePermissions instance.
func NewRulePermissions() *RulePermissions {
	return &RulePermissions{
		RuleBase: RuleBase{name: "permissions"},
	}
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RulePermissions) VisitJobPre(n *Job) error {
	rule.checkPermissions(n.Permissions)
	return nil
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RulePermissions) VisitWorkflowPre(n *Workflow) error {
	rule.checkPermissions(n.Permissions)
	return nil
}

func (rule *RulePermissions) checkPermissions(p *Permissions) {
	if p == nil {
		return
	}

	if p.All != nil {
		switch p.All.Value {
		case "write-all", "read-all":
			// OK
		default:
			rule.errorf(p.All.Pos, "%q is invalid for permission for all the scopes. available values are \"read-all\" and \"write-all\"", p.All.Value)
		}
		return
	}

	for _, p := range p.Scopes {
		n := p.Name.Value // Permission names are case-sensitive
		if _, ok := allPermissionScopes[n]; !ok {
			ss := make([]string, 0, len(allPermissionScopes))
			for s := range allPermissionScopes {
				ss = append(ss, s)
			}
			rule.errorf(p.Name.Pos, "unknown permission scope %q. all available permission scopes are %s", n, sortedQuotes(ss))
		}
		switch p.Value.Value {
		case "read", "write", "none":
			// OK
		default:
			rule.errorf(p.Value.Pos, "%q is invalid for permission of scope %q. available values are \"read\", \"write\" or \"none\"", p.Value.Value, n)
		}
	}
}
