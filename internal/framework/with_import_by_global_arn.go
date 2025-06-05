// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// WithImportByGlobalARN is intended to be embedded in global resources which import state via the "arn" attribute.
// See https://developer.hashicorp.com/terraform/plugin/framework/resources/import.
type WithImportByGlobalARN struct {
	attributeName  string
	duplicateAttrs []string
}

func (w *WithImportByGlobalARN) SetARNAttributeName(attr string, duplicateAttrs []string) {
	w.attributeName = attr
	w.duplicateAttrs = duplicateAttrs
}

func (w *WithImportByGlobalARN) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	if request.ID != "" {
		_, err := arn.Parse(request.ID)
		if err != nil {
			response.Diagnostics.AddError(
				"Invalid Resource Import ID Value",
				"The import ID could not be parsed as an ARN.\n\n"+
					fmt.Sprintf("Value: %q\nError: %s", request.ID, err),
			)
			return
		}

		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(w.attributeName), request.ID)...)
		for _, attr := range w.duplicateAttrs {
			response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(attr), request.ID)...)
		}

		return
	}

	if identity := request.Identity; identity != nil {
		arnPath := path.Root(w.attributeName)
		var arnVal string
		identity.GetAttribute(ctx, arnPath, &arnVal)

		_, err := arn.Parse(arnVal)
		if err != nil {
			response.Diagnostics.AddAttributeError(
				arnPath,
				"Invalid Import Attribute Value",
				fmt.Sprintf("Import attribute %q is not a valid ARN, got: %s", arnPath, arnVal),
			)
			return
		}

		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(w.attributeName), arnVal)...)
		for _, attr := range w.duplicateAttrs {
			response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(attr), arnVal)...)
		}
	}
}
