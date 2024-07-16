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

// @SDKDataSource("aws_location_place_index")
func DataSourcePlaceIndex() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePlaceIndexRead,
		Schema: map[string]*schema.Schema{
			names.AttrCreateTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_source_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"intended_use": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"index_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"index_name": {
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

func dataSourcePlaceIndexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	input := &locationservice.DescribePlaceIndexInput{}

	if v, ok := d.GetOk("index_name"); ok {
		input.IndexName = aws.String(v.(string))
	}

	output, err := conn.DescribePlaceIndexWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Location Service Place Index: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting Location Service Place Index: empty response")
	}

	d.SetId(aws.StringValue(output.IndexName))
	d.Set(names.AttrCreateTime, aws.TimeValue(output.CreateTime).Format(time.RFC3339))
	d.Set("data_source", output.DataSource)

	if output.DataSourceConfiguration != nil {
		d.Set("data_source_configuration", []interface{}{flattenDataSourceConfiguration(output.DataSourceConfiguration)})
	} else {
		d.Set("data_source_configuration", nil)
	}

	d.Set(names.AttrDescription, output.Description)
	d.Set("index_arn", output.IndexArn)
	d.Set("index_name", output.IndexName)
	d.Set(names.AttrTags, KeyValueTags(ctx, output.Tags).IgnoreAWS().IgnoreConfig(meta.(*conns.AWSClient).IgnoreTagsConfig).Map())
	d.Set("update_time", aws.TimeValue(output.UpdateTime).Format(time.RFC3339))

	return diags
}
