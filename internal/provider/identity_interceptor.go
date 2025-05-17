// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/importer"
)

func arnIdentityResourceImporter(attrName string) *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
			if err := importer.ARN(ctx, rd, attrName); err != nil {
				return nil, err
			}

			return []*schema.ResourceData{rd}, nil
		},
	}
}
