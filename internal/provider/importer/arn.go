// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ARN(ctx context.Context, rd *schema.ResourceData, attrName string) error {
	arnARN, err := arn.Parse(rd.Id())
	if err != nil {
		return fmt.Errorf("could not parse import ID %q as ARN: %s", rd.Id(), err)
	}
	rd.Set(attrName, rd.Id())

	if region, ok := rd.GetOk("region"); ok {
		if region != arnARN.Region {
			return fmt.Errorf("the region passed for import %q does not match the region %q in the ARN %q", region, arnARN.Region, rd.Id())
		}
	} else {
		rd.Set("region", arnARN.Region)
	}

	return nil
}
