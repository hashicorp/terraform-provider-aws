// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// withNoOpUpdate is intended to be embedded in resources which have no need of a custom Update method.
// For example, resources where only `tags` can be updated and that is handled via transparent tagging.
type withNoOpUpdate[T any] struct{}

func (w *withNoOpUpdate[T]) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var t T
	response.Diagnostics.Append(request.Plan.Get(ctx, &t)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &t)...)
}
