// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// NestedObjectType extends the Type interface for types that represent nested Objects.
type NestedObjectType interface {
	attr.Type

	// NewObjectPtr returns a new, empty value as an object pointer (a Go struct pointer).
	NewObjectPtr(context.Context) (any, diag.Diagnostics)

	// NullValue returns a Null Value.
	NullValue(context.Context) (attr.Value, diag.Diagnostics)

	// ValueFromObjectPtr returns a Value given an object pointer (a Go struct pointer).
	ValueFromObjectPtr(context.Context, any) (attr.Value, diag.Diagnostics)
}

// NestedObjectValue extends the Value interface for values that represent nested Objects.
type NestedObjectValue interface {
	attr.Value

	// ToObjectPtr returns the value as an object pointer (a Go struct pointer).
	ToObjectPtr(context.Context) (any, diag.Diagnostics)
}
