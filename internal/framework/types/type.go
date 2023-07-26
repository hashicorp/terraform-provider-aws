// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

// TypeWithNewPtr extends the Type interface to include the ability to
// return a new, empty value as a pointer (usually a Go struct pointer).
type TypeWithNewPtr interface {
	attr.Type

	// NewPtr returns a new, empty value as a pointer (usually a Go struct pointer).
	NewPtr(context.Context) any
}
