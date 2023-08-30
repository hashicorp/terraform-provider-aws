// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfjson

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/zclconf/go-cty/cty"
)

// MetadataFunctionsFormatVersionConstraints defines the versions of the JSON
// metadata functions format that are supported by this package.
var MetadataFunctionsFormatVersionConstraints = "~> 1.0"

// MetadataFunctions is the top-level object returned when exporting function
// signatures
type MetadataFunctions struct {
	// The version of the format. This should always match the
	// MetadataFunctionsFormatVersionConstraints in this package, else
	// unmarshaling will fail.
	FormatVersion string `json:"format_version"`

	// The signatures of the functions available in a Terraform version.
	Signatures map[string]*FunctionSignature `json:"function_signatures,omitempty"`
}

// Validate checks to ensure that MetadataFunctions is present, and the
// version matches the version supported by this library.
func (f *MetadataFunctions) Validate() error {
	if f == nil {
		return errors.New("metadata functions data is nil")
	}

	if f.FormatVersion == "" {
		return errors.New("unexpected metadata functions data, format version is missing")
	}

	constraint, err := version.NewConstraint(MetadataFunctionsFormatVersionConstraints)
	if err != nil {
		return fmt.Errorf("invalid version constraint: %w", err)
	}

	version, err := version.NewVersion(f.FormatVersion)
	if err != nil {
		return fmt.Errorf("invalid format version %q: %w", f.FormatVersion, err)
	}

	if !constraint.Check(version) {
		return fmt.Errorf("unsupported metadata functions format version: %q does not satisfy %q",
			version, constraint)
	}

	return nil
}

func (f *MetadataFunctions) UnmarshalJSON(b []byte) error {
	type rawFunctions MetadataFunctions
	var functions rawFunctions

	err := json.Unmarshal(b, &functions)
	if err != nil {
		return err
	}

	*f = *(*MetadataFunctions)(&functions)

	return f.Validate()
}

// FunctionSignature represents a function signature.
type FunctionSignature struct {
	// Description is an optional human-readable description
	// of the function
	Description string `json:"description,omitempty"`

	// ReturnType is the ctyjson representation of the function's
	// return types based on supplying all parameters using
	// dynamic types. Functions can have dynamic return types.
	ReturnType cty.Type `json:"return_type"`

	// Parameters describes the function's fixed positional parameters.
	Parameters []*FunctionParameter `json:"parameters,omitempty"`

	// VariadicParameter describes the function's variadic
	// parameter if it is supported.
	VariadicParameter *FunctionParameter `json:"variadic_parameter,omitempty"`
}

// FunctionParameter represents a parameter to a function.
type FunctionParameter struct {
	// Name is an optional name for the argument.
	Name string `json:"name,omitempty"`

	// Description is an optional human-readable description
	// of the argument
	Description string `json:"description,omitempty"`

	// IsNullable is true if null is acceptable value for the argument
	IsNullable bool `json:"is_nullable,omitempty"`

	// A type that any argument for this parameter must conform to.
	Type cty.Type `json:"type"`
}
