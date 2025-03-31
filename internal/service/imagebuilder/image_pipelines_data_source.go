// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/namevaluesfilters"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_imagebuilder_image_pipelines", name="Image Pipelines")
func dataSourceImagePipelines() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceImagePipelinesRead,

		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrFilter: namevaluesfilters.Schema(),
			names.AttrNames: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceImagePipelinesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.ListImagePipelinesInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = namevaluesfilters.New(v.(*schema.Set)).ImageBuilderFilters()
	}

	imagePipelines, err := findImagePipelines(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Image Pipelines: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set(names.AttrARNs, tfslices.ApplyToAll(imagePipelines, func(v awstypes.ImagePipeline) string {
		return aws.ToString(v.Arn)
	}))
	d.Set(names.AttrNames, tfslices.ApplyToAll(imagePipelines, func(v awstypes.ImagePipeline) string {
		return aws.ToString(v.Name)
	}))

	return diags
}

func findImagePipelines(ctx context.Context, conn *imagebuilder.Client, input *imagebuilder.ListImagePipelinesInput) ([]awstypes.ImagePipeline, error) {
	var output []awstypes.ImagePipeline

	pages := imagebuilder.NewListImagePipelinesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.ImagePipelineList...)
	}

	return output, nil
}
