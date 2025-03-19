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

// Metadata should return the full name of the ephemeral resource, such as
// examplecloud_thing.
func (*EphemeralResourceWithConfigure) Metadata(_ context.Context, request ephemeral.MetadataRequest, response *ephemeral.MetadataResponse) {
	// This method is implemented in the wrappers.
	panic("not implemented") // lintignore:R009
}

// Configure enables provider-level data or clients to be set in the
// provider-defined EphemeralResource type.
func (e *EphemeralResourceWithConfigure) Configure(_ context.Context, request ephemeral.ConfigureRequest, _ *ephemeral.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		e.meta = v
	}
}
