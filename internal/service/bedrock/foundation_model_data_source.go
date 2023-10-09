// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/bedrock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKDataSource("aws_bedrock_foundation_model")
func DataSourceFoundationModel() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFoundationModelRead,
		Schema: map[string]*schema.Schema{
			"model_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"model_arn": {
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
	}
}

func dataSourceFoundationModelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).BedrockConn(ctx)

	modelId := d.Get("model_id").(string)
	input := &bedrock.GetFoundationModelInput{
		ModelIdentifier: &modelId,
	}

	model, err := conn.GetFoundationModelWithContext(ctx, input)
	if err != nil {
		return diag.Errorf("reading Bedrock Foundation Model: %s", err)
	}

	d.SetId(modelId)
	d.Set("model_arn", model.ModelDetails.ModelArn)
	d.Set("model_id", model.ModelDetails.ModelId)
	d.Set("model_name", model.ModelDetails.ModelName)
	d.Set("provider_name", model.ModelDetails.ProviderName)
	d.Set("customizations_supported", aws.StringValueSlice(model.ModelDetails.CustomizationsSupported))
	d.Set("inference_types_supported", aws.StringValueSlice(model.ModelDetails.InferenceTypesSupported))
	d.Set("input_modalities", aws.StringValueSlice(model.ModelDetails.InputModalities))
	d.Set("output_modalities", aws.StringValueSlice(model.ModelDetails.OutputModalities))
	d.Set("response_streaming_supported", model.ModelDetails.ResponseStreamingSupported)

	return nil
}
