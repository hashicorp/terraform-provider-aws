// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/importer"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// WithImportRegionalSingleton is intended to be embedded in resources which import state via the "region" attribute.
// See https://developer.hashicorp.com/terraform/plugin/framework/resources/import.
type WithImportRegionalSingleton struct {
	identity inttypes.Identity
}

func (w *WithImportRegionalSingleton) SetIdentitySpec(identity inttypes.Identity) {
	w.identity = identity
}

func (w *WithImportRegionalSingleton) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	client := importer.Client(ctx)
	if client == nil {
		response.Diagnostics.AddError(
			"Unexpected Error",
			"An unexpected error occurred while importing a resource. "+
				"This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+
				"Importer context was nil.",
		)
		return
	}
	importer.RegionalSingleton(ctx, client, request, &w.identity, response)
}
