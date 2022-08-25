package ssm

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceMaintenanceWindowTask() *schema.Resource {
	return &schema.Resource{
		Create: resourceMaintenanceWindowTaskCreate,
		Read:   resourceMaintenanceWindowTaskRead,
		Update: resourceMaintenanceWindowTaskUpdate,
		Delete: resourceMaintenanceWindowTaskDelete,
		Importer: &schema.ResourceImporter{
			State: resourceMaintenanceWindowTaskImport,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"window_task_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"window_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"cutoff_behavior": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ssm.MaintenanceWindowTaskCutoffBehavior_Values(), false),
			},

			"max_concurrency": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([1-9][0-9]*|[1-9][0-9]%|[1-9]%|100%)$`), "must be a number without leading zeros or a percentage between 1% and 100% without leading zeros and ending with the percentage symbol"),
			},

			"max_errors": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([1-9][0-9]*|[0]|[1-9][0-9]%|[0-9]%|100%)$`), "must be zero, a number without leading zeros, or a percentage between 1% and 100% without leading zeros and ending with the percentage symbol"),
			},

			"task_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ssm.MaintenanceWindowTaskType_Values(), false),
			},

			"task_arn": {
				Type:     schema.TypeString,
				Required: true,
			},

			"service_role_arn": {
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
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 50,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_\-.]{3,128}$`),
					"Only alphanumeric characters, hyphens, dots & underscores allowed."),
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},

			"priority": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
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
										ValidateFunc: validation.StringMatch(regexp.MustCompile("([$]LATEST|[$]DEFAULT|^[1-9][0-9]*$)"), "see https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_MaintenanceWindowAutomationParameters.html"),
									},
									"parameter": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},

												"values": {
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
									"comment": {
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
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(ssm.DocumentHashType_Values(), false),
									},
									"document_version": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringMatch(regexp.MustCompile(`([$]LATEST|[$]DEFAULT|^[1-9][0-9]*$)`), "must be $DEFAULT, $LATEST, or a version number"),
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
														Type:         schema.TypeString,
														ValidateFunc: validation.StringInSlice(ssm.NotificationEvent_Values(), false),
													},
												},

												"notification_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(ssm.NotificationType_Values(), false),
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

									"parameter": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},

												"values": {
													Type:     schema.TypeList,
													Required: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},

									"service_role_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},

									"timeout_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(30, 2592000),
									},
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

									"name": {
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
		},
	}
}

func expandTaskInvocationParameters(config []interface{}) *ssm.MaintenanceWindowTaskInvocationParameters {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	params := &ssm.MaintenanceWindowTaskInvocationParameters{}
	for _, v := range config {
		paramConfig := v.(map[string]interface{})
		if attr, ok := paramConfig["automation_parameters"]; ok && len(attr.([]interface{})) > 0 && attr.([]interface{})[0] != nil {
			params.Automation = expandTaskInvocationAutomationParameters(attr.([]interface{}))
		}
		if attr, ok := paramConfig["lambda_parameters"]; ok && len(attr.([]interface{})) > 0 && attr.([]interface{})[0] != nil {
			params.Lambda = expandTaskInvocationLambdaParameters(attr.([]interface{}))
		}
		if attr, ok := paramConfig["run_command_parameters"]; ok && len(attr.([]interface{})) > 0 && attr.([]interface{})[0] != nil {
			params.RunCommand = expandTaskInvocationRunCommandParameters(attr.([]interface{}))
		}
		if attr, ok := paramConfig["step_functions_parameters"]; ok && len(attr.([]interface{})) > 0 && attr.([]interface{})[0] != nil {
			params.StepFunctions = expandTaskInvocationStepFunctionsParameters(attr.([]interface{}))
		}
	}
	return params
}

func flattenTaskInvocationParameters(parameters *ssm.MaintenanceWindowTaskInvocationParameters) []interface{} {
	result := make(map[string]interface{})
	if parameters.Automation != nil {
		result["automation_parameters"] = flattenTaskInvocationAutomationParameters(parameters.Automation)
	}

	if parameters.Lambda != nil {
		result["lambda_parameters"] = flattenTaskInvocationLambdaParameters(parameters.Lambda)
	}

	if parameters.RunCommand != nil {
		result["run_command_parameters"] = flattenTaskInvocationRunCommandParameters(parameters.RunCommand)
	}

	if parameters.StepFunctions != nil {
		result["step_functions_parameters"] = flattenTaskInvocationStepFunctionsParameters(parameters.StepFunctions)
	}

	return []interface{}{result}
}

func expandTaskInvocationAutomationParameters(config []interface{}) *ssm.MaintenanceWindowAutomationParameters {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	params := &ssm.MaintenanceWindowAutomationParameters{}
	configParam := config[0].(map[string]interface{})
	if attr, ok := configParam["document_version"]; ok && len(attr.(string)) != 0 {
		params.DocumentVersion = aws.String(attr.(string))
	}
	if attr, ok := configParam["parameter"]; ok && len(attr.(*schema.Set).List()) > 0 {
		params.Parameters = expandTaskInvocationCommonParameters(attr.(*schema.Set).List())
	}

	return params
}

func flattenTaskInvocationAutomationParameters(parameters *ssm.MaintenanceWindowAutomationParameters) []interface{} {
	result := make(map[string]interface{})

	if parameters.DocumentVersion != nil {
		result["document_version"] = aws.StringValue(parameters.DocumentVersion)
	}
	if parameters.Parameters != nil {
		result["parameter"] = flattenTaskInvocationCommonParameters(parameters.Parameters)
	}

	return []interface{}{result}
}

func expandTaskInvocationLambdaParameters(config []interface{}) *ssm.MaintenanceWindowLambdaParameters {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	params := &ssm.MaintenanceWindowLambdaParameters{}
	configParam := config[0].(map[string]interface{})
	if attr, ok := configParam["client_context"]; ok && len(attr.(string)) != 0 {
		params.ClientContext = aws.String(attr.(string))
	}
	if attr, ok := configParam["payload"]; ok && len(attr.(string)) != 0 {
		params.Payload = []byte(attr.(string))
	}
	if attr, ok := configParam["qualifier"]; ok && len(attr.(string)) != 0 {
		params.Qualifier = aws.String(attr.(string))
	}
	return params
}

func flattenTaskInvocationLambdaParameters(parameters *ssm.MaintenanceWindowLambdaParameters) []interface{} {
	result := make(map[string]interface{})

	if parameters.ClientContext != nil {
		result["client_context"] = aws.StringValue(parameters.ClientContext)
	}
	if parameters.Payload != nil {
		result["payload"] = string(parameters.Payload)
	}
	if parameters.Qualifier != nil {
		result["qualifier"] = aws.StringValue(parameters.Qualifier)
	}
	return []interface{}{result}
}

func expandTaskInvocationRunCommandParameters(config []interface{}) *ssm.MaintenanceWindowRunCommandParameters {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	params := &ssm.MaintenanceWindowRunCommandParameters{}
	configParam := config[0].(map[string]interface{})
	if attr, ok := configParam["comment"]; ok && len(attr.(string)) != 0 {
		params.Comment = aws.String(attr.(string))
	}
	if attr, ok := configParam["document_hash"]; ok && len(attr.(string)) != 0 {
		params.DocumentHash = aws.String(attr.(string))
	}
	if attr, ok := configParam["document_hash_type"]; ok && len(attr.(string)) != 0 {
		params.DocumentHashType = aws.String(attr.(string))
	}
	if attr, ok := configParam["document_version"]; ok && len(attr.(string)) != 0 {
		params.DocumentVersion = aws.String(attr.(string))
	}
	if attr, ok := configParam["notification_config"]; ok && len(attr.([]interface{})) > 0 {
		params.NotificationConfig = expandTaskInvocationRunCommandParametersNotificationConfig(attr.([]interface{}))
	}
	if attr, ok := configParam["output_s3_bucket"]; ok && len(attr.(string)) != 0 {
		params.OutputS3BucketName = aws.String(attr.(string))
	}
	if attr, ok := configParam["output_s3_key_prefix"]; ok && len(attr.(string)) != 0 {
		params.OutputS3KeyPrefix = aws.String(attr.(string))
	}
	if attr, ok := configParam["parameter"]; ok && len(attr.(*schema.Set).List()) > 0 {
		params.Parameters = expandTaskInvocationCommonParameters(attr.(*schema.Set).List())
	}
	if attr, ok := configParam["service_role_arn"]; ok && len(attr.(string)) != 0 {
		params.ServiceRoleArn = aws.String(attr.(string))
	}
	if attr, ok := configParam["timeout_seconds"]; ok && attr.(int) != 0 {
		params.TimeoutSeconds = aws.Int64(int64(attr.(int)))
	}

	if attr, ok := configParam["cloudwatch_config"]; ok && len(attr.([]interface{})) > 0 {
		params.CloudWatchOutputConfig = expandTaskInvocationRunCommandParametersCloudWatchConfig(attr.([]interface{}))
	}
	return params
}

func flattenTaskInvocationRunCommandParameters(parameters *ssm.MaintenanceWindowRunCommandParameters) []interface{} {
	result := make(map[string]interface{})

	if parameters.Comment != nil {
		result["comment"] = aws.StringValue(parameters.Comment)
	}
	if parameters.DocumentHash != nil {
		result["document_hash"] = aws.StringValue(parameters.DocumentHash)
	}
	if parameters.DocumentHashType != nil {
		result["document_hash_type"] = aws.StringValue(parameters.DocumentHashType)
	}
	if parameters.DocumentVersion != nil {
		result["document_version"] = aws.StringValue(parameters.DocumentVersion)
	}
	if parameters.NotificationConfig != nil {
		result["notification_config"] = flattenTaskInvocationRunCommandParametersNotificationConfig(parameters.NotificationConfig)
	}
	if parameters.OutputS3BucketName != nil {
		result["output_s3_bucket"] = aws.StringValue(parameters.OutputS3BucketName)
	}
	if parameters.OutputS3KeyPrefix != nil {
		result["output_s3_key_prefix"] = aws.StringValue(parameters.OutputS3KeyPrefix)
	}
	if parameters.Parameters != nil {
		result["parameter"] = flattenTaskInvocationCommonParameters(parameters.Parameters)
	}
	if parameters.ServiceRoleArn != nil {
		result["service_role_arn"] = aws.StringValue(parameters.ServiceRoleArn)
	}
	if parameters.TimeoutSeconds != nil {
		result["timeout_seconds"] = aws.Int64Value(parameters.TimeoutSeconds)
	}
	if parameters.CloudWatchOutputConfig != nil {
		result["cloudwatch_config"] = flattenTaskInvocationRunCommandParametersCloudWatchConfig(parameters.CloudWatchOutputConfig)
	}

	return []interface{}{result}
}

func expandTaskInvocationStepFunctionsParameters(config []interface{}) *ssm.MaintenanceWindowStepFunctionsParameters {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	params := &ssm.MaintenanceWindowStepFunctionsParameters{}
	configParam := config[0].(map[string]interface{})

	if attr, ok := configParam["input"]; ok && len(attr.(string)) != 0 {
		params.Input = aws.String(attr.(string))
	}
	if attr, ok := configParam["name"]; ok && len(attr.(string)) != 0 {
		params.Name = aws.String(attr.(string))
	}

	return params
}

func flattenTaskInvocationStepFunctionsParameters(parameters *ssm.MaintenanceWindowStepFunctionsParameters) []interface{} {
	result := make(map[string]interface{})

	if parameters.Input != nil {
		result["input"] = aws.StringValue(parameters.Input)
	}
	if parameters.Name != nil {
		result["name"] = aws.StringValue(parameters.Name)
	}
	return []interface{}{result}
}

func expandTaskInvocationRunCommandParametersNotificationConfig(config []interface{}) *ssm.NotificationConfig {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	params := &ssm.NotificationConfig{}
	configParam := config[0].(map[string]interface{})

	if attr, ok := configParam["notification_arn"]; ok && len(attr.(string)) != 0 {
		params.NotificationArn = aws.String(attr.(string))
	}
	if attr, ok := configParam["notification_events"]; ok && len(attr.([]interface{})) > 0 {
		params.NotificationEvents = flex.ExpandStringList(attr.([]interface{}))
	}
	if attr, ok := configParam["notification_type"]; ok && len(attr.(string)) != 0 {
		params.NotificationType = aws.String(attr.(string))
	}

	return params
}

func flattenTaskInvocationRunCommandParametersNotificationConfig(config *ssm.NotificationConfig) []interface{} {
	result := make(map[string]interface{})

	if config.NotificationArn != nil {
		result["notification_arn"] = aws.StringValue(config.NotificationArn)
	}
	if config.NotificationEvents != nil {
		result["notification_events"] = flex.FlattenStringList(config.NotificationEvents)
	}
	if config.NotificationType != nil {
		result["notification_type"] = aws.StringValue(config.NotificationType)
	}

	return []interface{}{result}
}

func expandTaskInvocationRunCommandParametersCloudWatchConfig(config []interface{}) *ssm.CloudWatchOutputConfig {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	params := &ssm.CloudWatchOutputConfig{}
	configParam := config[0].(map[string]interface{})

	if attr, ok := configParam["cloudwatch_log_group_name"]; ok && len(attr.(string)) != 0 {
		params.CloudWatchLogGroupName = aws.String(attr.(string))
	}
	if attr, ok := configParam["cloudwatch_output_enabled"]; ok {
		params.CloudWatchOutputEnabled = aws.Bool(attr.(bool))
	}

	return params
}

func flattenTaskInvocationRunCommandParametersCloudWatchConfig(config *ssm.CloudWatchOutputConfig) []interface{} {
	result := make(map[string]interface{})

	if config.CloudWatchLogGroupName != nil {
		result["cloudwatch_log_group_name"] = aws.StringValue(config.CloudWatchLogGroupName)
	}
	if config.CloudWatchOutputEnabled != nil {
		result["cloudwatch_output_enabled"] = aws.BoolValue(config.CloudWatchOutputEnabled)
	}

	return []interface{}{result}
}

func expandTaskInvocationCommonParameters(config []interface{}) map[string][]*string {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	params := make(map[string][]*string)

	for _, v := range config {
		paramConfig := v.(map[string]interface{})
		params[paramConfig["name"].(string)] = flex.ExpandStringList(paramConfig["values"].([]interface{}))
	}

	return params
}

func flattenTaskInvocationCommonParameters(parameters map[string][]*string) []interface{} {
	attributes := make([]interface{}, 0, len(parameters))

	keys := make([]string, 0, len(parameters))
	for k := range parameters {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		values := make([]string, 0)
		for _, value := range parameters[key] {
			values = append(values, aws.StringValue(value))
		}
		params := map[string]interface{}{
			"name":   key,
			"values": values,
		}
		attributes = append(attributes, params)
	}

	return attributes
}

func resourceMaintenanceWindowTaskCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn

	log.Printf("[INFO] Registering SSM Maintenance Window Task")

	params := &ssm.RegisterTaskWithMaintenanceWindowInput{
		WindowId: aws.String(d.Get("window_id").(string)),
		TaskType: aws.String(d.Get("task_type").(string)),
		TaskArn:  aws.String(d.Get("task_arn").(string)),
	}

	if v, ok := d.GetOk("max_errors"); ok {
		params.MaxErrors = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_concurrency"); ok {
		params.MaxConcurrency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cutoff_behavior"); ok {
		params.CutoffBehavior = aws.String(v.(string))
	}

	if v, ok := d.GetOk("targets"); ok {
		params.Targets = expandTargets(v.([]interface{}))
	}

	if v, ok := d.GetOk("service_role_arn"); ok {
		params.ServiceRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		params.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("priority"); ok {
		params.Priority = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("task_invocation_parameters"); ok {
		params.TaskInvocationParameters = expandTaskInvocationParameters(v.([]interface{}))
	}

	resp, err := conn.RegisterTaskWithMaintenanceWindow(params)
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(resp.WindowTaskId))

	return resourceMaintenanceWindowTaskRead(d, meta)
}

func resourceMaintenanceWindowTaskRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn
	windowID := d.Get("window_id").(string)

	params := &ssm.GetMaintenanceWindowTaskInput{
		WindowId:     aws.String(windowID),
		WindowTaskId: aws.String(d.Id()),
	}
	resp, err := conn.GetMaintenanceWindowTask(params)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ssm.ErrCodeDoesNotExistException) {
		log.Printf("[WARN] Maintenance Window (%s) Task (%s) not found, removing from state", windowID, d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error getting Maintenance Window (%s) Task (%s): %s", windowID, d.Id(), err)
	}

	windowTaskID := aws.StringValue(resp.WindowTaskId)
	d.Set("window_id", resp.WindowId)
	d.Set("window_task_id", windowTaskID)
	d.Set("max_concurrency", resp.MaxConcurrency)
	d.Set("max_errors", resp.MaxErrors)
	d.Set("task_type", resp.TaskType)
	d.Set("service_role_arn", resp.ServiceRoleArn)
	d.Set("task_arn", resp.TaskArn)
	d.Set("priority", resp.Priority)
	d.Set("name", resp.Name)
	d.Set("description", resp.Description)
	d.Set("cutoff_behavior", resp.CutoffBehavior)

	if resp.TaskInvocationParameters != nil {
		if err := d.Set("task_invocation_parameters", flattenTaskInvocationParameters(resp.TaskInvocationParameters)); err != nil {
			return fmt.Errorf("Error setting task_invocation_parameters error: %#v", err)
		}
	}

	if err := d.Set("targets", flattenTargets(resp.Targets)); err != nil {
		return fmt.Errorf("Error setting targets error: %#v", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ssm",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("windowtask/%s", windowTaskID),
	}.String()
	d.Set("arn", arn)

	return nil
}

func resourceMaintenanceWindowTaskUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn
	windowID := d.Get("window_id").(string)

	params := &ssm.UpdateMaintenanceWindowTaskInput{
		Priority:     aws.Int64(int64(d.Get("priority").(int))),
		WindowId:     aws.String(windowID),
		WindowTaskId: aws.String(d.Id()),
		TaskArn:      aws.String(d.Get("task_arn").(string)),
		Replace:      aws.Bool(true),
	}

	if v, ok := d.GetOk("service_role_arn"); ok {
		params.ServiceRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_errors"); ok {
		params.MaxErrors = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_concurrency"); ok {
		params.MaxConcurrency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("targets"); ok {
		params.Targets = expandTargets(v.([]interface{}))
	} else {
		params.MaxConcurrency = nil
		params.MaxErrors = nil
	}

	if v, ok := d.GetOk("cutoff_behavior"); ok {
		params.CutoffBehavior = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		params.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("task_invocation_parameters"); ok {
		params.TaskInvocationParameters = expandTaskInvocationParameters(v.([]interface{}))
	}

	_, err := conn.UpdateMaintenanceWindowTask(params)
	if err != nil {
		return fmt.Errorf("error updating Maintenance Window (%s) Task (%s): %w", windowID, d.Id(), err)
	}

	return resourceMaintenanceWindowTaskRead(d, meta)
}

func resourceMaintenanceWindowTaskDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn

	log.Printf("[INFO] Deregistering SSM Maintenance Window Task: %s", d.Id())

	params := &ssm.DeregisterTaskFromMaintenanceWindowInput{
		WindowId:     aws.String(d.Get("window_id").(string)),
		WindowTaskId: aws.String(d.Id()),
	}

	_, err := conn.DeregisterTaskFromMaintenanceWindow(params)
	if tfawserr.ErrCodeEquals(err, ssm.ErrCodeDoesNotExistException) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deregistering SSM Maintenance Window Task (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceMaintenanceWindowTaskImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
