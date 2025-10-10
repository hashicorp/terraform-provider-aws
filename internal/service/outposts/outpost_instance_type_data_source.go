// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package outposts

import (
	"context"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_outposts_outpost_instance_type", name="Outpost Instance Type")
func dataSourceOutpostInstanceType() *schema.Resource {
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

func dataSourceOutpostInstanceTypeRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OutpostsClient(ctx)

	input := &outposts.GetOutpostInstanceTypesInput{
		OutpostId: aws.String(d.Get(names.AttrARN).(string)), // Accepts both ARN and ID; prefer ARN which is more common
	}

	var outpostID string
	var foundInstanceTypes []string

	pages := outposts.NewGetOutpostInstanceTypesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "getting Outpost Instance Types: %s", err)
		}

		outpostID = aws.ToString(page.OutpostId)

		for _, outputInstanceType := range page.InstanceTypes {
			foundInstanceTypes = append(foundInstanceTypes, aws.ToString(outputInstanceType.InstanceType))
		}
	}

	if len(foundInstanceTypes) == 0 {
		return sdkdiag.AppendErrorf(diags, "no Outpost Instance Types found matching criteria; try different search")
	}

	var resultInstanceType string

	// Check requested instance type
	if v, ok := d.GetOk(names.AttrInstanceType); ok {
		if slices.Contains(foundInstanceTypes, v.(string)) {
			resultInstanceType = v.(string)
		}
	}

	// Search preferred instance types in their given order and set result
	// instance type for first match found
	if l := d.Get("preferred_instance_types").([]any); len(l) > 0 {
		for _, elem := range l {
			preferredInstanceType, ok := elem.(string)

			if !ok {
				continue
			}

			if slices.Contains(foundInstanceTypes, preferredInstanceType) {
				resultInstanceType = preferredInstanceType
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
