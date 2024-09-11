// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ivs_stream_key")
func DataSourceStreamKey() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceStreamKeyRead,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"channel_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrValue: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameStreamKey = "Stream Key Data Source"
)

func dataSourceStreamKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IVSClient(ctx)

	channelArn := d.Get("channel_arn").(string)

	out, err := FindStreamKeyByChannelID(ctx, conn, channelArn)
	if err != nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionReading, DSNameStreamKey, channelArn, err)
	}

	d.SetId(aws.ToString(out.Arn))

	d.Set(names.AttrARN, out.Arn)
	d.Set("channel_arn", out.ChannelArn)
	d.Set(names.AttrValue, out.Value)

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	//lintignore:AWSR002
	if err := d.Set(names.AttrTags, KeyValueTags(ctx, out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionSetting, DSNameStreamKey, d.Id(), err)
	}

	return diags
}
