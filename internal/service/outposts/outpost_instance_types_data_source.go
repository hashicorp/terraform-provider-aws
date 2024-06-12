// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package outposts

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_outposts_outpost_instance_types")
func DataSourceOutpostInstanceTypes() *schema.Resource {
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

func dataSourceOutpostInstanceTypesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OutpostsConn(ctx)

	input := &outposts.GetOutpostInstanceTypesInput{
		OutpostId: aws.String(d.Get(names.AttrARN).(string)), // Accepts both ARN and ID; prefer ARN which is more common
	}

	var outpostID string
	var instanceTypes []string

	for {
		output, err := conn.GetOutpostInstanceTypesWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "getting Outpost Instance Types: %s", err)
		}

		if output == nil {
			break
		}

		outpostID = aws.StringValue(output.OutpostId)

		for _, outputInstanceType := range output.InstanceTypes {
			instanceTypes = append(instanceTypes, aws.StringValue(outputInstanceType.InstanceType))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if err := d.Set("instance_types", instanceTypes); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting instance_types: %s", err)
	}

	d.SetId(outpostID)

	return diags
}
