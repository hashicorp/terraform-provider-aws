// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_location_map")
func DataSourceMap() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceMapRead,
		Schema: map[string]*schema.Schema{
			names.AttrConfiguration: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"style": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrCreateTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"map_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"map_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceMapRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	input := &locationservice.DescribeMapInput{}

	if v, ok := d.GetOk("map_name"); ok {
		input.MapName = aws.String(v.(string))
	}

	output, err := conn.DescribeMapWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Location Service Map: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting Location Service Map: empty response")
	}

	d.SetId(aws.StringValue(output.MapName))

	if output.Configuration != nil {
		d.Set(names.AttrConfiguration, []interface{}{flattenConfiguration(output.Configuration)})
	} else {
		d.Set(names.AttrConfiguration, nil)
	}

	d.Set(names.AttrCreateTime, aws.TimeValue(output.CreateTime).Format(time.RFC3339))
	d.Set(names.AttrDescription, output.Description)
	d.Set("map_arn", output.MapArn)
	d.Set("map_name", output.MapName)
	d.Set("update_time", aws.TimeValue(output.UpdateTime).Format(time.RFC3339))
	d.Set(names.AttrTags, KeyValueTags(ctx, output.Tags).IgnoreAWS().IgnoreConfig(meta.(*conns.AWSClient).IgnoreTagsConfig).Map())

	return diags
}
