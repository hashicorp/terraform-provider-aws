package actionlint

import (
	"fmt"
	"strconv"
	"strings"
)

//go:generate go run ./scripts/generate-availability ./availability.go

type typedExpr struct {
	ty  ExprType
	pos Pos
}

func isExprAssigned(str *String) bool {
	s := strings.TrimSpace(str.Value)
	return strings.HasPrefix(s, "${{") && strings.HasSuffix(s, "}}")
}

// RuleExpression is a rule checker to check expression syntax in string values of workflow syntax.
// It checks syntax and semantics of the expressions including type checks and functions/contexts
// definitions. For more details see
// - https://docs.github.com/en/actions/learn-github-actions/contexts
// - https://docs.github.com/en/actions/learn-github-actions/expressions
type RuleExpression struct {
	RuleBase
	matrixTy         *ObjectType
	stepsTy          *ObjectType
	needsTy          *ObjectType
	secretsTy        *ObjectType
	inputsTy         *ObjectType
	dispatchInputsTy *ObjectType
	jobsTy           *ObjectType
	workflow         *Workflow
	localActions     *LocalActionsCache
	localWorkflows   *LocalReusableWorkflowCache
}

// NewRuleExpression creates new RuleExpression instance.
func NewRuleExpression(actionsCache *LocalActionsCache, workflowCache *LocalReusableWorkflowCache) *RuleExpression {
	return &RuleExpression{
		RuleBase:         RuleBase{name: "expression"},
		matrixTy:         nil,
		stepsTy:          nil,
		needsTy:          nil,
		secretsTy:        nil,
		inputsTy:         nil,
		dispatchInputsTy: nil,
		jobsTy:           nil,
		workflow:         nil,
		localActions:     actionsCache,
		localWorkflows:   workflowCache,
	}
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RuleExpression) VisitWorkflowPre(n *Workflow) error {
	rule.checkString(n.Name, "")

	for _, e := range n.On {
		switch e := e.(type) {
		case *WebhookEvent:
			rule.checkStrings(e.Types, "")
			rule.checkWebhookEventFilter(e.Branches)
			rule.checkWebhookEventFilter(e.BranchesIgnore)
			rule.checkWebhookEventFilter(e.Tags)
			rule.checkWebhookEventFilter(e.TagsIgnore)
			rule.checkWebhookEventFilter(e.Paths)
			rule.checkWebhookEventFilter(e.PathsIgnore)
			rule.checkStrings(e.Workflows, "")
		case *ScheduledEvent:
			rule.checkStrings(e.Cron, "")
		case *WorkflowDispatchEvent:
			ity := NewEmptyStrictObjectType()
			for id, i := range e.Inputs {
				rule.checkString(i.Description, "")
				rule.checkString(i.Default, "")
				rule.checkBool(i.Required, "")
				rule.checkStrings(i.Options, "")

				var ty ExprType
				switch i.Type {
				case WorkflowDispatchEventInputTypeBoolean:
					ty = BoolType{}
				case WorkflowDispatchEventInputTypeString, WorkflowDispatchEventInputTypeChoice, WorkflowDispatchEventInputTypeEnvironment:
					ty = StringType{}
				default:
					ty = AnyType{}
				}
				ity.Props[id] = ty
			}
			rule.dispatchInputsTy = ity
		case *RepositoryDispatchEvent:
			rule.checkStrings(e.Types, "")
		case *WorkflowCallEvent:
			ity := NewEmptyStrictObjectType()

			// Set `inputs` context object before checking inputs since input's default values can refer `inputs` context.
			//   inputs:
			//     input1:
			//       type: string
			//     input2:
			//       type: string
			//       default: ${{ inputs.input1 }}
			rule.inputsTy = ity

			for _, i := range e.Inputs {
				rule.checkString(i.Description, "")
				// Check default value before setting type to `ity` because refering myself should cause an error.
				//   inputs:
				//     recursive:
				//       type: string
				//       default: ${{ inputs.recursive }}
				ts := rule.checkString(i.Default, "on.workflow_call.inputs.<inputs_id>.default")

				var ty ExprType
				switch i.Type {
				case WorkflowCallEventInputTypeString:
					ty = StringType{}
				case WorkflowCallEventInputTypeBoolean:
					ty = BoolType{}
					if len(ts) == 1 && isExprAssigned(i.Default) {
						switch ts[0].ty.(type) {
						case BoolType, AnyType:
							// ok
						default:
							rule.errorf(i.Default.Pos, "type of input %q must be bool but found type %s", i.Name.Value, ts[0].ty.String())
						}
					}
				case WorkflowCallEventInputTypeNumber:
					ty = NumberType{}
					if len(ts) == 1 && isExprAssigned(i.Default) {
						switch ts[0].ty.(type) {
						case NumberType, AnyType:
							// ok
						default:
							rule.errorf(i.Default.Pos, "type of input %q must be number but found type %s", i.Name.Value, ts[0].ty.String())
						}
					}
				default:
					ty = AnyType{}
				}
				ity.Props[i.ID] = ty
			}

			// When no secret is passed, secrets may be inherited from a caller of the workflow.
			// So `secrets` context must be typed as { string => string }. `e.Secrets` is nil when `secrets:` does not
			// exist. When `e.Secrets` is an empty map, `secrets:` exists but it has no child.
			if e.Secrets != nil {
				sty := NewEmptyStrictObjectType()
				for id, s := range e.Secrets {
					sty.Props[id] = StringType{}
					rule.checkString(s.Description, "")
				}
				rule.secretsTy = sty
			}

			for _, o := range e.Outputs {
				rule.checkString(o.Description, "")
				// o.Value will be checked in VisitWorkflowPost
			}
		}
	}

	rule.checkString(n.RunName, "run-name")
	rule.checkEnv(n.Env, "env")

	rule.checkDefaults(n.Defaults, "")
	rule.checkConcurrency(n.Concurrency, "concurrency")

	rule.workflow = n
	return nil
}

// VisitWorkflowPost is callback when visiting Workflow node after visiting its children
func (rule *RuleExpression) VisitWorkflowPost(n *Workflow) error {
	if e, ok := n.FindWorkflowCallEvent(); ok {
		rule.checkWorkflowCallOutputs(e.Outputs, n.Jobs)
	}
	rule.workflow = nil
	return nil
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleExpression) VisitJobPre(n *Job) error {
	// Type of needs must be resolved before resolving type of matrix because `needs` context can
	// be used in matrix configuration.
	rule.needsTy = rule.calcNeedsType(n)

	// Set matrix type at start of VisitJobPre() because matrix values are available in
	// jobs.<job_id> section. For example:
	//   jobs:
	//     foo:
	//       strategy:
	//         matrix:
	//           os: [ubuntu-latest, macos-latest, windows-latest]
	//       runs-on: ${{ matrix.os }}
	if n.Strategy != nil && n.Strategy.Matrix != nil {
		rule.matrixTy = rule.guessTypeOfMatrix(n.Strategy.Matrix)
	}

	rule.checkString(n.Name, "jobs.<job_id>.name")
	rule.checkStrings(n.Needs, "")

	if n.RunsOn != nil {
		if n.RunsOn.Expression != nil {
			if ty := rule.checkOneExpression(n.RunsOn.Expression, "runner label at \"runs-on\" section", "jobs.<job_id>.runs-on"); ty != nil {
				switch ty.(type) {
				case *ArrayType, StringType, AnyType:
					// OK
				default:
					rule.errorf(n.RunsOn.Expression.Pos, "type of expression at \"runs-on\" must be string or array but found type %q", ty.String())
				}
			}
		} else {
			for _, l := range n.RunsOn.Labels {
				rule.checkString(l, "jobs.<job_id>.runs-on")
			}
		}
	}

	rule.checkConcurrency(n.Concurrency, "jobs.<job_id>.concurrency")

	rule.checkEnv(n.Env, "jobs.<job_id>.env")

	rule.checkDefaults(n.Defaults, "jobs.<job_id>.defaults.run")
	rule.checkIfCondition(n.If, "jobs.<job_id>.if")

	if n.Strategy != nil {
		if n.Strategy.Matrix != nil {
			for _, r := range n.Strategy.Matrix.Rows {
				for _, v := range r.Values {
					rule.checkRawYAMLValue(v)
				}
			}
			rule.checkMatrixCombinations(n.Strategy.Matrix.Include, "include")
			rule.checkMatrixCombinations(n.Strategy.Matrix.Exclude, "exclude")
		}
		rule.checkBool(n.Strategy.FailFast, "jobs.<job_id>.strategy")
		rule.checkInt(n.Strategy.MaxParallel, "jobs.<job_id>.strategy")
	}

	rule.checkBool(n.ContinueOnError, "jobs.<job_id>.continue-on-error")
	rule.checkFloat(n.TimeoutMinutes, "jobs.<job_id>.timeout-minutes")
	rule.checkContainer(n.Container, "jobs.<job_id>.container", "")

	for _, s := range n.Services {
		rule.checkContainer(s.Container, "jobs.<job_id>.services", "<service_id>")
	}

	rule.checkWorkflowCall(n.WorkflowCall)

	rule.stepsTy = NewEmptyStrictObjectType()

	return nil
}

// VisitJobPost is callback when visiting Job node after visiting its children
func (rule *RuleExpression) VisitJobPost(n *Job) error {
	// 'environment' and 'outputs' sections are evaluated after all steps are run
	if n.Environment != nil {
		rule.checkString(n.Environment.Name, "jobs.<job_id>.environment")
		rule.checkString(n.Environment.URL, "jobs.<job_id>.environment.url")
	}
	for _, output := range n.Outputs {
		rule.checkString(output.Value, "jobs.<job_id>.outputs.<output_id>")
	}

	rule.matrixTy = nil
	rule.stepsTy = nil
	rule.needsTy = nil

	return nil
}

// VisitStep is callback when visiting Step node.
func (rule *RuleExpression) VisitStep(n *Step) error {
	rule.checkString(n.Name, "jobs.<job_id>.steps.name")
	rule.checkIfCondition(n.If, "jobs.<job_id>.steps.if")

	var spec *String
	switch e := n.Exec.(type) {
	case *ExecRun:
		rule.checkScriptString(e.Run, "jobs.<job_id>.steps.run")
		rule.checkString(e.Shell, "")
		rule.checkString(e.WorkingDirectory, "jobs.<job_id>.steps.working-directory")
	case *ExecAction:
		rule.checkString(e.Uses, "")
		for n, i := range e.Inputs {
			if e.Uses != nil && strings.HasPrefix(e.Uses.Value, "actions/github-script@") && n == "script" {
				rule.checkScriptString(i.Value, "jobs.<job_id>.steps.with")
			} else {
				rule.checkString(i.Value, "jobs.<job_id>.steps.with")
			}
		}
		rule.checkString(e.Entrypoint, "")
		rule.checkString(e.Args, "")
		spec = e.Uses
	}

	rule.checkEnv(n.Env, "jobs.<job_id>.steps.env") // env: at step level can refer 'env' context (#158)
	rule.checkBool(n.ContinueOnError, "jobs.<job_id>.steps.continue-on-error")
	rule.checkFloat(n.TimeoutMinutes, "jobs.<job_id>.steps.timeout-minutes")

	if n.ID != nil {
		// Step ID is case insensitive
		id := strings.ToLower(n.ID.Value)
		if strings.Contains(id, "${{") && strings.Contains(id, "}}") {
			rule.checkString(n.ID, "")
			rule.stepsTy.Loose()
		}
		rule.stepsTy.Props[id] = NewStrictObjectType(map[string]ExprType{
			"outputs":    rule.getActionOutputsType(spec),
			"conclusion": StringType{},
			"outcome":    StringType{},
		})
	}

	return nil
}

// Get type of `outputs.<output name>`
func (rule *RuleExpression) getActionOutputsType(spec *String) *ObjectType {
	if spec == nil {
		return NewMapObjectType(StringType{})
	}

	if strings.HasPrefix(spec.Value, "./") {
		meta, err := rule.localActions.FindMetadata(spec.Value)
		if err != nil {
			rule.error(spec.Pos, err.Error())
			return NewMapObjectType(StringType{})
		}
		if meta == nil {
			return NewMapObjectType(StringType{})
		}

		return typeOfActionOutputs(meta)
	}

	// github-script action allows to set any outputs through calling `core.setOutput` directly.
	// So any `outputs.*` properties should be accepted (#104)
	if strings.HasPrefix(spec.Value, "actions/github-script@") {
		return NewEmptyObjectType()
	}

	// When the action run at this step is a popular action, we know what outputs are set by it.
	// Set the output names to `steps.{step_id}.outputs.{name}`.
	if meta, ok := PopularActions[spec.Value]; ok {
		return typeOfActionOutputs(meta)
	}

	return NewMapObjectType(StringType{})
}

func (rule *RuleExpression) getWorkflowCallOutputsType(call *WorkflowCall) *ObjectType {
	if call.Uses == nil {
		return NewMapObjectType(StringType{})
	}

	m, err := rule.localWorkflows.FindMetadata(call.Uses.Value)
	if err != nil {
		rule.error(call.Uses.Pos, err.Error())
		return NewMapObjectType(StringType{})
	}
	if m == nil {
		return NewMapObjectType(StringType{})
	}

	p := make(map[string]ExprType, len(m.Outputs))
	for n := range m.Outputs {
		p[n] = StringType{}
	}
	return NewStrictObjectType(p)
}

func (rule *RuleExpression) checkOneExpression(s *String, what, workflowKey string) ExprType {
	// checkString is not available since it checks types for embedding values into a string
	if s == nil {
		return nil
	}

	ts, ok := rule.checkExprsIn(s.Value, s.Pos, s.Quoted, false, workflowKey)
	if !ok {
		return nil
	}

	if len(ts) != 1 {
		// This case should be unreachable since only one ${{ }} is included is checked by parser
		rule.errorf(s.Pos, "one ${{ }} expression should be included in %q value but got %d expressions", what, len(ts))
		return nil
	}

	return ts[0].ty
}

func (rule *RuleExpression) checkObjectTy(ty ExprType, pos *Pos, what string) ExprType {
	if ty == nil {
		return nil
	}
	switch ty.(type) {
	case *ObjectType, AnyType:
		return ty
	default:
		rule.errorf(pos, "type of expression at %q must be object but found type %s", what, ty.String())
		return nil
	}
}

func (rule *RuleExpression) checkArrayTy(ty ExprType, pos *Pos, what string) ExprType {
	if ty == nil {
		return nil
	}
	switch ty.(type) {
	case *ArrayType, AnyType:
		return ty
	default:
		rule.errorf(pos, "type of expression at %q must be array but found type %s", what, ty.String())
		return nil
	}
}

func (rule *RuleExpression) checkNumberTy(ty ExprType, pos *Pos, what string) ExprType {
	if ty == nil {
		return nil
	}
	switch ty.(type) {
	case NumberType, AnyType:
		return ty
	default:
		rule.errorf(pos, "type of expression at %q must be number but found type %s", what, ty.String())
		return nil
	}
}

func (rule *RuleExpression) checkObjectExpression(s *String, what, workflowKey string) ExprType {
	ty := rule.checkOneExpression(s, what, workflowKey)
	if ty == nil {
		return nil
	}
	return rule.checkObjectTy(ty, s.Pos, what)
}

func (rule *RuleExpression) checkArrayExpression(s *String, what, workflowKey string) ExprType {
	ty := rule.checkOneExpression(s, what, workflowKey)
	if ty == nil {
		return nil
	}
	return rule.checkArrayTy(ty, s.Pos, what)
}

func (rule *RuleExpression) checkNumberExpression(s *String, what, workflowKey string) ExprType {
	ty := rule.checkOneExpression(s, what, workflowKey)
	if ty == nil {
		return nil
	}
	return rule.checkNumberTy(ty, s.Pos, what)
}

func (rule *RuleExpression) checkMatrixCombinations(cs *MatrixCombinations, what string) {
	if cs == nil {
		return
	}

	if cs.Expression != nil {
		if ty, ok := rule.checkArrayExpression(cs.Expression, what, "jobs.<job_id>.strategy").(*ArrayType); ok {
			rule.checkObjectTy(ty.Elem, cs.Expression.Pos, what)
		}
		return
	}

	what = fmt.Sprintf("matrix combination at element of %s section", what)
	for _, combi := range cs.Combinations {
		if combi.Expression != nil {
			rule.checkObjectExpression(combi.Expression, what, "jobs.<job_id>.strategy")
			continue
		}
		for _, a := range combi.Assigns {
			rule.checkRawYAMLValue(a.Value)
		}
	}
}

func (rule *RuleExpression) checkEnv(env *Env, workflowKey string) {
	if env == nil {
		return
	}

	if env.Vars != nil {
		for _, e := range env.Vars {
			rule.checkString(e.Value, workflowKey)
		}
		return
	}

	// When form of "env: ${{...}}"
	rule.checkObjectExpression(env.Expression, "env", workflowKey)
}

func (rule *RuleExpression) checkContainer(c *Container, workflowKey, childWorkflowKeyPrefix string) {
	if c == nil {
		return
	}
	childWorkflowKey := workflowKey
	if childWorkflowKeyPrefix != "" {
		childWorkflowKey += "." + childWorkflowKeyPrefix
	}
	rule.checkString(c.Image, workflowKey)
	if c.Credentials != nil {
		k := childWorkflowKey + ".credentials" // e.g. jobs.<job_id>.container.credentials
		rule.checkString(c.Credentials.Username, k)
		rule.checkString(c.Credentials.Password, k)
	}
	rule.checkEnv(c.Env, workflowKey+".env.<env_id>") // e.g. jobs.<job_id>.container.env.<env_id>
	rule.checkStrings(c.Ports, workflowKey)
	rule.checkStrings(c.Volumes, workflowKey)
	rule.checkString(c.Options, workflowKey)
}

func (rule *RuleExpression) checkConcurrency(c *Concurrency, workflowKey string) {
	if c == nil {
		return
	}
	rule.checkString(c.Group, workflowKey)
	rule.checkBool(c.CancelInProgress, workflowKey)
}

func (rule *RuleExpression) checkDefaults(d *Defaults, workflowKey string) {
	if d == nil || d.Run == nil {
		return
	}
	rule.checkString(d.Run.Shell, workflowKey)
	rule.checkString(d.Run.WorkingDirectory, workflowKey)
}

func (rule *RuleExpression) checkWorkflowCall(c *WorkflowCall) {
	if c == nil || c.Uses == nil {
		return
	}

	rule.checkString(c.Uses, "")

	m, err := rule.localWorkflows.FindMetadata(c.Uses.Value)
	if err != nil {
		rule.error(c.Uses.Pos, err.Error())
	}

	for n, i := range c.Inputs {
		ts := rule.checkString(i.Value, "jobs.<job_id>.with.<with_id>")

		if m == nil {
			continue
		}

		mi, ok := m.Inputs[n]
		if !ok || mi == nil {
			continue
		}
		if _, ok := mi.Type.(AnyType); ok {
			continue
		}

		v := strings.TrimSpace(i.Value.Value)

		var ty ExprType = StringType{}
		switch len(ts) {
		case 0:
			switch v {
			case "null":
				ty = NullType{}
			case "true", "false":
				ty = BoolType{}
			default:
				if _, err := strconv.ParseFloat(v, 64); err == nil {
					ty = NumberType{}
				}
			}
		case 1:
			if isExprAssigned(i.Value) {
				ty = ts[0].ty
			}
		}

		if !mi.Type.Assignable(ty) {
			rule.errorf(
				i.Value.Pos,
				"input %q is typed as %s by reusable workflow %q. %s value cannot be assigned",
				mi.Name,
				mi.Type.String(),
				c.Uses.Value,
				ty.String(),
			)
		}
	}

	for _, s := range c.Secrets {
		rule.checkString(s.Value, "jobs.<job_id>.secrets.<secrets_id>")
	}
}

func (rule *RuleExpression) checkWebhookEventFilter(f *WebhookEventFilter) {
	if f == nil {
		return
	}
	rule.checkStrings(f.Values, "")
}

func (rule *RuleExpression) checkStrings(ss []*String, workflowKey string) {
	for _, s := range ss {
		rule.checkString(s, workflowKey)
	}
}

func (rule *RuleExpression) checkIfCondition(str *String, workflowKey string) {
	if str == nil {
		return
	}

	// Note:
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idif
	//
	// > When you use expressions in an if conditional, you may omit the expression syntax (${{ }})
	// > because GitHub automatically evaluates the if conditional as an expression, unless the
	// > expression contains any operators. If the expression contains any operators, the expression
	// > must be contained within ${{ }} to explicitly mark it for evaluation.
	//
	// This document is actually wrong. I confirmed that any strings without surrounding in ${{ }}
	// are evaluated.
	//
	// - run: echo 'run'
	//   if: '!false'
	// - run: echo 'not run'
	//   if: '!true'
	// - run: echo 'run'
	//   if: false || true
	// - run: echo 'run'
	//   if: true && true
	// - run: echo 'not run'
	//   if: true && false

	var condTy ExprType
	if strings.Contains(str.Value, "${{") && strings.Contains(str.Value, "}}") {
		ts := rule.checkString(str, workflowKey)

		if len(ts) == 1 {
			if isExprAssigned(str) {
				condTy = ts[0].ty
			}
		}
	} else {
		src := str.Value + "}}" // }} is necessary since lexer lexes it as end of tokens
		line, col := str.Pos.Line, str.Pos.Col

		p := NewExprParser()
		expr, err := p.Parse(NewExprLexer(src))
		if err != nil {
			rule.exprError(err, line, col)
			return
		}

		if ty, ok := rule.checkSemanticsOfExprNode(expr, line, col, false, workflowKey); ok {
			condTy = ty
		}
	}

	if condTy != nil && !(BoolType{}).Assignable(condTy) {
		rule.errorf(str.Pos, "\"if\" condition should be type \"bool\" but got type %q", condTy.String())
	}
}

func (rule *RuleExpression) checkTemplateEvaluatedType(ts []typedExpr) {
	for _, t := range ts {
		switch t.ty.(type) {
		case *ObjectType, *ArrayType, NullType:
			rule.errorf(&t.pos, "object, array, and null values should not be evaluated in template with ${{ }} but evaluating the value of type %s", t.ty)
		}
	}
}

func (rule *RuleExpression) checkString(str *String, workflowKey string) []typedExpr {
	if str == nil {
		return nil
	}

	ts, ok := rule.checkExprsIn(str.Value, str.Pos, str.Quoted, false, workflowKey)
	if !ok {
		return nil
	}

	rule.checkTemplateEvaluatedType(ts)
	return ts
}

func (rule *RuleExpression) checkScriptString(str *String, workflowKey string) {
	if str == nil {
		return
	}

	ts, ok := rule.checkExprsIn(str.Value, str.Pos, str.Quoted, true, workflowKey)
	if !ok {
		return
	}

	rule.checkTemplateEvaluatedType(ts)
}

func (rule *RuleExpression) checkBool(b *Bool, workflowKey string) {
	if b == nil || b.Expression == nil {
		return
	}

	ty := rule.checkOneExpression(b.Expression, "bool value", workflowKey)
	if ty == nil {
		return
	}

	switch ty.(type) {
	case BoolType, AnyType:
		// ok
	default:
		rule.errorf(b.Expression.Pos, "type of expression must be bool but found type %s", ty.String())
	}
}

func (rule *RuleExpression) checkInt(i *Int, workflowKey string) {
	if i == nil {
		return
	}
	rule.checkNumberExpression(i.Expression, "integer value", workflowKey)
}

func (rule *RuleExpression) checkFloat(f *Float, workflowKey string) {
	if f == nil {
		return
	}
	rule.checkNumberExpression(f.Expression, "float number value", workflowKey)
}

func (rule *RuleExpression) checkExprsIn(s string, pos *Pos, quoted, checkUntrusted bool, workflowKey string) ([]typedExpr, bool) {
	// TODO: Line number is not correct when the string contains newlines.

	line, col := pos.Line, pos.Col
	if quoted {
		col++ // when the string is quoted like 'foo' or "foo", column should be incremented
	}
	offset := 0
	ts := []typedExpr{}
	for {
		idx := strings.Index(s, "${{")
		if idx == -1 {
			break
		}

		start := idx + 3 // 3 means removing "${{"
		s = s[start:]
		offset += start
		col := col + offset

		ty, offsetAfter, ok := rule.checkSemantics(s, line, col, checkUntrusted, workflowKey)
		if !ok {
			return nil, false
		}
		if ty == nil || offsetAfter == 0 {
			return nil, true
		}
		ts = append(ts, typedExpr{ty, Pos{line, col - 3}})

		s = s[offsetAfter:]
		offset += offsetAfter
	}

	return ts, true
}

func (rule *RuleExpression) checkRawYAMLValue(v RawYAMLValue) {
	switch v := v.(type) {
	case *RawYAMLObject:
		for _, p := range v.Props {
			rule.checkRawYAMLValue(p)
		}
	case *RawYAMLArray:
		for _, v := range v.Elems {
			rule.checkRawYAMLValue(v)
		}
	case *RawYAMLString:
		rule.checkExprsIn(v.Value, v.Pos(), false, false, "jobs.<job_id>.strategy")
	default:
		panic("unreachable")
	}
}

func (rule *RuleExpression) exprError(err *ExprError, lineBase, colBase int) {
	pos := convertExprLineColToPos(err.Line, err.Column, lineBase, colBase)
	rule.error(pos, err.Message)
}

func (rule *RuleExpression) checkSemanticsOfExprNode(expr ExprNode, line, col int, checkUntrusted bool, workflowKey string) (ExprType, bool) {
	c := NewExprSemanticsChecker(checkUntrusted)
	if rule.matrixTy != nil {
		c.UpdateMatrix(rule.matrixTy)
	}
	if rule.stepsTy != nil {
		c.UpdateSteps(rule.stepsTy)
	}
	if rule.needsTy != nil {
		c.UpdateNeeds(rule.needsTy)
	}
	if rule.secretsTy != nil {
		c.UpdateSecrets(rule.secretsTy)
	}
	if rule.inputsTy != nil {
		c.UpdateInputs(rule.inputsTy)
	}
	if rule.dispatchInputsTy != nil {
		c.UpdateDispatchInputs(rule.dispatchInputsTy)
	}
	if rule.jobsTy != nil {
		c.UpdateJobs(rule.jobsTy)
	}
	if workflowKey != "" {
		ctx, sp := WorkflowKeyAvailability(workflowKey)
		if len(ctx) == 0 {
			rule.debug("No context avaiability was found for workflow key %q", workflowKey)
		}
		c.SetContextAvailability(ctx)
		c.SetSpecialFunctionAvailability(sp)
	}

	ty, errs := c.Check(expr)
	for _, err := range errs {
		rule.exprError(err, line, col)
	}

	return ty, len(errs) == 0
}

func (rule *RuleExpression) checkSemantics(src string, line, col int, checkUntrusted bool, workflowKey string) (ExprType, int, bool) {
	l := NewExprLexer(src)
	p := NewExprParser()
	expr, err := p.Parse(l)
	if err != nil {
		rule.exprError(err, line, col)
		return nil, l.Offset(), false
	}
	t, ok := rule.checkSemanticsOfExprNode(expr, line, col, checkUntrusted, workflowKey)
	return t, l.Offset(), ok
}

func (rule *RuleExpression) calcNeedsType(job *Job) *ObjectType {
	// https://docs.github.com/en/actions/learn-github-actions/contexts#needs-context
	o := NewEmptyStrictObjectType()
	rule.populateDependantNeedsTypes(o, job, job)
	return o
}

func (rule *RuleExpression) populateDependantNeedsTypes(out *ObjectType, job *Job, root *Job) {
	for _, id := range job.Needs {
		i := strings.ToLower(id.Value) // ID is case insensitive
		if i == root.ID.Value {
			continue // When cyclic dependency exists. This does not happen normally.
		}
		if _, ok := out.Props[i]; ok {
			continue // Already added
		}

		j, ok := rule.workflow.Jobs[i]
		if !ok {
			continue
		}

		var outputs *ObjectType
		if j.WorkflowCall == nil {
			outputs = NewEmptyStrictObjectType()
			for name := range j.Outputs {
				outputs.Props[name] = StringType{}
			}
		} else {
			outputs = rule.getWorkflowCallOutputsType(j.WorkflowCall)
		}

		out.Props[i] = NewStrictObjectType(map[string]ExprType{
			"outputs": outputs,
			"result":  StringType{},
		})

		// Do not collect outputs type from parent of parent recursively. (#151)
	}
}

func (rule *RuleExpression) guessTypeOfMatrixExpression(expr *String) *ObjectType {
	ty := rule.checkObjectExpression(expr, "matrix", "jobs.<job_id>.strategy")
	if ty == nil {
		return NewEmptyObjectType()
	}
	matTy, ok := ty.(*ObjectType)
	if !ok {
		return NewEmptyObjectType()
	}

	// Consider properties in include section elements since 'include' section adds matrix values
	incTy, ok := matTy.Props["include"]
	if ok {
		delete(matTy.Props, "include")
		if a, ok := incTy.(*ArrayType); ok {
			if o, ok := a.Elem.(*ObjectType); ok {
				for n, p := range o.Props {
					t, ok := matTy.Props[n]
					if !ok {
						matTy.Props[n] = p
						continue
					}
					matTy.Props[n] = t.Merge(p)
				}
			}
		}
	}

	delete(matTy.Props, "exclude")

	return matTy
}

func (rule *RuleExpression) guessTypeOfMatrix(m *Matrix) *ObjectType {
	if m.Expression != nil {
		return rule.guessTypeOfMatrixExpression(m.Expression)
	}

	o := NewEmptyStrictObjectType()

	for n, r := range m.Rows {
		o.Props[n] = rule.guessTypeOfMatrixRow(r)
	}

	// Note: Type check in 'include' section duplicates with checkMatrixCombinations() method

	if m.Include == nil {
		return o
	}

	if m.Include.Expression != nil {
		if a, ok := rule.checkOneExpression(m.Include.Expression, "include", "jobs.<job_id>.strategy").(*ArrayType); ok {
			if ret, ok := o.Merge(a.Elem).(*ObjectType); ok {
				return ret
			}
		}
		return NewEmptyObjectType()
	}

	for _, combi := range m.Include.Combinations {
		if combi.Expression != nil {
			ty := rule.checkOneExpression(m.Include.Expression, "matrix combination at element of include section", "jobs.<job_id>.strategy")
			if ty == nil {
				continue
			}
			if merged, ok := o.Merge(ty).(*ObjectType); ok {
				o = merged
			} else {
				o.Loose()
			}
			continue
		}

		for n, assign := range combi.Assigns {
			ty := guessTypeOfRawYAMLValue(assign.Value)
			if t, ok := o.Props[n]; ok {
				// When the combination exists in 'matrix' section, merge type with existing one
				ty = t.Merge(ty)
			}
			o.Props[n] = ty
		}
	}

	// Note: m.Exclude is not considered when guessing type of matrix

	return o
}

func (rule *RuleExpression) guessTypeOfMatrixRow(r *MatrixRow) ExprType {
	if r.Expression != nil {
		if a, ok := rule.checkArrayExpression(r.Expression, "matrix row", "jobs.<job_id>.strategy").(*ArrayType); ok {
			return a
		}
		return AnyType{}
	}

	var ty ExprType
	for _, v := range r.Values {
		t := guessTypeOfRawYAMLValue(v)
		if ty == nil {
			ty = t
		} else {
			ty = ty.Merge(t)
		}
	}

	if ty == nil {
		return AnyType{} // No element
	}

	return ty
}

func (rule *RuleExpression) checkWorkflowCallOutputs(outputs map[string]*WorkflowCallEventOutput, jobs map[string]*Job) {
	if len(outputs) == 0 || len(jobs) == 0 {
		return
	}

	props := make(map[string]ExprType, len(jobs))
	for n, j := range jobs {
		var o *ObjectType
		if j.WorkflowCall != nil {
			// Outputs are not defined in jobs.<job_id> section when it is reusable workflow call.
			o = NewEmptyObjectType()
		} else {
			p := make(map[string]ExprType, len(j.Outputs))
			for n := range j.Outputs {
				p[n] = StringType{}
			}
			o = NewStrictObjectType(p)
		}
		props[n] = NewStrictObjectType(map[string]ExprType{
			"outputs": o,
		})
	}
	rule.jobsTy = NewStrictObjectType(props)

	for _, o := range outputs {
		rule.checkString(o.Value, "on.workflow_call.outputs.<output_id>.value")
	}
}

func guessTypeOfRawYAMLValue(v RawYAMLValue) ExprType {
	switch v := v.(type) {
	case *RawYAMLObject:
		m := make(map[string]ExprType, len(v.Props))
		for k, p := range v.Props {
			m[k] = guessTypeOfRawYAMLValue(p)
		}
		return NewStrictObjectType(m)
	case *RawYAMLArray:
		if len(v.Elems) == 0 {
			return &ArrayType{AnyType{}, false}
		}
		elem := guessTypeOfRawYAMLValue(v.Elems[0])
		for _, v := range v.Elems[1:] {
			elem = elem.Merge(guessTypeOfRawYAMLValue(v))
		}
		return &ArrayType{elem, false}
	case *RawYAMLString:
		return guessTypeFromString(v.Value)
	default:
		panic("unreachable")
	}
}

func guessTypeFromString(s string) ExprType {
	// Note that keywords are case sensitive. TRUE, FALSE, NULL are invalid named value.
	if s == "true" || s == "false" {
		return BoolType{}
	}
	if s == "null" {
		return NullType{}
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return NumberType{}
	}
	return StringType{}
}

func convertExprLineColToPos(line, col, lineBase, colBase int) *Pos {
	// Line and column in ExprError are 1-based
	return &Pos{
		Line: line - 1 + lineBase,
		Col:  col - 1 + colBase,
	}
}

func typeOfActionOutputs(meta *ActionMetadata) *ObjectType {
	// Some action sets outputs dynamically. Such outputs are not defined in action.yml. actionlint
	// cannot check such outputs statically so it allows any props (#18)
	if meta.SkipOutputs {
		return NewEmptyObjectType()
	}
	props := make(map[string]ExprType, len(meta.Outputs))
	for n := range meta.Outputs {
		props[strings.ToLower(n)] = StringType{}
	}
	return NewStrictObjectType(props)
}
