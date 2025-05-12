// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// WithImportByGlobalARN is intended to be embedded in global resources which import state via the "arn" attribute.
// See https://developer.hashicorp.com/terraform/plugin/framework/resources/import.
type WithImportByGlobalARN struct{}

func (w *WithImportByGlobalARN) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	_, err := arn.Parse(request.ID)
	if err != nil {
		response.Diagnostics.AddError(
			"Invalid Resource Import ID Value",
			"The import ID could not be parsed as an ARN.\n\n"+
				fmt.Sprintf("Value: %q\nError: %s", request.ID, err),
		)
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), request, response)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), request.ID)...)
}
