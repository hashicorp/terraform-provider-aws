// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ObjectMapType extends the Type interface for types that represent mapped Objects.
type ObjectMapType interface {
	attr.Type

	// NullValue returns a Null Value.
	NullValue(context.Context) (attr.Value, diag.Diagnostics)

	// ValueFromMap returns a Value given an object map
	ValueFromMap(context.Context, basetypes.MapValue) (basetypes.MapValuable, diag.Diagnostics)

	// ValueFromRawMap returns a Value given an object map
	ValueFromRawMap(context.Context, any) (basetypes.MapValuable, diag.Diagnostics)

	// NewObjectMap returns a new Value
	New(context.Context) (any, diag.Diagnostics)
}

// ObjectMapValue extends the Value interface for values that represent mapped Objects.
type ObjectMapValue interface {
	attr.Value

	// ToObjectPtr returns the value as an object pointer (Go *struct).
	ToObjectMap(context.Context) (any, diag.Diagnostics)
}

// valueWithValues extends the Value interface for values that have a map Elements method.
type valueWithMapElements interface {
	attr.Value

	Elements() map[string]attr.Value
}
