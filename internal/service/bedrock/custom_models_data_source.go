// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/bedrock"
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
						"model_id": {
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
		return diag.Errorf("reading Bedrock Foundation Models: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	if err := d.Set("model_summaries", flattenCustomModelSummaries(models.ModelSummaries)); err != nil {
		return diag.Errorf("setting model_summaries: %s", err)
	}

	return nil
}

func flattenCustomModelSummaries(models []*bedrock.CustomModelSummary) []map[string]interface{} {
	if len(models) == 0 {
		return []map[string]interface{}{}
	}

	l := make([]map[string]interface{}, 0, len(models))

	for _, model := range models {
		m := map[string]interface{}{
			"base_model_arn":  aws.StringValue(model.BaseModelArn),
			"base_model_name": aws.StringValue(model.BaseModelName),
			"model_arn":       aws.StringValue(model.ModelArn),
			"model_name":      aws.StringValue(model.ModelName),
			"creation_time":   aws.TimeValue(model.CreationTime).Format(time.RFC3339),
		}
		l = append(l, m)
	}

	return l
}
