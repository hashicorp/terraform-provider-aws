// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

type EphemeralResourceWithConfigure struct {
	withMeta
}

func (e *EphemeralResourceWithConfigure) Configure(_ context.Context, request ephemeral.ConfigureRequest, _ *ephemeral.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		e.meta = v
	}
}
