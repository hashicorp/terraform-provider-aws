// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datapipeline

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datapipeline"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_datapipeline_pipeline_definition")
func DataSourcePipelineDefinition() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePipelineDefinitionRead,

		Schema: map[string]*schema.Schema{
			"parameter_object": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKey: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"string_value": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"parameter_value": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"string_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"pipeline_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"pipeline_object": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrField: {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKey: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ref_value": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"string_value": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourcePipelineDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DataPipelineClient(ctx)

	pipelineID := d.Get("pipeline_id").(string)
	input := &datapipeline.GetPipelineDefinitionInput{
		PipelineId: aws.String(pipelineID),
	}

	resp, err := conn.GetPipelineDefinition(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting DataPipeline Definition (%s): %s", pipelineID, err)
	}

	if err = d.Set("parameter_object", flattenPipelineDefinitionParameterObjects(resp.ParameterObjects)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting `%s` for DataPipeline Pipeline Definition (%s): %s", "parameter_object", pipelineID, err)
	}
	if err = d.Set("parameter_value", flattenPipelineDefinitionParameterValues(resp.ParameterValues)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting `%s` for DataPipeline Pipeline Definition (%s): %s", "parameter_object", pipelineID, err)
	}
	if err = d.Set("pipeline_object", flattenPipelineDefinitionObjects(resp.PipelineObjects)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting `%s` for DataPipeline Pipeline Definition (%s): %s", "parameter_object", pipelineID, err)
	}
	d.SetId(pipelineID)

	return diags
}
