package actionlint

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Pos represents position in the file.
type Pos struct {
	// Line is a line number of the position. This value is 1-based.
	Line int
	// Col is a column number of the position. This value is 1-based.
	Col int
}

func (p *Pos) String() string {
	return fmt.Sprintf("line:%d,col:%d", p.Line, p.Col)
}

// IsBefore returns if the position is before the other position. If they are equal, this function returns false.
func (p *Pos) IsBefore(other *Pos) bool {
	if p.Line < other.Line {
		return true
	}
	if p.Line > other.Line {
		return false
	}
	return p.Col < other.Col
}

// String represents generic string value in YAML file with position.
type String struct {
	// Value is a raw value of the string.
	Value string
	// Quoted represents the string is quoted with ' or " in the YAML source.
	Quoted bool
	// Pos is a position of the string in source.
	Pos *Pos
}

// Bool represents generic boolean value in YAML file with position.
type Bool struct {
	// Value is a raw value of the bool string.
	Value bool
	// Expression is a string when expression syntax ${{ }} is used for this section.
	Expression *String
	// Pos is a position in source.
	Pos *Pos
}

func (b *Bool) String() string {
	if b.Expression != nil {
		return b.Expression.Value
	}
	if b.Value {
		return "true"
	}
	return "false"
}

// Int represents generic integer value in YAML file with position.
type Int struct {
	// Value is a raw value of the integer string.
	Value int
	// Expression is a string when expression syntax ${{ }} is used for this section.
	Expression *String
	// Pos is a position in source.
	Pos *Pos
}

// Float represents generic float value in YAML file with position.
type Float struct {
	// Value is a raw value of the float string.
	Value float64
	// Expression is a string when expression syntax ${{ }} is used for this section.
	Expression *String
	// Pos is a position in source.
	Pos *Pos
}

// Event interface represents workflow events in 'on' section
type Event interface {
	// EventName returns name of the event to trigger this workflow.
	EventName() string
}

// WebhookEventFilter is a filter for Webhook events such as 'branches', 'paths-ignore', ...
// Webhook events are filtered by those filters. Some filters are exclusive.
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#using-filters
type WebhookEventFilter struct {
	// Name is a name of filter such like 'branches', 'tags'
	Name *String
	// Values is a list of filter values.
	Values []*String
}

// IsEmpty returns true when it has no value. This may mean the WebhookEventFilter instance itself is nil.
func (f *WebhookEventFilter) IsEmpty() bool {
	return f == nil || len(f.Values) == 0
}

// WebhookEvent represents event type based on webhook events.
// Some events can't have 'types' field. Only 'push' and 'pull' events can have 'tags', 'tags-ignore',
// 'paths' and 'paths-ignore' fields. Only 'workflow_run' event can have 'workflows' field.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#onevent_nametypes
type WebhookEvent struct {
	// Hook is a name of the webhook event.
	Hook *String
	// Types is list of types of the webhook event. Only the types enumerated here will trigger
	// the workflow.
	Types []*String
	// Branches is 'branches' filter. This value is nil when it is omitted.
	Branches *WebhookEventFilter
	// BranchesIgnore is 'branches-ignore' filter. This value is nil when it is omitted.
	BranchesIgnore *WebhookEventFilter
	// Tags is 'tags' filter. This value is nil when it is omitted.
	Tags *WebhookEventFilter
	// TagsIgnore is 'tags-ignore' filter. This value is nil when it is omitted.
	TagsIgnore *WebhookEventFilter
	// Paths is 'paths' filter. This value is nil when it is omitted.
	Paths *WebhookEventFilter
	// PathsIgnore is 'paths-ignore' filter. This value is nil when it is omitted.
	PathsIgnore *WebhookEventFilter
	// Workflows is list of workflow names which are triggered by 'workflow_run' event.
	Workflows []*String
	// Pos is a position in source.
	Pos *Pos
}

// EventName returns name of the event to trigger this workflow.
func (e *WebhookEvent) EventName() string {
	return e.Hook.Value
}

// ScheduledEvent is event scheduled by workflow.
// https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#scheduled-events
type ScheduledEvent struct {
	// Cron is list of cron strings which schedules workflow.
	Cron []*String
	// Pos is a position in source.
	Pos *Pos
}

// EventName returns name of the event to trigger this workflow.
func (e *ScheduledEvent) EventName() string {
	return "schedule"
}

// WorkflowDispatchEventInputType is a type for input types of workflow_dispatch events.
// https://github.blog/changelog/2021-11-10-github-actions-input-types-for-manual-workflows/
type WorkflowDispatchEventInputType uint8

const (
	// WorkflowDispatchEventInputTypeNone represents no type is specified to the input of workflow_dispatch event.
	WorkflowDispatchEventInputTypeNone WorkflowDispatchEventInputType = iota
	// WorkflowDispatchEventInputTypeString is string type of input of workflow_dispatch event.
	WorkflowDispatchEventInputTypeString
	// WorkflowDispatchEventInputTypeBoolean is boolean type of input of workflow_dispatch event.
	WorkflowDispatchEventInputTypeBoolean
	// WorkflowDispatchEventInputTypeChoice is choice type of input of workflow_dispatch event.
	WorkflowDispatchEventInputTypeChoice
	// WorkflowDispatchEventInputTypeEnvironment is environment type of input of workflow_dispatch event.
	WorkflowDispatchEventInputTypeEnvironment
)

// DispatchInput is input specified on dispatching workflow manually.
// https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#workflow_dispatch
type DispatchInput struct {
	// Name is a name of input value specified on dispatching workflow manually.
	Name *String
	// Description is a description of input value specified on dispatching workflow manually.
	Description *String
	// Required is a flag to show if this input is mandatory or not on dispatching workflow manually.
	Required *Bool
	// Default is a default value of input value on dispatching workflow manually.
	Default *String
	// Type is a type of the input
	// https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#workflow_dispatch
	Type WorkflowDispatchEventInputType
	// Options is list of options of choice type
	Options []*String
}

// WorkflowDispatchEvent is event on dispatching workflow manually.
// https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#workflow_dispatch
type WorkflowDispatchEvent struct {
	// Inputs is map from input names to input attributes. Keys are in lower case since they are case insensitive.
	Inputs map[string]*DispatchInput
	// Pos is a position in source.
	Pos *Pos
}

// EventName returns name of the event to trigger this workflow.
func (e *WorkflowDispatchEvent) EventName() string {
	return "workflow_dispatch"
}

// RepositoryDispatchEvent is repository_dispatch event configuration.
// https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#repository_dispatch
type RepositoryDispatchEvent struct {
	// Types is list of types which can trigger workflow.
	Types []*String
	// Pos is a position in source.
	Pos *Pos
}

// EventName returns name of the event to trigger this workflow.
func (e *RepositoryDispatchEvent) EventName() string {
	return "repository_dispatch"
}

// WorkflowCallEventInputType is a type of inputs at workflow_call event.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#onworkflow_callinput_idtype
type WorkflowCallEventInputType uint8

const (
	// WorkflowCallEventInputTypeInvalid represents invalid type input as default value of the type.
	WorkflowCallEventInputTypeInvalid WorkflowCallEventInputType = iota
	// WorkflowCallEventInputTypeBoolean represents boolean type input.
	WorkflowCallEventInputTypeBoolean
	// WorkflowCallEventInputTypeNumber represents number type input.
	WorkflowCallEventInputTypeNumber
	// WorkflowCallEventInputTypeString represents string type input.
	WorkflowCallEventInputTypeString
)

// WorkflowCallEventInput is an input configuration of workflow_call event.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#onworkflow_callinputs
type WorkflowCallEventInput struct {
	// Name is a name of the input.
	Name *String
	// Description is a description of the input.
	Description *String
	// Default is a default value of the input. Nil means no default value.
	Default *String
	// Required represents if the input is required or optional. When this value is nil, it was not explicitly specified.
	// In the case the default value is 'not required'.
	Required *Bool
	// Type of the input, which must be one of 'boolean', 'number' or 'string'. This property is required.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#onworkflow_callinput_idtype
	Type WorkflowCallEventInputType
	// ID is an ID of the input. Input ID is in lower case because it is case-insensitive.
	ID string
}

// IsRequired returns if the input is marked as required or not.
// require
func (i *WorkflowCallEventInput) IsRequired() bool {
	return i.Required != nil && i.Required.Value
}

// WorkflowCallEventSecret is a secret configuration of workflow_call event.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#onworkflow_callsecrets
type WorkflowCallEventSecret struct {
	// Name is a name of the secret.
	Name *String
	// Description is a description of the secret.
	Description *String
	// Required represents if the secret is required or optional. When this value is nil, it means optional.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#onworkflow_callsecretssecret_idrequired
	Required *Bool
}

// WorkflowCallEventOutput is an output configuration of workflow_call event.
// https://docs.github.com/en/actions/using-workflows/reusing-workflows#using-outputs-from-a-reusable-workflow
type WorkflowCallEventOutput struct {
	// Name is a name of the output.
	Name *String
	// Description is a description of the output.
	Description *String
	// Value is an expression for the value of the output.
	Value *String
}

// WorkflowCallEvent is workflow_call event configuration.
// https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#workflow-reuse-events
type WorkflowCallEvent struct {
	// Inputs is an array of inputs of the workflow_call event. This value is not a map unlike other fields of this
	// struct since its order is important when checking the default values of inputs.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#onworkflow_callinputs
	Inputs []*WorkflowCallEventInput
	// Secrets is a map from name of secret to secret configuration. When 'secrets' is omitted, nil is set to this
	// field. Keys are in lower case since they are case-insensitive.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#onworkflow_callsecrets
	Secrets map[string]*WorkflowCallEventSecret
	// Outputs is a map from name of output to output configuration. Keys are in lower case since they are case-insensitive.
	// https://docs.github.com/en/actions/using-workflows/reusing-workflows#using-outputs-from-a-reusable-workflow
	Outputs map[string]*WorkflowCallEventOutput
	// Pos is a position in source.
	Pos *Pos
}

// EventName returns name of the event to trigger this workflow.
func (e *WorkflowCallEvent) EventName() string {
	return "workflow_call"
}

// PermissionScope is struct for respective permission scope like "issues", "checks", ...
// https://docs.github.com/en/actions/security-guides/automatic-token-authentication#permissions-for-the-github_token
type PermissionScope struct {
	// Name is name of the scope.
	Name *String
	// Value is permission value of the scope.
	Value *String
}

// Permissions is set of permission configurations in workflow file. All permissions can be set at
// once. Or each permission can be configured respectively.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#permissions
type Permissions struct {
	// All represents a permission value for all the scopes at once.
	All *String
	// Scopes is mappings from scope name to its permission configuration
	Scopes map[string]*PermissionScope
	// Pos is a position in source.
	Pos *Pos
}

// DefaultsRun is configuration that shell is how to be run.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#defaultsrun
type DefaultsRun struct {
	// Shell is shell name to be run.
	Shell *String
	// WorkingDirectory is a default working directory path.
	WorkingDirectory *String
	// Pos is a position in source.
	Pos *Pos
}

// Defaults is set of default configurations to run shell.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#defaults
type Defaults struct {
	// Run is configuration of how to run shell.
	Run *DefaultsRun
	// Pos is a position in source.
	Pos *Pos
}

// Concurrency is a configuration of concurrency of the workflow.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#concurrency
type Concurrency struct {
	// Group is name of the concurrency group.
	Group *String
	// CancelInProgress is a flag that shows if canceling this workflow cancels other jobs in progress.
	CancelInProgress *Bool
	// Pos is a position in source.
	Pos *Pos
}

// Environment is a configuration of environment.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idenvironment
type Environment struct {
	// Name is a name of environment which the workflow uses.
	Name *String
	// URL is the URL mapped to 'environment_url' in the deployments API. Empty value means no value was specified.
	URL *String
	// Pos is a position in source.
	Pos *Pos
}

// ExecKind is kind of how the step is executed. A step runs some action or runs some shell script.
type ExecKind uint8

const (
	// ExecKindAction is kind for step to run action
	ExecKindAction ExecKind = iota
	// ExecKindRun is kind for step to run shell script
	ExecKindRun
)

// Exec is an interface how the step is executed. Step in workflow runs either an action or a script
type Exec interface {
	// Kind returns kind of the step execution.
	Kind() ExecKind
}

// ExecRun is configuration how to run shell script at the step.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsrun
type ExecRun struct {
	// Run is script to run.
	Run *String
	// Shell represents optional 'shell' field. Nil means nothing specified.
	Shell *String
	// WorkingDirectory represents optional 'working-directory' field. Nil means nothing specified.
	WorkingDirectory *String
	// RunPos is position of 'run' section
	RunPos *Pos
}

// Kind returns kind of the step execution.
func (e *ExecRun) Kind() ExecKind {
	return ExecKindRun
}

// Input is an input field for running an action.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepswith
type Input struct {
	// Name is a name of the input.
	Name *String
	// Value is a value of the input.
	Value *String
}

// ExecAction is configuration how to run action at the step.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsuses
type ExecAction struct {
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsuses
	Uses *String
	// Inputs represents inputs to the action to execute in 'with' section. Keys are in lower case since they are case-insensitive.
	Inputs map[string]*Input
	// Entrypoint represents optional 'entrypoint' field in 'with' section. Nil field means nothing specified
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepswithentrypoint
	Entrypoint *String
	// Args represents optional 'args' field in 'with' section. Nil field means nothing specified
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepswithargs
	Args *String
}

// Kind returns kind of the step execution.
func (e *ExecAction) Kind() ExecKind {
	return ExecKindAction
}

// RawYAMLValueKind is kind of raw YAML values
type RawYAMLValueKind int

const (
	// RawYAMLValueKindObject is kind for an object value of raw YAML value.
	RawYAMLValueKindObject = RawYAMLValueKind(yaml.MappingNode)
	// RawYAMLValueKindArray is kind for an array value of raw YAML value.
	RawYAMLValueKindArray = RawYAMLValueKind(yaml.SequenceNode)
	// RawYAMLValueKindString is kind for a string value of raw YAML value.
	RawYAMLValueKindString = RawYAMLValueKind(yaml.ScalarNode)
)

// RawYAMLValue is a value at matrix variation. Any value can be put at matrix variations
// including mappings and arrays.
type RawYAMLValue interface {
	// Kind returns kind of raw YAML value.
	Kind() RawYAMLValueKind
	// Equals returns if the other value is equal to the value.
	Equals(other RawYAMLValue) bool
	// Pos returns the start position of the value in the source file
	Pos() *Pos
	// String returns string representation of the value
	String() string
}

// RawYAMLObject is raw YAML mapping value.
type RawYAMLObject struct {
	// Props is map from property names to their values. Keys are in lower case since they are case-insensitive.
	Props map[string]RawYAMLValue
	pos   *Pos
}

// Kind returns kind of raw YAML value.
func (o *RawYAMLObject) Kind() RawYAMLValueKind {
	return RawYAMLValueKindObject
}

// Equals returns if the other value is equal to the value.
func (o *RawYAMLObject) Equals(other RawYAMLValue) bool {
	switch other := other.(type) {
	case *RawYAMLObject:
		for n, p1 := range o.Props {
			if p2, ok := other.Props[n]; !ok || !p1.Equals(p2) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// Pos returns the start position of the value in the source file
func (o *RawYAMLObject) Pos() *Pos {
	return o.pos
}

func (o *RawYAMLObject) String() string {
	qs := make([]string, 0, len(o.Props))
	for n, p := range o.Props {
		qs = append(qs, fmt.Sprintf("%q: %s", n, p.String()))
	}
	return fmt.Sprintf("{%s}", strings.Join(qs, ", "))
}

// RawYAMLArray is raw YAML sequence value.
type RawYAMLArray struct {
	// Elems is list of elements of the array value.
	Elems []RawYAMLValue
	pos   *Pos
}

// Kind returns kind of raw YAML value.
func (a *RawYAMLArray) Kind() RawYAMLValueKind {
	return RawYAMLValueKindArray
}

// Equals returns if the other value is equal to the value.
func (a *RawYAMLArray) Equals(other RawYAMLValue) bool {
	switch other := other.(type) {
	case *RawYAMLArray:
		if len(a.Elems) != len(other.Elems) {
			return false
		}
		for i, e1 := range a.Elems {
			if !e1.Equals(other.Elems[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// Pos returns the start position of the value in the source file
func (a *RawYAMLArray) Pos() *Pos {
	return a.pos
}

func (a *RawYAMLArray) String() string {
	qs := make([]string, 0, len(a.Elems))
	for _, v := range a.Elems {
		qs = append(qs, v.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(qs, ", "))
}

// RawYAMLString is raw YAML scalar value.
type RawYAMLString struct {
	// Note: Might be useful to add kind to check the string value is int/float/bool/null.

	// Value is string representation of the scalar node.
	Value string
	pos   *Pos
}

// Kind returns kind of raw YAML value.
func (s *RawYAMLString) Kind() RawYAMLValueKind {
	return RawYAMLValueKindString
}

// Equals returns if the other value is equal to the value.
func (s *RawYAMLString) Equals(other RawYAMLValue) bool {
	switch other := other.(type) {
	case *RawYAMLString:
		return s.Value == other.Value
	default:
		return false
	}
}

// Pos returns the start position of the value in the source file
func (s *RawYAMLString) Pos() *Pos {
	return s.pos
}

func (s *RawYAMLString) String() string {
	return strconv.Quote(s.Value)
}

// MatrixRow is one row of matrix. One matrix row can take multiple values. Those variations are
// stored as row of values in this struct.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix
type MatrixRow struct {
	// Name is a name of matrix value.
	Name *String
	// Values is variations of values which the matrix value can take.
	Values []RawYAMLValue
	// Expression is a string when expression syntax ${{ }} is used for this section.
	Expression *String
}

// MatrixAssign represents which value should be taken in the row of the matrix.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix
type MatrixAssign struct {
	// Key is a name of the matrix value.
	Key *String
	// Value is the value selected from values in row.
	Value RawYAMLValue
}

// MatrixCombination is combination of matrix value assignments to define one of matrix variations.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix
type MatrixCombination struct {
	Assigns map[string]*MatrixAssign
	// Expression is a string when expression syntax ${{ }} is used for this section.
	Expression *String
}

// MatrixCombinations is list of combinations of matrix assignments used for 'include' and 'exclude'
// sections.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix
type MatrixCombinations struct {
	Combinations []*MatrixCombination
	// Expression is a string when expression syntax ${{ }} is used for this section.
	Expression *String
}

// ContainsExpression returns if the combinations section includes at least one expression node.
func (cs *MatrixCombinations) ContainsExpression() bool {
	if cs.Expression != nil {
		return true
	}
	for _, c := range cs.Combinations {
		if c.Expression != nil {
			return true
		}
	}
	return false
}

// Matrix is matrix variations configuration of a job.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix
type Matrix struct {
	// Values stores mappings from name to values. Keys are in lower case since they are case-insensitive.
	Rows map[string]*MatrixRow
	// Include is list of combinations of matrix values and additional values on running matrix combinations.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#example-including-additional-values-into-combinations
	Include *MatrixCombinations
	// Exclude is list of combinations of matrix values which should not be run. Combinations in
	// this list will be removed from combinations of matrix to run.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#example-excluding-configurations-from-a-matrix
	Exclude *MatrixCombinations
	// Expression is a string when expression syntax ${{ }} is used for this section.
	Expression *String
	// Pos is a position in source.
	Pos *Pos
}

// Strategy is strategy configuration of how the job is run.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategy
type Strategy struct {
	// Matrix is matrix of combinations of values. Each combination will run the job once.
	Matrix *Matrix
	// FailFast is flag to show if other jobs should stop when one job fails.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategyfail-fast
	FailFast *Bool
	// MaxParallel is how many jobs should be run at once.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategymax-parallel
	MaxParallel *Int
	// Pos is a position in source.
	Pos *Pos
}

// EnvVar represents key-value of environment variable setup.
type EnvVar struct {
	// Name is name of the environment variable.
	Name *String
	// Value is string value of the environment variable.
	Value *String
}

// Env represents set of environment variables.
type Env struct {
	// Vars is mapping from env var name to env var value.
	Vars map[string]*EnvVar
	// Expression is an expression string which contains ${{ ... }}. When this value is not empty,
	// Vars should be nil.
	Expression *String
}

// Step is step configuration. Step runs one action or one shell script.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idsteps
type Step struct {
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsid
	ID *String
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsif
	If *String
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsname
	Name *String
	Exec Exec
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsenv
	Env *Env
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepscontinue-on-error
	ContinueOnError *Bool
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepstimeout-minutes
	TimeoutMinutes *Float
	// Pos is a position in source.
	Pos *Pos
}

// Credentials is credentials configuration.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idcontainercredentials
type Credentials struct {
	// Username is username for authentication.
	Username *String
	// Password is password for authentication.
	Password *String
	// Pos is a position in source.
	Pos *Pos
}

// Container is configuration of how to run the container.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idcontainer
type Container struct {
	// Image is specification of Docker image.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idcontainerimage
	Image *String
	// Credentials is credentials configuration of the Docker container.
	Credentials *Credentials
	// Env is environment variables setup in the container.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idcontainerenv
	Env *Env
	// Ports is list of port number mappings of the container.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idcontainerports
	Ports []*String
	// Volumes are list of volumes to be mounted to the container.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idcontainervolumes
	Volumes []*String
	// Options is options string to run the container.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idcontaineroptions
	Options *String
	// Pos is a position in source.
	Pos *Pos
}

// Service is configuration to run a service like database.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idservices
type Service struct {
	// Name is name of the service.
	Name *String
	// Container is configuration of container which runs the service.
	Container *Container
}

// Output is output entry of the job.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idoutputs
type Output struct {
	// Name is name of output.
	Name *String
	// Value is value of output.
	Value *String
}

// Runner is struct for runner configuration.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idruns-on
type Runner struct {
	// Labels is list label names to select a runner to run a job. There are preset labels and user
	// defined labels. Runner matching to the labels is selected.
	Labels []*String
	// Expression is a string when expression syntax ${{ }} is used for this section. Related issue is #164.
	Expression *String
}

// WorkflowCallInput is a normal input for workflow call.
type WorkflowCallInput struct {
	// Name is a name of the input.
	Name *String
	// Value is a value of the input.
	Value *String
}

// WorkflowCallSecret is a secret input for workflow call.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idwith
type WorkflowCallSecret struct {
	// Name is a name of the secret
	Name *String
	// Value is a value of the secret
	Value *String
}

// WorkflowCall is a struct to represent workflow call at jobs.<job_id>.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_iduses
type WorkflowCall struct {
	// Uses is a workflow specification to be called. This field is mandatory.
	Uses *String
	// Inputs is a map from input name to input value at 'with:'. Keys are in lower case since input names
	// are case-insensitive.
	Inputs map[string]*WorkflowCallInput
	// Secrets is a map from secret name to secret value at 'secrets:'. Keys are in lower case since input
	// names are case-insensitive.
	Secrets map[string]*WorkflowCallSecret
	// InheritSecrets is true when 'secrets: inherit' is specified. In this case, Secrets must be empty.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#onworkflow_callsecretsinherit
	InheritSecrets bool
}

// Job is configuration of how to run a job.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobs
type Job struct {
	// ID is an ID of the job, which is key of job configuration objects.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_id
	ID *String
	// Name is a name of job that user can specify freely.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idname
	Name *String
	// Needs is list of job IDs which should be run before running this job.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idneeds
	Needs []*String
	// RunsOn is runner configuration which run the job.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idruns-on
	RunsOn *Runner
	// Permissions is permission configuration for running the job.
	Permissions *Permissions
	// Environment is environment specification where the job runs.
	Environment *Environment
	// Concurrency is concurrency configuration on running the job.
	Concurrency *Concurrency
	// Outputs is map from output name to output specifications.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idoutputs
	Outputs map[string]*Output
	// Env is environment variables setup while running the job.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idenv
	Env *Env
	// Defaults is default configurations of how to run scripts.
	Defaults *Defaults
	// If is a condition whether this job should be run.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idif
	If *String
	// Steps is list of steps to be run in the job.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idsteps
	Steps []*Step
	// TimeoutMinutes is timeout value of running the job in minutes.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idtimeout-minutes
	TimeoutMinutes *Float
	// Strategy is strategy configuration of running the job.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategy
	Strategy *Strategy
	// ContinueOnError is a flag to show if execution should continue on error.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idcontinue-on-error
	ContinueOnError *Bool
	// Container is container configuration to run the job.
	Container *Container
	// Services is map from service names to service configurations. Keys are in lower case since they are case-insensitive.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idservices
	Services map[string]*Service
	// WorkflowCall is a workflow call by 'uses:'.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_iduses
	WorkflowCall *WorkflowCall
	// Pos is a position in source.
	Pos *Pos
}

// Workflow is root of workflow syntax tree, which represents one workflow configuration file.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions
type Workflow struct {
	// Name is name of the workflow. This field can be nil when user didn't specify the name explicitly.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#name
	Name *String
	// RunName is the name of workflow runs. This field can be set dynamically using ${{ }}.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#run-name
	RunName *String
	// On is list of events which can trigger this workflow.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#onpushpull_requestbranchestags
	On []Event
	// Permissions is configuration of permissions of this workflow.
	Permissions *Permissions
	// Env is a default set of environment variables while running this workflow.
	// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#env
	Env *Env
	// Defaults is default configuration of how to run scripts.
	Defaults *Defaults
	// Concurrency is concurrency configuration of entire workflow. Each jobs also can their own
	// concurrency configurations.
	Concurrency *Concurrency
	// Jobs is mappings from job ID to the job object. Keys are in lower case since they are case-insensitive.
	Jobs map[string]*Job
}

// FindWorkflowCallEvent returns workflow_call event node if exists
func (w *Workflow) FindWorkflowCallEvent() (*WorkflowCallEvent, bool) {
	for _, e := range w.On {
		if e, ok := e.(*WorkflowCallEvent); ok {
			return e, true
		}
	}
	return nil, false
}
