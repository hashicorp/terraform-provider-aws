// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfprotov5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Function describes the definition of a function. Result must be defined.
type Function struct {
	// Parameters is the ordered list of positional function parameters.
	Parameters []*FunctionParameter

	// VariadicParameter is an optional final parameter which accepts zero or
	// more argument values, in which Terraform will send an ordered list of the
	// parameter type.
	VariadicParameter *FunctionParameter

	// Return is the function result.
	Return *FunctionReturn

	// Summary is the shortened human-readable documentation for the function.
	Summary string

	// Description is the longer human-readable documentation for the function.
	Description string

	// DescriptionKind indicates the formatting and encoding that the
	// Description field is using.
	DescriptionKind StringKind

	// DeprecationMessage is the human-readable documentation if the function
	// is deprecated. This message should be practitioner oriented to explain
	// how their configuration should be updated.
	DeprecationMessage string
}

// FunctionMetadata describes metadata for a function in the GetMetadata RPC.
type FunctionMetadata struct {
	// Name is the name of the function.
	Name string
}

// FunctionParameter describes the definition of a function parameter. Type must
// be defined.
type FunctionParameter struct {
	// AllowNullValue when enabled denotes that a null argument value can be
	// passed to the provider. When disabled, Terraform returns an error if the
	// argument value is null.
	AllowNullValue bool

	// AllowUnknownValues when enabled denotes that any unknown argument value
	// (recursively checked for collections) can be passed to the provider. When
	// disabled and an unknown value is present, Terraform skips the function
	// call entirely and returns an unknown value result from the function.
	AllowUnknownValues bool

	// Description is the human-readable documentation for the parameter.
	Description string

	// DescriptionKind indicates the formatting and encoding that the
	// Description field is using.
	DescriptionKind StringKind

	// Name is the human-readable display name for the parameter. Parameters
	// are by definition positional and this name is only used in documentation.
	Name string

	// Type indicates the type of data the parameter expects.
	Type tftypes.Type
}

// FunctionReturn describes the definition of a function result. Type must be
// defined.
type FunctionReturn struct {
	// Type indicates the type of return data.
	Type tftypes.Type
}

// FunctionServer is an interface containing the methods a function
// implementation needs to fill.
type FunctionServer interface {
	// CallFunction is called when Terraform wants to execute the logic of a
	// function referenced in the configuration.
	CallFunction(context.Context, *CallFunctionRequest) (*CallFunctionResponse, error)

	// GetFunctions is called when Terraform wants to lookup which functions a
	// provider supports when not calling GetProviderSchema.
	GetFunctions(context.Context, *GetFunctionsRequest) (*GetFunctionsResponse, error)
}

// CallFunctionRequest is the request Terraform sends when it wants to execute
// the logic of function referenced in the configuration.
type CallFunctionRequest struct {
	// Name is the function name being called.
	Name string

	// Arguments is the configuration value of each argument the practitioner
	// supplied for the function call. The ordering and value of each element
	// matches the function parameters and their associated type.  If the
	// function definition includes a final variadic parameter, its value is an
	// ordered list of the variadic parameter type.
	Arguments []*DynamicValue
}

// CallFunctionResponse is the response from the provider with the result of
// executing the logic of the function.
type CallFunctionResponse struct {
	// Error reports errors related to the execution of the
	// function logic. Returning a nil error indicates a successful response
	// with no errors presented to practitioners.
	Error *FunctionError

	// Result is the return value from the called function, matching the result
	// type in the function definition.
	Result *DynamicValue
}

// GetFunctionsRequest is the request Terraform sends when it wants to lookup
// which functions a provider supports when not calling GetProviderSchema.
type GetFunctionsRequest struct{}

// GetFunctionsResponse is the response from the provider about the implemented
// functions.
type GetFunctionsResponse struct {
	// Diagnostics report errors or warnings related to the provider
	// implementation. Returning an empty slice indicates a successful response
	// with no warnings or errors presented to practitioners.
	Diagnostics []*Diagnostic

	// Functions is a map of function names to their definition.
	//
	// Unlike data resources and managed resources, the name should NOT be
	// prefixed with the provider name and an underscore. Configuration
	// references to functions use a separate namespacing syntax that already
	// includes the provider name.
	Functions map[string]*Function
}
