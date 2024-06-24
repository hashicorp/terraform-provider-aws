// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssm_maintenance_window_task", name="Maintenance Window Task")
func resourceMaintenanceWindowTask() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMaintenanceWindowTaskCreate,
		ReadWithoutTimeout:   resourceMaintenanceWindowTaskRead,
		UpdateWithoutTimeout: resourceMaintenanceWindowTaskUpdate,
		DeleteWithoutTimeout: resourceMaintenanceWindowTaskDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceMaintenanceWindowTaskImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cutoff_behavior": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.MaintenanceWindowTaskCutoffBehavior](),
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"max_concurrency": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^([1-9][0-9]*|[1-9][0-9]%|[1-9]%|100%)$`), "must be a number without leading zeros or a percentage between 1% and 100% without leading zeros and ending with the percentage symbol"),
			},
			"max_errors": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^([1-9][0-9]*|[0]|[1-9][0-9]%|[0-9]%|100%)$`), "must be zero, a number without leading zeros, or a percentage between 1% and 100% without leading zeros and ending with the percentage symbol"),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]{3,128}$`),
					"Only alphanumeric characters, hyphens, dots & underscores allowed."),
			},
			names.AttrPriority: {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			names.AttrServiceRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			"targets": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValues: {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 50,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"task_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"task_invocation_parameters": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"automation_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"document_version": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringMatch(regexache.MustCompile("([$]LATEST|[$]DEFAULT|^[1-9][0-9]*$)"), "see https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_MaintenanceWindowAutomationParameters.html"),
									},
									names.AttrParameter: {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrName: {
													Type:     schema.TypeString,
													Required: true,
												},
												names.AttrValues: {
													Type:     schema.TypeList,
													Required: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
								},
							},
						},
						"lambda_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_context": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 8000),
									},
									"payload": {
										Type:         schema.TypeString,
										Optional:     true,
										Sensitive:    true,
										ValidateFunc: validation.StringLenBetween(0, 4096),
									},
									"qualifier": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
								},
							},
						},
						"run_command_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cloudwatch_config": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cloudwatch_log_group_name": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
												"cloudwatch_output_enabled": {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
									names.AttrComment: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 100),
									},
									"document_hash": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 256),
									},
									"document_hash_type": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.DocumentHashType](),
									},
									"document_version": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringMatch(regexache.MustCompile(`([$]LATEST|[$]DEFAULT|^[1-9][0-9]*$)`), "must be $DEFAULT, $LATEST, or a version number"),
									},
									"notification_config": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"notification_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"notification_events": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Schema{
														Type:             schema.TypeString,
														ValidateDiagFunc: enum.Validate[awstypes.NotificationEvent](),
													},
												},
												"notification_type": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.NotificationType](),
												},
											},
										},
									},
									"output_s3_bucket": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"output_s3_key_prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrParameter: {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrName: {
													Type:     schema.TypeString,
													Required: true,
												},
												names.AttrValues: {
													Type:     schema.TypeList,
													Required: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									names.AttrServiceRoleARN: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"timeout_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(30, 2592000),
									},
								},
							},
						},
						"step_functions_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"input": {
										Type:         schema.TypeString,
										Optional:     true,
										Sensitive:    true,
										ValidateFunc: validation.StringLenBetween(0, 4096),
									},
									names.AttrName: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 80),
									},
								},
							},
						},
					},
				},
			},
			"task_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.MaintenanceWindowTaskType](),
			},
			"window_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"window_task_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMaintenanceWindowTaskCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	input := &ssm.RegisterTaskWithMaintenanceWindowInput{
		TaskArn:  aws.String(d.Get("task_arn").(string)),
		TaskType: awstypes.MaintenanceWindowTaskType(d.Get("task_type").(string)),
		WindowId: aws.String(d.Get("window_id").(string)),
	}

	if v, ok := d.GetOk("cutoff_behavior"); ok {
		input.CutoffBehavior = awstypes.MaintenanceWindowTaskCutoffBehavior(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_concurrency"); ok {
		input.MaxConcurrency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_errors"); ok {
		input.MaxErrors = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPriority); ok {
		input.Priority = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrServiceRoleARN); ok {
		input.ServiceRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("targets"); ok {
		input.Targets = expandTargets(v.([]interface{}))
	}

	if v, ok := d.GetOk("task_invocation_parameters"); ok {
		input.TaskInvocationParameters = expandTaskInvocationParameters(v.([]interface{}))
	}

	output, err := conn.RegisterTaskWithMaintenanceWindow(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSM Maintenance Window Task: %s", err)
	}

	d.SetId(aws.ToString(output.WindowTaskId))

	return append(diags, resourceMaintenanceWindowTaskRead(ctx, d, meta)...)
}

func resourceMaintenanceWindowTaskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	output, err := findMaintenanceWindowTaskByTwoPartKey(ctx, conn, d.Get("window_id").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Maintenance Window Task %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Maintenance Window Task (%s): %s", d.Id(), err)
	}

	windowTaskID := aws.ToString(output.WindowTaskId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ssm",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "windowtask/" + windowTaskID,
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("cutoff_behavior", output.CutoffBehavior)
	d.Set(names.AttrDescription, output.Description)
	d.Set("max_concurrency", output.MaxConcurrency)
	d.Set("max_errors", output.MaxErrors)
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrPriority, output.Priority)
	d.Set(names.AttrServiceRoleARN, output.ServiceRoleArn)
	if err := d.Set("targets", flattenTargets(output.Targets)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting targets: %s", err)
	}
	d.Set("task_arn", output.TaskArn)
	if output.TaskInvocationParameters != nil {
		if err := d.Set("task_invocation_parameters", flattenTaskInvocationParameters(output.TaskInvocationParameters)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting task_invocation_parameters: %s", err)
		}
	}
	d.Set("task_type", output.TaskType)
	d.Set("window_id", output.WindowId)
	d.Set("window_task_id", windowTaskID)

	return diags
}

func resourceMaintenanceWindowTaskUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	input := &ssm.UpdateMaintenanceWindowTaskInput{
		Priority:     aws.Int32(int32(d.Get(names.AttrPriority).(int))),
		Replace:      aws.Bool(true),
		TaskArn:      aws.String(d.Get("task_arn").(string)),
		WindowId:     aws.String(d.Get("window_id").(string)),
		WindowTaskId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("cutoff_behavior"); ok {
		input.CutoffBehavior = awstypes.MaintenanceWindowTaskCutoffBehavior(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_concurrency"); ok {
		input.MaxConcurrency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_errors"); ok {
		input.MaxErrors = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrServiceRoleARN); ok {
		input.ServiceRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("task_invocation_parameters"); ok {
		input.TaskInvocationParameters = expandTaskInvocationParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("targets"); ok {
		input.Targets = expandTargets(v.([]interface{}))
	} else {
		input.MaxConcurrency = nil
		input.MaxErrors = nil
	}

	_, err := conn.UpdateMaintenanceWindowTask(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Maintenance Window Task (%s): %s", d.Id(), err)
	}

	return append(diags, resourceMaintenanceWindowTaskRead(ctx, d, meta)...)
}

func resourceMaintenanceWindowTaskDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	log.Printf("[INFO] Deleting SSM Maintenance Window Task: %s", d.Id())
	_, err := conn.DeregisterTaskFromMaintenanceWindow(ctx, &ssm.DeregisterTaskFromMaintenanceWindowInput{
		WindowId:     aws.String(d.Get("window_id").(string)),
		WindowTaskId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.DoesNotExistException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSM Maintenance Window Task (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceMaintenanceWindowTaskImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <window-id>/<window-task-id>", d.Id())
	}

	windowID := idParts[0]
	windowTaskID := idParts[1]

	d.Set("window_id", windowID)
	d.SetId(windowTaskID)

	return []*schema.ResourceData{d}, nil
}

func findMaintenanceWindowTaskByTwoPartKey(ctx context.Context, conn *ssm.Client, windowID, windowTaskID string) (*ssm.GetMaintenanceWindowTaskOutput, error) {
	input := &ssm.GetMaintenanceWindowTaskInput{
		WindowId:     aws.String(windowID),
		WindowTaskId: aws.String(windowTaskID),
	}

	output, err := conn.GetMaintenanceWindowTask(ctx, input)

	if errs.IsA[*awstypes.DoesNotExistException](err) {
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

func expandTaskInvocationParameters(tfList []interface{}) *awstypes.MaintenanceWindowTaskInvocationParameters {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.MaintenanceWindowTaskInvocationParameters{}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})
		if v, ok := tfMap["automation_parameters"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			apiObject.Automation = expandTaskInvocationAutomationParameters(v.([]interface{}))
		}
		if v, ok := tfMap["lambda_parameters"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			apiObject.Lambda = expandTaskInvocationLambdaParameters(v.([]interface{}))
		}
		if v, ok := tfMap["run_command_parameters"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			apiObject.RunCommand = expandTaskInvocationRunCommandParameters(v.([]interface{}))
		}
		if v, ok := tfMap["step_functions_parameters"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			apiObject.StepFunctions = expandTaskInvocationStepFunctionsParameters(v.([]interface{}))
		}
	}

	return apiObject
}

func flattenTaskInvocationParameters(apiObject *awstypes.MaintenanceWindowTaskInvocationParameters) []interface{} {
	tfMap := make(map[string]interface{})

	if apiObject.Automation != nil {
		tfMap["automation_parameters"] = flattenTaskInvocationAutomationParameters(apiObject.Automation)
	}

	if apiObject.Lambda != nil {
		tfMap["lambda_parameters"] = flattenTaskInvocationLambdaParameters(apiObject.Lambda)
	}

	if apiObject.RunCommand != nil {
		tfMap["run_command_parameters"] = flattenTaskInvocationRunCommandParameters(apiObject.RunCommand)
	}

	if apiObject.StepFunctions != nil {
		tfMap["step_functions_parameters"] = flattenTaskInvocationStepFunctionsParameters(apiObject.StepFunctions)
	}

	return []interface{}{tfMap}
}

func expandTaskInvocationAutomationParameters(tfList []interface{}) *awstypes.MaintenanceWindowAutomationParameters {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.MaintenanceWindowAutomationParameters{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["document_version"]; ok && len(v.(string)) != 0 {
		apiObject.DocumentVersion = aws.String(v.(string))
	}
	if v, ok := tfMap[names.AttrParameter]; ok && len(v.(*schema.Set).List()) > 0 {
		apiObject.Parameters = expandTaskInvocationCommonParameters(v.(*schema.Set).List())
	}

	return apiObject
}

func flattenTaskInvocationAutomationParameters(apiObject *awstypes.MaintenanceWindowAutomationParameters) []interface{} {
	tfMap := make(map[string]interface{})

	if apiObject.DocumentVersion != nil {
		tfMap["document_version"] = aws.ToString(apiObject.DocumentVersion)
	}
	if apiObject.Parameters != nil {
		tfMap[names.AttrParameter] = flattenTaskInvocationCommonParameters(apiObject.Parameters)
	}

	return []interface{}{tfMap}
}

func expandTaskInvocationLambdaParameters(tfList []interface{}) *awstypes.MaintenanceWindowLambdaParameters {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.MaintenanceWindowLambdaParameters{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["client_context"]; ok && len(v.(string)) != 0 {
		apiObject.ClientContext = aws.String(v.(string))
	}
	if v, ok := tfMap["payload"]; ok && len(v.(string)) != 0 {
		apiObject.Payload = []byte(v.(string))
	}
	if v, ok := tfMap["qualifier"]; ok && len(v.(string)) != 0 {
		apiObject.Qualifier = aws.String(v.(string))
	}

	return apiObject
}

func flattenTaskInvocationLambdaParameters(apiObject *awstypes.MaintenanceWindowLambdaParameters) []interface{} {
	tfMap := make(map[string]interface{})

	if apiObject.ClientContext != nil {
		tfMap["client_context"] = aws.ToString(apiObject.ClientContext)
	}
	if apiObject.Payload != nil {
		tfMap["payload"] = string(apiObject.Payload)
	}
	if apiObject.Qualifier != nil {
		tfMap["qualifier"] = aws.ToString(apiObject.Qualifier)
	}

	return []interface{}{tfMap}
}

func expandTaskInvocationRunCommandParameters(tfList []interface{}) *awstypes.MaintenanceWindowRunCommandParameters {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.MaintenanceWindowRunCommandParameters{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["cloudwatch_config"]; ok && len(v.([]interface{})) > 0 {
		apiObject.CloudWatchOutputConfig = expandTaskInvocationRunCommandParametersCloudWatchConfig(v.([]interface{}))
	}
	if v, ok := tfMap[names.AttrComment]; ok && len(v.(string)) != 0 {
		apiObject.Comment = aws.String(v.(string))
	}
	if v, ok := tfMap["document_hash"]; ok && len(v.(string)) != 0 {
		apiObject.DocumentHash = aws.String(v.(string))
	}
	if v, ok := tfMap["document_hash_type"]; ok && len(v.(string)) != 0 {
		apiObject.DocumentHashType = awstypes.DocumentHashType(v.(string))
	}
	if v, ok := tfMap["document_version"]; ok && len(v.(string)) != 0 {
		apiObject.DocumentVersion = aws.String(v.(string))
	}
	if v, ok := tfMap["notification_config"]; ok && len(v.([]interface{})) > 0 {
		apiObject.NotificationConfig = expandTaskInvocationRunCommandParametersNotificationConfig(v.([]interface{}))
	}
	if v, ok := tfMap["output_s3_bucket"]; ok && len(v.(string)) != 0 {
		apiObject.OutputS3BucketName = aws.String(v.(string))
	}
	if v, ok := tfMap["output_s3_key_prefix"]; ok && len(v.(string)) != 0 {
		apiObject.OutputS3KeyPrefix = aws.String(v.(string))
	}
	if v, ok := tfMap[names.AttrParameter]; ok && len(v.(*schema.Set).List()) > 0 {
		apiObject.Parameters = expandTaskInvocationCommonParameters(v.(*schema.Set).List())
	}
	if v, ok := tfMap[names.AttrServiceRoleARN]; ok && len(v.(string)) != 0 {
		apiObject.ServiceRoleArn = aws.String(v.(string))
	}
	if v, ok := tfMap["timeout_seconds"]; ok && v.(int) != 0 {
		apiObject.TimeoutSeconds = aws.Int32(int32(v.(int)))
	}

	return apiObject
}

func flattenTaskInvocationRunCommandParameters(apiObject *awstypes.MaintenanceWindowRunCommandParameters) []interface{} {
	tfMap := make(map[string]interface{})

	if apiObject.CloudWatchOutputConfig != nil {
		tfMap["cloudwatch_config"] = flattenTaskInvocationRunCommandParametersCloudWatchConfig(apiObject.CloudWatchOutputConfig)
	}
	if apiObject.Comment != nil {
		tfMap[names.AttrComment] = aws.ToString(apiObject.Comment)
	}
	if apiObject.DocumentHash != nil {
		tfMap["document_hash"] = aws.ToString(apiObject.DocumentHash)
	}
	tfMap["document_hash_type"] = apiObject.DocumentHashType
	if apiObject.DocumentVersion != nil {
		tfMap["document_version"] = aws.ToString(apiObject.DocumentVersion)
	}
	if apiObject.NotificationConfig != nil {
		tfMap["notification_config"] = flattenTaskInvocationRunCommandParametersNotificationConfig(apiObject.NotificationConfig)
	}
	if apiObject.OutputS3BucketName != nil {
		tfMap["output_s3_bucket"] = aws.ToString(apiObject.OutputS3BucketName)
	}
	if apiObject.OutputS3KeyPrefix != nil {
		tfMap["output_s3_key_prefix"] = aws.ToString(apiObject.OutputS3KeyPrefix)
	}
	if apiObject.Parameters != nil {
		tfMap[names.AttrParameter] = flattenTaskInvocationCommonParameters(apiObject.Parameters)
	}
	if apiObject.ServiceRoleArn != nil {
		tfMap[names.AttrServiceRoleARN] = aws.ToString(apiObject.ServiceRoleArn)
	}
	if apiObject.TimeoutSeconds != nil {
		tfMap["timeout_seconds"] = aws.ToInt32(apiObject.TimeoutSeconds)
	}

	return []interface{}{tfMap}
}

func expandTaskInvocationStepFunctionsParameters(tfList []interface{}) *awstypes.MaintenanceWindowStepFunctionsParameters {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.MaintenanceWindowStepFunctionsParameters{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["input"]; ok && len(v.(string)) != 0 {
		apiObject.Input = aws.String(v.(string))
	}
	if v, ok := tfMap[names.AttrName]; ok && len(v.(string)) != 0 {
		apiObject.Name = aws.String(v.(string))
	}

	return apiObject
}

func flattenTaskInvocationStepFunctionsParameters(apiObject *awstypes.MaintenanceWindowStepFunctionsParameters) []interface{} {
	tfMap := make(map[string]interface{})

	if apiObject.Input != nil {
		tfMap["input"] = aws.ToString(apiObject.Input)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}

	return []interface{}{tfMap}
}

func expandTaskInvocationRunCommandParametersNotificationConfig(tfList []interface{}) *awstypes.NotificationConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.NotificationConfig{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["notification_arn"]; ok && len(v.(string)) != 0 {
		apiObject.NotificationArn = aws.String(v.(string))
	}
	if v, ok := tfMap["notification_events"]; ok && len(v.([]interface{})) > 0 {
		apiObject.NotificationEvents = flex.ExpandStringyValueList[awstypes.NotificationEvent](v.([]interface{}))
	}
	if v, ok := tfMap["notification_type"]; ok && len(v.(string)) != 0 {
		apiObject.NotificationType = awstypes.NotificationType(v.(string))
	}

	return apiObject
}

func flattenTaskInvocationRunCommandParametersNotificationConfig(apiObject *awstypes.NotificationConfig) []interface{} {
	tfMap := make(map[string]interface{})

	if apiObject.NotificationArn != nil {
		tfMap["notification_arn"] = aws.ToString(apiObject.NotificationArn)
	}
	if apiObject.NotificationEvents != nil {
		tfMap["notification_events"] = apiObject.NotificationEvents
	}
	tfMap["notification_type"] = apiObject.NotificationType

	return []interface{}{tfMap}
}

func expandTaskInvocationRunCommandParametersCloudWatchConfig(tfList []interface{}) *awstypes.CloudWatchOutputConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.CloudWatchOutputConfig{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["cloudwatch_log_group_name"]; ok && len(v.(string)) != 0 {
		apiObject.CloudWatchLogGroupName = aws.String(v.(string))
	}
	if v, ok := tfMap["cloudwatch_output_enabled"]; ok {
		apiObject.CloudWatchOutputEnabled = v.(bool)
	}

	return apiObject
}

func flattenTaskInvocationRunCommandParametersCloudWatchConfig(apiObject *awstypes.CloudWatchOutputConfig) []interface{} {
	tfMap := make(map[string]interface{})

	if apiObject.CloudWatchLogGroupName != nil {
		tfMap["cloudwatch_log_group_name"] = aws.ToString(apiObject.CloudWatchLogGroupName)
	}
	tfMap["cloudwatch_output_enabled"] = apiObject.CloudWatchOutputEnabled

	return []interface{}{tfMap}
}

func expandTaskInvocationCommonParameters(tfList []interface{}) map[string][]string {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := make(map[string][]string)

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})
		apiObject[tfMap[names.AttrName].(string)] = flex.ExpandStringValueList(tfMap[names.AttrValues].([]interface{}))
	}

	return apiObject
}

func flattenTaskInvocationCommonParameters(apiObject map[string][]string) []interface{} {
	tfList := make([]interface{}, 0, len(apiObject))

	keys := tfmaps.Keys(apiObject)
	slices.Sort(keys)

	for _, key := range keys {
		tfList = append(tfList, map[string]interface{}{
			names.AttrName:   key,
			names.AttrValues: apiObject[key],
		})
	}

	return tfList
}
