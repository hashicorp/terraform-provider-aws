// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/bedrock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_bedrock_model_invocation_logging_configuration")
func ResourceModelInvocationLoggingConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceModelInvocationLoggingConfigurationCreate,
		ReadWithoutTimeout:   resourceModelInvocationLoggingConfigurationRead,
		UpdateWithoutTimeout: resourceModelInvocationLoggingConfigurationCreate,
		DeleteWithoutTimeout: resourceModelInvocationLoggingConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"logging_config": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloud_watch_config": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"log_group_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"role_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringMatch(regexache.MustCompile(`^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$`), "Minimum length of 0. Maximum length of 2048."),
									},
									"large_data_delivery_s3_config": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"key_prefix": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"embedding_data_delivery_enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"image_data_delivery_enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"s3_config": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"key_prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"text_data_delivery_enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceModelInvocationLoggingConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BedrockConn(ctx)

	input := &bedrock.PutModelInvocationLoggingConfigurationInput{}

	if v, ok := d.GetOk("logging_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.LoggingConfig = expandLoggingConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.PutModelInvocationLoggingConfigurationWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "failed to put model invocation logging configuration: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	return append(diags, resourceModelInvocationLoggingConfigurationRead(ctx, d, meta)...)
}

func resourceModelInvocationLoggingConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BedrockConn(ctx)

	output, err := conn.GetModelInvocationLoggingConfigurationWithContext(ctx, &bedrock.GetModelInvocationLoggingConfigurationInput{})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "failed to get model invocation logging configuration: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	if output.LoggingConfig != nil {
		if err := d.Set("logging_config", []interface{}{flattenLoggingConfig(output.LoggingConfig)}); err != nil {
			return diag.Errorf("setting logging)config: %s", err)
		}
	} else {
		d.Set("logging_config", nil)
	}
	return diags
}

func resourceModelInvocationLoggingConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BedrockConn(ctx)

	_, err := conn.DeleteModelInvocationLoggingConfigurationWithContext(ctx, &bedrock.DeleteModelInvocationLoggingConfigurationInput{})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "failed to delete model invocation logging configuration: %s", err)
	}

	return diags
}
