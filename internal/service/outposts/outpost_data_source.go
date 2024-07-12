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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_outposts_outpost")
func DataSourceOutpost() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOutpostRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"lifecycle_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"site_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"site_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"supported_hardware_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceOutpostRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OutpostsConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	input := &outposts.ListOutpostsInput{}

	var results []*outposts.Outpost

	err := conn.ListOutpostsPagesWithContext(ctx, input, func(page *outposts.ListOutpostsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, outpost := range page.Outposts {
			if outpost == nil {
				continue
			}

			if v, ok := d.GetOk(names.AttrID); ok && v.(string) != aws.StringValue(outpost.OutpostId) {
				continue
			}

			if v, ok := d.GetOk(names.AttrName); ok && v.(string) != aws.StringValue(outpost.Name) {
				continue
			}

			if v, ok := d.GetOk(names.AttrARN); ok && v.(string) != aws.StringValue(outpost.OutpostArn) {
				continue
			}

			if v, ok := d.GetOk(names.AttrOwnerID); ok && v.(string) != aws.StringValue(outpost.OwnerId) {
				continue
			}

			results = append(results, outpost)
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Outposts Outposts: %s", err)
	}

	if len(results) == 0 {
		return sdkdiag.AppendErrorf(diags, "no Outposts Outpost found matching criteria; try different search")
	}

	if len(results) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple Outposts Outpost found matching criteria; try different search")
	}

	outpost := results[0]

	d.SetId(aws.StringValue(outpost.OutpostId))
	d.Set(names.AttrARN, outpost.OutpostArn)
	d.Set(names.AttrAvailabilityZone, outpost.AvailabilityZone)
	d.Set("availability_zone_id", outpost.AvailabilityZoneId)
	d.Set(names.AttrDescription, outpost.Description)
	d.Set("lifecycle_status", outpost.LifeCycleStatus)
	d.Set(names.AttrName, outpost.Name)
	d.Set(names.AttrOwnerID, outpost.OwnerId)
	d.Set("site_arn", outpost.SiteArn)
	d.Set("site_id", outpost.SiteId)
	d.Set("supported_hardware_type", outpost.SupportedHardwareType)

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, outpost.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
