package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceSchedule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceScheduleCreate,
		ReadWithoutTimeout:   resourceScheduleRead,
		UpdateWithoutTimeout: resourceScheduleUpdate,
		DeleteWithoutTimeout: resourceScheduleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z-_.]+$`), `The name must consist of alphanumerics, hyphens, underscores and dot(.).`),
				)),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 64-resource.UniqueIDSuffixLength),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z-_.]+$`), `The name must consist of alphanumerics, hyphens, and underscores.`),
				)),
			},
			"client_token": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z-_]+$`), `must consist of alphanumerics, hyphens and underscores.`),
				)),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"end_date": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"flexible_time_window": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maximum_window_in_minutes": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 1440),
						},
						"mode": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(toStringSlice(types.FlexibleTimeWindowMode("").Values()), false),
						},
					},
				},
			},
			"group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z-_.]+$`), `must consist of alphanumerics, hyphens, underscores and dot(.).`),
				)),
			},
			"kms_key_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 2048),
					validation.StringMatch(regexp.MustCompile(`^arn:aws(-[a-z]+)?:kms:[a-z0-9\-]+:\d{12}:(key|alias)\/[0-9a-zA-Z-_]*$`), `must be arn of KMS key or alias`),
				)),
			},
			"schedule_expression": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"schedule_expression_timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"start_date": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"state": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(toStringSlice(types.ScheduleState("").Values()), false),
			},
			"target": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"dead_letter_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"input": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"retry_policy": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"maximum_event_age_in_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(60, 86400),
									},
									"maximum_retry_attempts": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(0, 185),
									},
								},
							},
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},

						"ecs_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"capacity_provider_strategy": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 6,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"base": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(0, 100000),
												},
												"capacity_provider": {
													Type:     schema.TypeString,
													Required: true,
												},
												"weight": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(0, 1000),
												},
											},
										},
									},
									"enable_ecs_managed_tags": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"enable_execute_command": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"group": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"launch_type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(toStringSlice(types.LaunchType("").Values()), false),
									},
									"network_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"assign_public_ip": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(toStringSlice(types.AssignPublicIp("").Values()), false),
												},
												"security_groups": {
													Type:     schema.TypeSet,
													Optional: true,
													MaxItems: 5,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"subnets": {
													Type:     schema.TypeSet,
													Required: true,
													MaxItems: 16,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"placement_constraint": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 10,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"expression": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 2000),
												},
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(toStringSlice(types.PlacementConstraintType("").Values()), false),
												},
											},
										},
									},
									"placement_strategy": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 5,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"field": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 255),
												},
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(toStringSlice(types.PlacementStrategyType("").Values()), false),
												},
											},
										},
									},
									"platform_version": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 64),
									},
									"propagate_tags": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(toStringSlice(types.PropagateTags("").Values()), false),
									},
									"reference_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 1024),
									},
									"tags": tftags.TagsSchema(),
									"task_count": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 10),
										Default:      1,
									},
									"task_definition_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"event_bridge_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"detail_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									"source": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
								},
							},
						},
						"kinesis_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"partition_key": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
								},
							},
						},
						"sage_maker_pipeline_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"pipeline_parameter_list": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 200,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 256),
												},
												"value": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},
											},
										},
									},
								},
							},
						},
						"sqs_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"message_group_id": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
								},
							},
						},
					},
				},
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modification_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func toStringSlice[T any](origin []T) []string {
	out := make([]string, 0, len(origin))

	for _, v := range origin {
		out = append(out, fmt.Sprintf("%v", v))
	}

	return out
}

const (
	ResNameSchedule = "Schedule"
)

func resourceScheduleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchedulerClient

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))

	in := &scheduler.CreateScheduleInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("client_token"); ok {
		in.ClientToken = aws.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("end_date"); ok {
		tv, _ := time.Parse(time.RFC3339, v.(string))
		in.EndDate = aws.Time(tv)
	}
	if v, ok := d.Get("flexible_time_window").([]map[string]interface{}); ok && len(v) > 0 {
		in.FlexibleTimeWindow = expandFlexibleTimeWindow(v[0])
	}
	if v, ok := d.GetOk("group_name"); ok {
		in.GroupName = aws.String(v.(string))
	}
	if v, ok := d.GetOk("kms_key_arn"); ok {
		in.KmsKeyArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("schedule_expression"); ok {
		in.ScheduleExpression = aws.String(v.(string))
	}
	if v, ok := d.GetOk("schedule_expression_timezone"); ok {
		in.ScheduleExpressionTimezone = aws.String(v.(string))
	}
	if v, ok := d.GetOk("start_date"); ok {
		tv, _ := time.Parse(time.RFC3339, v.(string))
		in.StartDate = aws.Time(tv)
	}
	if v, ok := d.GetOk("state"); ok {
		in.State = types.ScheduleState(v.(string))
	}
	if v, ok := d.Get("target").([]map[string]interface{}); ok && len(v) > 0 {
		defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
		in.Target = expandTarget(v[0], defaultTagsConfig)
	}

	out, err := conn.CreateSchedule(ctx, in)
	if err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionCreating, ResNameSchedule, name, err)
	}

	if out == nil || out.ScheduleArn == nil {
		return create.DiagError(names.Scheduler, create.ErrActionCreating, ResNameSchedule, name, errors.New("empty output"))
	}

	d.SetId(name)
	return resourceScheduleRead(ctx, d, meta)
}

func resourceScheduleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchedulerClient

	out, err := findSchedule(ctx, conn, d.Id(), d.Get("group_name").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Scheduler Schedule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionReading, ResNameSchedule, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	if v := out.CreationDate; v != nil {
		d.Set("creation_date", aws.ToTime(out.CreationDate).Format(time.RFC3339))
	}
	if v := out.LastModificationDate; v != nil {
		d.Set("last_modification_date", aws.ToTime(out.LastModificationDate).Format(time.RFC3339))
	}
	d.Set("name", out.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.ToString(out.Name)))

	d.Set("description", out.Description)
	if v := out.EndDate; v != nil {
		d.Set("end_date", aws.ToTime(out.EndDate).Format(time.RFC3339))
	}
	d.Set("flexible_time_window", flattenFlexibleTimeWindow(out.FlexibleTimeWindow))
	d.Set("group_name", out.GroupName)
	d.Set("kms_key_arn", out.KmsKeyArn)
	d.Set("schedule_expression_timezone", out.ScheduleExpressionTimezone)
	if v := out.StartDate; v != nil {
		d.Set("start_date", aws.ToTime(out.StartDate).Format(time.RFC3339))
	}
	d.Set("state", out.State)
	d.Set("target", flattenTarget(out.Target))

	return nil
}

func resourceScheduleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchedulerClient

	in := &scheduler.UpdateScheduleInput{
		Name: aws.String(d.Id()),
	}

	updateRequires := []string{"client_token", "description", "end_date", "flexible_time_window", "group_name", "kms_key_arn",
		"schedule_expression", "schedule_expression_timezone", "start_date", "state", "target"}
	if d.HasChanges(updateRequires...) {
		if v, ok := d.GetOk("client_token"); ok {
			in.ClientToken = aws.String(v.(string))
		}
		if v, ok := d.GetOk("description"); ok {
			in.Description = aws.String(v.(string))
		}
		if v, ok := d.GetOk("end_date"); ok {
			tv, _ := time.Parse(time.RFC3339, v.(string))
			in.EndDate = aws.Time(tv)
		}
		if v, ok := d.Get("flexible_time_window").([]map[string]interface{}); ok && len(v) > 0 {
			in.FlexibleTimeWindow = expandFlexibleTimeWindow(v[0])
		}
		if v, ok := d.GetOk("group_name"); ok {
			in.GroupName = aws.String(v.(string))
		}
		if v, ok := d.GetOk("kms_key_arn"); ok {
			in.KmsKeyArn = aws.String(v.(string))
		}
		if v, ok := d.GetOk("schedule_expression"); ok {
			in.ScheduleExpression = aws.String(v.(string))
		}
		if v, ok := d.GetOk("schedule_expression_timezone"); ok {
			in.ScheduleExpressionTimezone = aws.String(v.(string))
		}
		if v, ok := d.GetOk("start_date"); ok {
			tv, _ := time.Parse(time.RFC3339, v.(string))
			in.StartDate = aws.Time(tv)
		}
		if v, ok := d.GetOk("state"); ok {
			in.State = types.ScheduleState(v.(string))
		}
		if v, ok := d.Get("target").([]map[string]interface{}); ok && len(v) > 0 {
			defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
			in.Target = expandTarget(v[0], defaultTagsConfig)
		}

		if _, err := conn.UpdateSchedule(ctx, in); err != nil {
			return create.DiagError(names.Scheduler, create.ErrActionUpdating, ResNameSchedule, d.Id(), err)
		}
	}

	return resourceScheduleRead(ctx, d, meta)
}

func resourceScheduleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchedulerClient

	log.Printf("[INFO] Deleting EventBridge Scheduler Schedule %s", d.Id())

	param := &scheduler.DeleteScheduleInput{
		Name: aws.String(d.Id()),
	}
	if groupName, ok := d.Get("group_name").(string); ok && groupName != "" {
		param.GroupName = aws.String(groupName)
	}
	_, err := conn.DeleteSchedule(ctx, param)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}
		return create.DiagError(names.Scheduler, create.ErrActionDeleting, ResNameSchedule, d.Id(), err)
	}

	return nil
}

func expandFlexibleTimeWindow(tfMap map[string]interface{}) *types.FlexibleTimeWindow {
	if tfMap == nil {
		return nil
	}

	config := &types.FlexibleTimeWindow{}

	if v, ok := tfMap["maximum_window_in_minutes"].(int32); ok {
		config.MaximumWindowInMinutes = aws.Int32(v)
	}
	if v, ok := tfMap["mode"].(types.FlexibleTimeWindowMode); ok && v != "" {
		config.Mode = v
	}

	return config
}

func flattenFlexibleTimeWindow(apiObject *types.FlexibleTimeWindow) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.MaximumWindowInMinutes; v != nil {
		tfMap["maximum_window_in_minutes"] = aws.ToInt32(v)
	}
	if v := apiObject.Mode; v != "" {
		tfMap["mode"] = v
	}

	return tfMap
}

func expandTarget(tfMap map[string]interface{}, defaultTags *tftags.DefaultConfig) *types.Target {
	if tfMap == nil {
		return nil
	}

	config := &types.Target{}

	if v, ok := tfMap["arn"].(string); ok && v != "" {
		config.Arn = aws.String(v)
	}
	if v, ok := tfMap["dead_letter_config"].([]map[string]interface{}); ok && len(v) > 0 {
		config.DeadLetterConfig = expandDeadLetterConfig(v[0])
	}
	if v, ok := tfMap["input"].(string); ok && v != "" {
		config.Input = aws.String(v)
	}
	if v, ok := tfMap["retry_policy"].([]map[string]interface{}); ok && len(v) > 0 {
		config.RetryPolicy = expandRetryPolicy(v[0])
	}
	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		config.RoleArn = aws.String(v)
	}
	if v, ok := tfMap["ecs_parameters"].([]map[string]interface{}); ok && len(v) > 0 {
		config.EcsParameters = expandECSParameters(v[0], defaultTags)
	}
	if v, ok := tfMap["event_bridge_parameters"].([]map[string]interface{}); ok && len(v) > 0 {
		config.EventBridgeParameters = expandEventBridgeParameters(v[0])
	}
	if v, ok := tfMap["kinesis_parameters"].([]map[string]interface{}); ok && len(v) > 0 {
		config.KinesisParameters = expandKinesisParameters(v[0])
	}
	if v, ok := tfMap["sage_maker_pipeline_parameters"].([]map[string]interface{}); ok && len(v) > 0 {
		config.SageMakerPipelineParameters = expandSageMakerPipelineParameters(v[0])
	}
	if v, ok := tfMap["sqs_parameters"].([]map[string]interface{}); ok && len(v) > 0 {
		config.SqsParameters = expandSQSParameters(v[0])
	}

	return config
}

func flattenTarget(apiObject *types.Target) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Arn; v != nil {
		tfMap["arn"] = aws.ToString(v)
	}
	if v := apiObject.DeadLetterConfig; v != nil {
		tfMap["dead_letter_config"] = flattenDeadLetterConfig(v)
	}
	if v := apiObject.Input; v != nil {
		tfMap["input"] = aws.ToString(v)
	}
	if v := apiObject.RetryPolicy; v != nil {
		tfMap["retry_policy"] = flattenRetryPolicy(v)
	}
	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.ToString(v)
	}
	if v := apiObject.EcsParameters; v != nil {
		tfMap["ecs_parameters"] = flattenECSParameters(v)
	}
	if v := apiObject.EventBridgeParameters; v != nil {
		tfMap["event_bridge_parameters"] = flattenEventBridgeParameters(v)
	}
	if v := apiObject.KinesisParameters; v != nil {
		tfMap["kinesis_parameters"] = flattenKinesisParameters(v)
	}
	if v := apiObject.SageMakerPipelineParameters; v != nil {
		tfMap["sage_maker_pipeline_parameters"] = flattenSageMakerPipelineParameters(v)
	}
	if v := apiObject.SqsParameters; v != nil {
		tfMap["sqs_parameters"] = flattenSQSParameters(v)
	}

	return tfMap
}

func expandDeadLetterConfig(tfMap map[string]interface{}) *types.DeadLetterConfig {
	if tfMap == nil {
		return nil
	}

	config := &types.DeadLetterConfig{}

	if v, ok := tfMap["arn"].(string); ok && v != "" {
		config.Arn = aws.String(v)
	}

	return config
}

func flattenDeadLetterConfig(apiObject *types.DeadLetterConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Arn; v != nil {
		tfMap["arn"] = aws.ToString(v)
	}

	return tfMap
}

func expandRetryPolicy(tfMap map[string]interface{}) *types.RetryPolicy {
	if tfMap == nil {
		return nil
	}

	config := &types.RetryPolicy{}

	if v, ok := tfMap["maximum_event_age_in_seconds"].(int32); ok {
		config.MaximumEventAgeInSeconds = aws.Int32(v)
	}
	if v, ok := tfMap["maximum_retry_attempts"].(int32); ok {
		config.MaximumRetryAttempts = aws.Int32(v)
	}

	return config
}

func flattenRetryPolicy(apiObject *types.RetryPolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.MaximumEventAgeInSeconds; v != nil {
		tfMap["maximum_event_age_in_seconds"] = aws.ToInt32(v)
	}
	if v := apiObject.MaximumRetryAttempts; v != nil {
		tfMap["maximum_retry_attempts"] = aws.ToInt32(v)
	}

	return tfMap
}

func expandECSParameters(tfMap map[string]interface{}, defaultTags *tftags.DefaultConfig) *types.EcsParameters {
	if tfMap == nil {
		return nil
	}

	config := &types.EcsParameters{}

	if v, ok := tfMap["capacity_provider_strategy"].([]map[string]interface{}); ok {
		config.CapacityProviderStrategy = expandCapacityProviderStrategy(v)
	}
	if v, ok := tfMap["enable_ecs_managed_tags"].(bool); ok {
		config.EnableECSManagedTags = aws.Bool(v)
	}
	if v, ok := tfMap["enable_execute_command"].(bool); ok {
		config.EnableExecuteCommand = aws.Bool(v)
	}
	if v, ok := tfMap["group"].(string); ok {
		config.Group = aws.String(v)
	}
	if v, ok := tfMap["launch_type"].(string); ok {
		config.LaunchType = types.LaunchType(v)
	}
	if v, ok := tfMap["network_configuration"].([]map[string]interface{}); ok && len(v) > 0 {
		config.NetworkConfiguration = expandNetworkConfiguration(v[0])
	}
	if v, ok := tfMap["placement_constraint"].([]map[string]interface{}); ok && len(v) > 0 {
		config.PlacementConstraints = expandPlacementConstraints(v)
	}
	if v, ok := tfMap["placement_strategy"].([]map[string]interface{}); ok && len(v) > 0 {
		config.PlacementStrategy = expandPlacementStrategy(v)
	}
	if v, ok := tfMap["platform_version"].(string); ok {
		config.PlatformVersion = aws.String(v)
	}
	if v, ok := tfMap["propagate_tags"].(string); ok {
		config.PropagateTags = types.PropagateTags(v)
	}
	if v, ok := tfMap["reference_id"].(string); ok {
		config.ReferenceId = aws.String(v)
	}
	tags := defaultTags.MergeTags(tftags.New(tfMap["tags"].(map[string]interface{})))
	if len(tags) > 0 {
		config.Tags = []map[string]string{tags.Map()}
	}
	if v, ok := tfMap["task_count"].(int32); ok {
		config.TaskCount = aws.Int32(v)
	}
	if v, ok := tfMap["task_definition_arn"].(string); ok {
		config.TaskDefinitionArn = aws.String(v)
	}

	return config
}

func flattenECSParameters(apiObject *types.EcsParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["capacity_provider_strategy"] = flattenCapacityProviderStrategy(apiObject.CapacityProviderStrategy)
	if v := apiObject.EnableECSManagedTags; v != nil {
		tfMap["enable_ecs_managed_tags"] = aws.ToBool(v)
	}
	if v := apiObject.EnableExecuteCommand; v != nil {
		tfMap["enable_execute_command"] = aws.ToBool(v)
	}
	if v := apiObject.Group; v != nil {
		tfMap["group"] = aws.ToString(v)
	}
	if v := apiObject.LaunchType; v != "" {
		tfMap["launch_type"] = string(v)
	}
	tfMap["network_configuration"] = flattenNetworkConfiguration(apiObject.NetworkConfiguration)
	tfMap["placement_constraint"] = flattenPlacementConstraints(apiObject.PlacementConstraints)
	tfMap["placement_strategy"] = flattenPlacementStrategy(apiObject.PlacementStrategy)
	if v := apiObject.PlatformVersion; v != nil {
		tfMap["platform_version"] = aws.ToString(v)
	}
	if v := apiObject.PropagateTags; v != "" {
		tfMap["propagate_tags"] = string(v)
	}
	if v := apiObject.ReferenceId; v != nil {
		tfMap["reference_id"] = aws.ToString(v)
	}
	if v := apiObject.Tags; len(v) > 0 {
		tfMap["tags"] = tftags.New(v).Map()
	}
	if v := apiObject.TaskCount; v != nil {
		tfMap["task_count"] = aws.ToInt32(v)
	}
	if v := apiObject.TaskDefinitionArn; v != nil {
		tfMap["task_definition_arn"] = aws.ToString(v)
	}

	return tfMap
}

func expandCapacityProviderStrategy(tfMap []map[string]interface{}) []types.CapacityProviderStrategyItem {
	if len(tfMap) == 0 {
		return nil
	}

	config := make([]types.CapacityProviderStrategyItem, 0, len(tfMap))

	for _, v := range tfMap {
		element := types.CapacityProviderStrategyItem{}
		if v, ok := v["base"].(int32); ok {
			element.Base = v
		}
		if v, ok := v["capacity_provider"].(string); ok && v != "" {
			element.CapacityProvider = aws.String(v)
		}
		if v, ok := v["weight"].(int32); ok {
			element.Weight = v
		}
		config = append(config, element)
	}

	return config
}

func flattenCapacityProviderStrategy(apiObject []types.CapacityProviderStrategyItem) []map[string]interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	tfMap := make([]map[string]interface{}, 0, len(apiObject))

	for _, param := range apiObject {
		element := map[string]interface{}{}

		element["base"] = param.Base
		if v := param.CapacityProvider; v != nil {
			element["capacity_provider"] = aws.ToString(v)
		}
		element["weight"] = param.Weight

		tfMap = append(tfMap, element)
	}

	return tfMap
}

func expandNetworkConfiguration(tfMap map[string]interface{}) *types.NetworkConfiguration {
	if tfMap == nil {
		return nil
	}

	config := &types.NetworkConfiguration{}

	if v, ok := tfMap["assign_public_ip"].(string); ok {
		config.AwsvpcConfiguration.AssignPublicIp = types.AssignPublicIp(v)
	}
	if v, ok := tfMap["security_groups"].(*schema.Set); ok {
		config.AwsvpcConfiguration.SecurityGroups = flex.ExpandStringValueSet(v)
	}
	if v, ok := tfMap["subnets"].(*schema.Set); ok {
		config.AwsvpcConfiguration.Subnets = flex.ExpandStringValueSet(v)
	}

	return config
}

func flattenNetworkConfiguration(apiObject *types.NetworkConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AwsvpcConfiguration.AssignPublicIp; v != "" {
		tfMap["assign_public_ip"] = string(v)
	}
	tfMap["security_groups"] = flex.FlattenStringValueSet(apiObject.AwsvpcConfiguration.SecurityGroups)
	tfMap["subnets"] = flex.FlattenStringValueSet(apiObject.AwsvpcConfiguration.Subnets)

	return tfMap
}

func expandPlacementConstraints(tfMap []map[string]interface{}) []types.PlacementConstraint {
	if len(tfMap) == 0 {
		return nil
	}

	config := make([]types.PlacementConstraint, 0, len(tfMap))

	for _, v := range tfMap {
		element := types.PlacementConstraint{}
		if v, ok := v["expression"].(string); ok && v != "" {
			element.Expression = aws.String(v)
		}
		if v, ok := v["type"].(string); ok && v != "" {
			element.Type = types.PlacementConstraintType(v)
		}
		config = append(config, element)
	}

	return config
}

func flattenPlacementConstraints(apiObject []types.PlacementConstraint) []map[string]interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	tfMap := make([]map[string]interface{}, 0, len(apiObject))

	for _, param := range apiObject {
		element := map[string]interface{}{}

		if v := param.Expression; v != nil {
			element["expression"] = aws.ToString(v)
		}
		element["type"] = string(param.Type)

		tfMap = append(tfMap, element)
	}

	return tfMap
}

func expandPlacementStrategy(tfMap []map[string]interface{}) []types.PlacementStrategy {
	if len(tfMap) == 0 {
		return nil
	}

	config := make([]types.PlacementStrategy, 0, len(tfMap))

	for _, v := range tfMap {
		element := types.PlacementStrategy{}
		if v, ok := v["field"].(string); ok && v != "" {
			element.Field = aws.String(v)
		}
		if v, ok := v["type"].(string); ok && v != "" {
			element.Type = types.PlacementStrategyType(v)
		}
		config = append(config, element)
	}

	return config
}

func flattenPlacementStrategy(apiObject []types.PlacementStrategy) []map[string]interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	tfMap := make([]map[string]interface{}, 0, len(apiObject))

	for _, param := range apiObject {
		element := map[string]interface{}{}

		if v := param.Field; v != nil {
			element["field"] = aws.ToString(v)
		}
		element["type"] = string(param.Type)

		tfMap = append(tfMap, element)
	}

	return tfMap
}

func expandEventBridgeParameters(tfMap map[string]interface{}) *types.EventBridgeParameters {
	if tfMap == nil {
		return nil
	}

	config := &types.EventBridgeParameters{}

	if v, ok := tfMap["detail_type"].(string); ok && v != "" {
		config.DetailType = aws.String(v)
	}
	if v, ok := tfMap["source"].(string); ok && v != "" {
		config.Source = aws.String(v)
	}

	return config
}

func flattenEventBridgeParameters(apiObject *types.EventBridgeParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DetailType; v != nil {
		tfMap["detail_type"] = aws.ToString(v)
	}
	if v := apiObject.Source; v != nil {
		tfMap["source"] = aws.ToString(v)
	}

	return tfMap
}

func expandKinesisParameters(tfMap map[string]interface{}) *types.KinesisParameters {
	if tfMap == nil {
		return nil
	}

	config := &types.KinesisParameters{}

	if v, ok := tfMap["partition_key"].(string); ok && v != "" {
		config.PartitionKey = aws.String(v)
	}

	return config
}

func flattenKinesisParameters(apiObject *types.KinesisParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.PartitionKey; v != nil {
		tfMap["partition_key"] = aws.ToString(v)
	}

	return tfMap
}

func expandSageMakerPipelineParameters(tfMap map[string]interface{}) *types.SageMakerPipelineParameters {
	if len(tfMap) == 0 {
		return nil
	}

	ppMap, ok := tfMap["pipeline_parameter_list"].([]map[string]interface{})
	if !ok || len(ppMap) == 0 {
		return nil
	}

	config := &types.SageMakerPipelineParameters{
		PipelineParameterList: make([]types.SageMakerPipelineParameter, 0, len(ppMap)),
	}

	for _, v := range ppMap {
		element := &types.SageMakerPipelineParameter{}
		if v, ok := v["name"].(string); ok && v != "" {
			element.Name = aws.String(v)
		}
		if v, ok := v["value"].(string); ok && v != "" {
			element.Value = aws.String(v)
		}
		config.PipelineParameterList = append(config.PipelineParameterList, *element)
	}

	return config
}

func flattenSageMakerPipelineParameters(apiObject *types.SageMakerPipelineParameters) map[string]interface{} {
	if apiObject == nil || len(apiObject.PipelineParameterList) == 0 {
		return nil
	}

	ppMap := make([]map[string]interface{}, 0, len(apiObject.PipelineParameterList))

	for _, param := range apiObject.PipelineParameterList {
		element := map[string]interface{}{}

		if v := param.Name; v != nil {
			element["name"] = aws.ToString(v)
		}
		if v := param.Value; v != nil {
			element["value"] = aws.ToString(v)
		}
		ppMap = append(ppMap, element)
	}

	return map[string]interface{}{"pipeline_parameter_list": ppMap}
}

func expandSQSParameters(tfMap map[string]interface{}) *types.SqsParameters {
	if tfMap == nil {
		return nil
	}

	config := &types.SqsParameters{}

	if v, ok := tfMap["message_group_id"].(string); ok && v != "" {
		config.MessageGroupId = aws.String(v)
	}

	return config
}

func flattenSQSParameters(apiObject *types.SqsParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.MessageGroupId; v != nil {
		tfMap["message_group_id"] = aws.ToString(v)
	}

	return tfMap
}
