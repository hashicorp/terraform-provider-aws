// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediaconvert

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_media_convert_queue", name="Queue")
func DataSourceQueue() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceQueueRead,

		Schema: map[string]*schema.Schema{
			"arn": {
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn, err := GetAccountClient(ctx, meta.(*conns.AWSClient))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Media Convert Account Client: %s", err)
	}

	id := d.Get("id").(string)
	queue, err := FindQueueByName(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Media Convert Queue (%s): %s", id, err)
	}

	arn, name := aws.StringValue(queue.Arn), aws.StringValue(queue.Name)
	d.SetId(name)
	d.Set("arn", arn)
	d.Set("name", name)
	d.Set("status", queue.Status)

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Media Convert Queue (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
