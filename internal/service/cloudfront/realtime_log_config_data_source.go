// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudfront_realtime_log_config", name="Real-time Log Config")
func dataSourceRealtimeLogConfig() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRealtimeLogConfigRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEndpoint: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kinesis_stream_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrRoleARN: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrStreamARN: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"stream_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"fields": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"sampling_rate": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceRealtimeLogConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	name := d.Get(names.AttrName).(string)
	logConfig, err := findRealtimeLogConfigByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Real-time Log Config (%s): %s", name, err)
	}

	arn := aws.ToString(logConfig.ARN)
	d.SetId(arn)
	d.Set(names.AttrARN, arn)
	if err := d.Set(names.AttrEndpoint, flattenEndPoints(logConfig.EndPoints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint: %s", err)
	}
	d.Set("fields", logConfig.Fields)
	d.Set(names.AttrName, logConfig.Name)
	d.Set("sampling_rate", logConfig.SamplingRate)

	return diags
}

func findRealtimeLogConfigByName(ctx context.Context, conn *cloudfront.Client, name string) (*awstypes.RealtimeLogConfig, error) {
	input := &cloudfront.GetRealtimeLogConfigInput{
		Name: aws.String(name),
	}

	return findRealtimeLogConfig(ctx, conn, input)
}
