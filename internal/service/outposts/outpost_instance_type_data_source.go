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

// @SDKDataSource("aws_outposts_outpost_instance_type")
func DataSourceOutpostInstanceType() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOutpostInstanceTypeRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrInstanceType: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"preferred_instance_types"},
			},
			"preferred_instance_types": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{names.AttrInstanceType},
				Elem:          &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceOutpostInstanceTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OutpostsConn(ctx)

	input := &outposts.GetOutpostInstanceTypesInput{
		OutpostId: aws.String(d.Get(names.AttrARN).(string)), // Accepts both ARN and ID; prefer ARN which is more common
	}

	var outpostID string
	var foundInstanceTypes []string

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
			foundInstanceTypes = append(foundInstanceTypes, aws.StringValue(outputInstanceType.InstanceType))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if len(foundInstanceTypes) == 0 {
		return sdkdiag.AppendErrorf(diags, "no Outpost Instance Types found matching criteria; try different search")
	}

	var resultInstanceType string

	// Check requested instance type
	if v, ok := d.GetOk(names.AttrInstanceType); ok {
		for _, foundInstanceType := range foundInstanceTypes {
			if foundInstanceType == v.(string) {
				resultInstanceType = v.(string)
				break
			}
		}
	}

	// Search preferred instance types in their given order and set result
	// instance type for first match found
	if l := d.Get("preferred_instance_types").([]interface{}); len(l) > 0 {
		for _, elem := range l {
			preferredInstanceType, ok := elem.(string)

			if !ok {
				continue
			}

			for _, foundInstanceType := range foundInstanceTypes {
				if foundInstanceType == preferredInstanceType {
					resultInstanceType = preferredInstanceType
					break
				}
			}

			if resultInstanceType != "" {
				break
			}
		}
	}

	if resultInstanceType == "" && len(foundInstanceTypes) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple Outpost Instance Types found matching criteria; try different search")
	}

	if resultInstanceType == "" && len(foundInstanceTypes) == 1 {
		resultInstanceType = foundInstanceTypes[0]
	}

	if resultInstanceType == "" {
		return sdkdiag.AppendErrorf(diags, "no Outpost Instance Types found matching criteria; try different search")
	}

	d.Set(names.AttrInstanceType, resultInstanceType)

	d.SetId(outpostID)

	return diags
}
