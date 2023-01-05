package actionlint

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var reFormatPlaceholder = regexp.MustCompile(`{\d+}`)

func ordinal(i int) string {
	suffix := "th"
	switch i % 10 {
	case 1:
		if i%100 != 11 {
			suffix = "st"
		}
	case 2:
		if i%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if i%100 != 13 {
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", i, suffix)
}

// Functions

// FuncSignature is a signature of function, which holds return and arguments types.
type FuncSignature struct {
	// Name is a name of the function.
	Name string
	// Ret is a return type of the function.
	Ret ExprType
	// Params is a list of parameter types of the function. The final element of this list might
	// be repeated as variable length arguments.
	Params []ExprType
	// VariableLengthParams is a flag to handle variable length parameters. When this flag is set to
	// true, it means that the last type of params might be specified multiple times (including zero
	// times). Setting true implies length of Params is more than 0.
	VariableLengthParams bool
}

func (sig *FuncSignature) String() string {
	ts := make([]string, 0, len(sig.Params))
	for _, p := range sig.Params {
		ts = append(ts, p.String())
	}
	elip := ""
	if sig.VariableLengthParams {
		elip = "..."
	}
	return fmt.Sprintf("%s(%s%s) -> %s", sig.Name, strings.Join(ts, ", "), elip, sig.Ret.String())
}

// BuiltinFuncSignatures is a set of all builtin function signatures. All function names are in
// lower case because function names are compared in case insensitive.
// https://docs.github.com/en/actions/learn-github-actions/expressions#functions
var BuiltinFuncSignatures = map[string][]*FuncSignature{
	"contains": {
		{
			Name: "contains",
			Ret:  BoolType{},
			Params: []ExprType{
				StringType{},
				StringType{},
			},
		},
		{
			Name: "contains",
			Ret:  BoolType{},
			Params: []ExprType{
				&ArrayType{Elem: AnyType{}},
				AnyType{},
			},
		},
	},
	"startswith": {
		{
			Name: "startsWith",
			Ret:  BoolType{},
			Params: []ExprType{
				StringType{},
				StringType{},
			},
		},
	},
	"endswith": {
		{
			Name: "endsWith",
			Ret:  BoolType{},
			Params: []ExprType{
				StringType{},
				StringType{},
			},
		},
	},
	"format": {
		{
			Name: "format",
			Ret:  StringType{},
			Params: []ExprType{
				StringType{},
				AnyType{}, // variable length
			},
			VariableLengthParams: true,
		},
	},
	"join": {
		{
			Name: "join",
			Ret:  StringType{},
			Params: []ExprType{
				&ArrayType{Elem: StringType{}},
				StringType{},
			},
		},
		{
			Name: "join",
			Ret:  StringType{},
			Params: []ExprType{
				StringType{},
				StringType{},
			},
		},
		// When the second parameter is omitted, values are concatenated with ','.
		{
			Name: "join",
			Ret:  StringType{},
			Params: []ExprType{
				&ArrayType{Elem: StringType{}},
			},
		},
		{
			Name: "join",
			Ret:  StringType{},
			Params: []ExprType{
				StringType{},
			},
		},
	},
	"tojson": {{
		Name: "toJSON",
		Ret:  StringType{},
		Params: []ExprType{
			AnyType{},
		},
	}},
	"fromjson": {{
		Name: "fromJSON",
		Ret:  AnyType{},
		Params: []ExprType{
			StringType{},
		},
	}},
	"hashfiles": {{
		Name: "hashFiles",
		Ret:  StringType{},
		Params: []ExprType{
			StringType{},
		},
		VariableLengthParams: true,
	}},
	"success": {{
		Name:   "success",
		Ret:    BoolType{},
		Params: []ExprType{},
	}},
	"always": {{
		Name:   "always",
		Ret:    BoolType{},
		Params: []ExprType{},
	}},
	"cancelled": {{
		Name:   "cancelled",
		Ret:    BoolType{},
		Params: []ExprType{},
	}},
	"failure": {{
		Name:   "failure",
		Ret:    BoolType{},
		Params: []ExprType{},
	}},
}

// Global variables

// BuiltinGlobalVariableTypes defines types of all global variables. All context variables are
// documented at https://docs.github.com/en/actions/learn-github-actions/contexts
var BuiltinGlobalVariableTypes = map[string]ExprType{
	// https://docs.github.com/en/actions/learn-github-actions/contexts#github-context
	"github": NewStrictObjectType(map[string]ExprType{
		"action":            StringType{},
		"action_path":       StringType{},
		"action_ref":        StringType{},
		"action_repository": StringType{},
		"action_status":     StringType{},
		"actor":             StringType{},
		"api_url":           StringType{},
		"base_ref":          StringType{},
		"env":               StringType{},
		"event":             NewEmptyObjectType(), // Note: Stricter type check for this payload would be possible
		"event_name":        StringType{},
		"event_path":        StringType{},
		"graphql_url":       StringType{},
		"head_ref":          StringType{},
		"job":               StringType{},
		"ref":               StringType{},
		"ref_name":          StringType{},
		"ref_protected":     StringType{},
		"ref_type":          StringType{},
		"path":              StringType{},
		"repository":        StringType{},
		"repository_owner":  StringType{},
		"repositoryurl":     StringType{}, // repositoryUrl
		"retention_days":    NumberType{},
		"run_id":            StringType{},
		"run_number":        StringType{},
		"run_attempt":       StringType{},
		"server_url":        StringType{},
		"sha":               StringType{},
		"token":             StringType{},
		"triggering_actor":  StringType{},
		"workflow":          StringType{},
		"workspace":         StringType{},
	}),
	// https://docs.github.com/en/actions/learn-github-actions/contexts#env-context
	"env": NewMapObjectType(StringType{}), // env.<env_name>
	// https://docs.github.com/en/actions/learn-github-actions/contexts#job-context
	"job": NewStrictObjectType(map[string]ExprType{
		"container": NewStrictObjectType(map[string]ExprType{
			"id":      StringType{},
			"network": StringType{},
		}),
		"services": NewMapObjectType(
			NewStrictObjectType(map[string]ExprType{
				"id":      StringType{}, // job.services.<service id>.id
				"network": StringType{},
				"ports":   NewMapObjectType(StringType{}),
			}),
		),
		"status": StringType{},
	}),
	// https://docs.github.com/en/actions/learn-github-actions/contexts#steps-context
	"steps": NewEmptyStrictObjectType(), // This value will be updated contextually
	// https://docs.github.com/en/actions/learn-github-actions/contexts#runner-context
	"runner": NewStrictObjectType(map[string]ExprType{
		"name":       StringType{},
		"os":         StringType{},
		"arch":       StringType{},
		"temp":       StringType{},
		"tool_cache": StringType{},
		"debug":      StringType{},
		"workspace":  StringType{},
	}),
	// https://docs.github.com/en/actions/learn-github-actions/contexts#secrets-context
	"secrets": NewMapObjectType(StringType{}),
	// https://docs.github.com/en/actions/learn-github-actions/contexts#strategy-context
	"strategy": NewObjectType(map[string]ExprType{
		"fail-fast":    BoolType{},
		"job-index":    NumberType{},
		"job-total":    NumberType{},
		"max-parallel": NumberType{},
	}),
	// https://docs.github.com/en/actions/learn-github-actions/contexts#matrix-context
	"matrix": NewEmptyStrictObjectType(), // This value will be updated contextually
	// https://docs.github.com/en/actions/learn-github-actions/contexts#needs-context
	"needs": NewEmptyStrictObjectType(), // This value will be updated contextually
	// https://docs.github.com/en/actions/learn-github-actions/contexts#inputs-context
	// https://docs.github.com/en/actions/learn-github-actions/reusing-workflows
	"inputs": NewEmptyStrictObjectType(),
}

// Semantics checker

// ExprSemanticsChecker is a semantics checker for expression syntax. It checks types of values
// in given expression syntax tree. It additionally checks other semantics like arguments of
// format() built-in function. To know the details of the syntax, see
//
// - https://docs.github.com/en/actions/learn-github-actions/contexts
// - https://docs.github.com/en/actions/learn-github-actions/expressions
type ExprSemanticsChecker struct {
	funcs                 map[string][]*FuncSignature
	vars                  map[string]ExprType
	errs                  []*ExprError
	varsCopied            bool
	githubVarCopied       bool
	untrusted             *UntrustedInputChecker
	availableContexts     []string
	availableSpecialFuncs []string
}

// NewExprSemanticsChecker creates new ExprSemanticsChecker instance. When checkUntrustedInput is
// set to true, the checker will make use of possibly untrusted inputs error.
func NewExprSemanticsChecker(checkUntrustedInput bool) *ExprSemanticsChecker {
	c := &ExprSemanticsChecker{
		funcs:           BuiltinFuncSignatures,
		vars:            BuiltinGlobalVariableTypes,
		varsCopied:      false,
		githubVarCopied: false,
	}
	if checkUntrustedInput {
		c.untrusted = NewUntrustedInputChecker(BuiltinUntrustedInputs)
	}
	return c
}

func errorAtExpr(e ExprNode, msg string) *ExprError {
	t := e.Token()
	return &ExprError{
		Message: msg,
		Offset:  t.Offset,
		Line:    t.Line,
		Column:  t.Column,
	}
}

func errorfAtExpr(e ExprNode, format string, args ...interface{}) *ExprError {
	return errorAtExpr(e, fmt.Sprintf(format, args...))
}

func (sema *ExprSemanticsChecker) errorf(e ExprNode, format string, args ...interface{}) {
	sema.errs = append(sema.errs, errorfAtExpr(e, format, args...))
}

func (sema *ExprSemanticsChecker) ensureVarsCopied() {
	if sema.varsCopied {
		return
	}

	// Make shallow copy of current variables map not to pollute global variable
	copied := make(map[string]ExprType, len(sema.vars))
	for k, v := range sema.vars {
		copied[k] = v
	}
	sema.vars = copied
	sema.varsCopied = true
}

func (sema *ExprSemanticsChecker) ensureGithubVarCopied() {
	if sema.githubVarCopied {
		return
	}
	sema.ensureVarsCopied()

	sema.vars["github"] = sema.vars["github"].DeepCopy()
}

// UpdateMatrix updates matrix object to given object type. Since matrix values change according to
// 'matrix' section of job configuration, the type needs to be updated.
func (sema *ExprSemanticsChecker) UpdateMatrix(ty *ObjectType) {
	sema.ensureVarsCopied()
	sema.vars["matrix"] = ty
}

// UpdateSteps updates 'steps' context object to given object type.
func (sema *ExprSemanticsChecker) UpdateSteps(ty *ObjectType) {
	sema.ensureVarsCopied()
	sema.vars["steps"] = ty
}

// UpdateNeeds updates 'needs' context object to given object type.
func (sema *ExprSemanticsChecker) UpdateNeeds(ty *ObjectType) {
	sema.ensureVarsCopied()
	sema.vars["needs"] = ty
}

// UpdateSecrets updates 'secrets' context object to given object type.
func (sema *ExprSemanticsChecker) UpdateSecrets(ty *ObjectType) {
	sema.ensureVarsCopied()

	// Merges automatically supplied secrets with manually defined secrets.
	// ACTIONS_STEP_DEBUG and ACTIONS_RUNNER_DEBUG seem supplied from caller of the workflow (#130)
	copied := NewStrictObjectType(map[string]ExprType{
		"github_token":         StringType{},
		"actions_step_debug":   StringType{},
		"actions_runner_debug": StringType{},
	})
	for n, v := range ty.Props {
		copied.Props[n] = v
	}
	sema.vars["secrets"] = copied
}

// UpdateInputs updates 'inputs' context object to given object type.
func (sema *ExprSemanticsChecker) UpdateInputs(ty *ObjectType) {
	sema.ensureVarsCopied()
	sema.vars["inputs"] = ty
}

// UpdateDispatchInputs updates 'github.event.inputs' and 'inputs' objects to given object type.
// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#workflow_dispatch
func (sema *ExprSemanticsChecker) UpdateDispatchInputs(ty *ObjectType) {
	sema.UpdateInputs(ty)

	// Update `github.event.inputs`.
	// Unlike `inputs.*`, type of `github.event.inputs.*` is always string unlike `inputs.*`. We need
	// to create a new type from `ty` (e.g. {foo: boolean, bar: number} -> {foo: string, bar: string})

	p := make(map[string]ExprType, len(ty.Props))
	for n := range ty.Props {
		p[n] = StringType{}
	}
	ty = NewStrictObjectType(p)

	sema.ensureGithubVarCopied()
	sema.vars["github"].(*ObjectType).Props["event"].(*ObjectType).Props["inputs"] = ty
}

// UpdateJobs updates 'jobs' context object to given object type.
func (sema *ExprSemanticsChecker) UpdateJobs(ty *ObjectType) {
	sema.ensureVarsCopied()
	sema.vars["jobs"] = ty
}

// SetContextAvailability sets available context names while semantics checks. Some contexts limit
// where they can be used.
// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
//
// Elements of 'avail' parameter must be in lower case to check context names in case-insensitive.
//
// If this method is not called before checks, ExprSemanticsChecker considers any contexts are
// available by default.
// Available contexts for workflow keys can be obtained from actionlint.ContextAvailability.
func (sema *ExprSemanticsChecker) SetContextAvailability(avail []string) {
	sema.availableContexts = avail
}

func (sema *ExprSemanticsChecker) checkAvailableContext(n *VariableNode) {
	if len(sema.availableContexts) == 0 {
		return
	}

	ctx := strings.ToLower(n.Name)
	for _, c := range sema.availableContexts {
		if c == ctx {
			return
		}
	}

	s := "contexts are"
	if len(sema.availableContexts) == 1 {
		s = "context is"
	}
	sema.errorf(
		n,
		"context %q is not allowed here. available %s %s. see https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability for more details",
		n.Name,
		s,
		quotes(sema.availableContexts),
	)
}

// SetSpecialFunctionAvailability sets names of available special functions while semantics checks.
// Some functions limit where they can be used.
// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
//
// Elements of 'avail' parameter must be in lower case to check function names in case-insensitive.
//
// If this method is not called before checks, ExprSemanticsChecker considers no special function is
// allowed by default. Allowed functions can be obtained from actionlint.SpecialFunctionNames global
// constant.
//
// Available function names for workflow keys can be obtained from actionlint.ContextAvailability.
func (sema *ExprSemanticsChecker) SetSpecialFunctionAvailability(avail []string) {
	sema.availableSpecialFuncs = avail
}

func (sema *ExprSemanticsChecker) checkSpecialFunctionAvailability(n *FuncCallNode) {
	f := strings.ToLower(n.Callee)

	allowed, ok := SpecialFunctionNames[f]
	if !ok {
		return // This function is not special
	}

	for _, sp := range sema.availableSpecialFuncs {
		if sp == f {
			return
		}
	}

	sema.errorf(
		n,
		"calling function %q is not allowed here. %q is only available in %s. see https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability for more details",
		n.Callee,
		n.Callee,
		quotes(allowed),
	)
}

func (sema *ExprSemanticsChecker) visitUntrustedCheckerOnLeaveNode(n ExprNode) {
	if sema.untrusted != nil {
		sema.untrusted.OnVisitNodeLeave(n)
	}
}

func (sema *ExprSemanticsChecker) checkVariable(n *VariableNode) ExprType {
	v, ok := sema.vars[n.Name]
	if !ok {
		ss := make([]string, 0, len(sema.vars))
		for n := range sema.vars {
			ss = append(ss, n)
		}
		sema.errorf(n, "undefined variable %q. available variables are %s", n.Token().Value, sortedQuotes(ss))
		return AnyType{}
	}

	sema.checkAvailableContext(n)
	return v
}

func (sema *ExprSemanticsChecker) checkObjectDeref(n *ObjectDerefNode) ExprType {
	switch ty := sema.check(n.Receiver).(type) {
	case AnyType:
		return AnyType{}
	case *ObjectType:
		if t, ok := ty.Props[n.Property]; ok {
			return t
		}
		if ty.Mapped != nil {
			return ty.Mapped
		}
		if ty.IsStrict() {
			sema.errorf(n, "property %q is not defined in object type %s", n.Property, ty.String())
		}
		return AnyType{}
	case *ArrayType:
		if !ty.Deref {
			sema.errorf(n, "receiver of object dereference %q must be type of object but got %q", n.Property, ty.String())
			return AnyType{}
		}
		switch et := ty.Elem.(type) {
		case AnyType:
			// When element type is any, map the any type to any. Reuse `ty`
			return ty
		case *ObjectType:
			// Map element type of delererenced array
			var elem ExprType = AnyType{}
			if t, ok := et.Props[n.Property]; ok {
				elem = t
			} else if et.Mapped != nil {
				elem = et.Mapped
			} else if et.IsStrict() {
				sema.errorf(n, "property %q is not defined in object type %s as element of filtered array", n.Property, et.String())
			}
			return &ArrayType{elem, true}
		default:
			sema.errorf(
				n,
				"property filtered by %q at object filtering must be type of object but got %q",
				n.Property,
				ty.Elem.String(),
			)
			return AnyType{}
		}
	default:
		sema.errorf(n, "receiver of object dereference %q must be type of object but got %q", n.Property, ty.String())
		return AnyType{}
	}
}

func (sema *ExprSemanticsChecker) checkArrayDeref(n *ArrayDerefNode) ExprType {
	switch ty := sema.check(n.Receiver).(type) {
	case AnyType:
		return &ArrayType{AnyType{}, true}
	case *ArrayType:
		ty.Deref = true
		return ty
	case *ObjectType:
		// Object filtering is available for objects, not only arrays (#66)

		if ty.Mapped != nil {
			// For map object or loose object at receiver of .*
			switch mty := ty.Mapped.(type) {
			case AnyType:
				return &ArrayType{AnyType{}, true}
			case *ObjectType:
				return &ArrayType{mty, true}
			default:
				sema.errorf(n, "elements of object at receiver of object filtering `.*` must be type of object but got %q. the type of receiver was %q", mty.String(), ty.String())
				return AnyType{}
			}
		}

		// For strict object at receiver of .*
		found := false
		for _, t := range ty.Props {
			if _, ok := t.(*ObjectType); ok {
				found = true
				break
			}
		}
		if !found {
			sema.errorf(n, "object type %q cannot be filtered by object filtering `.*` since it has no object element", ty.String())
			return AnyType{}
		}

		return &ArrayType{AnyType{}, true}
	default:
		sema.errorf(n, "receiver of object filtering `.*` must be type of array or object but got %q", ty.String())
		return AnyType{}
	}
}

func (sema *ExprSemanticsChecker) checkIndexAccess(n *IndexAccessNode) ExprType {
	// Note: Index must be visited before Index to make UntrustedInputChecker work correctly even if
	// the expression has some nest like foo[aaa.bbb].bar. Nest happens in top-down order and
	// properties/indices access check is done in bottom-up order. So, as far as we visit nested
	// index nodes before visiting operand, the index is recursively checked first.
	idx := sema.check(n.Index)

	switch ty := sema.check(n.Operand).(type) {
	case AnyType:
		return AnyType{}
	case *ArrayType:
		switch idx.(type) {
		case AnyType, NumberType:
			return ty.Elem
		default:
			sema.errorf(n.Index, "index access of array must be type of number but got %q", idx.String())
			return AnyType{}
		}
	case *ObjectType:
		switch idx.(type) {
		case AnyType:
			return AnyType{}
		case StringType:
			// Index access with string literal like foo['bar']
			if lit, ok := n.Index.(*StringNode); ok {
				if prop, ok := ty.Props[lit.Value]; ok {
					return prop
				}
				if ty.Mapped != nil {
					return ty.Mapped
				}
				if ty.IsStrict() {
					sema.errorf(n, "property %q is not defined in object type %s", lit.Value, ty.String())
				}
			}
			if ty.Mapped != nil {
				return ty.Mapped
			}
			return AnyType{} // Fallback
		default:
			sema.errorf(n.Index, "property access of object must be type of string but got %q", idx.String())
			return AnyType{}
		}
	default:
		sema.errorf(n, "index access operand must be type of object or array but got %q", ty.String())
		return AnyType{}
	}
}

func checkFuncSignature(n *FuncCallNode, sig *FuncSignature, args []ExprType) *ExprError {
	lp, la := len(sig.Params), len(args)
	if sig.VariableLengthParams && (lp > la) || !sig.VariableLengthParams && lp != la {
		atLeast := ""
		if sig.VariableLengthParams {
			atLeast = "at least "
		}
		return errorfAtExpr(
			n,
			"number of arguments is wrong. function %q takes %s%d parameters but %d arguments are given",
			sig.String(),
			atLeast,
			lp,
			la,
		)
	}

	for i := 0; i < len(sig.Params); i++ {
		p, a := sig.Params[i], args[i]
		if !p.Assignable(a) {
			return errorfAtExpr(
				n.Args[i],
				"%s argument of function call is not assignable. %q cannot be assigned to %q. called function type is %q",
				ordinal(i+1),
				a.String(),
				p.String(),
				sig.String(),
			)
		}
	}

	// Note: Unlike many languages, this check does not allow 0 argument for the variable length
	// parameter since it is useful for checking hashFiles() and format().
	if sig.VariableLengthParams {
		rest := args[lp:]
		p := sig.Params[lp-1]
		for i, a := range rest {
			if !p.Assignable(a) {
				return errorfAtExpr(
					n.Args[lp+i],
					"%s argument of function call is not assignable. %q cannot be assigned to %q. called function type is %q",
					ordinal(lp+i+1),
					a.String(),
					p.String(),
					sig.String(),
				)
			}
		}
	}

	return nil
}

func (sema *ExprSemanticsChecker) checkBuiltinFunctionCall(n *FuncCallNode, sig *FuncSignature) {
	sema.checkSpecialFunctionAvailability(n)

	// Special checks for specific built-in functions
	switch n.Callee {
	case "format":
		lit, ok := n.Args[0].(*StringNode)
		if !ok {
			return
		}
		l := len(n.Args) - 1 // -1 means removing first format string argument

		// Find all placeholders in format string
		holders := make(map[int]struct{}, l)
		for _, m := range reFormatPlaceholder.FindAllString(lit.Value, -1) {
			i, _ := strconv.Atoi(m[1 : len(m)-1])
			holders[i] = struct{}{}
		}

		for i := 0; i < l; i++ {
			_, ok := holders[i]
			if !ok {
				sema.errorf(n, "format string %q does not contain placeholder {%d}. remove argument which is unused in the format string", lit.Value, i)
				continue
			}
			delete(holders, i) // forget it to check unused placeholders
		}

		for i := range holders {
			sema.errorf(n, "format string %q contains placeholder {%d} but only %d arguments are given to format", lit.Value, i, l)
		}
	}
}

func (sema *ExprSemanticsChecker) checkFuncCall(n *FuncCallNode) ExprType {
	// Check function name in case insensitive. For example, toJson and toJSON are the same function.
	callee := strings.ToLower(n.Callee)
	sigs, ok := sema.funcs[callee]
	if !ok {
		ss := make([]string, 0, len(sema.funcs))
		for n := range sema.funcs {
			ss = append(ss, n)
		}
		sema.errorf(n, "undefined function %q. available functions are %s", n.Callee, sortedQuotes(ss))
		return AnyType{}
	}

	tys := make([]ExprType, 0, len(n.Args))
	for _, a := range n.Args {
		tys = append(tys, sema.check(a))
	}

	// Check all overloads
	errs := []*ExprError{}
	for _, sig := range sigs {
		err := checkFuncSignature(n, sig, tys)
		if err == nil {
			// When one of overload pass type check, overload was resolved correctly
			sema.checkBuiltinFunctionCall(n, sig)
			return sig.Ret
		}
		errs = append(errs, err)
	}

	// All candidates failed
	sema.errs = append(sema.errs, errs...)

	return AnyType{}
}

func (sema *ExprSemanticsChecker) checkNotOp(n *NotOpNode) ExprType {
	ty := sema.check(n.Operand)
	if !(BoolType{}).Assignable(ty) {
		sema.errorf(n, "type of operand of ! operator %q is not assignable to type \"bool\"", ty.String())
	}
	return BoolType{}
}

func (sema *ExprSemanticsChecker) checkCompareOp(n *CompareOpNode) ExprType {
	sema.check(n.Left)
	sema.check(n.Right)
	// Note: Comparing values is very loose. Any value can be compared with any value without an
	// error.
	// https://docs.github.com/en/actions/learn-github-actions/expressions#operators
	return BoolType{}
}

func (sema *ExprSemanticsChecker) checkLogicalOp(n *LogicalOpNode) ExprType {
	lty := sema.check(n.Left)
	rty := sema.check(n.Right)
	return lty.Merge(rty)
}

func (sema *ExprSemanticsChecker) check(expr ExprNode) ExprType {
	defer sema.visitUntrustedCheckerOnLeaveNode(expr) // Call this method in bottom-up order

	switch e := expr.(type) {
	case *VariableNode:
		return sema.checkVariable(e)
	case *NullNode:
		return NullType{}
	case *BoolNode:
		return BoolType{}
	case *StringNode:
		return StringType{}
	case *IntNode, *FloatNode:
		return NumberType{}
	case *ObjectDerefNode:
		return sema.checkObjectDeref(e)
	case *ArrayDerefNode:
		return sema.checkArrayDeref(e)
	case *IndexAccessNode:
		return sema.checkIndexAccess(e)
	case *FuncCallNode:
		return sema.checkFuncCall(e)
	case *NotOpNode:
		return sema.checkNotOp(e)
	case *CompareOpNode:
		return sema.checkCompareOp(e)
	case *LogicalOpNode:
		return sema.checkLogicalOp(e)
	default:
		panic("unreachable")
	}
}

// Check checks semantics of given expression syntax tree. It returns the type of the expression as
// the first return value when the check was successfully done. And it returns all errors found
// while checking the expression as the second return value.
func (sema *ExprSemanticsChecker) Check(expr ExprNode) (ExprType, []*ExprError) {
	sema.errs = []*ExprError{}
	if sema.untrusted != nil {
		sema.untrusted.Init()
	}
	ty := sema.check(expr)
	errs := sema.errs
	if sema.untrusted != nil {
		sema.untrusted.OnVisitEnd()
		errs = append(errs, sema.untrusted.Errs()...)
	}
	return ty, errs
}
