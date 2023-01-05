package actionlint

import (
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron"
)

//go:generate go run ./scripts/generate-webhook-events ./all_webhooks.go

// RuleEvents is a rule to check 'on' field in workflow.
// https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows
type RuleEvents struct {
	RuleBase
}

// NewRuleEvents creates new RuleEvents instance.
func NewRuleEvents() *RuleEvents {
	return &RuleEvents{
		RuleBase: RuleBase{name: "events"},
	}
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RuleEvents) VisitWorkflowPre(n *Workflow) error {
	for _, e := range n.On {
		rule.checkEvent(e)
	}
	return nil
}

func (rule *RuleEvents) checkEvent(event Event) {
	switch e := event.(type) {
	case *ScheduledEvent:
		for _, c := range e.Cron {
			rule.checkCron(c)
		}
	case *WorkflowDispatchEvent:
		rule.checkWorkflowDispatchEvent(e)
	case *RepositoryDispatchEvent:
		// Nothing to do
	case *WorkflowCallEvent:
		rule.checkWorkflowCallEvent(e)
	case *WebhookEvent:
		rule.checkWebhookEvent(e)
	default:
		panic("unreachable")
	}
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#onschedule
func (rule *RuleEvents) checkCron(spec *String) {
	p := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := p.Parse(spec.Value)
	if err != nil {
		rule.errorf(spec.Pos, "invalid CRON format %q in schedule event: %s", spec.Value, err.Error())
		return
	}

	start := sched.Next(time.Unix(0, 0))
	next := sched.Next(start)
	diff := next.Sub(start).Seconds()

	// (#14) https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#scheduled-events
	//
	// > The shortest interval you can run scheduled workflows is once every 5 minutes.
	if diff < 60.0*5 {
		rule.errorf(spec.Pos, "scheduled job runs too frequently. it runs once per %g seconds. the shortest interval is once every 5 minutes", diff)
	}
}

func (rule *RuleEvents) filterNotAvailable(pos *Pos, filter, hook string, available []string) {
	e := "events"
	if len(available) < 2 {
		e = "event"
	}
	rule.errorf(pos, "%q filter is not available for %s event. it is only for %s %s", filter, hook, strings.Join(available, ", "), e)
}

func (rule *RuleEvents) checkExclusiveFilters(filter, ignore *WebhookEventFilter, hook string, available []string) {
	ok := false
	for _, a := range available {
		if a == hook {
			ok = true
			break
		}
	}

	if ok {
		if !filter.IsEmpty() && !ignore.IsEmpty() {
			p := filter.Name.Pos
			if p.IsBefore(ignore.Name.Pos) {
				p = ignore.Name.Pos
			}
			rule.errorf(p, "both %q and %q filters cannot be used for the same event %q. note: use '!' to negate patterns", filter.Name.Value, ignore.Name.Value, hook)
		}
	} else {
		if !filter.IsEmpty() {
			rule.filterNotAvailable(filter.Name.Pos, filter.Name.Value, hook, available)
		}
		if !ignore.IsEmpty() {
			rule.filterNotAvailable(ignore.Name.Pos, ignore.Name.Value, hook, available)
		}
	}
}

// https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#webhook-events
func (rule *RuleEvents) checkWebhookEvent(event *WebhookEvent) {
	hook := event.Hook.Value

	types, ok := AllWebhookTypes[hook]
	if !ok {
		rule.errorf(event.Pos, "unknown Webhook event %q. see https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#webhook-events for list of all Webhook event names", hook)
		return
	}

	rule.checkTypes(event.Hook, event.Types, types)

	if hook == "workflow_run" {
		if len(event.Workflows) == 0 {
			rule.error(event.Pos, "no workflow is configured for \"workflow_run\" event")
		}
	} else {
		if len(event.Workflows) != 0 {
			rule.errorf(event.Pos, "\"workflows\" cannot be configured for %q event. it is only for workflow_run event", hook)
		}
	}

	// Some filters are available with specific events and exclusive
	// - on.<push|pull_request|pull_request_target>.<paths|paths-ignore>
	// - on.push.<branches|tags|branches-ignore|tags-ignore>
	// - on.<pull_request|pull_request_target>.<branches|branches-ignore>
	// - on.workflow_run.<branches|branches-ignore>
	rule.checkExclusiveFilters(
		event.Paths,
		event.PathsIgnore,
		hook,
		[]string{"push", "pull_request", "pull_request_target"},
	)
	rule.checkExclusiveFilters(
		event.Branches,
		event.BranchesIgnore,
		hook,
		[]string{"push", "pull_request", "pull_request_target", "workflow_run"},
	)
	rule.checkExclusiveFilters(
		event.Tags,
		event.TagsIgnore,
		hook,
		[]string{"push"},
	)
}

func (rule *RuleEvents) checkTypes(hook *String, types []*String, expected []string) {
	if len(expected) == 0 && len(types) > 0 {
		rule.errorf(hook.Pos, "\"types\" cannot be specified for %q Webhook event", hook.Value)
		return
	}

	for _, ty := range types {
		valid := false
		for _, e := range expected {
			if ty.Value == e {
				valid = true
				break
			}
		}
		if !valid {
			rule.errorf(
				ty.Pos,
				"invalid activity type %q for %q Webhook event. available types are %s",
				ty.Value,
				hook.Value,
				sortedQuotes(expected),
			)
		}
	}
}

// https://docs.github.com/en/actions/learn-github-actions/reusing-workflows
func (rule *RuleEvents) checkWorkflowCallEvent(event *WorkflowCallEvent) {
	for _, i := range event.Inputs {
		if i.Default == nil {
			continue
		}
		// ${{ }} is available in the default value
		if !strings.Contains(i.Default.Value, "${{") {
			switch i.Type {
			case WorkflowCallEventInputTypeNumber:
				if _, err := strconv.ParseFloat(i.Default.Value, 64); err != nil {
					rule.errorf(
						i.Default.Pos,
						"input of workflow_call event %q is typed as number but its default value %q cannot be parsed as a float number: %s",
						i.Name.Value,
						i.Default.Value,
						err,
					)
				}
			case WorkflowCallEventInputTypeBoolean:
				if d := strings.ToLower(i.Default.Value); d != "true" && d != "false" {
					rule.errorf(
						i.Default.Pos,
						"input of workflow_call event %q is typed as boolean. its default value must be true or false but got %q",
						i.Name.Value,
						i.Default.Value,
					)
				}
			}
		}
		if i.IsRequired() {
			rule.errorf(
				i.Default.Pos,
				"input %q of workflow_call event has the default value %q, but it is also required. if an input is marked as required, its default value will never be used",
				i.Name.Value,
				i.Default.Value,
			)
		}
	}
}

// https://github.blog/changelog/2021-11-10-github-actions-input-types-for-manual-workflows/
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#onworkflow_dispatchinputs
func (rule *RuleEvents) checkWorkflowDispatchEvent(event *WorkflowDispatchEvent) {
	for n, i := range event.Inputs {
		if i.Type == WorkflowDispatchEventInputTypeChoice {
			if len(i.Options) == 0 {
				rule.errorf(i.Name.Pos, "input type of %q is \"choice\" but \"options\" is not set", n)
				continue
			}
			seen := make(map[string]struct{}, len(i.Options))
			for _, o := range i.Options {
				if _, ok := seen[o.Value]; ok {
					rule.errorf(o.Pos, "option %q is duplicated in options of %q input", o.Value, n)
					continue
				}
				seen[o.Value] = struct{}{}
			}
			if i.Default != nil {
				var b quotesBuilder
				for _, o := range i.Options {
					b.append(o.Value)
				}
				if _, ok := seen[i.Default.Value]; !ok {
					rule.errorf(i.Default.Pos, "default value %q of %q input is not included in its options %q", i.Default.Value, n, b.build())
				}
			}
		} else {
			if len(i.Options) > 0 {
				rule.errorf(i.Name.Pos, "\"options\" can not be set to %q input because its input type is not \"choice\"", n)
			}
			switch i.Type {
			case WorkflowDispatchEventInputTypeBoolean:
				if i.Default != nil {
					if d := strings.ToLower(i.Default.Value); d != "true" && d != "false" {
						rule.errorf(i.Default.Pos, "type of %q input is \"boolean\". its default value %q must be \"true\" or \"false\"", n, i.Default.Value)
					}
				}
			default:
				// TODO: Can some check be done for WorkflowDispatchEventInputTypeEnvironment?
				// What is suitable for default value of the type? (Or is a default value never suitable?)
			}
		}
	}
}
