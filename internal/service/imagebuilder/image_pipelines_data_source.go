// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/namevaluesfilters"
	namevaluesfiltersv2 "github.com/hashicorp/terraform-provider-aws/internal/namevaluesfilters/v2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_imagebuilder_image_pipelines")
func DataSourceImagePipelines() *schema.Resource {
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

func dataSourceImagePipelinesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.ListImagePipelinesInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = namevaluesfiltersv2.New(v.(*schema.Set)).ImageBuilderFilters()
	}

	out, err := conn.ListImagePipelines(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Image Pipelines: %s", err)
	}

	var arns, nms []string

	for _, r := range out.ImagePipelineList {
		arns = append(arns, aws.ToString(r.Arn))
		nms = append(nms, aws.ToString(r.Name))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrARNs, arns)
	d.Set(names.AttrNames, nms)

	return diags
}
