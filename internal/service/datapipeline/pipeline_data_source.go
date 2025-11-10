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

// @SDKDataSource("aws_datapipeline_pipeline", name="Pipeline")
// @Tags
// @Testing(tagsIdentifierAttribute="id", tagsResourceType="Pipeline")
func dataSourcePipeline() *schema.Resource {
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

func dataSourcePipelineRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DataPipelineClient(ctx)

	pipelineId := d.Get("pipeline_id").(string)

	v, err := findPipeline(ctx, conn, pipelineId)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing DataPipeline Pipeline (%s): %s", pipelineId, err)
	}

	d.SetId(pipelineId)
	d.Set(names.AttrName, v.Name)
	d.Set(names.AttrDescription, v.Description)

	setTagsOut(ctx, v.Tags)

	return diags
}
