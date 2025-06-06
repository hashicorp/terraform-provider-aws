// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// WithImportByGlobalARN is intended to be embedded in global resources which import state via the "arn" attribute.
// See https://developer.hashicorp.com/terraform/plugin/framework/resources/import.
type WithImportByGlobalARN struct {
	importByARN
}

func (w *WithImportByGlobalARN) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	w.importState(ctx, request, response)
}
