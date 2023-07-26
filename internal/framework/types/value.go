// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ValueWithToPtr extends the Value interface to include the ability to
// return the value as a pointer (usually a Go struct pointer).
type ValueWithToPtr interface {
	attr.Value

	// ToPtr returns the value as a pointer (usually a Go struct pointer).
	ToPtr(context.Context) (any, diag.Diagnostics)
}
