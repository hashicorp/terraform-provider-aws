// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// TypeWithValueFromPtr extends the Type interface to include the ability to
// return a Value converted from a pointer (usually a Go struct pointer).
type TypeWithValueFromPtr interface {
	attr.Type

	// ValueFromPtr returns a Value converted from a pointer (usually a Go struct pointer).
	ValueFromPtr(context.Context, any) (attr.Value, diag.Diagnostics)
}
