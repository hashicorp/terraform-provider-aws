package actionlint

import (
	"strings"
)

// UntrustedInputMap is a recursive map to match context object property dereferences.
// Root of this map represents each context names and their ancestors represent recursive properties.
type UntrustedInputMap struct {
	Name     string
	Parent   *UntrustedInputMap
	Children map[string]*UntrustedInputMap
}

func (m *UntrustedInputMap) String() string {
	var b strings.Builder
	m.buildPath(&b)
	return b.String()
}

// Find child object property in this map
func (m *UntrustedInputMap) findObjectProp(name string) (*UntrustedInputMap, bool) {
	if m != nil && m.Children != nil {
		if c, ok := m.Children[name]; ok {
			return c, true
		}
	}
	return nil, false
}

// Find child array element in this map. This is special case with object filter where its receiver is an array
func (m *UntrustedInputMap) findArrayElem() (*UntrustedInputMap, bool) {
	return m.findObjectProp("*")
}

// Build path like `github.event.commits.*.body` by following parents
func (m *UntrustedInputMap) buildPath(b *strings.Builder) {
	if m.Parent != nil && m.Parent.Name != "" {
		m.Parent.buildPath(b)
		b.WriteRune('.')
	}
	b.WriteString(m.Name)
}

// NewUntrustedInputMap creates new instance of UntrustedInputMap. It is used for node of search
// tree of untrusted input checker.
func NewUntrustedInputMap(name string, children ...*UntrustedInputMap) *UntrustedInputMap {
	m := &UntrustedInputMap{
		Name:     name,
		Parent:   nil,
		Children: nil, // Leaf of the tree is nil
	}
	if len(children) > 0 {
		m.Children = make(map[string]*UntrustedInputMap, len(children))
		for _, c := range children {
			c.Parent = m
			m.Children[c.Name] = c
		}
	}
	return m
}

// UntrustedInputSearchRoots is a list of untrusted inputs. It forms tree structure to detect
// untrusted inputs in nested object property access, array index access, and object filters
// efficiently. Each value of this map represents a root of the search so their names are context
// names.
type UntrustedInputSearchRoots map[string]*UntrustedInputMap

// AddRoot adds a new root to search for detecting untrusted input.
func (ms UntrustedInputSearchRoots) AddRoot(m *UntrustedInputMap) {
	ms[m.Name] = m
}

// TODO: Automatically generate BuiltinUntrustedInputs from https://github.com/github/codeql/blob/main/javascript/ql/src/experimental/Security/CWE-094/ExpressionInjection.ql

// BuiltinUntrustedInputs is list of untrusted inputs. These inputs are detected as untrusted in
// `run:` scripts. See the URL for more details.
// - https://securitylab.github.com/research/github-actions-untrusted-input/
// - https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions
// - https://github.com/github/codeql/blob/main/javascript/ql/src/experimental/Security/CWE-094/ExpressionInjection.ql
var BuiltinUntrustedInputs = UntrustedInputSearchRoots{
	"github": NewUntrustedInputMap("github",
		NewUntrustedInputMap("event",
			NewUntrustedInputMap("issue",
				NewUntrustedInputMap("title"),
				NewUntrustedInputMap("body"),
			),
			NewUntrustedInputMap("pull_request",
				NewUntrustedInputMap("title"),
				NewUntrustedInputMap("body"),
				NewUntrustedInputMap("head",
					NewUntrustedInputMap("ref"),
					NewUntrustedInputMap("label"),
					NewUntrustedInputMap("repo",
						NewUntrustedInputMap("default_branch"),
					),
				),
			),
			NewUntrustedInputMap("comment",
				NewUntrustedInputMap("body"),
			),
			NewUntrustedInputMap("review",
				NewUntrustedInputMap("body"),
			),
			NewUntrustedInputMap("review_comment",
				NewUntrustedInputMap("body"),
			),
			NewUntrustedInputMap("pages",
				NewUntrustedInputMap("*",
					NewUntrustedInputMap("page_name"),
				),
			),
			NewUntrustedInputMap("commits",
				NewUntrustedInputMap("*",
					NewUntrustedInputMap("message"),
					NewUntrustedInputMap("author",
						NewUntrustedInputMap("email"),
						NewUntrustedInputMap("name"),
					),
				),
			),
			NewUntrustedInputMap("head_commit",
				NewUntrustedInputMap("message"),
				NewUntrustedInputMap("author",
					NewUntrustedInputMap("email"),
					NewUntrustedInputMap("name"),
				),
			),
			NewUntrustedInputMap("discussion",
				NewUntrustedInputMap("title"),
				NewUntrustedInputMap("body"),
			),
		),
		NewUntrustedInputMap("head_ref"),
	),
}

// UntrustedInputChecker is a checker to detect untrusted inputs in an expression syntax tree.
// This checker checks object property accesses, array index accesses, and object filters. And
// detects paths to untrusted inputs. Found errors are stored in this instance and can be get via
// Errs method.
//
// Note: To avoid breaking the state of checking property accesses on nested property accesses like
// foo[aaa.bbb].bar, IndexAccessNode.Index must be visited before IndexAccessNode.Operand.
type UntrustedInputChecker struct {
	roots           UntrustedInputSearchRoots
	filteringObject bool
	cur             []*UntrustedInputMap
	start           ExprNode
	errs            []*ExprError
}

// NewUntrustedInputChecker creates a new UntrustedInputChecker instance. The roots argument is a
// search tree which defines untrusted input paths as trees.
func NewUntrustedInputChecker(roots UntrustedInputSearchRoots) *UntrustedInputChecker {
	return &UntrustedInputChecker{
		roots:           roots,
		filteringObject: false,
		cur:             nil,
		start:           nil,
		errs:            []*ExprError{},
	}
}

// Reset the state for next search
func (u *UntrustedInputChecker) reset() {
	u.start = nil
	u.filteringObject = false
	u.cur = u.cur[:0]
}

func (u *UntrustedInputChecker) compact() {
	delta := 0
	for i, c := range u.cur {
		if c == nil {
			delta++
			continue
		}
		if delta > 0 {
			u.cur[i-delta] = c
		}
	}
	u.cur = u.cur[:len(u.cur)-delta]
}

func (u *UntrustedInputChecker) onVar(v *VariableNode) {
	c, ok := u.roots[v.Name] // Find root context (currently only "github" exists)
	if !ok {
		return
	}
	u.start = v
	u.cur = append(u.cur, c)
}

func (u *UntrustedInputChecker) onPropAccess(name string) {
	compact := false
	for i, cur := range u.cur {
		c, ok := cur.findObjectProp(name)
		if !ok {
			u.cur[i] = nil
			compact = true
			continue
		}
		u.cur[i] = c // depth + 1
	}
	if compact {
		u.compact()
	}
}

func (u *UntrustedInputChecker) onIndexAccess() {
	if u.filteringObject {
		u.filteringObject = false
		return // For example, match `github.event.*.body[0]` as `github.event.commits[0].body`
	}

	compact := false
	for i, cur := range u.cur {
		if c, ok := cur.findArrayElem(); ok {
			u.cur[i] = c
			continue
		}
		u.cur[i] = nil
		compact = true
	}
	if compact {
		u.compact()
	}
}

func (u *UntrustedInputChecker) onObjectFilter() {
	u.filteringObject = true

	compact := false
	for i, cur := range u.cur {
		// Object filter for arrays
		if c, ok := cur.findArrayElem(); ok {
			u.cur[i] = c
			continue
		}

		if len(cur.Children) == 0 {
			u.cur[i] = nil
			compact = true
		}

		// Object filter for objects
		first := true
		for _, c := range cur.Children {
			if first {
				u.cur[i] = c
				first = false
			} else {
				u.cur = append(u.cur, c)
			}
		}
	}
	if compact {
		u.compact()
	}
}

func (u *UntrustedInputChecker) end() {
	var inputs []string
	for _, cur := range u.cur {
		if cur.Children != nil {
			continue // When `Children` is nil, the node is a leaf
		}
		var b strings.Builder
		cur.buildPath(&b)
		inputs = append(inputs, b.String())
	}

	if len(inputs) == 1 {
		err := errorfAtExpr(
			u.start,
			"%q is potentially untrusted. avoid using it directly in inline scripts. instead, pass it through an environment variable. see https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions for more details",
			inputs[0],
		)
		u.errs = append(u.errs, err)
	} else if len(inputs) > 1 {
		// When multiple untrusted inputs are detected, it means the expression extracts multiple properties with object
		// filter syntax. Show all properties in error message.
		err := errorfAtExpr(
			u.start,
			"object filter extracts potentially untrusted properties %s. avoid using the value directly in inline scripts. instead, pass the value through an environment variable. see https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions for more details",
			sortedQuotes(inputs),
		)
		u.errs = append(u.errs, err)
	}

	u.reset()
}

// OnVisitNodeLeave is a callback which should be called on visiting node after visiting its children.
func (u *UntrustedInputChecker) OnVisitNodeLeave(n ExprNode) {
	switch n := n.(type) {
	case *VariableNode:
		u.end()
		u.onVar(n)
	case *ObjectDerefNode:
		u.onPropAccess(n.Property)
	case *IndexAccessNode:
		if lit, ok := n.Index.(*StringNode); ok {
			// Special case like github['event']['issue']['title']
			u.onPropAccess(lit.Value)
			break
		}
		u.onIndexAccess()
	case *ArrayDerefNode:
		u.onObjectFilter()
	default:
		u.end()
	}
}

// OnVisitEnd is a callback which should be called after visiting whole syntax tree. This callback
// is necessary to handle the case where an untrusted input access is at root of expression.
func (u *UntrustedInputChecker) OnVisitEnd() {
	u.end()
}

// Errs returns errors detected by this checker. This method should be called after visiting all
// nodes in a syntax tree.
func (u *UntrustedInputChecker) Errs() []*ExprError {
	return u.errs
}

// Init initializes a state of checker.
func (u *UntrustedInputChecker) Init() {
	u.errs = u.errs[:0]
	u.reset()
}
