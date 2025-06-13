// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/importer"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// TODO: Needs a better name
type ImportByIdentityer interface {
	SetIdentitySpec(identity inttypes.Identity)
}

var _ ImportByIdentityer = &WithImportByARN{}

type WithImportByARN struct {
	identity inttypes.Identity
}

func (w *WithImportByARN) SetIdentitySpec(identity inttypes.Identity) {
	w.identity = identity
}

func (w *WithImportByARN) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
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
	if w.identity.IsGlobalResource {
		importer.GlobalARN(ctx, client, request, &w.identity, response)
	} else if w.identity.IsGlobalARNFormat {
		importer.RegionalARNWithGlobalFormat(ctx, client, request, &w.identity, response)
	} else {
		importer.RegionalARN(ctx, client, request, &w.identity, response)
	}
}
