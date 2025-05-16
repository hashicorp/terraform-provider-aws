// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func arnIdentityResourceImporter() *schema.ResourceImporter {
	return &_arnIdentityResourceImporter
}

var (
	_arnIdentityResourceImporter = schema.ResourceImporter{
		StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
			arnARN, err := arn.Parse(rd.Id())
			if err != nil {
				return nil, fmt.Errorf("could not parse import ID %q as ARN: %s", rd.Id(), err)
			}

			if region, ok := rd.GetOk("region"); ok {
				if region != arnARN.Region {
					return nil, fmt.Errorf("the region passed for import %q does not match the region %q in the ARN %q", region, arnARN.Region, rd.Id())
				}
			} else {
				rd.Set("region", arnARN.Region)
			}

			return []*schema.ResourceData{rd}, nil
		},
	}
)
