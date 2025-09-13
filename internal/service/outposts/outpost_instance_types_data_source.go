// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package outposts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_outposts_outpost_instance_types", name="Outpost Instance Types")
func dataSourceOutpostInstanceTypes() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOutpostInstanceTypesRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"instance_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceOutpostInstanceTypesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OutpostsClient(ctx)

	input := &outposts.GetOutpostInstanceTypesInput{
		OutpostId: aws.String(d.Get(names.AttrARN).(string)), // Accepts both ARN and ID; prefer ARN which is more common
	}

	var outpostID string
	var instanceTypes []string

	pages := outposts.NewGetOutpostInstanceTypesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "getting Outpost Instance Types: %s", err)
		}

		outpostID = aws.ToString(page.OutpostId)

		for _, outputInstanceType := range page.InstanceTypes {
			instanceTypes = append(instanceTypes, aws.ToString(outputInstanceType.InstanceType))
		}
	}

	if err := d.Set("instance_types", instanceTypes); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting instance_types: %s", err)
	}

	d.SetId(outpostID)

	return diags
}
