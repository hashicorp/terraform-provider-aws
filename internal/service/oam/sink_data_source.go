// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oam

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

// @SDKDataSource("aws_oam_sink")
func DataSourceSink() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSinkRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sink_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sink_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameSink = "Sink Data Source"
)

func dataSourceSinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	sinkIdentifier := d.Get("sink_identifier").(string)

	out, err := findSinkByID(ctx, conn, sinkIdentifier)
	if err != nil {
		return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionReading, DSNameSink, sinkIdentifier, err)
	}

	d.SetId(aws.ToString(out.Arn))

	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrName, out.Name)
	d.Set("sink_id", out.Id)

	tags, err := listTags(ctx, conn, d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionReading, DSNameSink, d.Id(), err)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionSetting, DSNameSink, d.Id(), err)
	}

	return nil
}
