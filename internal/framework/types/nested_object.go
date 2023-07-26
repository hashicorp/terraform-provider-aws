// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// NestedObjectValue extends the Value interface for values that represent nested Objects.
type NestedObjectValue interface {
	attr.Value

	// ToPtr returns the value as a pointer (usually a Go struct pointer).
	ToPtr(context.Context) (any, diag.Diagnostics)
}
