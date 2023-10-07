// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKDataSource("aws_bedrock_custom_models")
func DataSourceCustomModels() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCustomModelsRead,
		Schema: map[string]*schema.Schema{
			"model_summaries": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base_model_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"base_model_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"model_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"model_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"creation_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCustomModelsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).BedrockConn(ctx)

	models, err := conn.ListCustomModelsWithContext(ctx, nil)
	if err != nil {
		return diag.Errorf("reading Bedrock Custom Models: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	if err := d.Set("model_summaries", flattenCustomModelSummaries(models.ModelSummaries)); err != nil {
		return diag.Errorf("setting model_summaries: %s", err)
	}

	return nil
}
