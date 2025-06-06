// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/fwprovider/importer"
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
	if w.identity.IsGlobalResource {
		importer.GlobalARN(ctx, request, &w.identity, response)
	} else {
		importer.RegionalARN(ctx, request, &w.identity, response)
	}
}
