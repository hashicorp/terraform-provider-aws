// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/importer"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

var _ ImportByIdentityer = &WithImportByParameterizedIdentity{}

// WithImportByParameterizedIdentity is intended to be embedded in resources which import state via a parameterized Identity.
// See https://developer.hashicorp.com/terraform/plugin/framework/resources/import.
type WithImportByParameterizedIdentity struct {
	identity   inttypes.Identity
	importSpec inttypes.FrameworkImport
}

func (w *WithImportByParameterizedIdentity) SetIdentitySpec(identity inttypes.Identity, importSpec inttypes.FrameworkImport) {
	w.identity = identity
	w.importSpec = importSpec
}

func (w *WithImportByParameterizedIdentity) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
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

	if w.identity.IsSingleParameter {
		if w.identity.IsGlobalResource {
			importer.GlobalSingleParameterized(ctx, client, request, &w.identity, &w.importSpec, response)
			return
		} else {
			importer.RegionalSingleParameterized(ctx, client, request, &w.identity, &w.importSpec, response)
			return
		}
	} else {
		if w.identity.IsGlobalResource {
			importer.GlobalMultipleParameterized(ctx, client, request, &w.identity, &w.importSpec, response)
			return
		} else {
			importer.RegionalMultipleParameterized(ctx, client, request, &w.identity, &w.importSpec, response)
			return
		}
	}
}
