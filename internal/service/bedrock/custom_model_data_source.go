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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_bedrock_custom_model")
// @Tags(identifierAttribute="model_arn")
func DataSourceCustomModel() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCustomModelRead,
		Schema: map[string]*schema.Schema{
			"model_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"base_model_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hyper_parameters": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"job_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_tags": tftags.TagsSchemaComputed(),
			"model_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"model_kms_key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"model_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"output_data_config": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"training_data_config": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"training_metrics": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"training_loss": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
					},
				},
			},
			"validation_data_config": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"validator": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"validation_metrics": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"validation_loss": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchemaComputed(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceCustomModelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).BedrockConn(ctx)

	modelId := d.Get("model_id").(string)
	input := &bedrock.GetCustomModelInput{
		ModelIdentifier: &modelId,
	}
	model, err := conn.GetCustomModelWithContext(ctx, input)
	if err != nil {
		return diag.Errorf("reading Bedrock Custom Model: %s", err)
	}

	d.SetId(modelId)
	d.Set("base_model_arn", aws.StringValue(model.BaseModelArn))
	d.Set("creation_time", aws.TimeValue(model.CreationTime).Format(time.RFC3339))
	d.Set("hyper_parameters", model.HyperParameters)
	d.Set("job_arn", aws.StringValue(model.JobArn))
	d.Set("job_name", aws.StringValue(model.JobName))
	d.Set("model_arn", aws.StringValue(model.ModelArn))
	d.Set("model_kms_key_arn", aws.StringValue(model.ModelKmsKeyArn))
	d.Set("model_name", aws.StringValue(model.ModelName))
	d.Set("output_data_config", aws.StringValue(model.OutputDataConfig.S3Uri))
	d.Set("training_data_config", aws.StringValue(model.TrainingDataConfig.S3Uri))
	if err := d.Set("training_metrics", flattenTrainingMetrics(model.TrainingMetrics)); err != nil {
		return diag.Errorf("setting training_metrics: %s", err)
	}
	if err := d.Set("validation_data_config", flattenValidationDataConfig(model.ValidationDataConfig)); err != nil {
		return diag.Errorf("setting validation_metrics: %s", err)
	}
	if err := d.Set("validation_metrics", flattenValidationMetrics(model.ValidationMetrics)); err != nil {
		return diag.Errorf("setting validation_metrics: %s", err)
	}

	jobTags, err := listTags(ctx, conn, *model.JobArn)
	if err != nil {
		return diag.Errorf("reading Tags for Job: %s", err)
	}
	d.Set("job_tags", jobTags.IgnoreAWS().Map())

	return nil
}
