// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_cloudfront_origin_access_control")
func DataSourceOriginAccessControl() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOriginAccessControlRead,

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Computed: true,
				Default:  nil,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"origin_access_control_origin_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_behavior": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceOriginAccessControlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn(ctx)
	id := d.Get("id").(string)
	params := &cloudfront.GetOriginAccessControlInput{
		Id: aws.String(id),
	}

	resp, err := conn.GetOriginAccessControlWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Origin Access Control (%s): %s", id, err)
	}

	// Update other attributes outside DistributionConfig
	d.SetId(aws.StringValue(resp.OriginAccessControl.Id))
	d.Set("etag", resp.ETag)
	d.Set("description", resp.OriginAccessControl.OriginAccessControlConfig.Description)
	d.Set("name", resp.OriginAccessControl.OriginAccessControlConfig.Name)
	d.Set("origin_access_control_origin_type", resp.OriginAccessControl.OriginAccessControlConfig.OriginAccessControlOriginType)
	d.Set("signing_behavior", resp.OriginAccessControl.OriginAccessControlConfig.SigningBehavior)
	d.Set("signing_protocol", resp.OriginAccessControl.OriginAccessControlConfig.SigningProtocol)
	return diags
}
