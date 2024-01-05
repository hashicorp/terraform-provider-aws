// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_cloudfront_distribution")
func DataSourceDistribution() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDistributionRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"aliases": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"in_progress_validation_batches": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"web_acl_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchema(),
		},
	}
}

func dataSourceDistributionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindDistributionByID(ctx, conn, d.Get("id").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting CloudFront Distribution (%s): %s", d.Id(), err)
	}

	d.SetId(aws.StringValue(output.Distribution.Id))
	d.Set("etag", output.ETag)
	if distribution := output.Distribution; distribution != nil {
		d.Set("arn", distribution.ARN)
		d.Set("domain_name", distribution.DomainName)
		d.Set("in_progress_validation_batches", distribution.InProgressInvalidationBatches)
		d.Set("last_modified_time", aws.String(distribution.LastModifiedTime.String()))
		d.Set("status", distribution.Status)
		d.Set("hosted_zone_id", meta.(*conns.AWSClient).CloudFrontDistributionHostedZoneID())
		if distributionConfig := distribution.DistributionConfig; distributionConfig != nil {
			d.Set("enabled", distributionConfig.Enabled)
			d.Set("web_acl_id", distributionConfig.WebACLId)
			if aliases := distributionConfig.Aliases; aliases != nil {
				d.Set("aliases", aliases.Items)
			}
		}
	}
	tags, err := listTags(ctx, conn, d.Get("arn").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for CloudFront Distribution (%s): %s", d.Id(), err)
	}
	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
