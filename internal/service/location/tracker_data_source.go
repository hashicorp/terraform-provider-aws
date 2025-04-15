// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_location_tracker", name="Tracker")
func DataSourceTracker() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTrackerRead,
		Schema: map[string]*schema.Schema{
			names.AttrCreateTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"position_filtering": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"tracker_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tracker_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTrackerRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	input := &location.DescribeTrackerInput{
		TrackerName: aws.String(d.Get("tracker_name").(string)),
	}

	output, err := conn.DescribeTracker(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Location Service Tracker: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting Location Service Tracker: empty response")
	}

	d.SetId(aws.ToString(output.TrackerName))
	d.Set(names.AttrCreateTime, aws.ToTime(output.CreateTime).Format(time.RFC3339))
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrKMSKeyID, output.KmsKeyId)
	d.Set("position_filtering", output.PositionFiltering)
	d.Set(names.AttrTags, KeyValueTags(ctx, output.Tags).IgnoreAWS().IgnoreConfig(meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)).Map())
	d.Set("tracker_arn", output.TrackerArn)
	d.Set("tracker_name", output.TrackerName)
	d.Set("update_time", aws.ToTime(output.UpdateTime).Format(time.RFC3339))

	return diags
}
