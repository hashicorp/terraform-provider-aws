// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// MappedObjectType extends the Type interface for types that represent mapped Objects.
type MappedObjectType interface {
	attr.Type

	// NewObjectPtr returns a new, empty value as an object pointer (Go *struct).
	NewObjectPtr(context.Context) (any, diag.Diagnostics)

	// NullValue returns a Null Value.
	NullValue(context.Context) (attr.Value, diag.Diagnostics)
}

// MappedObjectValue extends the Value interface for values that represent mapped Objects.
type MappedObjectValue interface {
	attr.Value

	// ToObjectPtr returns the value as an object pointer (Go *struct).
	ToObjectMap(context.Context) (any, diag.Diagnostics)
}

// valueWithValues extends the Value interface for values that have a map Elements method.
type valueWithMapElements interface {
	attr.Value

	Elements() map[string]attr.Value
}
