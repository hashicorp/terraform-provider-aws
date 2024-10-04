// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// WithNoOpDelete is intended to be embedded in resources which have no need of a custom Delete method.
type WithNoOpDelete struct{}

func (w *WithNoOpDelete) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
}
