// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
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
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
			}
		},
	}
}

func dataSourceWebACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)
	diags = validateDataSourceInput(d)
	if diags.HasError() {
		return diags
	}

	if v, ok := d.GetOk("resource_arn"); ok {
		return getWebAclDataSourceByArn(ctx, conn, v.(string), d, diags)
	}

	name := d.Get("name").(string)

	var foundWebACL *wafv2.WebACLSummary
	input := &wafv2.ListWebACLsInput{
		Scope: aws.String(d.Get("scope").(string)),
		Limit: aws.Int64(100),
	}

	for {
		resp, err := conn.ListWebACLsWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACLs: %s", err)
		}

		if resp == nil || resp.WebACLs == nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACLs")
		}

		for _, webACL := range resp.WebACLs {
			if aws.StringValue(webACL.Name) == name {
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
	d.Set("description", foundWebACL.Description)

	return diags
}

func getWebAclDataSourceByArn(ctx context.Context, conn *wafv2.WAFV2, arn string, d *schema.ResourceData, diags diag.Diagnostics) diag.Diagnostics {
	input := &wafv2.GetWebACLForResourceInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetWebACLForResourceWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACLs: %s", err)
	}

	if output == nil || output.WebACL == nil {
		return sdkdiag.AppendErrorf(diags, "WAFv2 WebACL not found for source: %s", arn)
	}

	d.SetId(*output.WebACL.Id)
	d.Set("arn", *output.WebACL.ARN)
	d.Set("name", *output.WebACL.Name)
	d.Set("description", *output.WebACL.Description)
	return diags
}

func validateDataSourceInput(d *schema.ResourceData) diag.Diagnostics {
	if _, ok := d.GetOk("resource_arn"); ok {
		return nil
	}

	if _, ok := d.GetOk("name"); !ok {
		return diag.Errorf("name is required")
	}

	if v, ok := d.GetOk("scope"); !ok {
		return diag.Errorf("scope is required")
	} else {
		_, err := validation.StringInSlice(wafv2.Scope_Values(), false)(v, "scope")
		if len(err) == 0 {
			return nil
		}
		return diag.FromErr(err[0])
	}
}
