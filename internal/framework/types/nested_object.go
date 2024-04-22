// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// NestedObjectType extends the Type interface for types that represent nested Objects.
// The nested objects are either a single object or a collection of objects (List or Set).
// It isn't generic on the Go struct type as it's referenced within AutoFlEx.
type NestedObjectType interface {
	attr.Type

	// NewObjectPtr returns a new, empty value as an object pointer (Go *struct).
	NewObjectPtr(context.Context) (any, diag.Diagnostics)

	// NullValue returns a Null Value.
	NullValue(context.Context) (attr.Value, diag.Diagnostics)

	// ValueFromObjectPtr returns a Value given an object pointer (Go *struct).
	ValueFromObjectPtr(context.Context, any) (attr.Value, diag.Diagnostics)
}

// NestedObjectCollectionType extends the NestedObjectType interface for types that represent
// collections (Lists or Sets) of nested Objects.
// It isn't generic on the Go struct type as it's referenced within AutoFlEx.
type NestedObjectCollectionType interface {
	NestedObjectType

	// NewObjectSlice returns a new value as an object slice (Go []*struct).
	NewObjectSlice(context.Context, int, int) (any, diag.Diagnostics)

	// ValueFromObjectSlice returns a Value given an object pointer (Go []*struct).
	ValueFromObjectSlice(context.Context, any) (attr.Value, diag.Diagnostics)
}

// NestedObjectValue extends the Value interface for values that represent nested Objects.
// The nested objects are either a single object or a collection of objects (List or Set).
// It isn't generic on the Go struct type as it's referenced within AutoFlEx.
type NestedObjectValue interface {
	attr.Value

	// ToObjectPtr returns the value as an object pointer (Go *struct).
	ToObjectPtr(context.Context) (any, diag.Diagnostics)
}

// NestedObjectCollectionValue extends the NestedObjectValue interface for values that represent collections of nested Objects.
// It isn't generic on the Go struct type as it's referenced within AutoFlEx.
type NestedObjectCollectionValue interface {
	NestedObjectValue

	// ToObjectSlice returns the value as an object slice (Go []*struct).
	ToObjectSlice(context.Context) (any, diag.Diagnostics)
}

// valueWithElements extends the Value interface for values that have an Elements method.
type valueWithElements interface {
	attr.Value

	Elements() []attr.Value
}
