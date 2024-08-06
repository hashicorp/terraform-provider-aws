// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datapipeline

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_datapipeline_pipeline")
func DataSourcePipeline() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePipelineRead,

		Schema: map[string]*schema.Schema{
			"pipeline_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourcePipelineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DataPipelineClient(ctx)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	pipelineId := d.Get("pipeline_id").(string)

	v, err := PipelineRetrieve(ctx, pipelineId, conn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing DataPipeline Pipeline (%s): %s", pipelineId, err)
	}

	d.Set(names.AttrName, v.Name)
	d.Set(names.AttrDescription, v.Description)

	tags := KeyValueTags(ctx, v.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set(names.AttrTags, tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.SetId(pipelineId)

	return diags
}
