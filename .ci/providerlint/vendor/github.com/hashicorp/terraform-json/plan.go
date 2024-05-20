// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/go-version"
)

// PlanFormatVersionConstraints defines the versions of the JSON plan format
// that are supported by this package.
var PlanFormatVersionConstraints = ">= 0.1, < 2.0"

// ResourceMode is a string representation of the resource type found
// in certain fields in the plan.
type ResourceMode string

const (
	// DataResourceMode is the resource mode for data sources.
	DataResourceMode ResourceMode = "data"

	// ManagedResourceMode is the resource mode for managed resources.
	ManagedResourceMode ResourceMode = "managed"
)

// Plan represents the entire contents of an output Terraform plan.
type Plan struct {
	// useJSONNumber opts into the behavior of calling
	// json.Decoder.UseNumber prior to decoding the plan, which turns
	// numbers into json.Numbers instead of float64s. Set it using
	// Plan.UseJSONNumber.
	useJSONNumber bool

	// The version of the plan format. This should always match the
	// PlanFormatVersion constant in this package, or else an unmarshal
	// will be unstable.
	FormatVersion string `json:"format_version,omitempty"`

	// The version of Terraform used to make the plan.
	TerraformVersion string `json:"terraform_version,omitempty"`

	// The variables set in the root module when creating the plan.
	Variables map[string]*PlanVariable `json:"variables,omitempty"`

	// The common state representation of resources within this plan.
	// This is a product of the existing state merged with the diff for
	// this plan.
	PlannedValues *StateValues `json:"planned_values,omitempty"`

	// The change operations for resources and data sources within this plan
	// resulting from resource drift.
	ResourceDrift []*ResourceChange `json:"resource_drift,omitempty"`

	// The change operations for resources and data sources within this
	// plan.
	ResourceChanges []*ResourceChange `json:"resource_changes,omitempty"`

	// The change operations for outputs within this plan.
	OutputChanges map[string]*Change `json:"output_changes,omitempty"`

	// The Terraform state prior to the plan operation. This is the
	// same format as PlannedValues, without the current diff merged.
	PriorState *State `json:"prior_state,omitempty"`

	// The Terraform configuration used to make the plan.
	Config *Config `json:"configuration,omitempty"`

	// RelevantAttributes represents any resource instances and their
	// attributes which may have contributed to the planned changes
	RelevantAttributes []ResourceAttribute `json:"relevant_attributes,omitempty"`

	// Checks contains the results of any conditional checks executed, or
	// planned to be executed, during this plan.
	Checks []CheckResultStatic `json:"checks,omitempty"`

	// Timestamp contains the static timestamp that Terraform considers to be
	// the time this plan executed, in UTC.
	Timestamp string `json:"timestamp,omitempty"`
}

// ResourceAttribute describes a full path to a resource attribute
type ResourceAttribute struct {
	// Resource describes resource instance address (e.g. null_resource.foo)
	Resource string `json:"resource"`
	// Attribute describes the attribute path using a lossy representation
	// of cty.Path. (e.g. ["id"] or ["objects", 0, "val"]).
	Attribute []json.RawMessage `json:"attribute"`
}

// UseJSONNumber controls whether the Plan will be decoded using the
// json.Number behavior or the float64 behavior. When b is true, the Plan will
// represent numbers in PlanOutputs as json.Numbers. When b is false, the
// Plan will represent numbers in PlanOutputs as float64s.
func (p *Plan) UseJSONNumber(b bool) {
	p.useJSONNumber = b
}

// Validate checks to ensure that the plan is present, and the
// version matches the version supported by this library.
func (p *Plan) Validate() error {
	if p == nil {
		return errors.New("plan is nil")
	}

	if p.FormatVersion == "" {
		return errors.New("unexpected plan input, format version is missing")
	}

	constraint, err := version.NewConstraint(PlanFormatVersionConstraints)
	if err != nil {
		return fmt.Errorf("invalid version constraint: %w", err)
	}

	version, err := version.NewVersion(p.FormatVersion)
	if err != nil {
		return fmt.Errorf("invalid format version %q: %w", p.FormatVersion, err)
	}

	if !constraint.Check(version) {
		return fmt.Errorf("unsupported plan format version: %q does not satisfy %q",
			version, constraint)
	}

	return nil
}

func isStringInSlice(slice []string, s string) bool {
	for _, el := range slice {
		if el == s {
			return true
		}
	}
	return false
}

func (p *Plan) UnmarshalJSON(b []byte) error {
	type rawPlan Plan
	var plan rawPlan

	dec := json.NewDecoder(bytes.NewReader(b))
	if p.useJSONNumber {
		dec.UseNumber()
	}
	err := dec.Decode(&plan)
	if err != nil {
		return err
	}

	*p = *(*Plan)(&plan)

	return p.Validate()
}

// ResourceChange is a description of an individual change action
// that Terraform plans to use to move from the prior state to a new
// state matching the configuration.
type ResourceChange struct {
	// The absolute resource address.
	Address string `json:"address,omitempty"`

	// The absolute address that this resource instance had
	// at the conclusion of a previous plan.
	PreviousAddress string `json:"previous_address,omitempty"`

	// The module portion of the above address. Omitted if the instance
	// is in the root module.
	ModuleAddress string `json:"module_address,omitempty"`

	// The resource mode.
	Mode ResourceMode `json:"mode,omitempty"`

	// The resource type, example: "aws_instance" for aws_instance.foo.
	Type string `json:"type,omitempty"`

	// The resource name, example: "foo" for aws_instance.foo.
	Name string `json:"name,omitempty"`

	// The instance key for any resources that have been created using
	// "count" or "for_each". If neither of these apply the key will be
	// empty.
	//
	// This value can be either an integer (int) or a string.
	Index interface{} `json:"index,omitempty"`

	// The name of the provider this resource belongs to. This allows
	// the provider to be interpreted unambiguously in the unusual
	// situation where a provider offers a resource type whose name
	// does not start with its own name, such as the "googlebeta"
	// provider offering "google_compute_instance".
	ProviderName string `json:"provider_name,omitempty"`

	// An identifier used during replacement operations, and can be
	// used to identify the exact resource being replaced in state.
	DeposedKey string `json:"deposed,omitempty"`

	// The data describing the change that will be made to this object.
	Change *Change `json:"change,omitempty"`
}

// Change is the representation of a proposed change for an object.
type Change struct {
	// The action to be carried out by this change.
	Actions Actions `json:"actions,omitempty"`

	// Before and After are representations of the object value both
	// before and after the action. For create and delete actions,
	// either Before or After is unset (respectively). For no-op
	// actions, both values will be identical. After will be incomplete
	// if there are values within it that won't be known until after
	// apply.
	Before interface{} `json:"before,"`
	After  interface{} `json:"after,omitempty"`

	// A deep object of booleans that denotes any values that are
	// unknown in a resource. These values were previously referred to
	// as "computed" values.
	//
	// If the value cannot be found in this map, then its value should
	// be available within After, so long as the operation supports it.
	AfterUnknown interface{} `json:"after_unknown,omitempty"`

	// BeforeSensitive and AfterSensitive are object values with similar
	// structure to Before and After, but with all sensitive leaf values
	// replaced with true, and all non-sensitive leaf values omitted. These
	// objects should be combined with Before and After to prevent accidental
	// display of sensitive values in user interfaces.
	BeforeSensitive interface{} `json:"before_sensitive,omitempty"`
	AfterSensitive  interface{} `json:"after_sensitive,omitempty"`

	// Importing contains the import metadata about this operation. If importing
	// is present (ie. not null) then the change is an import operation in
	// addition to anything mentioned in the actions field. The actual contents
	// of the Importing struct is subject to change, so downstream consumers
	// should treat any values in here as strictly optional.
	Importing *Importing `json:"importing,omitempty"`

	// GeneratedConfig contains any HCL config generated for this resource
	// during planning as a string.
	//
	// If this is populated, then Importing should also be populated but this
	// might change in the future. However, not all Importing changes will
	// contain generated config.
	GeneratedConfig string `json:"generated_config,omitempty"`

	// ReplacePaths contains a set of paths that point to attributes/elements
	// that are causing the overall resource to be replaced rather than simply
	// updated.
	//
	// This field is always a slice of indexes, where an index in this context
	// is either an integer pointing to a child of a set/list, or a string
	// pointing to the child of a map, object, or block.
	ReplacePaths []interface{} `json:"replace_paths,omitempty"`
}

// Importing is a nested object for the resource import metadata.
type Importing struct {
	// The original ID of this resource used to target it as part of planned
	// import operation.
	ID string `json:"id,omitempty"`
}

// PlanVariable is a top-level variable in the Terraform plan.
type PlanVariable struct {
	// The value for this variable at plan time.
	Value interface{} `json:"value,omitempty"`
}
