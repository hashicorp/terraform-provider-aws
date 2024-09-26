// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// WithNoOpRead is intended to be embedded in resources which have no need of a custom Read method.
type WithNoOpRead struct{}

func (w *WithNoOpRead) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
}
