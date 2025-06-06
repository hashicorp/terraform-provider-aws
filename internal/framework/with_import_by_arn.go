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

type importByARN struct {
	attributeName  string
	duplicateAttrs []string
}

func (i *importByARN) SetARNAttributeName(attr string, duplicateAttrs []string) {
	i.attributeName = attr
	i.duplicateAttrs = duplicateAttrs
}

func (i *importByARN) importState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) arn.ARN {
	var (
		arnARN arn.ARN
		arnVal string
	)
	if arnVal = request.ID; arnVal != "" {
		var err error
		arnARN, err = arn.Parse(arnVal)
		if err != nil {
			response.Diagnostics.AddError(
				"Invalid Resource Import ID Value",
				"The import ID could not be parsed as an ARN.\n\n"+
					fmt.Sprintf("Value: %q\nError: %s", arnVal, err),
			)
			return arn.ARN{}
		}
	} else if identity := request.Identity; identity != nil {
		arnPath := path.Root(i.attributeName)
		identity.GetAttribute(ctx, arnPath, &arnVal)

		var err error
		arnARN, err = arn.Parse(arnVal)
		if err != nil {
			response.Diagnostics.AddAttributeError(
				arnPath,
				"Invalid Import Attribute Value",
				fmt.Sprintf("Import attribute %q is not a valid ARN, got: %s", arnPath, arnVal),
			)
			return arn.ARN{}
		}
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(i.attributeName), arnVal)...)
	for _, attr := range i.duplicateAttrs {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(attr), arnVal)...)
	}

	if identity := response.Identity; identity != nil {
		response.Diagnostics.Append(identity.SetAttribute(ctx, path.Root(i.attributeName), arnVal)...)
	}

	return arnARN
}
