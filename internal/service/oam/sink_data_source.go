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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
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
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameSink = "Sink Data Source"
)

func dataSourceSinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	sinkIdentifier := d.Get("sink_identifier").(string)

	out, err := findSinkByID(ctx, conn, sinkIdentifier)
	if err != nil {
		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionReading, DSNameSink, sinkIdentifier, err)
	}

	d.SetId(aws.ToString(out.Arn))

	d.Set("arn", out.Arn)
	d.Set("name", out.Name)
	d.Set("sink_id", out.Id)

	tags, err := listTags(ctx, conn, d.Id())
	if err != nil {
		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionReading, DSNameSink, d.Id(), err)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.DiagError(names.ObservabilityAccessManager, create.ErrActionSetting, DSNameSink, d.Id(), err)
	}

	return nil
}
