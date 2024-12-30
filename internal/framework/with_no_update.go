// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// WithNoUpdate is intended to be embedded in resources which cannot be updated.
type WithNoUpdate struct{}

func (w *WithNoUpdate) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.Append(diag.NewErrorDiagnostic("not supported", "This resource's Update method should not have been called"))
}
