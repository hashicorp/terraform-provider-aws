// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKDataSource("aws_bedrock_foundation_models")
func DataSourceFoundationModels() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFoundationModelsRead,
		Schema: map[string]*schema.Schema{
			"model_summaries": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"model_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"model_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"model_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"provider_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"customizations_supported": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"inference_types_supported": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"input_modalities": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"output_modalities": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"response_streaming_supported": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceFoundationModelsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).BedrockConn(ctx)

	models, err := conn.ListFoundationModelsWithContext(ctx, nil)
	if err != nil {
		return diag.Errorf("reading Bedrock Foundation Models: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	if err := d.Set("model_summaries", flattenFoundationModelSummaries(models.ModelSummaries)); err != nil {
		return diag.Errorf("setting model_summaries: %s", err)
	}

	return nil
}
