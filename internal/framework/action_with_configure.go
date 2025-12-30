// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

type ActionWithConfigure struct {
	withMeta
}

// Metadata should return the full name of the action, such as
// aws_lambda_invoke.
func (*ActionWithConfigure) Metadata(_ context.Context, request action.MetadataRequest, response *action.MetadataResponse) {
	// This method is implemented in the wrappers.
	panic("not implemented") // lintignore:R009
}

// Configure enables provider-level data or clients to be set in the
// provider-defined Action type.
func (a *ActionWithConfigure) Configure(_ context.Context, request action.ConfigureRequest, _ *action.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		a.meta = v
	}
}
