// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_wafv2_web_acl")
func DataSourceWebACL() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWebACLRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"description": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"name": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"resource_arn": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"scope": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(wafv2.Scope_Values(), false),
				},
			}
		},
	}
}

func dataSourceWebACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)
	diags = validateWebACLDataSourceInput(d)
	if diags.HasError() {
		return diags
	}

	if v, ok := d.GetOk("resource_arn"); ok {
		if d.Get("scope").(string) == "REGIONAL" {
			return getWebACLDataSourceByResourceARN(ctx, d, conn, diags, v.(string))
		}
		arn, diags := getWebACLFromCloudfrontARN(ctx, d, meta.(*conns.AWSClient).CloudFrontConn(ctx), diags)
		if diags.HasError() {
			return diags
		}
		return getWebACLDataSourceByField(ctx, d, conn, diags, *arn, false)
	}

	name := d.Get("name").(string)

	return getWebACLDataSourceByField(ctx, d, conn, diags, name, true)
}

func getWebACLFromCloudfrontARN(ctx context.Context, d *schema.ResourceData, conn *cloudfront.CloudFront, diags diag.Diagnostics) (*string, diag.Diagnostics) {
	arn := d.Get("resource_arn").(string)
	arnParts := strings.Split(arn, "/")
	id := arnParts[len(arnParts)-1]

	input := &cloudfront.GetDistributionConfigInput{
		Id: aws.String(id),
	}

	output, err := conn.GetDistributionConfigWithContext(ctx, input)
	if err != nil {
		return nil, sdkdiag.AppendErrorf(diags, "reading CloudFront distribution: %s", err)
	}

	return output.DistributionConfig.WebACLId, nil
}

func getWebACLDataSourceByField(ctx context.Context, d *schema.ResourceData, conn *wafv2.WAFV2, diags diag.Diagnostics, check string, isName bool) diag.Diagnostics {
	var foundWebACL *wafv2.WebACLSummary
	input := &wafv2.ListWebACLsInput{
		Scope: aws.String(d.Get("scope").(string)),
		Limit: aws.Int64(100),
	}

	name := d.Get("name").(string)

	for {
		resp, err := conn.ListWebACLsWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACLs: %s", err)
		}

		if resp == nil || resp.WebACLs == nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACLs")
		}

		for _, webACL := range resp.WebACLs {
			if (isName && aws.StringValue(webACL.Name) == check) || (!isName && aws.StringValue(webACL.ARN) == check) {
				foundWebACL = webACL
				break
			}
		}

		if resp.NextMarker == nil || foundWebACL != nil {
			break
		}
		input.NextMarker = resp.NextMarker
	}

	if foundWebACL == nil {
		return sdkdiag.AppendErrorf(diags, "WAFv2 WebACL not found for name: %s", name)
	}

	d.SetId(aws.StringValue(foundWebACL.Id))
	d.Set("arn", foundWebACL.ARN)
	d.Set("name", foundWebACL.Name)
	d.Set("description", foundWebACL.Description)

	return diags
}

func getWebACLDataSourceByResourceARN(ctx context.Context, d *schema.ResourceData, conn *wafv2.WAFV2, diags diag.Diagnostics, arn string) diag.Diagnostics {
	input := &wafv2.GetWebACLForResourceInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetWebACLForResourceWithContext(ctx, input)
	if err != nil {
		sdkdiag.AppendWarningf(diags, "reading WAFv2 WebACLs: %s", err)
	}

	if output == nil || output.WebACL == nil {
		return diags
	}

	d.SetId(aws.StringValue(output.WebACL.Id))
	d.Set("arn", output.WebACL.ARN)
	d.Set("name", output.WebACL.Name)
	d.Set("description", output.WebACL.Description)
	return diags
}

func validateWebACLDataSourceInput(d *schema.ResourceData) diag.Diagnostics {
	if _, ok := d.GetOk("resource_arn"); ok {
		return nil
	}

	if _, ok := d.GetOk("name"); ok {
		return nil
	}

	return diag.Errorf("name or resource_arn are required")
}
