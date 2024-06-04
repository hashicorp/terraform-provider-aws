// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediaconvert

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

// @SDKDataSource("aws_media_convert_queue", name="Queue")
// @Tags(identifierAttribute="arn")
func dataSourceQueue() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceQueueRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertClient(ctx)

	id := d.Get(names.AttrID).(string)
	queue, err := findQueueByName(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Media Convert Queue (%s): %s", id, err)
	}

	name := aws.ToString(queue.Name)
	d.SetId(name)
	d.Set(names.AttrARN, queue.Arn)
	d.Set(names.AttrName, name)
	d.Set(names.AttrStatus, queue.Status)

	return diags
}
