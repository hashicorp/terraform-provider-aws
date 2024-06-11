// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_flow_definition", name="Flow Definition")
// @Tags(identifierAttribute="arn")
func ResourceFlowDefinition() *schema.Resource {
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
										StateFunc: func(v interface{}) string {
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
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.AwsManagedHumanLoopRequestSource_Values(), false),
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFlowDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	name := d.Get("flow_definition_name").(string)
	input := &sagemaker.CreateFlowDefinitionInput{
		FlowDefinitionName: aws.String(name),
		HumanLoopConfig:    expandFlowDefinitionHumanLoopConfig(d.Get("human_loop_config").([]interface{})),
		RoleArn:            aws.String(d.Get(names.AttrRoleARN).(string)),
		OutputConfig:       expandFlowDefinitionOutputConfig(d.Get("output_config").([]interface{})),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("human_loop_activation_config"); ok && (len(v.([]interface{})) > 0) {
		loopConfig, err := expandFlowDefinitionHumanLoopActivationConfig(v.([]interface{}))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating SageMaker Flow Definition Human Loop Activation Config (%s): %s", name, err)
		}
		input.HumanLoopActivationConfig = loopConfig
	}

	if v, ok := d.GetOk("human_loop_request_source"); ok && (len(v.([]interface{})) > 0) {
		input.HumanLoopRequestSource = expandFlowDefinitionHumanLoopRequestSource(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating SageMaker Flow Definition: %s", input)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateFlowDefinitionWithContext(ctx, input)
	}, "ValidationException")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Flow Definition (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := WaitFlowDefinitionActive(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Flow Definition (%s) to become active: %s", d.Id(), err)
	}

	return append(diags, resourceFlowDefinitionRead(ctx, d, meta)...)
}

func resourceFlowDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	flowDefinition, err := FindFlowDefinitionByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker Flow Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Flow Definition (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(flowDefinition.FlowDefinitionArn)
	d.Set(names.AttrARN, arn)
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

func resourceFlowDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceFlowDefinitionRead(ctx, d, meta)...)
}

func resourceFlowDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	log.Printf("[DEBUG] Deleting SageMaker Flow Definition: %s", d.Id())
	_, err := conn.DeleteFlowDefinitionWithContext(ctx, &sagemaker.DeleteFlowDefinitionInput{
		FlowDefinitionName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Flow Definition (%s): %s", d.Id(), err)
	}

	if _, err := WaitFlowDefinitionDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Flow Definition (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func expandFlowDefinitionHumanLoopActivationConfig(l []interface{}) (*sagemaker.HumanLoopActivationConfig, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	loopConfig, err := expandFlowDefinitionHumanLoopActivationConditionsConfig(m["human_loop_activation_conditions_config"].([]interface{}))
	if err != nil {
		return nil, err
	}
	config := &sagemaker.HumanLoopActivationConfig{
		HumanLoopActivationConditionsConfig: loopConfig,
	}

	return config, nil
}

func flattenFlowDefinitionHumanLoopActivationConfig(config *sagemaker.HumanLoopActivationConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"human_loop_activation_conditions_config": flattenFlowDefinitionHumanLoopActivationConditionsConfig(config.HumanLoopActivationConditionsConfig),
	}

	return []map[string]interface{}{m}
}

func expandFlowDefinitionHumanLoopActivationConditionsConfig(l []interface{}) (*sagemaker.HumanLoopActivationConditionsConfig, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	v, err := protocol.DecodeJSONValue(m["human_loop_activation_conditions"].(string), protocol.NoEscape)
	if err != nil {
		return nil, err
	}

	config := &sagemaker.HumanLoopActivationConditionsConfig{
		HumanLoopActivationConditions: v,
	}

	return config, nil
}

func flattenFlowDefinitionHumanLoopActivationConditionsConfig(config *sagemaker.HumanLoopActivationConditionsConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	v, err := protocol.EncodeJSONValue(config.HumanLoopActivationConditions, protocol.NoEscape)
	if err != nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"human_loop_activation_conditions": v,
	}

	return []map[string]interface{}{m}
}

func expandFlowDefinitionOutputConfig(l []interface{}) *sagemaker.FlowDefinitionOutputConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.FlowDefinitionOutputConfig{
		S3OutputPath: aws.String(m["s3_output_path"].(string)),
	}

	if v, ok := m[names.AttrKMSKeyID].(string); ok && v != "" {
		config.KmsKeyId = aws.String(v)
	}

	return config
}

func flattenFlowDefinitionOutputConfig(config *sagemaker.FlowDefinitionOutputConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		names.AttrKMSKeyID: aws.StringValue(config.KmsKeyId),
		"s3_output_path":   aws.StringValue(config.S3OutputPath),
	}

	return []map[string]interface{}{m}
}

func expandFlowDefinitionHumanLoopRequestSource(l []interface{}) *sagemaker.HumanLoopRequestSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.HumanLoopRequestSource{
		AwsManagedHumanLoopRequestSource: aws.String(m["aws_managed_human_loop_request_source"].(string)),
	}

	return config
}

func flattenFlowDefinitionHumanLoopRequestSource(config *sagemaker.HumanLoopRequestSource) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"aws_managed_human_loop_request_source": aws.StringValue(config.AwsManagedHumanLoopRequestSource),
	}

	return []map[string]interface{}{m}
}

func expandFlowDefinitionHumanLoopConfig(l []interface{}) *sagemaker.HumanLoopConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.HumanLoopConfig{
		HumanTaskUiArn:  aws.String(m["human_task_ui_arn"].(string)),
		TaskCount:       aws.Int64(int64(m["task_count"].(int))),
		TaskDescription: aws.String(m["task_description"].(string)),
		TaskTitle:       aws.String(m["task_title"].(string)),
		WorkteamArn:     aws.String(m["workteam_arn"].(string)),
	}

	if v, ok := m["public_workforce_task_price"].([]interface{}); ok && len(v) > 0 {
		config.PublicWorkforceTaskPrice = expandFlowDefinitionPublicWorkforceTaskPrice(v)
	}

	if v, ok := m["task_keywords"].(*schema.Set); ok && v.Len() > 0 {
		config.TaskKeywords = flex.ExpandStringSet(v)
	}

	if v, ok := m["task_availability_lifetime_in_seconds"].(int); ok {
		config.TaskAvailabilityLifetimeInSeconds = aws.Int64(int64(v))
	}

	if v, ok := m["task_time_limit_in_seconds"].(int); ok {
		config.TaskTimeLimitInSeconds = aws.Int64(int64(v))
	}

	return config
}

func flattenFlowDefinitionHumanLoopConfig(config *sagemaker.HumanLoopConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"human_task_ui_arn": aws.StringValue(config.HumanTaskUiArn),
		"task_count":        aws.Int64Value(config.TaskCount),
		"task_description":  aws.StringValue(config.TaskDescription),
		"task_title":        aws.StringValue(config.TaskTitle),
		"workteam_arn":      aws.StringValue(config.WorkteamArn),
	}

	if config.PublicWorkforceTaskPrice != nil {
		m["public_workforce_task_price"] = flattenFlowDefinitionPublicWorkforceTaskPrice(config.PublicWorkforceTaskPrice)
	}

	if config.TaskKeywords != nil {
		m["task_keywords"] = flex.FlattenStringSet(config.TaskKeywords)
	}

	if config.TaskAvailabilityLifetimeInSeconds != nil {
		m["task_availability_lifetime_in_seconds"] = aws.Int64Value(config.TaskAvailabilityLifetimeInSeconds)
	}

	if config.TaskTimeLimitInSeconds != nil {
		m["task_time_limit_in_seconds"] = aws.Int64Value(config.TaskTimeLimitInSeconds)
	}

	return []map[string]interface{}{m}
}

func expandFlowDefinitionPublicWorkforceTaskPrice(l []interface{}) *sagemaker.PublicWorkforceTaskPrice {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.PublicWorkforceTaskPrice{}

	if v, ok := m["amount_in_usd"].([]interface{}); ok && len(v) > 0 {
		config.AmountInUsd = expandFlowDefinitionAmountInUsd(v)
	}

	return config
}

func expandFlowDefinitionAmountInUsd(l []interface{}) *sagemaker.USD {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.USD{}

	if v, ok := m["cents"].(int); ok {
		config.Cents = aws.Int64(int64(v))
	}

	if v, ok := m["dollars"].(int); ok {
		config.Dollars = aws.Int64(int64(v))
	}

	if v, ok := m["tenth_fractions_of_a_cent"].(int); ok {
		config.TenthFractionsOfACent = aws.Int64(int64(v))
	}

	return config
}

func flattenFlowDefinitionAmountInUsd(config *sagemaker.USD) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.Cents != nil {
		m["cents"] = aws.Int64Value(config.Cents)
	}

	if config.Dollars != nil {
		m["dollars"] = aws.Int64Value(config.Dollars)
	}

	if config.TenthFractionsOfACent != nil {
		m["tenth_fractions_of_a_cent"] = aws.Int64Value(config.TenthFractionsOfACent)
	}

	return []map[string]interface{}{m}
}

func flattenFlowDefinitionPublicWorkforceTaskPrice(config *sagemaker.PublicWorkforceTaskPrice) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.AmountInUsd != nil {
		m["amount_in_usd"] = flattenFlowDefinitionAmountInUsd(config.AmountInUsd)
	}

	return []map[string]interface{}{m}
}
