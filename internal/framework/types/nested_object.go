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

	// NewPtr returns a new, empty value as a pointer (usually a Go struct pointer).
	NewPtr(context.Context) (any, diag.Diagnostics)

	// NullValue returns a Null Value.
	NullValue(context.Context) (attr.Value, diag.Diagnostics)

	// ValueFromPtr returns a Value given a pointer (usually a Go struct pointer).
	ValueFromPtr(context.Context, any) (attr.Value, diag.Diagnostics)
}

// NestedObjectValue extends the Value interface for values that represent nested Objects.
type NestedObjectValue interface {
	attr.Value

	// ToPtr returns the value as a pointer (usually a Go struct pointer).
	ToPtr(context.Context) (any, diag.Diagnostics)
}
