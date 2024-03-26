// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var ReservedDataSourceFields = []string{
	"connection",
	"count",
	"depends_on",
	"lifecycle",
	"provider",
	"provisioner",
}

var ReservedResourceFields = []string{
	"connection",
	"count",
	"depends_on",
	"lifecycle",
	"provider",
	"provisioner",
}

// Resource is an abstraction for multiple Terraform concepts:
//
//   - Managed Resource: An infrastructure component with a schema, lifecycle
//     operations such as create, read, update, and delete
//     (CRUD), and optional implementation details such as
//     import support, upgrade state support, and difference
//     customization.
//   - Data Resource: Also known as a data source. An infrastructure component
//     with a schema and only the read lifecycle operation.
//   - Block: When implemented within a Schema type Elem field, a configuration
//     block that contains nested schema information such as attributes
//     and blocks.
//
// To fully implement managed resources, the Provider type ResourcesMap field
// should include a reference to an implementation of this type. To fully
// implement data resources, the Provider type DataSourcesMap field should
// include a reference to an implementation of this type.
//
// Each field further documents any constraints based on the Terraform concept
// being implemented.
type Resource struct {
	// Schema is the structure and type information for this component. This
	// field, or SchemaFunc, is required for all Resource concepts. To prevent
	// storing all schema information in memory for the lifecycle of a provider,
	// use SchemaFunc instead.
	//
	// The keys of this map are the names used in a practitioner configuration,
	// such as the attribute or block name. The values describe the structure
	// and type information of that attribute or block.
	Schema map[string]*Schema

	// SchemaFunc is the structure and type information for this component. This
	// field, or Schema, is required for all Resource concepts. Use this field
	// instead of Schema on top level Resource declarations to prevent storing
	// all schema information in memory for the lifecycle of a provider.
	//
	// The keys of this map are the names used in a practitioner configuration,
	// such as the attribute or block name. The values describe the structure
	// and type information of that attribute or block.
	SchemaFunc func() map[string]*Schema

	// SchemaVersion is the version number for this resource's Schema
	// definition. This field is only valid when the Resource is a managed
	// resource.
	//
	// The current SchemaVersion stored in the state for each resource.
	// Provider authors can increment this version number when Schema semantics
	// change in an incompatible manner. If the state's SchemaVersion is less
	// than the current SchemaVersion, the MigrateState and StateUpgraders
	// functionality is executed to upgrade the state information.
	//
	// When unset, SchemaVersion defaults to 0, so provider authors can start
	// their Versioning at any integer >= 1
	SchemaVersion int

	// MigrateState is responsible for updating an InstanceState with an old
	// version to the format expected by the current version of the Schema.
	// This field is only valid when the Resource is a managed resource.
	//
	// It is called during Refresh if the State's stored SchemaVersion is less
	// than the current SchemaVersion of the Resource.
	//
	// The function is yielded the state's stored SchemaVersion and a pointer to
	// the InstanceState that needs updating, as well as the configured
	// provider's configured meta interface{}, in case the migration process
	// needs to make any remote API calls.
	//
	// Deprecated: MigrateState is deprecated and any new changes to a resource's schema
	// should be handled by StateUpgraders. Existing MigrateState implementations
	// should remain for compatibility with existing state. MigrateState will
	// still be called if the stored SchemaVersion is less than the
	// first version of the StateUpgraders.
	MigrateState StateMigrateFunc

	// StateUpgraders contains the functions responsible for upgrading an
	// existing state with an old schema version to a newer schema. It is
	// called specifically by Terraform when the stored schema version is less
	// than the current SchemaVersion of the Resource. This field is only valid
	// when the Resource is a managed resource.
	//
	// StateUpgraders map specific schema versions to a StateUpgrader
	// function. The registered versions are expected to be ordered,
	// consecutive values. The initial value may be greater than 0 to account
	// for legacy schemas that weren't recorded and can be handled by
	// MigrateState.
	StateUpgraders []StateUpgrader

	// Create is called when the provider must create a new instance of a
	// managed resource. This field is only valid when the Resource is a
	// managed resource. Only one of Create, CreateContext, or
	// CreateWithoutTimeout should be implemented.
	//
	// The *ResourceData parameter contains the plan and state data for this
	// managed resource instance. The available data in the Get* methods is the
	// the proposed state, which is the merged data of the practitioner
	// configuration and any CustomizeDiff field logic.
	//
	// The SetId method must be called with a non-empty value for the managed
	// resource instance to be properly saved into the Terraform state and
	// avoid a "inconsistent result after apply" error.
	//
	// The interface{} parameter is the result of the Provider type
	// ConfigureFunc field execution. If the Provider does not define
	// a ConfigureFunc, this will be nil. This parameter is conventionally
	// used to store API clients and other provider instance specific data.
	//
	// The error return parameter, if not nil, will be converted into an error
	// diagnostic when passed back to Terraform.
	//
	// Deprecated: Use CreateContext or CreateWithoutTimeout instead. This
	// implementation does not support request cancellation initiated by
	// Terraform, such as a system or practitioner sending SIGINT (Ctrl-c).
	// This implementation also does not support warning diagnostics.
	Create CreateFunc

	// Read is called when the provider must refresh the state of a managed
	// resource instance or data resource instance. This field is only valid
	// when the Resource is a managed resource or data resource. Only one of
	// Read, ReadContext, or ReadWithoutTimeout should be implemented.
	//
	// The *ResourceData parameter contains the state data for this managed
	// resource instance or data resource instance.
	//
	// Managed resources can signal to Terraform that the managed resource
	// instance no longer exists and potentially should be recreated by calling
	// the SetId method with an empty string ("") parameter and without
	// returning an error.
	//
	// Data resources that are designed to return state for a singular
	// infrastructure component should conventionally return an error if that
	// infrastructure does not exist and omit any calls to the
	// SetId method.
	//
	// The interface{} parameter is the result of the Provider type
	// ConfigureFunc field execution. If the Provider does not define
	// a ConfigureFunc, this will be nil. This parameter is conventionally
	// used to store API clients and other provider instance specific data.
	//
	// The error return parameter, if not nil, will be converted into an error
	// diagnostic when passed back to Terraform.
	//
	// Deprecated: Use ReadContext or ReadWithoutTimeout instead. This
	// implementation does not support request cancellation initiated by
	// Terraform, such as a system or practitioner sending SIGINT (Ctrl-c).
	// This implementation also does not support warning diagnostics.
	Read ReadFunc

	// Update is called when the provider must update an instance of a
	// managed resource. This field is only valid when the Resource is a
	// managed resource. Only one of Update, UpdateContext, or
	// UpdateWithoutTimeout should be implemented.
	//
	// This implementation is optional. If omitted, all Schema must enable
	// the ForceNew field and any practitioner changes that would have
	// caused and update will instead destroy and recreate the infrastructure
	// compontent.
	//
	// The *ResourceData parameter contains the plan and state data for this
	// managed resource instance. The available data in the Get* methods is the
	// the proposed state, which is the merged data of the prior state,
	// practitioner configuration, and any CustomizeDiff field logic. The
	// available data for the GetChange* and HasChange* methods is the prior
	// state and proposed state.
	//
	// The interface{} parameter is the result of the Provider type
	// ConfigureFunc field execution. If the Provider does not define
	// a ConfigureFunc, this will be nil. This parameter is conventionally
	// used to store API clients and other provider instance specific data.
	//
	// The error return parameter, if not nil, will be converted into an error
	// diagnostic when passed back to Terraform.
	//
	// Deprecated: Use UpdateContext or UpdateWithoutTimeout instead. This
	// implementation does not support request cancellation initiated by
	// Terraform, such as a system or practitioner sending SIGINT (Ctrl-c).
	// This implementation also does not support warning diagnostics.
	Update UpdateFunc

	// Delete is called when the provider must destroy the instance of a
	// managed resource. This field is only valid when the Resource is a
	// managed resource. Only one of Delete, DeleteContext, or
	// DeleteWithoutTimeout should be implemented.
	//
	// The *ResourceData parameter contains the state data for this managed
	// resource instance.
	//
	// The interface{} parameter is the result of the Provider type
	// ConfigureFunc field execution. If the Provider does not define
	// a ConfigureFunc, this will be nil. This parameter is conventionally
	// used to store API clients and other provider instance specific data.
	//
	// The error return parameter, if not nil, will be converted into an error
	// diagnostic when passed back to Terraform.
	//
	// Deprecated: Use DeleteContext or DeleteWithoutTimeout instead. This
	// implementation does not support request cancellation initiated by
	// Terraform, such as a system or practitioner sending SIGINT (Ctrl-c).
	// This implementation also does not support warning diagnostics.
	Delete DeleteFunc

	// Exists is a function that is called to check if a resource still
	// exists. This field is only valid when the Resource is a managed
	// resource.
	//
	// If this returns false, then this will affect the diff
	// accordingly. If this function isn't set, it will not be called. You
	// can also signal existence in the Read method by calling d.SetId("")
	// if the Resource is no longer present and should be removed from state.
	// The *ResourceData passed to Exists should _not_ be modified.
	//
	// Deprecated: Remove in preference of ReadContext or ReadWithoutTimeout.
	Exists ExistsFunc

	// CreateContext is called when the provider must create a new instance of
	// a managed resource. This field is only valid when the Resource is a
	// managed resource. Only one of Create, CreateContext, or
	// CreateWithoutTimeout should be implemented.
	//
	// The Context parameter stores SDK information, such as loggers and
	// timeout deadlines. It also is wired to receive any cancellation from
	// Terraform such as a system or practitioner sending SIGINT (Ctrl-c).
	//
	// By default, CreateContext has a 20 minute timeout. Use the Timeouts
	// field to control the default duration or implement CreateWithoutTimeout
	// instead of CreateContext to remove the default timeout.
	//
	// The *ResourceData parameter contains the plan and state data for this
	// managed resource instance. The available data in the Get* methods is the
	// the proposed state, which is the merged data of the practitioner
	// configuration and any CustomizeDiff field logic.
	//
	// The SetId method must be called with a non-empty value for the managed
	// resource instance to be properly saved into the Terraform state and
	// avoid a "inconsistent result after apply" error.
	//
	// The interface{} parameter is the result of the Provider type
	// ConfigureFunc field execution. If the Provider does not define
	// a ConfigureFunc, this will be nil. This parameter is conventionally
	// used to store API clients and other provider instance specific data.
	//
	// The diagnostics return parameter, if not nil, can contain any
	// combination and multiple of warning and/or error diagnostics.
	CreateContext CreateContextFunc

	// ReadContext is called when the provider must refresh the state of a managed
	// resource instance or data resource instance. This field is only valid
	// when the Resource is a managed resource or data resource. Only one of
	// Read, ReadContext, or ReadWithoutTimeout should be implemented.
	//
	// The Context parameter stores SDK information, such as loggers and
	// timeout deadlines. It also is wired to receive any cancellation from
	// Terraform such as a system or practitioner sending SIGINT (Ctrl-c).
	//
	// By default, ReadContext has a 20 minute timeout. Use the Timeouts
	// field to control the default duration or implement ReadWithoutTimeout
	// instead of ReadContext to remove the default timeout.
	//
	// The *ResourceData parameter contains the state data for this managed
	// resource instance or data resource instance.
	//
	// Managed resources can signal to Terraform that the managed resource
	// instance no longer exists and potentially should be recreated by calling
	// the SetId method with an empty string ("") parameter and without
	// returning an error.
	//
	// Data resources that are designed to return state for a singular
	// infrastructure component should conventionally return an error if that
	// infrastructure does not exist and omit any calls to the
	// SetId method.
	//
	// The interface{} parameter is the result of the Provider type
	// ConfigureFunc field execution. If the Provider does not define
	// a ConfigureFunc, this will be nil. This parameter is conventionally
	// used to store API clients and other provider instance specific data.
	//
	// The diagnostics return parameter, if not nil, can contain any
	// combination and multiple of warning and/or error diagnostics.
	ReadContext ReadContextFunc

	// UpdateContext is called when the provider must update an instance of a
	// managed resource. This field is only valid when the Resource is a
	// managed resource. Only one of Update, UpdateContext, or
	// UpdateWithoutTimeout should be implemented.
	//
	// This implementation is optional. If omitted, all Schema must enable
	// the ForceNew field and any practitioner changes that would have
	// caused and update will instead destroy and recreate the infrastructure
	// compontent.
	//
	// The Context parameter stores SDK information, such as loggers and
	// timeout deadlines. It also is wired to receive any cancellation from
	// Terraform such as a system or practitioner sending SIGINT (Ctrl-c).
	//
	// By default, UpdateContext has a 20 minute timeout. Use the Timeouts
	// field to control the default duration or implement UpdateWithoutTimeout
	// instead of UpdateContext to remove the default timeout.
	//
	// The *ResourceData parameter contains the plan and state data for this
	// managed resource instance. The available data in the Get* methods is the
	// the proposed state, which is the merged data of the prior state,
	// practitioner configuration, and any CustomizeDiff field logic. The
	// available data for the GetChange* and HasChange* methods is the prior
	// state and proposed state.
	//
	// The interface{} parameter is the result of the Provider type
	// ConfigureFunc field execution. If the Provider does not define
	// a ConfigureFunc, this will be nil. This parameter is conventionally
	// used to store API clients and other provider instance specific data.
	//
	// The diagnostics return parameter, if not nil, can contain any
	// combination and multiple of warning and/or error diagnostics.
	UpdateContext UpdateContextFunc

	// DeleteContext is called when the provider must destroy the instance of a
	// managed resource. This field is only valid when the Resource is a
	// managed resource. Only one of Delete, DeleteContext, or
	// DeleteWithoutTimeout should be implemented.
	//
	// The Context parameter stores SDK information, such as loggers and
	// timeout deadlines. It also is wired to receive any cancellation from
	// Terraform such as a system or practitioner sending SIGINT (Ctrl-c).
	//
	// By default, DeleteContext has a 20 minute timeout. Use the Timeouts
	// field to control the default duration or implement DeleteWithoutTimeout
	// instead of DeleteContext to remove the default timeout.
	//
	// The *ResourceData parameter contains the state data for this managed
	// resource instance.
	//
	// The interface{} parameter is the result of the Provider type
	// ConfigureFunc field execution. If the Provider does not define
	// a ConfigureFunc, this will be nil. This parameter is conventionally
	// used to store API clients and other provider instance specific data.
	//
	// The diagnostics return parameter, if not nil, can contain any
	// combination and multiple of warning and/or error diagnostics.
	DeleteContext DeleteContextFunc

	// CreateWithoutTimeout is called when the provider must create a new
	// instance of a managed resource. This field is only valid when the
	// Resource is a managed resource. Only one of Create, CreateContext, or
	// CreateWithoutTimeout should be implemented.
	//
	// Most resources should prefer CreateContext with properly implemented
	// operation timeout values, however there are cases where operation
	// synchronization across concurrent resources is necessary in the resource
	// logic, such as a mutex, to prevent remote system errors. Since these
	// operations would have an indeterminate timeout that scales with the
	// number of resources, this allows resources to control timeout behavior.
	//
	// The Context parameter stores SDK information, such as loggers. It also
	// is wired to receive any cancellation from Terraform such as a system or
	// practitioner sending SIGINT (Ctrl-c).
	//
	// The *ResourceData parameter contains the plan and state data for this
	// managed resource instance. The available data in the Get* methods is the
	// the proposed state, which is the merged data of the practitioner
	// configuration and any CustomizeDiff field logic.
	//
	// The SetId method must be called with a non-empty value for the managed
	// resource instance to be properly saved into the Terraform state and
	// avoid a "inconsistent result after apply" error.
	//
	// The interface{} parameter is the result of the Provider type
	// ConfigureFunc field execution. If the Provider does not define
	// a ConfigureFunc, this will be nil. This parameter is conventionally
	// used to store API clients and other provider instance specific data.
	//
	// The diagnostics return parameter, if not nil, can contain any
	// combination and multiple of warning and/or error diagnostics.
	CreateWithoutTimeout CreateContextFunc

	// ReadWithoutTimeout is called when the provider must refresh the state of
	// a managed resource instance or data resource instance. This field is
	// only valid when the Resource is a managed resource or data resource.
	// Only one of Read, ReadContext, or ReadWithoutTimeout should be
	// implemented.
	//
	// Most resources should prefer ReadContext with properly implemented
	// operation timeout values, however there are cases where operation
	// synchronization across concurrent resources is necessary in the resource
	// logic, such as a mutex, to prevent remote system errors. Since these
	// operations would have an indeterminate timeout that scales with the
	// number of resources, this allows resources to control timeout behavior.
	//
	// The Context parameter stores SDK information, such as loggers. It also
	// is wired to receive any cancellation from Terraform such as a system or
	// practitioner sending SIGINT (Ctrl-c).
	//
	// The *ResourceData parameter contains the state data for this managed
	// resource instance or data resource instance.
	//
	// Managed resources can signal to Terraform that the managed resource
	// instance no longer exists and potentially should be recreated by calling
	// the SetId method with an empty string ("") parameter and without
	// returning an error.
	//
	// Data resources that are designed to return state for a singular
	// infrastructure component should conventionally return an error if that
	// infrastructure does not exist and omit any calls to the
	// SetId method.
	//
	// The interface{} parameter is the result of the Provider type
	// ConfigureFunc field execution. If the Provider does not define
	// a ConfigureFunc, this will be nil. This parameter is conventionally
	// used to store API clients and other provider instance specific data.
	//
	// The diagnostics return parameter, if not nil, can contain any
	// combination and multiple of warning and/or error diagnostics.
	ReadWithoutTimeout ReadContextFunc

	// UpdateWithoutTimeout is called when the provider must update an instance
	// of a managed resource. This field is only valid when the Resource is a
	// managed resource. Only one of Update, UpdateContext, or
	// UpdateWithoutTimeout should be implemented.
	//
	// Most resources should prefer UpdateContext with properly implemented
	// operation timeout values, however there are cases where operation
	// synchronization across concurrent resources is necessary in the resource
	// logic, such as a mutex, to prevent remote system errors. Since these
	// operations would have an indeterminate timeout that scales with the
	// number of resources, this allows resources to control timeout behavior.
	//
	// This implementation is optional. If omitted, all Schema must enable
	// the ForceNew field and any practitioner changes that would have
	// caused and update will instead destroy and recreate the infrastructure
	// compontent.
	//
	// The Context parameter stores SDK information, such as loggers. It also
	// is wired to receive any cancellation from Terraform such as a system or
	// practitioner sending SIGINT (Ctrl-c).
	//
	// The *ResourceData parameter contains the plan and state data for this
	// managed resource instance. The available data in the Get* methods is the
	// the proposed state, which is the merged data of the prior state,
	// practitioner configuration, and any CustomizeDiff field logic. The
	// available data for the GetChange* and HasChange* methods is the prior
	// state and proposed state.
	//
	// The interface{} parameter is the result of the Provider type
	// ConfigureFunc field execution. If the Provider does not define
	// a ConfigureFunc, this will be nil. This parameter is conventionally
	// used to store API clients and other provider instance specific data.
	//
	// The diagnostics return parameter, if not nil, can contain any
	// combination and multiple of warning and/or error diagnostics.
	UpdateWithoutTimeout UpdateContextFunc

	// DeleteWithoutTimeout is called when the provider must destroy the
	// instance of a managed resource. This field is only valid when the
	// Resource is a managed resource. Only one of Delete, DeleteContext, or
	// DeleteWithoutTimeout should be implemented.
	//
	// Most resources should prefer DeleteContext with properly implemented
	// operation timeout values, however there are cases where operation
	// synchronization across concurrent resources is necessary in the resource
	// logic, such as a mutex, to prevent remote system errors. Since these
	// operations would have an indeterminate timeout that scales with the
	// number of resources, this allows resources to control timeout behavior.
	//
	// The Context parameter stores SDK information, such as loggers. It also
	// is wired to receive any cancellation from Terraform such as a system or
	// practitioner sending SIGINT (Ctrl-c).
	//
	// The *ResourceData parameter contains the state data for this managed
	// resource instance.
	//
	// The interface{} parameter is the result of the Provider type
	// ConfigureFunc field execution. If the Provider does not define
	// a ConfigureFunc, this will be nil. This parameter is conventionally
	// used to store API clients and other provider instance specific data.
	//
	// The diagnostics return parameter, if not nil, can contain any
	// combination and multiple of warning and/or error diagnostics.
	DeleteWithoutTimeout DeleteContextFunc

	// CustomizeDiff is called after a difference (plan) has been generated
	// for the Resource and allows for customizations, such as setting values
	// not controlled by configuration, conditionally triggering resource
	// recreation, or implementing additional validation logic to abort a plan.
	// This field is only valid when the Resource is a managed resource.
	//
	// The Context parameter stores SDK information, such as loggers. It also
	// is wired to receive any cancellation from Terraform such as a system or
	// practitioner sending SIGINT (Ctrl-c).
	//
	// The *ResourceDiff parameter is similar to ResourceData but replaces the
	// Set method with other difference handling methods, such as SetNew,
	// SetNewComputed, and ForceNew. In general, only Schema with Computed
	// enabled can have those methods executed against them.
	//
	// The phases Terraform runs this in, and the state available via functions
	// like Get and GetChange, are as follows:
	//
	//  * New resource: One run with no state
	//  * Existing resource: One run with state
	//   * Existing resource, forced new: One run with state (before ForceNew),
	//     then one run without state (as if new resource)
	//  * Tainted resource: No runs (custom diff logic is skipped)
	//  * Destroy: No runs (standard diff logic is skipped on destroy diffs)
	//
	// This function needs to be resilient to support all scenarios.
	//
	// The interface{} parameter is the result of the Provider type
	// ConfigureFunc field execution. If the Provider does not define
	// a ConfigureFunc, this will be nil. This parameter is conventionally
	// used to store API clients and other provider instance specific data.
	//
	// The error return parameter, if not nil, will be converted into an error
	// diagnostic when passed back to Terraform.
	CustomizeDiff CustomizeDiffFunc

	// Importer is called when the provider must import an instance of a
	// managed resource. This field is only valid when the Resource is a
	// managed resource.
	//
	// If this is nil, then this resource does not support importing. If
	// this is non-nil, then it supports importing and ResourceImporter
	// must be validated. The validity of ResourceImporter is verified
	// by InternalValidate on Resource.
	Importer *ResourceImporter

	// If non-empty, this string is emitted as the details of a warning
	// diagnostic during validation (validate, plan, and apply operations).
	// This field is only valid when the Resource is a managed resource or
	// data resource.
	DeprecationMessage string

	// Timeouts configures the default time duration allowed before a create,
	// read, update, or delete operation is considered timed out, which returns
	// an error to practitioners. This field is only valid when the Resource is
	// a managed resource or data resource.
	//
	// When implemented, practitioners can add a timeouts configuration block
	// within their managed resource or data resource configuration to further
	// customize the create, read, update, or delete operation timeouts. For
	// example, a configuration may specify a longer create timeout for a
	// database resource due to its data size.
	//
	// The ResourceData that is passed to create, read, update, and delete
	// functionality can access the merged time duration of the Resource
	// default timeouts configured in this field and the practitioner timeouts
	// configuration via the Timeout method. Practitioner configuration
	// always overrides any default values set here, whether shorter or longer.
	Timeouts *ResourceTimeout

	// Description is used as the description for docs, the language server and
	// other user facing usage. It can be plain-text or markdown depending on the
	// global DescriptionKind setting. This field is valid for any Resource.
	Description string

	// UseJSONNumber should be set when state upgraders will expect
	// json.Numbers instead of float64s for numbers. This is added as a
	// toggle for backwards compatibility for type assertions, but should
	// be used in all new resources to avoid bugs with sufficiently large
	// user input. This field is only valid when the Resource is a managed
	// resource.
	//
	// See github.com/hashicorp/terraform-plugin-sdk/issues/655 for more
	// details.
	UseJSONNumber bool

	// EnableLegacyTypeSystemApplyErrors when enabled will prevent the SDK from
	// setting the legacy type system flag in the protocol during
	// ApplyResourceChange (Create, Update, and Delete) operations. Before
	// enabling this setting in a production release for a resource, the
	// resource should be exhaustively acceptance tested with the setting
	// enabled in an environment where it is easy to clean up resources,
	// potentially outside of Terraform, since these errors may be unavoidable
	// in certain cases.
	//
	// Disabling the legacy type system protocol flag is an unsafe operation
	// when using this SDK as there are certain unavoidable behaviors imposed
	// by the SDK, however this option is surfaced to allow provider developers
	// to try to discover fixable data inconsistency errors more easily.
	// Terraform, when encountering an enabled legacy type system protocol flag,
	// will demote certain schema and data consistency errors into warning logs
	// containing the text "legacy plugin SDK". Some errors for errant schema
	// definitions, such as when an attribute is not marked as Computed as
	// expected by Terraform, can only be resolved by migrating to
	// terraform-plugin-framework since that SDK does not impose behavior
	// changes with it enabled. However, data-based errors typically require
	// logic fixes that should be applicable for both SDKs to be resolved.
	EnableLegacyTypeSystemApplyErrors bool

	// EnableLegacyTypeSystemPlanErrors when enabled will prevent the SDK from
	// setting the legacy type system flag in the protocol during
	// PlanResourceChange operations. Before enabling this setting in a
	// production release for a resource, the resource should be exhaustively
	// acceptance tested with the setting enabled in an environment where it is
	// easy to clean up resources, potentially outside of Terraform, since these
	// errors may be unavoidable in certain cases.
	//
	// Disabling the legacy type system protocol flag is an unsafe operation
	// when using this SDK as there are certain unavoidable behaviors imposed
	// by the SDK, however this option is surfaced to allow provider developers
	// to try to discover fixable data inconsistency errors more easily.
	// Terraform, when encountering an enabled legacy type system protocol flag,
	// will demote certain schema and data consistency errors into warning logs
	// containing the text "legacy plugin SDK". Some errors for errant schema
	// definitions, such as when an attribute is not marked as Computed as
	// expected by Terraform, can only be resolved by migrating to
	// terraform-plugin-framework since that SDK does not impose behavior
	// changes with it enabled. However, data-based errors typically require
	// logic fixes that should be applicable for both SDKs to be resolved.
	EnableLegacyTypeSystemPlanErrors bool
}

// SchemaMap returns the schema information for this Resource whether it is
// defined via the SchemaFunc field or Schema field. The SchemaFunc field, if
// defined, takes precedence over the Schema field.
func (r *Resource) SchemaMap() map[string]*Schema {
	if r.SchemaFunc != nil {
		return r.SchemaFunc()
	}

	return r.Schema
}

// ShimInstanceStateFromValue converts a cty.Value to a
// terraform.InstanceState.
func (r *Resource) ShimInstanceStateFromValue(state cty.Value) (*terraform.InstanceState, error) {
	// Get the raw shimmed value. While this is correct, the set hashes don't
	// match those from the Schema.
	s := terraform.NewInstanceStateShimmedFromValue(state, r.SchemaVersion)

	// We now rebuild the state through the ResourceData, so that the set indexes
	// match what helper/schema expects.
	data, err := schemaMap(r.SchemaMap()).Data(s, nil)
	if err != nil {
		return nil, err
	}

	s = data.State()
	if s == nil {
		s = &terraform.InstanceState{}
	}
	return s, nil
}

// The following function types are of the legacy CRUD operations.
//
// Deprecated: Please use the context aware equivalents instead.
type CreateFunc func(*ResourceData, interface{}) error

// Deprecated: Please use the context aware equivalents instead.
type ReadFunc func(*ResourceData, interface{}) error

// Deprecated: Please use the context aware equivalents instead.
type UpdateFunc func(*ResourceData, interface{}) error

// Deprecated: Please use the context aware equivalents instead.
type DeleteFunc func(*ResourceData, interface{}) error

// Deprecated: Please use the context aware equivalents instead.
type ExistsFunc func(*ResourceData, interface{}) (bool, error)

// See Resource documentation.
type CreateContextFunc func(context.Context, *ResourceData, interface{}) diag.Diagnostics

// See Resource documentation.
type ReadContextFunc func(context.Context, *ResourceData, interface{}) diag.Diagnostics

// See Resource documentation.
type UpdateContextFunc func(context.Context, *ResourceData, interface{}) diag.Diagnostics

// See Resource documentation.
type DeleteContextFunc func(context.Context, *ResourceData, interface{}) diag.Diagnostics

// See Resource documentation.
type StateMigrateFunc func(
	int, *terraform.InstanceState, interface{}) (*terraform.InstanceState, error)

// Implementation of a single schema version state upgrade.
type StateUpgrader struct {
	// Version is the version schema that this Upgrader will handle, converting
	// it to Version+1.
	Version int

	// Type describes the schema that this function can upgrade. Type is
	// required to decode the schema if the state was stored in a legacy
	// flatmap format.
	Type cty.Type

	// Upgrade takes the JSON encoded state and the provider meta value, and
	// upgrades the state one single schema version. The provided state is
	// deocded into the default json types using a map[string]interface{}. It
	// is up to the StateUpgradeFunc to ensure that the returned value can be
	// encoded using the new schema.
	Upgrade StateUpgradeFunc
}

// Function signature for a schema version state upgrade handler.
//
// The Context parameter stores SDK information, such as loggers. It also
// is wired to receive any cancellation from Terraform such as a system or
// practitioner sending SIGINT (Ctrl-c).
//
// The map[string]interface{} parameter contains the previous schema version
// state data for a managed resource instance. The keys are top level attribute
// or block names mapped to values that can be type asserted similar to
// fetching values using the ResourceData Get* methods:
//
//   - TypeBool: bool
//   - TypeFloat: float
//   - TypeInt: int
//   - TypeList: []interface{}
//   - TypeMap: map[string]interface{}
//   - TypeSet: *Set
//   - TypeString: string
//
// In certain scenarios, the map may be nil, so checking for that condition
// upfront is recommended to prevent potential panics.
//
// The interface{} parameter is the result of the Provider type
// ConfigureFunc field execution. If the Provider does not define
// a ConfigureFunc, this will be nil. This parameter is conventionally
// used to store API clients and other provider instance specific data.
//
// The map[string]interface{} return parameter should contain the upgraded
// schema version state data for a managed resource instance. Values must
// align to the typing mentioned above.
type StateUpgradeFunc func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error)

// See Resource documentation.
type CustomizeDiffFunc func(context.Context, *ResourceDiff, interface{}) error

func (r *Resource) create(ctx context.Context, d *ResourceData, meta interface{}) diag.Diagnostics {
	if r.Create != nil {
		if err := r.Create(d, meta); err != nil {
			return diag.FromErr(err)
		}
		return nil
	}

	if r.CreateWithoutTimeout != nil {
		return r.CreateWithoutTimeout(ctx, d, meta)
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(TimeoutCreate))
	defer cancel()
	return r.CreateContext(ctx, d, meta)
}

func (r *Resource) read(ctx context.Context, d *ResourceData, meta interface{}) diag.Diagnostics {
	if r.Read != nil {
		if err := r.Read(d, meta); err != nil {
			return diag.FromErr(err)
		}
		return nil
	}

	if r.ReadWithoutTimeout != nil {
		return r.ReadWithoutTimeout(ctx, d, meta)
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(TimeoutRead))
	defer cancel()
	return r.ReadContext(ctx, d, meta)
}

func (r *Resource) update(ctx context.Context, d *ResourceData, meta interface{}) diag.Diagnostics {
	if r.Update != nil {
		if err := r.Update(d, meta); err != nil {
			return diag.FromErr(err)
		}
		return nil
	}

	if r.UpdateWithoutTimeout != nil {
		return r.UpdateWithoutTimeout(ctx, d, meta)
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(TimeoutUpdate))
	defer cancel()
	return r.UpdateContext(ctx, d, meta)
}

func (r *Resource) delete(ctx context.Context, d *ResourceData, meta interface{}) diag.Diagnostics {
	if r.Delete != nil {
		if err := r.Delete(d, meta); err != nil {
			return diag.FromErr(err)
		}
		return nil
	}

	if r.DeleteWithoutTimeout != nil {
		return r.DeleteWithoutTimeout(ctx, d, meta)
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(TimeoutDelete))
	defer cancel()
	return r.DeleteContext(ctx, d, meta)
}

// Apply creates, updates, and/or deletes a resource.
func (r *Resource) Apply(
	ctx context.Context,
	s *terraform.InstanceState,
	d *terraform.InstanceDiff,
	meta interface{}) (*terraform.InstanceState, diag.Diagnostics) {
	schema := schemaMap(r.SchemaMap())
	data, err := schema.Data(s, d)
	if err != nil {
		return s, diag.FromErr(err)
	}

	if s != nil && data != nil {
		data.providerMeta = s.ProviderMeta
	}

	// Instance Diff shoould have the timeout info, need to copy it over to the
	// ResourceData meta
	rt := ResourceTimeout{}
	if _, ok := d.Meta[TimeoutKey]; ok {
		if err := rt.DiffDecode(d); err != nil {
			logging.HelperSchemaError(ctx, "Error decoding ResourceTimeout", map[string]interface{}{logging.KeyError: err})
		}
	} else if s != nil {
		if _, ok := s.Meta[TimeoutKey]; ok {
			if err := rt.StateDecode(s); err != nil {
				logging.HelperSchemaError(ctx, "Error decoding ResourceTimeout", map[string]interface{}{logging.KeyError: err})
			}
		}
	} else {
		logging.HelperSchemaDebug(ctx, "No meta timeoutkey found in Apply()")
	}
	data.timeouts = &rt

	if s == nil {
		// The Terraform API dictates that this should never happen, but
		// it doesn't hurt to be safe in this case.
		s = new(terraform.InstanceState)
	}

	var diags diag.Diagnostics

	if d.Destroy || d.RequiresNew() {
		if s.ID != "" {
			// Destroy the resource since it is created
			logging.HelperSchemaTrace(ctx, "Calling downstream")
			diags = append(diags, r.delete(ctx, data, meta)...)
			logging.HelperSchemaTrace(ctx, "Called downstream")

			if diags.HasError() {
				return r.recordCurrentSchemaVersion(data.State()), diags
			}

			// Make sure the ID is gone.
			data.SetId("")
		}

		// If we're only destroying, and not creating, then return
		// now since we're done!
		if !d.RequiresNew() {
			return nil, diags
		}

		// Reset the data to be stateless since we just destroyed
		data, err = schema.Data(nil, d)
		if err != nil {
			return nil, append(diags, diag.FromErr(err)...)
		}

		// data was reset, need to re-apply the parsed timeouts
		data.timeouts = &rt
	}

	if data.Id() == "" {
		// We're creating, it is a new resource.
		data.MarkNewResource()
		logging.HelperSchemaTrace(ctx, "Calling downstream")
		diags = append(diags, r.create(ctx, data, meta)...)
		logging.HelperSchemaTrace(ctx, "Called downstream")
	} else {
		if !r.updateFuncSet() {
			return s, append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "doesn't support update",
			})
		}
		logging.HelperSchemaTrace(ctx, "Calling downstream")
		diags = append(diags, r.update(ctx, data, meta)...)
		logging.HelperSchemaTrace(ctx, "Called downstream")
	}

	return r.recordCurrentSchemaVersion(data.State()), diags
}

// Diff returns a diff of this resource.
func (r *Resource) Diff(
	ctx context.Context,
	s *terraform.InstanceState,
	c *terraform.ResourceConfig,
	meta interface{}) (*terraform.InstanceDiff, error) {

	t := &ResourceTimeout{}
	err := t.ConfigDecode(r, c)

	if err != nil {
		return nil, fmt.Errorf("[ERR] Error decoding timeout: %s", err)
	}

	instanceDiff, err := schemaMap(r.SchemaMap()).Diff(ctx, s, c, r.CustomizeDiff, meta, true)
	if err != nil {
		return instanceDiff, err
	}

	if instanceDiff != nil {
		if err := t.DiffEncode(instanceDiff); err != nil {
			logging.HelperSchemaError(ctx, "Error encoding timeout to instance diff", map[string]interface{}{logging.KeyError: err})
		}
	} else {
		logging.HelperSchemaDebug(ctx, "Instance Diff is nil in Diff()")
	}

	return instanceDiff, err
}

func (r *Resource) SimpleDiff(
	ctx context.Context,
	s *terraform.InstanceState,
	c *terraform.ResourceConfig,
	meta interface{}) (*terraform.InstanceDiff, error) {

	instanceDiff, err := schemaMap(r.SchemaMap()).Diff(ctx, s, c, r.CustomizeDiff, meta, false)
	if err != nil {
		return instanceDiff, err
	}

	if instanceDiff == nil {
		instanceDiff = terraform.NewInstanceDiff()
	}

	// Make sure the old value is set in each of the instance diffs.
	// This was done by the RequiresNew logic in the full legacy Diff.
	for k, attr := range instanceDiff.Attributes {
		if attr == nil {
			continue
		}
		if s != nil {
			attr.Old = s.Attributes[k]
		}
	}

	return instanceDiff, nil
}

// Validate validates the resource configuration against the schema.
func (r *Resource) Validate(c *terraform.ResourceConfig) diag.Diagnostics {
	diags := schemaMap(r.SchemaMap()).Validate(c)

	if r.DeprecationMessage != "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Deprecated Resource",
			Detail:   r.DeprecationMessage,
		})
	}

	return diags
}

// ReadDataApply loads the data for a data source, given a diff that
// describes the configuration arguments and desired computed attributes.
func (r *Resource) ReadDataApply(
	ctx context.Context,
	d *terraform.InstanceDiff,
	meta interface{},
) (*terraform.InstanceState, diag.Diagnostics) {
	// Data sources are always built completely from scratch
	// on each read, so the source state is always nil.
	data, err := schemaMap(r.SchemaMap()).Data(nil, d)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	logging.HelperSchemaTrace(ctx, "Calling downstream")
	diags := r.read(ctx, data, meta)
	logging.HelperSchemaTrace(ctx, "Called downstream")

	state := data.State()
	if state != nil && state.ID == "" {
		// Data sources can set an ID if they want, but they aren't
		// required to; we'll provide a placeholder if they don't,
		// to preserve the invariant that all resources have non-empty
		// ids.
		state.ID = "-"
	}

	return r.recordCurrentSchemaVersion(state), diags
}

// RefreshWithoutUpgrade reads the instance state, but does not call
// MigrateState or the StateUpgraders, since those are now invoked in a
// separate API call.
// RefreshWithoutUpgrade is part of the new plugin shims.
func (r *Resource) RefreshWithoutUpgrade(
	ctx context.Context,
	s *terraform.InstanceState,
	meta interface{}) (*terraform.InstanceState, diag.Diagnostics) {
	// If the ID is already somehow blank, it doesn't exist
	if s.ID == "" {
		return nil, nil
	}

	rt := ResourceTimeout{}
	if _, ok := s.Meta[TimeoutKey]; ok {
		if err := rt.StateDecode(s); err != nil {
			logging.HelperSchemaError(ctx, "Error decoding ResourceTimeout", map[string]interface{}{logging.KeyError: err})
		}
	}

	schema := schemaMap(r.SchemaMap())

	if r.Exists != nil {
		// Make a copy of data so that if it is modified it doesn't
		// affect our Read later.
		data, err := schema.Data(s, nil)
		if err != nil {
			return s, diag.FromErr(err)
		}
		data.timeouts = &rt

		if s != nil {
			data.providerMeta = s.ProviderMeta
		}

		logging.HelperSchemaTrace(ctx, "Calling downstream")
		exists, err := r.Exists(data, meta)
		logging.HelperSchemaTrace(ctx, "Called downstream")

		if err != nil {
			return s, diag.FromErr(err)
		}

		if !exists {
			return nil, nil
		}
	}

	data, err := schema.Data(s, nil)
	if err != nil {
		return s, diag.FromErr(err)
	}
	data.timeouts = &rt

	if s != nil {
		data.providerMeta = s.ProviderMeta
	}

	logging.HelperSchemaTrace(ctx, "Calling downstream")
	diags := r.read(ctx, data, meta)
	logging.HelperSchemaTrace(ctx, "Called downstream")

	state := data.State()
	if state != nil && state.ID == "" {
		state = nil
	}

	schema.handleDiffSuppressOnRefresh(ctx, s, state)
	return r.recordCurrentSchemaVersion(state), diags
}

func (r *Resource) createFuncSet() bool {
	return (r.Create != nil || r.CreateContext != nil || r.CreateWithoutTimeout != nil)
}

func (r *Resource) readFuncSet() bool {
	return (r.Read != nil || r.ReadContext != nil || r.ReadWithoutTimeout != nil)
}

func (r *Resource) updateFuncSet() bool {
	return (r.Update != nil || r.UpdateContext != nil || r.UpdateWithoutTimeout != nil)
}

func (r *Resource) deleteFuncSet() bool {
	return (r.Delete != nil || r.DeleteContext != nil || r.DeleteWithoutTimeout != nil)
}

// InternalValidate should be called to validate the structure
// of the resource.
//
// This should be called in a unit test for any resource to verify
// before release that a resource is properly configured for use with
// this library.
//
// Provider.InternalValidate() will automatically call this for all of
// the resources it manages, so you don't need to call this manually if it
// is part of a Provider.
func (r *Resource) InternalValidate(topSchemaMap schemaMap, writable bool) error {
	if r == nil {
		return errors.New("resource is nil")
	}

	if !writable {
		if r.createFuncSet() || r.updateFuncSet() || r.deleteFuncSet() {
			return fmt.Errorf("must not implement Create, Update or Delete")
		}

		// CustomizeDiff cannot be defined for read-only resources
		if r.CustomizeDiff != nil {
			return fmt.Errorf("cannot implement CustomizeDiff")
		}
	}

	schema := schemaMap(r.SchemaMap())
	tsm := topSchemaMap

	if r.isTopLevel() && writable {
		// All non-Computed attributes must be ForceNew if Update is not defined
		if !r.updateFuncSet() {
			nonForceNewAttrs := make([]string, 0)
			for k, v := range schema {
				if !v.ForceNew && !v.Computed {
					nonForceNewAttrs = append(nonForceNewAttrs, k)
				}
			}
			if len(nonForceNewAttrs) > 0 {
				return fmt.Errorf(
					"No Update defined, must set ForceNew on: %#v", nonForceNewAttrs)
			}
		} else {
			nonUpdateableAttrs := make([]string, 0)
			for k, v := range schema {
				if v.ForceNew || v.Computed && !v.Optional {
					nonUpdateableAttrs = append(nonUpdateableAttrs, k)
				}
			}
			updateableAttrs := len(schema) - len(nonUpdateableAttrs)
			if updateableAttrs == 0 {
				return fmt.Errorf(
					"All fields are ForceNew or Computed w/out Optional, Update is superfluous")
			}
		}

		tsm = schema

		// Destroy, and Read are required
		if !r.readFuncSet() {
			return fmt.Errorf("Read must be implemented")
		}
		if !r.deleteFuncSet() {
			return fmt.Errorf("Delete must be implemented")
		}

		// If we have an importer, we need to verify the importer.
		if r.Importer != nil {
			if err := r.Importer.InternalValidate(); err != nil {
				return err
			}
		}

		if f, ok := tsm["id"]; ok {
			// if there is an explicit ID, validate it...
			err := validateResourceID(f)
			if err != nil {
				return err
			}
		}

		for k := range tsm {
			if isReservedResourceFieldName(k) {
				return fmt.Errorf("%s is a reserved field name", k)
			}
		}
	}

	lastVersion := -1
	for _, u := range r.StateUpgraders {
		if lastVersion >= 0 && u.Version-lastVersion > 1 {
			return fmt.Errorf("missing schema version between %d and %d", lastVersion, u.Version)
		}

		if u.Version >= r.SchemaVersion {
			return fmt.Errorf("StateUpgrader version %d is >= current version %d", u.Version, r.SchemaVersion)
		}

		if !u.Type.IsObjectType() {
			return fmt.Errorf("StateUpgrader %d type is not cty.Object", u.Version)
		}

		if u.Upgrade == nil {
			return fmt.Errorf("StateUpgrader %d missing StateUpgradeFunc", u.Version)
		}

		lastVersion = u.Version
	}

	if lastVersion >= 0 && lastVersion != r.SchemaVersion-1 {
		return fmt.Errorf("missing StateUpgrader between %d and %d", lastVersion, r.SchemaVersion)
	}

	// Data source
	if r.isTopLevel() && !writable {
		tsm = schema
		for k := range tsm {
			if isReservedDataSourceFieldName(k) {
				return fmt.Errorf("%s is a reserved field name", k)
			}
		}
	}

	if r.SchemaFunc != nil && r.Schema != nil {
		return fmt.Errorf("SchemaFunc and Schema should not both be set")
	}

	// check context funcs are not set alongside their nonctx counterparts
	if r.CreateContext != nil && r.Create != nil {
		return fmt.Errorf("CreateContext and Create should not both be set")
	}
	if r.ReadContext != nil && r.Read != nil {
		return fmt.Errorf("ReadContext and Read should not both be set")
	}
	if r.UpdateContext != nil && r.Update != nil {
		return fmt.Errorf("UpdateContext and Update should not both be set")
	}
	if r.DeleteContext != nil && r.Delete != nil {
		return fmt.Errorf("DeleteContext and Delete should not both be set")
	}

	// check context funcs are not set alongside their without timeout counterparts
	if r.CreateContext != nil && r.CreateWithoutTimeout != nil {
		return fmt.Errorf("CreateContext and CreateWithoutTimeout should not both be set")
	}
	if r.ReadContext != nil && r.ReadWithoutTimeout != nil {
		return fmt.Errorf("ReadContext and ReadWithoutTimeout should not both be set")
	}
	if r.UpdateContext != nil && r.UpdateWithoutTimeout != nil {
		return fmt.Errorf("UpdateContext and UpdateWithoutTimeout should not both be set")
	}
	if r.DeleteContext != nil && r.DeleteWithoutTimeout != nil {
		return fmt.Errorf("DeleteContext and DeleteWithoutTimeout should not both be set")
	}

	// check non-context funcs are not set alongside the context without timeout counterparts
	if r.Create != nil && r.CreateWithoutTimeout != nil {
		return fmt.Errorf("Create and CreateWithoutTimeout should not both be set")
	}
	if r.Read != nil && r.ReadWithoutTimeout != nil {
		return fmt.Errorf("Read and ReadWithoutTimeout should not both be set")
	}
	if r.Update != nil && r.UpdateWithoutTimeout != nil {
		return fmt.Errorf("Update and UpdateWithoutTimeout should not both be set")
	}
	if r.Delete != nil && r.DeleteWithoutTimeout != nil {
		return fmt.Errorf("Delete and DeleteWithoutTimeout should not both be set")
	}

	return schema.InternalValidate(tsm)
}

func isReservedDataSourceFieldName(name string) bool {
	for _, reservedName := range ReservedDataSourceFields {
		if name == reservedName {
			return true
		}
	}
	return false
}

func validateResourceID(s *Schema) error {
	if s.Type != TypeString {
		return fmt.Errorf(`the "id" attribute must be of TypeString`)
	}

	if s.Required {
		return fmt.Errorf(`the "id" attribute cannot be marked Required`)
	}

	// ID should at least be computed. If unspecified it will be set to Computed and Optional,
	// but Optional is unnecessary if undesired.
	if !s.Computed {
		return fmt.Errorf(`the "id" attribute must be marked Computed`)
	}
	return nil
}

func isReservedResourceFieldName(name string) bool {
	for _, reservedName := range ReservedResourceFields {
		if name == reservedName {
			return true
		}
	}

	return false
}

// Data returns a ResourceData struct for this Resource. Each return value
// is a separate copy and can be safely modified differently.
//
// The data returned from this function has no actual affect on the Resource
// itself (including the state given to this function).
//
// This function is useful for unit tests and ResourceImporter functions.
func (r *Resource) Data(s *terraform.InstanceState) *ResourceData {
	result, err := schemaMap(r.SchemaMap()).Data(s, nil)
	if err != nil {
		// At the time of writing, this isn't possible (Data never returns
		// non-nil errors). We panic to find this in the future if we have to.
		// I don't see a reason for Data to ever return an error.
		panic(err)
	}

	// load the Resource timeouts
	result.timeouts = r.Timeouts
	if result.timeouts == nil {
		result.timeouts = &ResourceTimeout{}
	}

	// Set the schema version to latest by default
	result.meta = map[string]interface{}{
		"schema_version": strconv.Itoa(r.SchemaVersion),
	}

	return result
}

// TestResourceData Yields a ResourceData filled with this resource's schema for use in unit testing
//
// TODO: May be able to be removed with the above ResourceData function.
func (r *Resource) TestResourceData() *ResourceData {
	return &ResourceData{
		schema: r.SchemaMap(),
	}
}

// Returns true if the resource is "top level" i.e. not a sub-resource.
func (r *Resource) isTopLevel() bool {
	// TODO: This is a heuristic; replace with a definitive attribute?
	return (r.createFuncSet() || r.readFuncSet())
}

func (r *Resource) recordCurrentSchemaVersion(
	state *terraform.InstanceState) *terraform.InstanceState {
	if state != nil && r.SchemaVersion > 0 {
		if state.Meta == nil {
			state.Meta = make(map[string]interface{})
		}
		state.Meta["schema_version"] = strconv.Itoa(r.SchemaVersion)
	}
	return state
}

// Noop is a convenience implementation of resource function which takes
// no action and returns no error.
func Noop(*ResourceData, interface{}) error {
	return nil
}

// NoopContext is a convenience implementation of context aware resource function which takes
// no action and returns no error.
func NoopContext(context.Context, *ResourceData, interface{}) diag.Diagnostics {
	return nil
}

// RemoveFromState is a convenience implementation of a resource function
// which sets the resource ID to empty string (to remove it from state)
// and returns no error.
func RemoveFromState(d *ResourceData, _ interface{}) error {
	d.SetId("")
	return nil
}
