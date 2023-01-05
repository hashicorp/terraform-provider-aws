package actionlint

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func nodeKindName(k yaml.Kind) string {
	switch k {
	case yaml.DocumentNode:
		return "document"
	case yaml.SequenceNode:
		return "sequence"
	case yaml.MappingNode:
		return "mapping"
	case yaml.ScalarNode:
		return "scalar"
	case yaml.AliasNode:
		return "alias"
	default:
		if os.Getenv("ACTIONLINT_DEBUG") != "" {
			return "unknown"
		}
		panic("unreachable")
	}
}

func posAt(n *yaml.Node) *Pos {
	return &Pos{n.Line, n.Column}
}

func isNull(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.Tag == "!!null"
}

func newString(n *yaml.Node) *String {
	quoted := n.Style&(yaml.DoubleQuotedStyle|yaml.SingleQuotedStyle) != 0
	return &String{n.Value, quoted, posAt(n)}
}

type workflowKeyVal struct {
	id  string // Key in lower case since workflow keys are case-insensitive
	key *String
	val *yaml.Node
}

type parser struct {
	errors []*Error
}

func (p *parser) error(n *yaml.Node, m string) {
	p.errors = append(p.errors, &Error{m, "", n.Line, n.Column, "syntax-check"})
}

func (p *parser) errorAt(pos *Pos, m string) {
	p.errors = append(p.errors, &Error{m, "", pos.Line, pos.Col, "syntax-check"})
}

func (p *parser) errorfAt(pos *Pos, format string, args ...interface{}) {
	m := fmt.Sprintf(format, args...)
	p.errorAt(pos, m)
}

func (p *parser) errorf(n *yaml.Node, format string, args ...interface{}) {
	m := fmt.Sprintf(format, args...)
	p.error(n, m)
}

func (p *parser) unexpectedKey(s *String, sec string, expected []string) {
	l := len(expected)
	var m string
	if l == 1 {
		m = fmt.Sprintf("expected %q key for %q section but got %q", expected[0], sec, s.Value)
	} else if l > 1 {
		m = fmt.Sprintf("unexpected key %q for %q section. expected one of %v", s.Value, sec, sortedQuotes(expected))
	} else {
		m = fmt.Sprintf("unexpected key %q for %q section", s.Value, sec)
	}
	p.errorAt(s.Pos, m)
}

func (p *parser) checkNotEmpty(sec string, len int, n *yaml.Node) bool {
	if len == 0 {
		p.errorf(n, "%q section should not be empty", sec)
		return false
	}
	return true
}

func (p *parser) checkSequence(sec string, n *yaml.Node, allowEmpty bool) bool {
	if n.Kind != yaml.SequenceNode {
		p.errorf(n, "%q section must be sequence node but got %s node with %q tag", sec, nodeKindName(n.Kind), n.Tag)
		return false
	}
	return allowEmpty || p.checkNotEmpty(sec, len(n.Content), n)
}

func (p *parser) checkString(n *yaml.Node, allowEmpty bool) bool {
	// Do not check n.Tag is !!str because we don't need to check the node is string strictly.
	// In almost all cases, other nodes (like 42) are handled as string with its string representation.
	if n.Kind != yaml.ScalarNode {
		p.errorf(n, "expected scalar node for string value but found %s node with %q tag", nodeKindName(n.Kind), n.Tag)
		return false
	}
	if !allowEmpty && n.Value == "" {
		p.error(n, "string should not be empty")
		return false
	}
	return true
}

func (p *parser) missingExpression(n *yaml.Node, expecting string) {
	p.errorf(n, "expecting a single ${{...}} expression or %s, but found plain text node", expecting)
}

func (p *parser) parseExpression(n *yaml.Node, expecting string) *String {
	s := strings.TrimSpace(n.Value)
	if !strings.HasPrefix(s, "${{") || !strings.HasSuffix(s, "}}") {
		p.missingExpression(n, expecting)
		return nil
	}
	if strings.Count(n.Value, "${{") != 1 {
		p.missingExpression(n, expecting)
		return nil
	}
	return newString(n)
}

func (p *parser) mayParseExpression(n *yaml.Node) *String {
	if n.Tag != "!!str" {
		return nil
	}
	s := strings.TrimSpace(n.Value)
	if !strings.HasPrefix(s, "${{") || !strings.HasSuffix(s, "}}") || strings.Count(n.Value, "${{") != 1 {
		return nil
	}
	return newString(n)
}

func (p *parser) parseString(n *yaml.Node, allowEmpty bool) *String {
	if !p.checkString(n, allowEmpty) {
		return &String{"", false, posAt(n)}
	}
	return newString(n)
}

func (p *parser) parseStringSequence(sec string, n *yaml.Node, allowEmpty bool, allowElemEmpty bool) []*String {
	if ok := p.checkSequence(sec, n, allowEmpty); !ok {
		return nil
	}

	ss := make([]*String, 0, len(n.Content))
	for _, c := range n.Content {
		s := p.parseString(c, allowElemEmpty)
		if s != nil {
			ss = append(ss, s)
		}
	}
	return ss
}

func (p *parser) parseStringOrStringSequence(sec string, n *yaml.Node, allowEmpty bool, allowElemEmpty bool) []*String {
	switch n.Kind {
	case yaml.ScalarNode:
		if allowEmpty && n.Tag == "!!null" {
			return []*String{} // In the case of 'foo:'
		}
		return []*String{p.parseString(n, allowElemEmpty)}
	default:
		return p.parseStringSequence(sec, n, allowEmpty, allowElemEmpty)
	}
}

func (p *parser) parseBool(n *yaml.Node) *Bool {
	if n.Kind != yaml.ScalarNode || (n.Tag != "!!bool" && n.Tag != "!!str") {
		p.errorf(n, "expected bool value but found %s node with %q tag", nodeKindName(n.Kind), n.Tag)
		return nil
	}

	if n.Tag == "!!str" {
		e := p.parseExpression(n, "boolean literal \"true\" or \"false\"")
		return &Bool{
			Expression: e,
			Pos:        posAt(n),
		}
	}

	return &Bool{
		Value: n.Value == "true",
		Pos:   posAt(n),
	}
}

func (p *parser) parseInt(n *yaml.Node) *Int {
	if n.Kind != yaml.ScalarNode || (n.Tag != "!!int" && n.Tag != "!!str") {
		p.errorf(n, "expected scalar node for integer value but found %s node with %q tag", nodeKindName(n.Kind), n.Tag)
		return nil
	}

	if n.Tag == "!!str" {
		e := p.parseExpression(n, "integer literal")
		return &Int{
			Expression: e,
			Pos:        posAt(n),
		}
	}

	i, err := strconv.Atoi(n.Value)
	if err != nil {
		p.errorf(n, "invalid integer value: %q: %s", n.Value, err.Error())
		return nil
	}

	return &Int{
		Value: i,
		Pos:   posAt(n),
	}
}

func (p *parser) parseFloat(n *yaml.Node) *Float {
	if n.Kind != yaml.ScalarNode || (n.Tag != "!!float" && n.Tag != "!!int" && n.Tag != "!!str") {
		p.errorf(n, "expected scalar node for float value but found %s node with %q tag", nodeKindName(n.Kind), n.Tag)
		return nil
	}

	if n.Tag == "!!str" {
		e := p.parseExpression(n, "float number literal")
		return &Float{
			Expression: e,
			Pos:        posAt(n),
		}
	}

	f, err := strconv.ParseFloat(n.Value, 64)
	if err != nil {
		p.errorf(n, "invalid float value: %q: %s", n.Value, err.Error())
		return nil
	}

	return &Float{
		Value: f,
		Pos:   posAt(n),
	}
}

func (p *parser) parseMapping(what string, n *yaml.Node, allowEmpty bool) []workflowKeyVal {
	isNull := isNull(n)

	if !isNull && n.Kind != yaml.MappingNode {
		p.errorf(n, "%s is %s node but mapping node is expected", what, nodeKindName(n.Kind))
		return nil
	}

	if !allowEmpty && isNull {
		p.errorf(n, "%s should not be empty. please remove this section if it's unnecessary", what)
		return nil
	}

	l := len(n.Content) / 2
	keys := make(map[string]*Pos, l)
	m := make([]workflowKeyVal, 0, l)
	for i := 0; i < len(n.Content); i += 2 {
		k := p.parseString(n.Content[i], false)
		if k == nil {
			continue
		}

		// Keys of mappings are case insensitive. For example, following matrix is invalid.
		// matrix:
		//   foo: [1, 2, 3]
		//   FOO: [1, 2, 3]
		// To detect case insensitive duplicate keys, we use lowercase keys always
		id := strings.ToLower(k.Value)

		if pos, ok := keys[id]; ok {
			p.errorfAt(k.Pos, "key %q is duplicated in %s. previously defined at %s. note that key names are case insensitive", k.Value, what, pos.String())
			continue
		}
		m = append(m, workflowKeyVal{id, k, n.Content[i+1]})
		keys[id] = k.Pos
	}

	if !allowEmpty && len(m) == 0 {
		p.errorf(n, "%s should not be empty. please remove this section if it's unnecessary", what)
	}

	return m
}

func (p *parser) parseSectionMapping(sec string, n *yaml.Node, allowEmpty bool) []workflowKeyVal {
	return p.parseMapping(fmt.Sprintf("%q section", sec), n, allowEmpty)
}

func (p *parser) parseScheduleEvent(pos *Pos, n *yaml.Node) *ScheduledEvent {
	if ok := p.checkSequence("schedule", n, false); !ok {
		return nil
	}

	cron := make([]*String, 0, len(n.Content))
	for _, c := range n.Content {
		m := p.parseMapping("element of \"schedule\" section", c, false)
		if len(m) != 1 || m[0].id != "cron" {
			p.error(c, "element of \"schedule\" section must be mapping and must contain one key \"cron\"")
			continue
		}
		s := p.parseString(m[0].val, false)
		if s != nil {
			cron = append(cron, s)
		}
	}

	return &ScheduledEvent{cron, pos}
}

// https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#workflow_dispatch
func (p *parser) parseWorkflowDispatchEvent(pos *Pos, n *yaml.Node) *WorkflowDispatchEvent {
	ret := &WorkflowDispatchEvent{Pos: pos}

	for _, kv := range p.parseSectionMapping("workflow_dispatch", n, true) {
		if kv.id != "inputs" {
			p.unexpectedKey(kv.key, "workflow_dispatch", []string{"inputs"})
			continue
		}

		inputs := p.parseSectionMapping("inputs", kv.val, true)
		ret.Inputs = make(map[string]*DispatchInput, len(inputs))
		for _, input := range inputs {
			name, spec := input.key, input.val

			var desc *String
			var req *Bool
			var def *String
			var ty WorkflowDispatchEventInputType = WorkflowDispatchEventInputTypeNone
			var opts []*String

			for _, attr := range p.parseMapping("input settings of workflow_dispatch event", spec, true) {
				switch attr.id {
				case "description":
					desc = p.parseString(attr.val, true)
				case "required":
					req = p.parseBool(attr.val)
				case "default":
					def = p.parseString(attr.val, true)
				case "type":
					if !p.checkString(attr.val, false) {
						continue
					}
					switch attr.val.Value {
					case "string":
						ty = WorkflowDispatchEventInputTypeString
					case "boolean":
						ty = WorkflowDispatchEventInputTypeBoolean
					case "choice":
						ty = WorkflowDispatchEventInputTypeChoice
					case "environment":
						ty = WorkflowDispatchEventInputTypeEnvironment
					default:
						p.errorf(attr.val, "input type of workflow_dispatch event must be one of \"string\", \"boolean\", \"choice\", \"environment\" but got %q", attr.val.Value)
					}
				case "options":
					opts = p.parseStringSequence("options", attr.val, false, false)
				default:
					p.unexpectedKey(attr.key, "inputs", []string{"description", "required", "default"})
				}
			}

			ret.Inputs[input.id] = &DispatchInput{
				Name:        name,
				Description: desc,
				Required:    req,
				Default:     def,
				Type:        ty,
				Options:     opts,
			}
		}
	}

	return ret
}

// https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#repository_dispatch
func (p *parser) parseRepositoryDispatchEvent(pos *Pos, n *yaml.Node) *RepositoryDispatchEvent {
	ret := &RepositoryDispatchEvent{Pos: pos}

	// Note: Omitting 'types' is ok. In the case, all types trigger the workflow
	for _, kv := range p.parseSectionMapping("repository_dispatch", n, true) {
		if kv.id == "types" {
			ret.Types = p.parseStringOrStringSequence("types", kv.val, false, false)
		} else {
			p.unexpectedKey(kv.key, "repository_dispatch", []string{"types"})
		}
	}

	return ret
}

// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#using-filters
func (p *parser) parseWebhookEventFilter(name *String, n *yaml.Node) *WebhookEventFilter {
	v := p.parseStringOrStringSequence(name.Value, n, false, false)
	return &WebhookEventFilter{name, v}
}

func (p *parser) parseWebhookEvent(name *String, n *yaml.Node) *WebhookEvent {
	ret := &WebhookEvent{Hook: name, Pos: name.Pos}

	// Note: 'tags', 'tags-ignore', 'branches', 'branches-ignore' can be empty. Since there are
	// some cases where setting empty values to them is necessary.
	//
	// > If only define only tags filter (tags/tags-ignore) or only branches filter
	// > (branches/branches-ignore) for on.push, the workflow won’t run for events affecting the
	// > undefined Git ref.
	//
	// https://github.community/t/using-on-push-tags-ignore-and-paths-ignore-together/16931
	for _, kv := range p.parseSectionMapping(name.Value, n, true) {
		// Note: Glob pattern cannot be empty, but it is checked by 'glob' rule with better error
		// message. So parser allows empty patterns here.
		switch kv.id {
		case "types":
			ret.Types = p.parseStringOrStringSequence(kv.key.Value, kv.val, false, false)
		case "branches":
			ret.Branches = p.parseWebhookEventFilter(kv.key, kv.val)
		case "branches-ignore":
			ret.BranchesIgnore = p.parseWebhookEventFilter(kv.key, kv.val)
		case "tags":
			ret.Tags = p.parseWebhookEventFilter(kv.key, kv.val)
		case "tags-ignore":
			ret.TagsIgnore = p.parseWebhookEventFilter(kv.key, kv.val)
		case "paths":
			ret.Paths = p.parseWebhookEventFilter(kv.key, kv.val)
		case "paths-ignore":
			ret.PathsIgnore = p.parseWebhookEventFilter(kv.key, kv.val)
		case "workflows":
			ret.Workflows = p.parseStringOrStringSequence(kv.key.Value, kv.val, false, false)
		default:
			p.unexpectedKey(kv.key, name.Value, []string{
				"types",
				"branches",
				"branches-ignore",
				"tags",
				"tags-ignore",
				"paths",
				"paths-ignore",
				"workflows",
			})
		}
	}

	return ret
}

// - https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#workflow-reuse-events
// - https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#onworkflow_callinputs
// - https://docs.github.com/en/actions/learn-github-actions/reusing-workflows
func (p *parser) parseWorkflowCallEvent(pos *Pos, n *yaml.Node) *WorkflowCallEvent {
	ret := &WorkflowCallEvent{Pos: pos}

	for _, kv := range p.parseSectionMapping("workflow_call", n, true) {
		switch kv.id {
		case "inputs":
			inputs := p.parseSectionMapping("inputs", kv.val, true)
			ret.Inputs = make([]*WorkflowCallEventInput, 0, len(inputs))
			for _, kv := range inputs {
				name, spec := kv.key, kv.val
				input := &WorkflowCallEventInput{Name: name, ID: kv.id}
				sawType := false

				for _, attr := range p.parseMapping("input of workflow_call event", spec, true) {
					switch attr.id {
					case "description":
						input.Description = p.parseString(attr.val, true)
					case "required":
						input.Required = p.parseBool(attr.val)
					case "default":
						input.Default = p.parseString(attr.val, true)
					case "type":
						switch attr.val.Value {
						case "boolean":
							input.Type = WorkflowCallEventInputTypeBoolean
						case "number":
							input.Type = WorkflowCallEventInputTypeNumber
						case "string":
							input.Type = WorkflowCallEventInputTypeString
						default:
							p.errorf(attr.val, "invalid value %q for input type of workflow_call event. it must be one of \"boolean\", \"number\", or \"string\"", attr.val.Value)
						}
						sawType = true
					default:
						p.unexpectedKey(attr.key, "inputs at workflow_call event", []string{"description", "required", "default", "type"})
					}
				}

				if !sawType {
					p.errorfAt(name.Pos, "\"type\" is missing at %q input of workflow_call event", name.Value)
				}

				ret.Inputs = append(ret.Inputs, input)
			}
		case "secrets":
			secrets := p.parseSectionMapping("secrets", kv.val, true)
			ret.Secrets = make(map[string]*WorkflowCallEventSecret, len(secrets))
			for _, kv := range secrets {
				name, spec := kv.key, kv.val
				secret := &WorkflowCallEventSecret{Name: name}

				for _, attr := range p.parseMapping("secret of workflow_call event", spec, true) {
					switch attr.id {
					case "description":
						secret.Description = p.parseString(attr.val, true)
					case "required":
						secret.Required = p.parseBool(attr.val)
					default:
						p.unexpectedKey(attr.key, "secrets", []string{"description", "required"})
					}
				}

				ret.Secrets[kv.id] = secret
			}
		case "outputs":
			outputs := p.parseSectionMapping("outputs", kv.val, true)
			ret.Outputs = make(map[string]*WorkflowCallEventOutput, len(outputs))
			for _, kv := range outputs {
				name, spec := kv.key, kv.val
				output := &WorkflowCallEventOutput{Name: name}

				for _, attr := range p.parseMapping("output of workflow_call event", spec, true) {
					switch attr.id {
					case "description":
						output.Description = p.parseString(attr.val, true)
					case "value":
						output.Value = p.parseString(attr.val, false)
					default:
						p.unexpectedKey(attr.key, "outputs at workflow_call event", []string{"description", "value"})
					}
				}

				if output.Value == nil {
					p.errorfAt(name.Pos, "\"value\" is missing at %q output of workflow_call event", name.Value)
				}

				ret.Outputs[kv.id] = output
			}
		default:
			p.unexpectedKey(kv.key, "workflow_call", []string{"inputs", "secrets"})
		}
	}

	return ret
}

func (p *parser) parseEvents(pos *Pos, n *yaml.Node) []Event {
	switch n.Kind {
	case yaml.ScalarNode:
		switch n.Value {
		case "workflow_dispatch":
			return []Event{
				&WorkflowDispatchEvent{Pos: posAt(n)},
			}
		case "repository_dispatch":
			return []Event{
				&RepositoryDispatchEvent{Pos: posAt(n)},
			}
		case "schedule":
			p.errorAt(pos, "schedule event must be configured with mapping")
			return []Event{}
		case "workflow_call":
			return []Event{
				&WorkflowCallEvent{Pos: posAt(n)},
			}
		default:
			h := p.parseString(n, false)
			if h.Value == "" {
				return []Event{}
			}
			return []Event{
				&WebhookEvent{
					Hook: h,
					Pos:  posAt(n),
				},
			}
		}
	case yaml.MappingNode:
		kvs := p.parseSectionMapping("on", n, false)
		ret := make([]Event, 0, len(kvs))

		for _, kv := range kvs {
			pos := kv.key.Pos
			switch kv.id {
			case "schedule":
				if e := p.parseScheduleEvent(pos, kv.val); e != nil {
					ret = append(ret, e)
				}
			case "workflow_dispatch":
				ret = append(ret, p.parseWorkflowDispatchEvent(pos, kv.val))
			case "repository_dispatch":
				ret = append(ret, p.parseRepositoryDispatchEvent(pos, kv.val))
			case "workflow_call":
				ret = append(ret, p.parseWorkflowCallEvent(pos, kv.val))
			default:
				ret = append(ret, p.parseWebhookEvent(kv.key, kv.val))
			}
		}

		return ret
	case yaml.SequenceNode:
		l := len(n.Content)
		p.checkNotEmpty("on", l, n)
		ret := make([]Event, 0, l)

		for _, c := range n.Content {
			if s := p.parseString(c, false); s != nil {
				switch s.Value {
				case "schedule", "repository_dispatch":
					p.errorf(c, "%q event should not be listed in sequence. Use mapping for \"on\" section and configure the event as values of the mapping", s.Value)
				case "workflow_dispatch":
					ret = append(ret, &WorkflowDispatchEvent{Pos: posAt(c)})
				case "workflow_call":
					ret = append(ret, &WorkflowCallEvent{Pos: posAt(c)})
				default:
					ret = append(ret, &WebhookEvent{Hook: s, Pos: posAt(c)})
				}
			}
		}

		return ret
	default:
		p.errorf(n, "\"on\" section value is expected to be mapping or sequence but found %s node", nodeKindName(n.Kind))
		return nil
	}
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#permissions
func (p *parser) parsePermissions(pos *Pos, n *yaml.Node) *Permissions {
	ret := &Permissions{Pos: pos}

	if n.Kind == yaml.ScalarNode {
		ret.All = p.parseString(n, false)
	} else {
		m := p.parseSectionMapping("permissions", n, true)
		scopes := make(map[string]*PermissionScope, len(m))
		for _, kv := range m {
			scopes[kv.id] = &PermissionScope{
				Name:  kv.key,
				Value: p.parseString(kv.val, false),
			}
		}
		ret.Scopes = scopes
	}

	return ret
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#env
func (p *parser) parseEnv(n *yaml.Node) *Env {
	if n.Kind == yaml.ScalarNode {
		return &Env{
			Expression: p.parseExpression(n, "mapping value for \"env\" section"),
		}
	}

	m := p.parseMapping("env", n, false)
	vars := make(map[string]*EnvVar, len(m))

	for _, kv := range m {
		vars[kv.id] = &EnvVar{
			Name:  kv.key,
			Value: p.parseString(kv.val, true),
		}
	}

	return &Env{Vars: vars}
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#defaults
func (p *parser) parseDefaults(pos *Pos, n *yaml.Node) *Defaults {
	ret := &Defaults{Pos: pos}

	for _, kv := range p.parseSectionMapping("defaults", n, false) {
		if kv.id != "run" {
			p.unexpectedKey(kv.key, "defaults", []string{"run"})
			continue
		}
		ret.Run = &DefaultsRun{Pos: kv.key.Pos}

		for _, attr := range p.parseSectionMapping("run", kv.val, false) {
			switch attr.id {
			case "shell":
				ret.Run.Shell = p.parseString(attr.val, false)
			case "working-directory":
				ret.Run.WorkingDirectory = p.parseString(attr.val, false)
			default:
				p.unexpectedKey(attr.key, "run", []string{"shell", "working-directory"})
			}
		}
	}

	if ret.Run == nil {
		p.error(n, "\"defaults\" section should have \"run\" section")
	}

	return ret
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idconcurrency
func (p *parser) parseConcurrency(pos *Pos, n *yaml.Node) *Concurrency {
	ret := &Concurrency{Pos: pos}

	if n.Kind == yaml.ScalarNode {
		ret.Group = p.parseString(n, false)
	} else {
		groupFound := false
		for _, kv := range p.parseSectionMapping("concurrency", n, false) {
			switch kv.id {
			case "group":
				ret.Group = p.parseString(kv.val, false)
				groupFound = true
			case "cancel-in-progress":
				ret.CancelInProgress = p.parseBool(kv.val)
			default:
				p.unexpectedKey(kv.key, "concurrency", []string{"group", "cancel-in-progress"})
			}
		}
		if !groupFound {
			p.error(n, "group name is missing in \"concurrency\" section")
		}
	}

	return ret
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idenvironment
func (p *parser) parseEnvironment(pos *Pos, n *yaml.Node) *Environment {
	ret := &Environment{Pos: pos}

	if n.Kind == yaml.ScalarNode {
		ret.Name = p.parseString(n, false)
	} else {
		nameFound := false
		for _, kv := range p.parseSectionMapping("environment", n, false) {
			switch kv.id {
			case "name":
				ret.Name = p.parseString(kv.val, false)
				nameFound = true
			case "url":
				ret.URL = p.parseString(kv.val, false)
			default:
				p.unexpectedKey(kv.key, "environment", []string{"name", "url"})
			}
		}
		if !nameFound {
			p.error(n, "name is missing in \"environment\" section")
		}
	}

	return ret
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idoutputs
func (p *parser) parseOutputs(n *yaml.Node) map[string]*Output {
	outputs := p.parseSectionMapping("outputs", n, false)
	ret := make(map[string]*Output, len(outputs))
	for _, output := range outputs {
		ret[output.id] = &Output{
			Name:  output.key,
			Value: p.parseString(output.val, true),
		}
	}
	p.checkNotEmpty("outputs", len(ret), n)
	return ret
}

func (p *parser) parseRawYAMLValue(n *yaml.Node) RawYAMLValue {
	switch n.Kind {
	case yaml.ScalarNode:
		return &RawYAMLString{n.Value, posAt(n)}
	case yaml.SequenceNode:
		vs := make([]RawYAMLValue, 0, len(n.Content))
		for _, c := range n.Content {
			if v := p.parseRawYAMLValue(c); v != nil {
				vs = append(vs, v)
			}
		}
		return &RawYAMLArray{vs, posAt(n)}
	case yaml.MappingNode:
		parsed := p.parseMapping("matrix row value", n, true)
		m := make(map[string]RawYAMLValue, len(parsed))
		for _, kv := range parsed {
			if v := p.parseRawYAMLValue(kv.val); v != nil {
				m[kv.id] = v
			}
		}
		return &RawYAMLObject{m, posAt(n)}
	default:
		p.errorf(n, "unexpected %s node on parsing value in matrix row", nodeKindName(n.Kind))
		return nil
	}
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#example-including-additional-values-into-combinations
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#example-excluding-configurations-from-a-matrix
func (p *parser) parseMatrixCombinations(sec string, n *yaml.Node) *MatrixCombinations {
	if n.Kind == yaml.ScalarNode {
		return &MatrixCombinations{
			Expression: p.parseExpression(n, "array of matrix combination"),
		}
	}

	if ok := p.checkSequence(sec, n, false); !ok {
		return nil
	}

	ret := make([]*MatrixCombination, 0, len(n.Content))
	for _, c := range n.Content {
		if c.Kind == yaml.ScalarNode {
			if e := p.parseExpression(c, "mapping of matrix combination"); e != nil {
				ret = append(ret, &MatrixCombination{Expression: e})
			}
			continue
		}

		kvs := p.parseMapping(fmt.Sprintf("element in %q section", sec), c, false)
		assigns := make(map[string]*MatrixAssign, len(kvs))
		for _, kv := range kvs {
			if v := p.parseRawYAMLValue(kv.val); v != nil {
				assigns[kv.id] = &MatrixAssign{kv.key, v}
			}
		}
		ret = append(ret, &MatrixCombination{Assigns: assigns})
	}
	return &MatrixCombinations{Combinations: ret}
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix
func (p *parser) parseMatrix(pos *Pos, n *yaml.Node) *Matrix {
	if n.Kind == yaml.ScalarNode {
		return &Matrix{
			Expression: p.parseExpression(n, "matrix"),
			Pos:        posAt(n),
		}
	}

	ret := &Matrix{Pos: pos, Rows: make(map[string]*MatrixRow)}

	for _, kv := range p.parseSectionMapping("matrix", n, false) {
		switch kv.id {
		case "include":
			ret.Include = p.parseMatrixCombinations("include", kv.val)
		case "exclude":
			ret.Exclude = p.parseMatrixCombinations("exclude", kv.val)
		default:
			if kv.val.Kind == yaml.ScalarNode {
				ret.Rows[kv.id] = &MatrixRow{
					Expression: p.parseExpression(kv.val, "array value for matrix variations"),
				}
				continue
			}

			if ok := p.checkSequence("matrix values", kv.val, false); !ok {
				continue
			}

			values := make([]RawYAMLValue, 0, len(kv.val.Content))
			for _, c := range kv.val.Content {
				if v := p.parseRawYAMLValue(c); v != nil {
					values = append(values, v)
				}
			}

			ret.Rows[kv.id] = &MatrixRow{
				Name:   kv.key,
				Values: values,
			}
		}
	}

	return ret
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategy
func (p *parser) parseStrategy(pos *Pos, n *yaml.Node) *Strategy {
	ret := &Strategy{Pos: pos}

	for _, kv := range p.parseSectionMapping("strategy", n, false) {
		switch kv.id {
		case "matrix":
			ret.Matrix = p.parseMatrix(kv.key.Pos, kv.val)
		case "fail-fast":
			ret.FailFast = p.parseBool(kv.val)
		case "max-parallel":
			ret.MaxParallel = p.parseInt(kv.val)
		default:
			p.unexpectedKey(kv.key, "strategy", []string{"matrix", "fail-fast", "max-parallel"})
		}
	}

	return ret
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idcontainer
func (p *parser) parseContainer(sec string, pos *Pos, n *yaml.Node) *Container {
	ret := &Container{Pos: pos}

	if n.Kind == yaml.ScalarNode {
		// When you only specify a container image, you can omit the image keyword.
		ret.Image = p.parseString(n, false)
	} else {
		for _, kv := range p.parseSectionMapping(sec, n, false) {
			switch kv.id {
			case "image":
				ret.Image = p.parseString(kv.val, false)
			case "credentials":
				cred := &Credentials{Pos: kv.key.Pos}
				for _, c := range p.parseSectionMapping("credentials", kv.val, false) {
					switch c.id {
					case "username":
						cred.Username = p.parseString(c.val, false)
					case "password":
						cred.Password = p.parseString(c.val, false)
					default:
						p.unexpectedKey(c.key, "credentials", []string{"username", "password"})
					}
				}
				if cred.Username == nil || cred.Password == nil {
					p.errorAt(kv.key.Pos, "both \"username\" and \"password\" must be specified in \"credentials\" section")
					continue
				}
				ret.Credentials = cred
			case "env":
				ret.Env = p.parseEnv(kv.val)
			case "ports":
				ret.Ports = p.parseStringSequence("ports", kv.val, true, false)
			case "volumes":
				ret.Ports = p.parseStringSequence("volumes", kv.val, true, false)
			case "options":
				ret.Options = p.parseString(kv.val, true)
			default:
				p.unexpectedKey(kv.key, sec, []string{
					"image",
					"credentials",
					"env",
					"ports",
					"volumes",
					"options",
				})
			}
		}
	}

	return ret
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idsteps
func (p *parser) parseStep(n *yaml.Node) *Step {
	ret := &Step{Pos: posAt(n)}
	var workDir *String

	for _, kv := range p.parseMapping("element of \"steps\" section", n, false) {
		switch kv.id {
		case "id":
			ret.ID = p.parseString(kv.val, false)
		case "if":
			ret.If = p.parseString(kv.val, false)
		case "name":
			ret.Name = p.parseString(kv.val, true)
		case "env":
			ret.Env = p.parseEnv(kv.val)
		case "continue-on-error":
			ret.ContinueOnError = p.parseBool(kv.val)
		case "timeout-minutes":
			ret.TimeoutMinutes = p.parseFloat(kv.val)
		case "uses", "with":
			var exec *ExecAction
			if ret.Exec == nil {
				exec = &ExecAction{}
			} else if e, ok := ret.Exec.(*ExecAction); ok {
				exec = e
			} else {
				p.errorfAt(kv.key.Pos, "this step is for running shell command since it contains at least one of \"run\", \"shell\" keys, but also contains %q key which is used for running action", kv.key.Value)
				continue
			}
			if kv.id == "uses" {
				exec.Uses = p.parseString(kv.val, false)
			} else {
				// kv.key == "with"
				with := p.parseSectionMapping("with", kv.val, false)
				exec.Inputs = make(map[string]*Input, len(with))
				for _, input := range with {
					switch input.id {
					case "entrypoint":
						// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepswithentrypoint
						exec.Entrypoint = p.parseString(input.val, false)
					case "args":
						// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepswithargs
						exec.Args = p.parseString(input.val, true)
					default:
						exec.Inputs[input.id] = &Input{input.key, p.parseString(input.val, true)}
					}
				}
			}
			ret.Exec = exec
		case "run", "shell":
			var exec *ExecRun
			if ret.Exec == nil {
				exec = &ExecRun{}
			} else if e, ok := ret.Exec.(*ExecRun); ok {
				exec = e
			} else {
				p.errorfAt(kv.key.Pos, "this step is for running action since it contains at least one of \"uses\", \"with\" keys, but also contains %q key which is used for running shell command", kv.key.Value)
				continue
			}
			switch kv.id {
			case "run":
				exec.Run = p.parseString(kv.val, false)
				exec.RunPos = kv.key.Pos
			case "shell":
				exec.Shell = p.parseString(kv.val, false)
			}
			exec.WorkingDirectory = workDir
			ret.Exec = exec
		case "working-directory":
			workDir = p.parseString(kv.val, false)
			if e, ok := ret.Exec.(*ExecRun); ok {
				e.WorkingDirectory = workDir
			}
		default:
			p.unexpectedKey(kv.key, "step", []string{
				"id",
				"if",
				"name",
				"env",
				"continue-on-error",
				"timeout-minutes",
				"uses",
				"with",
				"run",
				"working-directory",
				"shell",
			})
		}
	}

	switch e := ret.Exec.(type) {
	case *ExecAction:
		if e.Uses == nil {
			p.error(n, "\"uses\" is required to run action in step")
		}
		if workDir != nil {
			p.errorAt(workDir.Pos, "\"working-directory\" is not available with \"uses\". it is only available with \"run\"")
		}
	case *ExecRun:
		if e.Run == nil {
			p.error(n, "\"run\" is required to run script in step")
		}
	default:
		p.error(n, "step must run script with \"run\" section or run action with \"uses\" section")
	}

	return ret
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idsteps
func (p *parser) parseSteps(n *yaml.Node) []*Step {
	if ok := p.checkSequence("steps", n, false); !ok {
		return nil
	}

	ret := make([]*Step, 0, len(n.Content))

	for _, c := range n.Content {
		if s := p.parseStep(c); s != nil {
			ret = append(ret, s)
		}
	}

	return ret
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_id
func (p *parser) parseJob(id *String, n *yaml.Node) *Job {
	ret := &Job{ID: id, Pos: id.Pos}
	call := &WorkflowCall{}

	// Only below keys are allowed on reusable workflow call
	// https://docs.github.com/en/actions/learn-github-actions/reusing-workflows#supported-keywords-for-jobs-that-call-a-reusable-workflow
	//   - jobs.<job_id>.name
	//   - jobs.<job_id>.uses
	//   - jobs.<job_id>.with
	//   - jobs.<job_id>.with.<input_id>
	//   - jobs.<job_id>.secrets
	//   - jobs.<job_id>.secrets.<secret_id>
	//   - jobs.<job_id>.needs
	//   - jobs.<job_id>.if
	//   - jobs.<job_id>.permissions

	// https://docs.github.com/en/actions/using-workflows/reusing-workflows#supported-keywords-for-jobs-that-call-a-reusable-workflow
	var stepsOnlyKey *String
	var callOnlyKey *String

	for _, kv := range p.parseMapping(fmt.Sprintf("%q job", id.Value), n, false) {
		k, v := kv.key, kv.val
		switch kv.id {
		case "name":
			ret.Name = p.parseString(v, true)
		case "needs":
			if v.Kind == yaml.ScalarNode {
				// needs: job1
				ret.Needs = []*String{p.parseString(v, false)}
			} else {
				// needs: [job1, job2]
				ret.Needs = p.parseStringSequence("needs", v, false, false)
			}
		case "runs-on":
			if expr := p.mayParseExpression(v); expr != nil {
				ret.RunsOn = &Runner{nil, expr}
			} else {
				labels := p.parseStringOrStringSequence("runs-on", v, false, false)
				ret.RunsOn = &Runner{labels, nil}
			}
			stepsOnlyKey = k
		case "permissions":
			ret.Permissions = p.parsePermissions(k.Pos, v)
		case "environment":
			ret.Environment = p.parseEnvironment(k.Pos, v)
			stepsOnlyKey = k
		case "concurrency":
			ret.Concurrency = p.parseConcurrency(k.Pos, v)
		case "outputs":
			ret.Outputs = p.parseOutputs(v)
			stepsOnlyKey = k
		case "env":
			ret.Env = p.parseEnv(v)
			stepsOnlyKey = k
		case "defaults":
			ret.Defaults = p.parseDefaults(k.Pos, v)
			stepsOnlyKey = k
		case "if":
			ret.If = p.parseString(v, false)
		case "steps":
			ret.Steps = p.parseSteps(v)
			stepsOnlyKey = k
		case "timeout-minutes":
			ret.TimeoutMinutes = p.parseFloat(v)
			stepsOnlyKey = k
		case "strategy":
			ret.Strategy = p.parseStrategy(k.Pos, v)
		case "continue-on-error":
			ret.ContinueOnError = p.parseBool(v)
			stepsOnlyKey = k
		case "container":
			ret.Container = p.parseContainer("container", k.Pos, v)
			stepsOnlyKey = k
		case "services":
			services := p.parseSectionMapping("services", v, false)
			ret.Services = make(map[string]*Service, len(services))
			for _, s := range services {
				ret.Services[s.id] = &Service{
					Name:      s.key,
					Container: p.parseContainer("services", s.key.Pos, s.val),
				}
			}
		case "uses":
			call.Uses = p.parseString(v, false)
			callOnlyKey = k
		case "with":
			with := p.parseSectionMapping("with", v, false)
			call.Inputs = make(map[string]*WorkflowCallInput, len(with))
			for _, i := range with {
				call.Inputs[i.id] = &WorkflowCallInput{
					Name:  i.key,
					Value: p.parseString(i.val, true),
				}
			}
			callOnlyKey = k
		case "secrets":
			if kv.val.Kind == yaml.ScalarNode {
				// `secrets: inherit` special case
				// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#onworkflow_callsecretsinherit
				if kv.val.Value == "inherit" {
					call.InheritSecrets = true
				} else {
					p.errorf(kv.val, "expected mapping node for secrets or \"inherit\" string node but found %q node", kv.val.Value)
				}
			} else {
				secrets := p.parseSectionMapping("secrets", v, false)
				call.Secrets = make(map[string]*WorkflowCallSecret, len(secrets))
				for _, s := range secrets {
					call.Secrets[s.id] = &WorkflowCallSecret{
						Name:  s.key,
						Value: p.parseString(s.val, true),
					}
				}
			}
			callOnlyKey = k
		default:
			p.unexpectedKey(kv.key, "job", []string{
				"name",
				"needs",
				"runs-on",
				"permissions",
				"environment",
				"concurrency",
				"outputs",
				"env",
				"defaults",
				"if",
				"steps",
				"timeout-minutes",
				"strategy",
				"continue-on-error",
				"container",
				"services",
				"uses",
				"with",
				"secrets",
			})
		}
	}

	if call.Uses != nil {
		if stepsOnlyKey != nil {
			p.errorfAt(
				stepsOnlyKey.Pos,
				"when a reusable workflow is called with \"uses\", %q is not available. only following keys are allowed: \"name\", \"uses\", \"with\", \"secrets\", \"needs\", \"if\", and \"permissions\" in job %q",
				stepsOnlyKey.Value,
				id.Value,
			)
		} else {
			ret.WorkflowCall = call
		}
	} else {
		// When not a reusable call
		if ret.Steps == nil {
			p.errorfAt(id.Pos, "\"steps\" section is missing in job %q", id.Value)
		}
		if ret.RunsOn == nil {
			p.errorfAt(id.Pos, "\"runs-on\" section is missing in job %q", id.Value)
		}
		if callOnlyKey != nil {
			p.errorfAt(
				callOnlyKey.Pos,
				"%q is only available for a reusable workflow call with \"uses\" but \"uses\" is not found in job %q",
				callOnlyKey.Value,
				id.Value,
			)
		}
	}

	return ret
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobs
func (p *parser) parseJobs(n *yaml.Node) map[string]*Job {
	jobs := p.parseSectionMapping("jobs", n, false)
	ret := make(map[string]*Job, len(jobs))
	for _, kv := range jobs {
		ret[kv.id] = p.parseJob(kv.key, kv.val)
	}
	return ret
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions
func (p *parser) parse(n *yaml.Node) *Workflow {
	w := &Workflow{}

	if n.Line == 0 {
		n.Line = 1
	}
	if n.Column == 0 {
		n.Column = 1
	}

	if len(n.Content) == 0 {
		p.error(n, "workflow is empty")
		return w
	}

	for _, kv := range p.parseMapping("workflow", n.Content[0], false) {
		k, v := kv.key, kv.val
		switch kv.id {
		case "name":
			w.Name = p.parseString(v, true)
		case "on":
			w.On = p.parseEvents(k.Pos, v)
		case "permissions":
			w.Permissions = p.parsePermissions(k.Pos, v)
		case "env":
			w.Env = p.parseEnv(v)
		case "defaults":
			w.Defaults = p.parseDefaults(k.Pos, v)
		case "concurrency":
			w.Concurrency = p.parseConcurrency(k.Pos, v)
		case "jobs":
			w.Jobs = p.parseJobs(v)
		case "run-name":
			w.RunName = p.parseString(v, false)
		default:
			p.unexpectedKey(k, "workflow", []string{
				"name",
				"run-name",
				"on",
				"permissions",
				"env",
				"defaults",
				"concurrency",
				"jobs",
			})
		}
	}

	if w.On == nil {
		p.error(n, "\"on\" section is missing in workflow")
	}
	if w.Jobs == nil {
		p.error(n, "\"jobs\" section is missing in workflow")
	}

	return w
}

// func dumpYAML(n *yaml.Node, level int) {
// 	fmt.Printf("%s%s (%s, %d,%d): %q\n", strings.Repeat(". ", level), nodeKindName(n.Kind), n.Tag, n.Line, n.Column, n.Value)
// 	for _, c := range n.Content {
// 		dumpYAML(c, level+1)
// 	}
// }

func handleYAMLError(err error) []*Error {
	re := regexp.MustCompile(`\bline (\d+):`)

	yamlErr := func(msg string) *Error {
		l := 0
		if ss := re.FindStringSubmatch(msg); len(ss) > 1 {
			l, _ = strconv.Atoi(ss[1])
		}
		msg = fmt.Sprintf("could not parse as YAML: %s", msg)
		return &Error{msg, "", l, 0, "yaml-syntax"}
	}

	if te, ok := err.(*yaml.TypeError); ok {
		errs := make([]*Error, 0, len(te.Errors))
		for _, msg := range te.Errors {
			errs = append(errs, yamlErr(msg))
		}
		return errs
	}

	return []*Error{yamlErr(err.Error())}
}

// Parse parses given source as byte sequence into workflow syntax tree. It returns all errors
// detected while parsing the input. It means that detecting one error does not stop parsing. Even
// if one or more errors are detected, parser will try to continue parsing and finding more errors.
func Parse(b []byte) (*Workflow, []*Error) {
	var n yaml.Node

	if err := yaml.Unmarshal(b, &n); err != nil {
		return nil, handleYAMLError(err)
	}

	// Uncomment for checking YAML tree
	// dumpYAML(&n, 0)

	p := &parser{}
	w := p.parse(&n)

	return w, p.errors
}
