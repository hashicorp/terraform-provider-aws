// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package privatestate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// PrivateState defines an interface for managing provider-defined resource private state data.
type PrivateState interface {
	// GetKey returns the private state data associated with the given key.
	GetKey(context.Context, string) ([]byte, diag.Diagnostics)
	// SetKey sets the private state data at the given key.
	SetKey(context.Context, string, []byte) diag.Diagnostics
}
