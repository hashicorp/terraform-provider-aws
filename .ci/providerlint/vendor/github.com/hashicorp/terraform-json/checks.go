// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfjson

// CheckKind is a string representation of the type of conditional check
// referenced in a check result.
type CheckKind string

const (
	// CheckKindResource indicates the check result is from a pre- or
	// post-condition on a resource or data source.
	CheckKindResource CheckKind = "resource"

	// CheckKindOutputValue indicates the check result is from an output
	// post-condition.
	CheckKindOutputValue CheckKind = "output_value"

	// CheckKindCheckBlock indicates the check result is from a check block.
	CheckKindCheckBlock CheckKind = "check"
)

// CheckStatus is a string representation of the status of a given conditional
// check.
type CheckStatus string

const (
	// CheckStatusPass indicates the check passed.
	CheckStatusPass CheckStatus = "pass"

	// CheckStatusFail indicates the check failed.
	CheckStatusFail CheckStatus = "fail"

	// CheckStatusError indicates the check errored. This is distinct from
	// CheckStatusFail in that it represents a logical or configuration error
	// within the check block that prevented the check from executing, as
	// opposed to the check was attempted and evaluated to false.
	CheckStatusError CheckStatus = "error"

	// CheckStatusUnknown indicates the result of the check was not known. This
	// could be because a value within the check could not be known at plan
	// time, or because the overall plan failed for an unrelated reason before
	// this check could be executed.
	CheckStatusUnknown CheckStatus = "unknown"
)

// CheckStaticAddress details the address of the object that performed a given
// check. The static address points to the overall resource, as opposed to the
// dynamic address which contains the instance key for any resource that has
// multiple instances.
type CheckStaticAddress struct {
	// ToDisplay is a formatted and ready to display representation of the
	// address.
	ToDisplay string `json:"to_display"`

	// Kind represents the CheckKind of this check.
	Kind CheckKind `json:"kind"`

	// Module is the module part of the address. This will be empty for any
	// resources in the root module.
	Module string `json:"module,omitempty"`

	// Mode is the ResourceMode of the resource that contains this check. This
	// field is only set is Kind equals CheckKindResource.
	Mode ResourceMode `json:"mode,omitempty"`

	// Type is the resource type for the resource that contains this check. This
	// field is only set if Kind equals CheckKindResource.
	Type string `json:"type,omitempty"`

	// Name is the name of the resource, check block, or output that contains
	// this check.
	Name string `json:"name,omitempty"`
}

// CheckDynamicAddress contains the InstanceKey field for any resources that
// have multiple instances. A complete address can be built by combining the
// CheckStaticAddress with the CheckDynamicAddress.
type CheckDynamicAddress struct {
	// ToDisplay is a formatted and ready to display representation of the
	// full address, including the additional information from the relevant
	// CheckStaticAddress.
	ToDisplay string `json:"to_display"`

	// Module is the module part of the address. This address will include the
	// instance key for any module expansions resulting from foreach or count
	// arguments. This field will be empty for any resources within the root
	// module.
	Module string `json:"module,omitempty"`

	// InstanceKey is the instance key for any instances of a given resource.
	//
	// InstanceKey will be empty if there was no foreach or count argument
	// defined on the containing object.
	InstanceKey interface{} `json:"instance_key,omitempty"`
}

// CheckResultStatic is the container for a "checkable object".
//
// A "checkable object" is a resource or data source, an output, or a check
// block.
type CheckResultStatic struct {
	// Address is the absolute address of the "checkable object"
	Address CheckStaticAddress `json:"address"`

	// Status is the overall status for all the checks within this object.
	Status CheckStatus `json:"status"`

	// Instances contains the results for dynamic object that belongs to this
	// static object. For example, any instances created from an object using
	// the foreach or count meta arguments.
	//
	// Check blocks and outputs will only contain a single instance, while
	// resources can contain 1 to many.
	Instances []CheckResultDynamic `json:"instances,omitempty"`
}

// CheckResultDynamic describes the check result for a dynamic object that
// results from the expansion of the containing object.
type CheckResultDynamic struct {
	// Address is the relative address of this instance given the Address in the
	// parent object.
	Address CheckDynamicAddress `json:"address"`

	// Status is the overall status for the checks within this dynamic object.
	Status CheckStatus `json:"status"`

	// Problems describes any additional optional details about this check if
	// the check failed.
	//
	// This will not include the errors resulting from this check block, as they
	// will be exposed as diagnostics in the original terraform execution. It
	// may contain any failure messages even if the overall status is
	// CheckStatusError, however, as the instance could contain multiple checks
	// that returned a mix of error and failure statuses.
	Problems []CheckResultProblem `json:"problems,omitempty"`
}

// CheckResultProblem describes one of potentially several problems that led to
// a check being classied as CheckStatusFail.
type CheckResultProblem struct {
	// Message is the condition error message provided by the original check
	// author.
	Message string `json:"message"`
}
