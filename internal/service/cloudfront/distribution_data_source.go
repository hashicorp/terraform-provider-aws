// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudfront_distribution", name="Distribution")
// @Tags(identifierAttribute="arn")
func dataSourceDistribution() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDistributionRead,

		Schema: map[string]*schema.Schema{
			"aliases": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrHostedZoneID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Required: true,
			},
			"in_progress_validation_batches": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"web_acl_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceDistributionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	id := d.Get(names.AttrID).(string)
	output, err := findDistributionByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Distribution (%s): %s", id, err)
	}

	d.SetId(aws.ToString(output.Distribution.Id))
	distribution := output.Distribution
	distributionConfig := distribution.DistributionConfig
	if aliases := distributionConfig.Aliases; aliases != nil {
		d.Set("aliases", aliases.Items)
	}
	d.Set(names.AttrARN, distribution.ARN)
	d.Set(names.AttrDomainName, distribution.DomainName)
	d.Set(names.AttrEnabled, distributionConfig.Enabled)
	d.Set("etag", output.ETag)
	d.Set(names.AttrHostedZoneID, meta.(*conns.AWSClient).CloudFrontDistributionHostedZoneID(ctx))
	d.Set("in_progress_validation_batches", distribution.InProgressInvalidationBatches)
	d.Set("last_modified_time", aws.String(distribution.LastModifiedTime.String()))
	d.Set(names.AttrStatus, distribution.Status)
	d.Set("web_acl_id", distributionConfig.WebACLId)

	return diags
}
