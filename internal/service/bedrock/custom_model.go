// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/bedrock"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_bedrock_custom_model", name="Custom-Model")
// @Tags(identifierAttribute="model_arn")
func ResourceCustomModel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomModelCreate,
		ReadWithoutTimeout:   resourceCustomModelRead,
		DeleteWithoutTimeout: resourceCustomModelDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(24 * time.Hour),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
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
			"job_tags": tftags.TagsSchemaForceNew(),
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
				Optional: true,
				ForceNew: true,
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
			"base_model_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
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
			"training_metrics": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"training_loss": {
							Type:     schema.TypeInt,
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
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchemaForceNew(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceCustomModelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BedrockConn(ctx)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	job_tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("job_tags").(map[string]interface{})))

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
		TrainingDataConfig: &bedrock.TrainingDataConfig{
			S3Uri: aws.String(trainingDataConfig),
		},
		OutputDataConfig: &bedrock.OutputDataConfig{
			S3Uri: aws.String(outputDataConfig),
		},
		CustomModelTags: getTagsIn(ctx),
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
	if len(job_tags) > 0 {
		input.JobTags = Tags(tftags.New(ctx, job_tags.IgnoreAWS()))
	}
	if v, ok := d.GetOk("vpc_config"); ok {
		input.VpcConfig = expandVPCConfig(v.([]interface{}))
	}
	if v, ok := d.GetOk("validation_data_config"); ok {
		input.ValidationDataConfig = expandValidationDataConfig(v.([]*string))
	}

	tflog.Info(ctx, "CreateModelCustomizationJobInput:", map[string]any{
		"BaseModelIdentifier":  input.BaseModelIdentifier,
		"ClientRequestToken":   input.ClientRequestToken,
		"CustomModelName":      input.CustomModelName,
		"CustomModelKmsKeyId":  input.CustomModelKmsKeyId,
		"JobName":              jobName,
		"RoleArn":              roleArn,
		"OutputDataConfig":     outputDataConfig,
		"TrainingDataConfig":   trainingDataConfig,
		"ValidationDataConfig": input.ValidationDataConfig,
		"VpcConfig":            input.VpcConfig,
	})

	jobStart, err := conn.CreateModelCustomizationJobWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Bedrock Custom Model Customization Job: %s", err)
	}

	// Successfully started job. Save the name as the id of the custom model.
	d.SetId(customModelName)
	// also store the job arn now incase we need to cancel and destroy.
	d.Set("job_arn", jobStart.JobArn)

	err = waitForModelCustomizationJob(ctx, conn, *jobStart.JobArn, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "failed to complete model customisation job: %s", err)
	}

	return append(diags, resourceCustomModelRead(ctx, d, meta)...)
}

func resourceCustomModelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BedrockConn(ctx)

	tflog.Info(ctx, "resourceCustomModelRead: Getting Custom Model...")
	modelId := d.Id()
	input := &bedrock.GetCustomModelInput{
		ModelIdentifier: &modelId,
	}
	model, err := conn.GetCustomModelWithContext(ctx, input)
	if err != nil {
		// If we got here, the state has the model name and the job arn.
		// Should we check for tainted state instead?
		tflog.Info(ctx, "resourceCustomModelRead: Error reading Bedrock Custom Model. Ignoring to allow destroy to attempt to cleanup.")
		//return diag.Errorf("reading Bedrock Custom Model: %s", err)
		return diags
	}

	d.Set("base_model_arn", model.BaseModelArn)
	d.Set("creation_time", aws.TimeValue(model.CreationTime).Format(time.RFC3339))
	d.Set("hyper_parameters", model.HyperParameters)
	d.Set("job_arn", model.JobArn)
	// This is nil in the model object - could be a bug
	// However this is already in state so we can skip setting this here and avoid a forced update due to value change.
	// d.Set("job_name", model.JobName)
	d.Set("model_arn", model.ModelArn)
	d.Set("model_kms_key_arn", model.ModelKmsKeyArn)
	d.Set("model_name", model.ModelName)
	d.Set("output_data_config", model.OutputDataConfig.S3Uri)
	d.Set("training_data_config", model.TrainingDataConfig.S3Uri)
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

	return diags
}

func resourceCustomModelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BedrockConn(ctx)

	modelId := d.Id()
	jobArn := d.Get("job_arn").(string)
	tflog.Info(ctx, fmt.Sprintf("Cancelling Bedrock customization job %s", jobArn))
	_, err := conn.StopModelCustomizationJobWithContext(ctx, &bedrock.StopModelCustomizationJobInput{
		JobIdentifier: &jobArn,
	})
	if err != nil {
		// ignore validatin errors - eg. already complete
		if _, ok := err.(*bedrock.ValidationException); !ok {
			return sdkdiag.AppendErrorf(diags, "stopping Bedrock Customization Job ID(%s): %s", jobArn, err)
		}
	}

	tflog.Info(ctx, fmt.Sprintf("Deleting Bedrock Custom Model: %s", d.Id()))
	_, err = conn.DeleteCustomModelWithContext(ctx, &bedrock.DeleteCustomModelInput{
		ModelIdentifier: &modelId,
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Bedrock Custom Model ID(%s): %s", d.Id(), err)
	}

	return diags
}

func waitForModelCustomizationJob(ctx context.Context, conn *bedrock.Bedrock, jobArn string, timeout time.Duration) error {
	return retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		jobEnd, err := conn.GetModelCustomizationJobWithContext(ctx, &bedrock.GetModelCustomizationJobInput{
			JobIdentifier: &jobArn,
		})
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("getting model customization job: %s", err))
		}

		tflog.Info(ctx, "GetModelCustomizationJobOuput:", map[string]any{
			"Status": jobEnd.Status,
		})

		switch *jobEnd.Status {
		case "InProgress":
			return retry.RetryableError(fmt.Errorf("expected instance to be Completed but was in state %s", *jobEnd.Status))
		case "Completed":
			return nil
		default:
			return retry.NonRetryableError(fmt.Errorf(*jobEnd.FailureMessage))
		}
	})
}
