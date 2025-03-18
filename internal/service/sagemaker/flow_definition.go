// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_flow_definition", name="Flow Definition")
// @Tags(identifierAttribute="arn")
func resourceFlowDefinition() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFlowDefinitionCreate,
		ReadWithoutTimeout:   resourceFlowDefinitionRead,
		UpdateWithoutTimeout: resourceFlowDefinitionUpdate,
		DeleteWithoutTimeout: resourceFlowDefinitionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"flow_definition_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9a-z](-*[0-9a-z])*$`), "Valid characters are a-z, 0-9, and - (hyphen)."),
				),
			},
			"human_loop_activation_config": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				MaxItems:     1,
				RequiredWith: []string{"human_loop_request_source", "human_loop_activation_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"human_loop_activation_conditions_config": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"human_loop_activation_conditions": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 10240),
											validation.StringIsJSON,
										),
										StateFunc: func(v any) string {
											json, _ := structure.NormalizeJsonString(v)
											return json
										},
										DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
									},
								},
							},
						},
					},
				},
			},
			"human_loop_config": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"human_task_ui_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"public_workforce_task_price": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"amount_in_usd": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cents": {
													Type:         schema.TypeInt,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.IntBetween(0, 99),
												},
												"dollars": {
													Type:         schema.TypeInt,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.IntBetween(0, 2),
												},
												"tenth_fractions_of_a_cent": {
													Type:         schema.TypeInt,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.IntBetween(0, 9),
												},
											},
										},
									},
								},
							},
						},
						"task_availability_lifetime_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 864000),
						},
						"task_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 3),
						},
						"task_description": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"task_keywords": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 1,
							MaxItems: 5,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 30),
									validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+( [0-9A-Za-z]+)*$`), ""),
								),
							},
						},
						"task_time_limit_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							Default:      3600,
							ValidateFunc: validation.IntBetween(30, 28800),
						},
						"task_title": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"workteam_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"human_loop_request_source": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				MaxItems:     1,
				RequiredWith: []string{"human_loop_request_source", "human_loop_activation_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aws_managed_human_loop_request_source": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.AwsManagedHumanLoopRequestSource](),
						},
					},
				},
			},
			"output_config": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKMSKeyID: {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"s3_output_path": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexache.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
								validation.StringLenBetween(1, 512),
							),
						},
					},
				},
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceFlowDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("flow_definition_name").(string)
	input := &sagemaker.CreateFlowDefinitionInput{
		FlowDefinitionName: aws.String(name),
		HumanLoopConfig:    expandFlowDefinitionHumanLoopConfig(d.Get("human_loop_config").([]any)),
		RoleArn:            aws.String(d.Get(names.AttrRoleARN).(string)),
		OutputConfig:       expandFlowDefinitionOutputConfig(d.Get("output_config").([]any)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("human_loop_activation_config"); ok && (len(v.([]any)) > 0) {
		loopConfig, err := expandFlowDefinitionHumanLoopActivationConfig(v.([]any))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Flow Definition Human Loop Activation Config (%s): %s", name, err)
		}
		input.HumanLoopActivationConfig = loopConfig
	}

	if v, ok := d.GetOk("human_loop_request_source"); ok && (len(v.([]any)) > 0) {
		input.HumanLoopRequestSource = expandFlowDefinitionHumanLoopRequestSource(v.([]any))
	}

	log.Printf("[DEBUG] Creating SageMaker AI Flow Definition: %#v", input)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (any, error) {
		return conn.CreateFlowDefinition(ctx, input)
	}, ErrCodeValidationException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Flow Definition (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitFlowDefinitionActive(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Flow Definition (%s) to become active: %s", d.Id(), err)
	}

	return append(diags, resourceFlowDefinitionRead(ctx, d, meta)...)
}

func resourceFlowDefinitionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	flowDefinition, err := findFlowDefinitionByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker AI Flow Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Flow Definition (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, flowDefinition.FlowDefinitionArn)
	d.Set(names.AttrRoleARN, flowDefinition.RoleArn)
	d.Set("flow_definition_name", flowDefinition.FlowDefinitionName)

	if err := d.Set("human_loop_activation_config", flattenFlowDefinitionHumanLoopActivationConfig(flowDefinition.HumanLoopActivationConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting human_loop_activation_config: %s", err)
	}

	if err := d.Set("human_loop_config", flattenFlowDefinitionHumanLoopConfig(flowDefinition.HumanLoopConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting human_loop_config: %s", err)
	}

	if err := d.Set("human_loop_request_source", flattenFlowDefinitionHumanLoopRequestSource(flowDefinition.HumanLoopRequestSource)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting human_loop_request_source: %s", err)
	}

	if err := d.Set("output_config", flattenFlowDefinitionOutputConfig(flowDefinition.OutputConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting output_config: %s", err)
	}

	return diags
}

func resourceFlowDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceFlowDefinitionRead(ctx, d, meta)...)
}

func resourceFlowDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	log.Printf("[DEBUG] Deleting SageMaker AI Flow Definition: %s", d.Id())
	_, err := conn.DeleteFlowDefinition(ctx, &sagemaker.DeleteFlowDefinitionInput{
		FlowDefinitionName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Flow Definition (%s): %s", d.Id(), err)
	}

	if _, err := waitFlowDefinitionDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Flow Definition (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func findFlowDefinitionByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeFlowDefinitionOutput, error) {
	input := &sagemaker.DescribeFlowDefinitionInput{
		FlowDefinitionName: aws.String(name),
	}

	output, err := conn.DescribeFlowDefinition(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandFlowDefinitionHumanLoopActivationConfig(l []any) (*awstypes.HumanLoopActivationConfig, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]any)

	loopConfig, err := expandFlowDefinitionHumanLoopActivationConditionsConfig(m["human_loop_activation_conditions_config"].([]any))
	if err != nil {
		return nil, err
	}
	config := &awstypes.HumanLoopActivationConfig{
		HumanLoopActivationConditionsConfig: loopConfig,
	}

	return config, nil
}

func flattenFlowDefinitionHumanLoopActivationConfig(config *awstypes.HumanLoopActivationConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"human_loop_activation_conditions_config": flattenFlowDefinitionHumanLoopActivationConditionsConfig(config.HumanLoopActivationConditionsConfig),
	}

	return []map[string]any{m}
}

func expandFlowDefinitionHumanLoopActivationConditionsConfig(l []any) (*awstypes.HumanLoopActivationConditionsConfig, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]any)
	output := &awstypes.HumanLoopActivationConditionsConfig{}

	if v, ok := m["human_loop_activation_conditions"]; ok && v.(string) != "" {
		out, err := structure.NormalizeJsonString(v)
		if err != nil {
			return nil, fmt.Errorf("Human Loop Activation Conditions (%s) is invalid JSON: %w", out, err)
		}

		output.HumanLoopActivationConditions = aws.String(out)
	}

	return output, nil
}

func flattenFlowDefinitionHumanLoopActivationConditionsConfig(config *awstypes.HumanLoopActivationConditionsConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"human_loop_activation_conditions": config.HumanLoopActivationConditions,
	}

	return []map[string]any{m}
}

func expandFlowDefinitionOutputConfig(l []any) *awstypes.FlowDefinitionOutputConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.FlowDefinitionOutputConfig{
		S3OutputPath: aws.String(m["s3_output_path"].(string)),
	}

	if v, ok := m[names.AttrKMSKeyID].(string); ok && v != "" {
		config.KmsKeyId = aws.String(v)
	}

	return config
}

func flattenFlowDefinitionOutputConfig(config *awstypes.FlowDefinitionOutputConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		names.AttrKMSKeyID: aws.ToString(config.KmsKeyId),
		"s3_output_path":   aws.ToString(config.S3OutputPath),
	}

	return []map[string]any{m}
}

func expandFlowDefinitionHumanLoopRequestSource(l []any) *awstypes.HumanLoopRequestSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.HumanLoopRequestSource{
		AwsManagedHumanLoopRequestSource: awstypes.AwsManagedHumanLoopRequestSource(m["aws_managed_human_loop_request_source"].(string)),
	}

	return config
}

func flattenFlowDefinitionHumanLoopRequestSource(config *awstypes.HumanLoopRequestSource) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"aws_managed_human_loop_request_source": config.AwsManagedHumanLoopRequestSource,
	}

	return []map[string]any{m}
}

func expandFlowDefinitionHumanLoopConfig(l []any) *awstypes.HumanLoopConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.HumanLoopConfig{
		HumanTaskUiArn:  aws.String(m["human_task_ui_arn"].(string)),
		TaskCount:       aws.Int32(int32(m["task_count"].(int))),
		TaskDescription: aws.String(m["task_description"].(string)),
		TaskTitle:       aws.String(m["task_title"].(string)),
		WorkteamArn:     aws.String(m["workteam_arn"].(string)),
	}

	if v, ok := m["public_workforce_task_price"].([]any); ok && len(v) > 0 {
		config.PublicWorkforceTaskPrice = expandFlowDefinitionPublicWorkforceTaskPrice(v)
	}

	if v, ok := m["task_keywords"].(*schema.Set); ok && v.Len() > 0 {
		config.TaskKeywords = flex.ExpandStringValueSet(v)
	}

	if v, ok := m["task_availability_lifetime_in_seconds"].(int); ok {
		config.TaskAvailabilityLifetimeInSeconds = aws.Int32(int32(v))
	}

	if v, ok := m["task_time_limit_in_seconds"].(int); ok {
		config.TaskTimeLimitInSeconds = aws.Int32(int32(v))
	}

	return config
}

func flattenFlowDefinitionHumanLoopConfig(config *awstypes.HumanLoopConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"human_task_ui_arn": aws.ToString(config.HumanTaskUiArn),
		"task_count":        aws.ToInt32(config.TaskCount),
		"task_description":  aws.ToString(config.TaskDescription),
		"task_title":        aws.ToString(config.TaskTitle),
		"workteam_arn":      aws.ToString(config.WorkteamArn),
	}

	if config.PublicWorkforceTaskPrice != nil {
		m["public_workforce_task_price"] = flattenFlowDefinitionPublicWorkforceTaskPrice(config.PublicWorkforceTaskPrice)
	}

	if config.TaskKeywords != nil {
		m["task_keywords"] = flex.FlattenStringValueSet(config.TaskKeywords)
	}

	if config.TaskAvailabilityLifetimeInSeconds != nil {
		m["task_availability_lifetime_in_seconds"] = aws.ToInt32(config.TaskAvailabilityLifetimeInSeconds)
	}

	if config.TaskTimeLimitInSeconds != nil {
		m["task_time_limit_in_seconds"] = aws.ToInt32(config.TaskTimeLimitInSeconds)
	}

	return []map[string]any{m}
}

func expandFlowDefinitionPublicWorkforceTaskPrice(l []any) *awstypes.PublicWorkforceTaskPrice {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.PublicWorkforceTaskPrice{}

	if v, ok := m["amount_in_usd"].([]any); ok && len(v) > 0 {
		config.AmountInUsd = expandFlowDefinitionAmountInUsd(v)
	}

	return config
}

func expandFlowDefinitionAmountInUsd(l []any) *awstypes.USD {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.USD{}

	if v, ok := m["cents"].(int); ok {
		config.Cents = aws.Int32(int32(v))
	}

	if v, ok := m["dollars"].(int); ok {
		config.Dollars = aws.Int32(int32(v))
	}

	if v, ok := m["tenth_fractions_of_a_cent"].(int); ok {
		config.TenthFractionsOfACent = aws.Int32(int32(v))
	}

	return config
}

func flattenFlowDefinitionAmountInUsd(config *awstypes.USD) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.Cents != nil {
		m["cents"] = aws.ToInt32(config.Cents)
	}

	if config.Dollars != nil {
		m["dollars"] = aws.ToInt32(config.Dollars)
	}

	if config.TenthFractionsOfACent != nil {
		m["tenth_fractions_of_a_cent"] = aws.ToInt32(config.TenthFractionsOfACent)
	}

	return []map[string]any{m}
}

func flattenFlowDefinitionPublicWorkforceTaskPrice(config *awstypes.PublicWorkforceTaskPrice) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.AmountInUsd != nil {
		m["amount_in_usd"] = flattenFlowDefinitionAmountInUsd(config.AmountInUsd)
	}

	return []map[string]any{m}
}
