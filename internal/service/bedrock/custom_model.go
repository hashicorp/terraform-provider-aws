// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/bedrock"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKResource("aws_bedrock_custom_model", name="Custom-Model")
func ResourceCustomModel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomModelCreate,
		ReadWithoutTimeout:   resourceCustomModelRead,
		DeleteWithoutTimeout: resourceCustomModelDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(40 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"job_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_model_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^(arn:aws(-[^:]+)?:bedrock:[a-z0-9-]{1,20}:(([0-9]{12}:custom-model/[a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}/[a-z0-9]{12})|(:foundation-model/[a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([a-z0-9-]{1,63}[.]){0,2}[a-z0-9-]{1,63}([:][a-z0-9-]{1,63}){0,2})))|([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2})|(([0-9a-zA-Z][_-]?)+)$`), "minimum length of 1. Maximum length of 2048."),
			},
			"client_request_token": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*$`), "minimum length of 1. Maximum length of 256."),
			},
			"custom_model_kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^arn:aws(-[^:]+)?:kms:[a-zA-Z0-9-]*:[0-9]{12}:((key/[a-zA-Z0-9-]{36})|(alias/[a-zA-Z0-9-_/]+))$`), "minimum length of 1. Maximum length of 2048."),
			},
			"custom_model_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^([0-9a-zA-Z][_-]?)+$`), "minimum length of 1. Maximum length of 63."),
			},
			"hyper_parameters": {
				Type:     schema.TypeMap,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"job_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"output_data_config": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^s3://[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9](/.*)?$`), "minimum length of 1. Maximum length of 1024."),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$`), "minimum length of 1. Maximum length of 2048."),
			},
			"training_data_config": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^s3://[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9](/.*)?$`), "minimum length of 1. Maximum length of 1024."),
			},
			"validation_data_config": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				ForceNew: true,
				// ValidateFunc: validation.StringMatch(regexache.MustCompile(`^s3://[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9](/.*)?$`), "minimum length of 1. Maximum length of 1024."),
			},
			"vpc_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							// ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[-0-9a-zA-Z]+$`), "minimum length of 1. Maximum length of 32."),
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							// ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[-0-9a-zA-Z]+$`), "minimum length of 1. Maximum length of 32."),
						},
					},
				},
			},
			// names.AttrTags:    tftags.TagsSchema(),
			// names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceCustomModelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BedrockConn(ctx)

	baseModelId := d.Get("base_model_id").(string)
	customModelName := d.Get("custom_model_name").(string)
	jobName := d.Get("job_name").(string)
	roleArn := d.Get("role_arn").(string)
	outputDataConfig := d.Get("output_data_config").(string)
	trainingDataConfig := d.Get("training_data_config").(string)

	input := &bedrock.CreateModelCustomizationJobInput{
		BaseModelIdentifier: aws.String(baseModelId),
		CustomModelName:     aws.String(customModelName),
		JobName:             aws.String(jobName),
		RoleArn:             aws.String(roleArn),
		OutputDataConfig: &bedrock.OutputDataConfig{
			S3Uri: aws.String(outputDataConfig),
		},
		TrainingDataConfig: &bedrock.TrainingDataConfig{
			S3Uri: aws.String(trainingDataConfig),
		},
	}

	if v, ok := d.GetOk("hyper_parameters"); ok && len(v.(map[string]interface{})) > 0 {
		input.HyperParameters = flex.ExpandStringMap(v.(map[string]interface{}))
	}
	if v, ok := d.GetOk("client_request_token"); ok {
		input.ClientRequestToken = aws.String(v.(string))
	}
	if v, ok := d.GetOk("custom_model_kms_key_id"); ok {
		input.CustomModelKmsKeyId = aws.String(v.(string))
	}

	// if v, ok := d.GetOk("hyper_parameters"); ok && v.(*schema.Set).Len() > 0 {
	// 	input.HyperParameters = expandHyperParameters(v.(*schema.Set).List())
	// }

	tflog.Info(ctx, "CreateModelCustomizationJobInput:", map[string]any{
		"BaseModelIdentifier": baseModelId,
		"CustomModelName":     customModelName,
		"JobName":             jobName,
		"RoleArn":             roleArn,
		"OutputDataConfig":    outputDataConfig,
		"TrainingDataConfig":  trainingDataConfig,
	})

	output, err := conn.CreateModelCustomizationJobWithContext(ctx, input)
	// _, err := tfresource.RetryWhen(ctx, propagationTimeout,
	// 	func() (interface{}, error) {
	// 		return conn.CreateAddonWithContext(ctx, input)
	// 	},
	// 	func(err error) (bool, error) {
	// 		if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "CREATE_FAILED") {
	// 			return true, err
	// 		}

	// 		if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "does not exist") {
	// 			return true, err
	// 		}

	// 		return false, err
	// 	},
	// )
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Bedrock Custom Model Customization Job: %s", err)
	}

	d.SetId(*output.JobArn)

	// if _, err := waitAddonCreated(ctx, conn, clusterName, addonName, d.Timeout(schema.TimeoutCreate)); err != nil {
	// 	// Creating addon w/o setting resolve_conflicts to "OVERWRITE"
	// 	// might result in a failed creation, if unmanaged version of addon is already deployed
	// 	// and there are configuration conflicts:
	// 	// ConfigurationConflict	Apply failed with 1 conflict: conflict with "kubectl"...
	// 	//
	// 	// Addon resource is tainted after failed creation, thus will be deleted and created again.
	// 	// Re-creating like this will resolve the error, but it will also purge any
	// 	// configurations that were applied by the user (that were conflicting). This might we an unwanted
	// 	// side effect and should be left for the user to decide how to handle it.
	// 	diags = sdkdiag.AppendErrorf(diags, "waiting for EKS Add-On (%s) create: %s", d.Id(), err)
	// 	return sdkdiag.AppendWarningf(diags, "Running terraform apply again will remove the kubernetes add-on and attempt to create it again effectively purging previous add-on configuration")
	// }

	return append(diags, resourceCustomModelRead(ctx, d, meta)...)
}

func resourceCustomModelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BedrockConn(ctx)

	modelId := d.Get("model_id").(string)
	input := &bedrock.GetCustomModelInput{
		ModelIdentifier: &modelId,
	}
	model, err := conn.GetCustomModelWithContext(ctx, input)
	if err != nil {
		return diag.Errorf("reading Bedrock Foundation Models: %s", err)
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

	// "base_model_arn": aws.StringValue(model.BaseModelArn),
	// 		"base_model_name": aws.StringValue(model.BaseModelName),
	// 		"model_arn": aws.StringValue(model.ModelArn),
	// 		"model_name": aws.StringValue(model.ModelName),
	// 		"creation_time": aws.TimeValue(model.CreationTime).Format(time.RFC3339),

	return diags
}

func resourceCustomModelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BedrockConn(ctx)

	modelId := d.Get("model_id").(string)

	input := &bedrock.DeleteCustomModelInput{
		ModelIdentifier: &modelId,
	}

	log.Printf("[DEBUG] Deleting Bedrock Custom Model: %s", d.Id())
	_, err := conn.DeleteCustomModelWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Bedrock Custom Model ID(%s): %s", d.Id(), err)
	}

	return diags
}

// func expandHyperParameters(data []interface{}) []*bedrock. {
// 	var streamingTargets []*chimesdkvoice.StreamingNotificationTarget

// 	for _, item := range data {
// 		streamingTargets = append(streamingTargets, &chimesdkvoice.StreamingNotificationTarget{
// 			NotificationTarget: aws.String(item.(string)),
// 		})
// 	}

// 	return streamingTargets
// }
