// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"channel_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameStreamKey = "Stream Key Data Source"
)

func dataSourceStreamKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IVSConn(ctx)

	channelArn := d.Get("channel_arn").(string)

	out, err := FindStreamKeyByChannelID(ctx, conn, channelArn)
	if err != nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionReading, DSNameStreamKey, channelArn, err)
	}

	d.SetId(aws.StringValue(out.Arn))

	d.Set("arn", out.Arn)
	d.Set("channel_arn", out.ChannelArn)
	d.Set("value", out.Value)

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	//lintignore:AWSR002
	if err := d.Set("tags", KeyValueTags(ctx, out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.AppendDiagError(diags, names.IVS, create.ErrActionSetting, DSNameStreamKey, d.Id(), err)
	}

	return diags
}
