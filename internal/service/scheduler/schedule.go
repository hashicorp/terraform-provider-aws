// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	awstypes "github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_scheduler_schedule", name="Schedule")
func resourceSchedule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceScheduleCreate,
		ReadWithoutTimeout:   resourceScheduleRead,
		UpdateWithoutTimeout: resourceScheduleUpdate,
		DeleteWithoutTimeout: resourceScheduleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"action_after_completion": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ActionAfterCompletion](),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(0, 512)),
			},
			"end_date": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsRFC3339Time),
			},
			"flexible_time_window": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maximum_window_in_minutes": {
							Type:             schema.TypeInt,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 1440)),
						},
						names.AttrMode: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.FlexibleTimeWindowMode](),
						},
					},
				},
			},
			names.AttrGroupName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringLenBetween(1, 64),
				),
			},
			names.AttrKMSKeyARN: {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), `The name must consist of alphanumerics, hyphens, and underscores.`),
				)),
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 64-sdkid.UniqueIDSuffixLength),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), `The name must consist of alphanumerics, hyphens, and underscores.`),
				)),
			},
			names.AttrScheduleExpression: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 256)),
			},
			"schedule_expression_timezone": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "UTC",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 50)),
			},
			"start_date": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsRFC3339Time),
			},
			names.AttrState: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.ScheduleStateEnabled,
				ValidateDiagFunc: enum.Validate[awstypes.ScheduleState](),
			},
			names.AttrTarget: {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrARN: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
						},
						"dead_letter_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrARN: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
									},
								},
							},
						},
						"ecs_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrCapacityProviderStrategy: {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 6,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"base": {
													Type:             schema.TypeInt,
													Optional:         true,
													ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 100000)),
												},
												"capacity_provider": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 255)),
												},
												names.AttrWeight: {
													Type:             schema.TypeInt,
													Optional:         true,
													ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 1000)),
												},
											},
										},
									},
									"enable_ecs_managed_tags": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"enable_execute_command": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"group": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 255)),
									},
									"launch_type": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.LaunchType](),
									},
									names.AttrNetworkConfiguration: {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"assign_public_ip": {
													Type:     schema.TypeBool,
													Optional: true,
													Default:  false,
												},
												names.AttrSecurityGroups: {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												names.AttrSubnets: {
													Type:     schema.TypeSet,
													Required: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"placement_constraints": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 10,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrExpression: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 2000)),
												},
												names.AttrType: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.PlacementConstraintType](),
												},
											},
										},
									},
									"placement_strategy": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 5,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrField: {
													Type:             schema.TypeString,
													Optional:         true,
													DiffSuppressFunc: sdkv2.SuppressEquivalentStringCaseInsensitive,
												},
												names.AttrType: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.PlacementStrategyType](),
												},
											},
										},
									},
									"platform_version": {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrPropagateTags: {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.PropagateTags](),
									},
									"reference_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrTags: tftags.TagsSchema(),
									"task_count": {
										Type:             schema.TypeInt,
										Optional:         true,
										Default:          1,
										ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 10)),
									},
									"task_definition_arn": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
									},
								},
							},
						},
						"eventbridge_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"detail_type": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 128)),
									},
									names.AttrSource: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 256)),
									},
								},
							},
						},
						"input": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, math.MaxInt)),
						},
						"kinesis_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"partition_key": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 256)),
									},
								},
							},
						},
						"retry_policy": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"maximum_event_age_in_seconds": {
										Type:             schema.TypeInt,
										Optional:         true,
										Default:          86400,
										ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(60, 86400)),
									},
									"maximum_retry_attempts": {
										Type:             schema.TypeInt,
										Optional:         true,
										Default:          185,
										ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 185)),
									},
								},
							},
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// Prevent transitive usage of this suppression. This was discovered
								// when attempting to update maximum_retry_attempts from 1 to 0.
								if k != "target.0.retry_policy.#" {
									return false
								}

								return verify.SuppressMissingOptionalConfigurationBlock(k, old, new, d)
							},
						},
						names.AttrRoleARN: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
						},
						"sagemaker_pipeline_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"pipeline_parameter": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 200,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrName: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 256)),
												},
												names.AttrValue: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
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
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 128)),
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

func resourceScheduleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchedulerClient(ctx)

	name := create.Name(ctx, d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	in := scheduler.CreateScheduleInput{
		Name:               aws.String(name),
		ScheduleExpression: aws.String(d.Get(names.AttrScheduleExpression).(string)),
	}

	if v, ok := d.Get("action_after_completion").(string); ok && v != "" {
		in.ActionAfterCompletion = awstypes.ActionAfterCompletion(v)
	}

	if v, ok := d.Get(names.AttrDescription).(string); ok && v != "" {
		in.Description = aws.String(v)
	}

	if v, ok := d.Get("end_date").(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)
		in.EndDate = aws.Time(v)
	}

	if v, ok := d.Get("flexible_time_window").([]any); ok && len(v) > 0 {
		in.FlexibleTimeWindow = expandFlexibleTimeWindow(v[0].(map[string]any))
	}

	if v, ok := d.Get(names.AttrGroupName).(string); ok && v != "" {
		in.GroupName = aws.String(v)
	}

	if v, ok := d.Get(names.AttrKMSKeyARN).(string); ok && v != "" {
		in.KmsKeyArn = aws.String(v)
	}

	if v, ok := d.Get("schedule_expression_timezone").(string); ok && v != "" {
		in.ScheduleExpressionTimezone = aws.String(v)
	}

	if v, ok := d.Get("start_date").(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)
		in.StartDate = aws.Time(v)
	}

	if v, ok := d.Get(names.AttrState).(string); ok && v != "" {
		in.State = awstypes.ScheduleState(v)
	}

	if v, ok := d.Get(names.AttrTarget).([]any); ok && len(v) > 0 {
		in.Target = expandTarget(ctx, v[0].(map[string]any))
	}

	out, err := retryWhenIAMNotPropagated(ctx, func(ctx context.Context) (*scheduler.CreateScheduleOutput, error) {
		return conn.CreateSchedule(ctx, &in)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Scheduler Schedule (%s): %s", name, err)
	}

	// When the schedule is created without specifying a group, it is assigned
	// to the "default" schedule group. The group name isn't explicitly available
	// in the output from CreateSchedule.
	//
	// To prevent having this implicit knowledge in the provider, derive the
	// group name from the resource ARN.
	id, err := scheduleResourceIDFromARN(aws.ToString(out.ScheduleArn))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(id)

	return append(diags, resourceScheduleRead(ctx, d, meta)...)
}

func resourceScheduleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics { // nosemgrep:ci.scheduler-in-func-name
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchedulerClient(ctx)

	groupName, scheduleName, err := scheduleParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	out, err := findScheduleByTwoPartKey(ctx, conn, groupName, scheduleName)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] EventBridge Scheduler Schedule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Scheduler Schedule (%s): %s", d.Id(), err)
	}

	d.Set("action_after_completion", out.ActionAfterCompletion)
	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrDescription, out.Description)
	if out.EndDate != nil {
		d.Set("end_date", aws.ToTime(out.EndDate).Format(time.RFC3339))
	} else {
		d.Set("end_date", nil)
	}
	if err := d.Set("flexible_time_window", []any{flattenFlexibleTimeWindow(out.FlexibleTimeWindow)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting flexible_time_window: %s", err)
	}
	d.Set(names.AttrGroupName, out.GroupName)
	d.Set(names.AttrKMSKeyARN, out.KmsKeyArn)
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(out.Name)))
	d.Set(names.AttrScheduleExpression, out.ScheduleExpression)
	d.Set("schedule_expression_timezone", out.ScheduleExpressionTimezone)
	if out.StartDate != nil {
		d.Set("start_date", aws.ToTime(out.StartDate).Format(time.RFC3339))
	} else {
		d.Set("start_date", nil)
	}
	d.Set(names.AttrState, out.State)
	if err := d.Set(names.AttrTarget, []any{flattenTarget(ctx, out.Target)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting target: %s", err)
	}

	return diags
}

func resourceScheduleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchedulerClient(ctx)

	groupName, scheduleName, err := scheduleParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	in := scheduler.UpdateScheduleInput{
		FlexibleTimeWindow: expandFlexibleTimeWindow(d.Get("flexible_time_window").([]any)[0].(map[string]any)),
		GroupName:          aws.String(groupName),
		Name:               aws.String(scheduleName),
		ScheduleExpression: aws.String(d.Get(names.AttrScheduleExpression).(string)),
		Target:             expandTarget(ctx, d.Get(names.AttrTarget).([]any)[0].(map[string]any)),
	}

	if v, ok := d.Get("action_after_completion").(string); ok && v != "" {
		in.ActionAfterCompletion = awstypes.ActionAfterCompletion(v)
	}

	if v, ok := d.Get(names.AttrDescription).(string); ok && v != "" {
		in.Description = aws.String(v)
	}

	if v, ok := d.Get("end_date").(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)
		in.EndDate = aws.Time(v)
	}

	if v, ok := d.Get(names.AttrKMSKeyARN).(string); ok && v != "" {
		in.KmsKeyArn = aws.String(v)
	}

	if v, ok := d.Get("schedule_expression_timezone").(string); ok && v != "" {
		in.ScheduleExpressionTimezone = aws.String(v)
	}

	if v, ok := d.Get("start_date").(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)
		in.StartDate = aws.Time(v)
	}

	if v, ok := d.Get(names.AttrState).(string); ok && v != "" {
		in.State = awstypes.ScheduleState(v)
	}

	_, err = retryWhenIAMNotPropagated(ctx, func(ctx context.Context) (*scheduler.UpdateScheduleOutput, error) {
		return conn.UpdateSchedule(ctx, &in)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EventBridge Scheduler Schedule (%s): %s", d.Id(), err)
	}

	return append(diags, resourceScheduleRead(ctx, d, meta)...)
}

func resourceScheduleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchedulerClient(ctx)

	groupName, scheduleName, err := scheduleParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting EventBridge Scheduler Schedule: %s", d.Id())
	in := scheduler.DeleteScheduleInput{
		GroupName: aws.String(groupName),
		Name:      aws.String(scheduleName),
	}
	_, err = conn.DeleteSchedule(ctx, &in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Scheduler Schedule (%s): %s", d.Id(), err)
	}

	return diags
}

const (
	iamPropagationTimeout = 2 * time.Minute
)

func retryWhenIAMNotPropagated[T any](ctx context.Context, f func(context.Context) (T, error)) (T, error) {
	return tfresource.RetryWhenIsAErrorMessageContains[T, *awstypes.ValidationException](ctx, iamPropagationTimeout, f, "The execution role you provide must allow AWS EventBridge Scheduler to assume the role.")
}

func findScheduleByTwoPartKey(ctx context.Context, conn *scheduler.Client, groupName, scheduleName string) (*scheduler.GetScheduleOutput, error) {
	in := scheduler.GetScheduleInput{
		GroupName: aws.String(groupName),
		Name:      aws.String(scheduleName),
	}

	return findSchedule(ctx, conn, &in)
}

func findSchedule(ctx context.Context, conn *scheduler.Client, input *scheduler.GetScheduleInput) (*scheduler.GetScheduleOutput, error) {
	output, err := conn.GetSchedule(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Arn == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

const scheduleResourceIDSeparator = "/" // nosemgrep:ci.scheduler-in-var-name,ci.scheduler-in-const-name

func scheduleCreateResourceID(groupName, scheduleName string) string {
	parts := []string{groupName, scheduleName}
	id := strings.Join(parts, scheduleResourceIDSeparator)

	return id
}

// scheduleResourceIDFromARN constructs a string of the form "group_name/schedule_name"
// from the given Schedule ARN.
func scheduleResourceIDFromARN(s string) (id string, err error) { // nosemgrep:ci.scheduler-in-func-name
	v, err := arn.Parse(s)
	if err != nil {
		return "", err
	}

	parts := strings.Split(v.Resource, "/")
	if len(parts) != 3 || parts[1] == "" || parts[2] == "" {
		err = errors.New("expected an schedule ARN")
		return
	}

	return scheduleCreateResourceID(parts[1], parts[2]), nil
}

func scheduleParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, scheduleResourceIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected schedule-group-name%[2]sschedule-name", id, scheduleResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func expandCapacityProviderStrategyItem(tfMap map[string]any) awstypes.CapacityProviderStrategyItem {
	if tfMap == nil {
		return awstypes.CapacityProviderStrategyItem{}
	}

	apiObject := awstypes.CapacityProviderStrategyItem{}

	if v, ok := tfMap["base"].(int); ok {
		apiObject.Base = int32(v)
	}

	if v, ok := tfMap["capacity_provider"].(string); ok && v != "" {
		apiObject.CapacityProvider = aws.String(v)
	}

	if v, ok := tfMap[names.AttrWeight].(int); ok {
		apiObject.Weight = int32(v)
	}

	return apiObject
}

func flattenCapacityProviderStrategyItem(apiObject awstypes.CapacityProviderStrategyItem) map[string]any {
	tfMap := map[string]any{}

	tfMap["base"] = apiObject.Base

	if v := apiObject.CapacityProvider; v != nil {
		tfMap["capacity_provider"] = aws.ToString(v)
	}

	tfMap[names.AttrWeight] = apiObject.Weight

	return tfMap
}

func expandDeadLetterConfig(tfMap map[string]any) *awstypes.DeadLetterConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DeadLetterConfig{}

	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		apiObject.Arn = aws.String(v)
	}

	return apiObject
}

func flattenDeadLetterConfig(apiObject *awstypes.DeadLetterConfig) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Arn; v != nil {
		tfMap[names.AttrARN] = aws.ToString(v)
	}

	return tfMap
}

func expandECSParameters(ctx context.Context, tfMap map[string]any) *awstypes.EcsParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.EcsParameters{}

	if v, ok := tfMap[names.AttrCapacityProviderStrategy].(*schema.Set); ok && v.Len() > 0 {
		for _, v := range v.List() {
			apiObject.CapacityProviderStrategy = append(apiObject.CapacityProviderStrategy, expandCapacityProviderStrategyItem(v.(map[string]any)))
		}
	}

	if v, ok := tfMap["enable_ecs_managed_tags"].(bool); ok {
		apiObject.EnableECSManagedTags = aws.Bool(v)
	}

	if v, ok := tfMap["enable_execute_command"].(bool); ok {
		apiObject.EnableExecuteCommand = aws.Bool(v)
	}

	if v, ok := tfMap["group"].(string); ok && v != "" {
		apiObject.Group = aws.String(v)
	}

	if v, ok := tfMap["launch_type"].(string); ok && v != "" {
		apiObject.LaunchType = awstypes.LaunchType(v)
	}

	if v, ok := tfMap[names.AttrNetworkConfiguration].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.NetworkConfiguration = expandNetworkConfiguration(v[0].(map[string]any))
	}

	if v, ok := tfMap["placement_constraints"].(*schema.Set); ok && v.Len() > 0 {
		for _, v := range v.List() {
			apiObject.PlacementConstraints = append(apiObject.PlacementConstraints, expandPlacementConstraint(v.(map[string]any)))
		}
	}

	if v, ok := tfMap["placement_strategy"].(*schema.Set); ok && v.Len() > 0 {
		for _, v := range v.List() {
			apiObject.PlacementStrategy = append(apiObject.PlacementStrategy, expandPlacementStrategy(v.(map[string]any)))
		}
	}

	if v, ok := tfMap["platform_version"].(string); ok && v != "" {
		apiObject.PlatformVersion = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPropagateTags].(string); ok && v != "" {
		apiObject.PropagateTags = awstypes.PropagateTags(v)
	}

	if v, ok := tfMap["reference_id"].(string); ok && v != "" {
		apiObject.ReferenceId = aws.String(v)
	}

	if tags := tftags.New(ctx, tfMap[names.AttrTags].(map[string]any)); len(tags) > 0 {
		for k, v := range tags.IgnoreAWS().Map() {
			apiObject.Tags = append(apiObject.Tags, map[string]string{
				names.AttrKey:   k,
				names.AttrValue: v,
			})
		}
	}

	if v, ok := tfMap["task_count"].(int); ok {
		apiObject.TaskCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap["task_definition_arn"].(string); ok && v != "" {
		apiObject.TaskDefinitionArn = aws.String(v)
	}

	return apiObject
}

func flattenECSParameters(ctx context.Context, apiObject *awstypes.EcsParameters) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.CapacityProviderStrategy; v != nil {
		var tfList []any

		for _, v := range v {
			tfList = append(tfList, flattenCapacityProviderStrategyItem(v))
		}

		tfMap[names.AttrCapacityProviderStrategy] = tfList
	}

	if v := apiObject.EnableECSManagedTags; v != nil {
		tfMap["enable_ecs_managed_tags"] = aws.ToBool(v)
	}

	if v := apiObject.EnableExecuteCommand; v != nil {
		tfMap["enable_execute_command"] = aws.ToBool(v)
	}

	if v := apiObject.Group; v != nil {
		tfMap["group"] = aws.ToString(v)
	}

	tfMap["launch_type"] = apiObject.LaunchType

	if v := apiObject.NetworkConfiguration; v != nil {
		tfMap[names.AttrNetworkConfiguration] = []any{flattenNetworkConfiguration(v)}
	}

	if v := apiObject.PlacementConstraints; len(v) > 0 {
		var tfList []any

		for _, v := range v {
			tfList = append(tfList, flattenPlacementConstraint(v))
		}

		tfMap["placement_constraints"] = tfList
	}

	if v := apiObject.PlacementStrategy; len(v) > 0 {
		var tfList []any

		for _, v := range v {
			tfList = append(tfList, flattenPlacementStrategy(v))
		}

		tfMap["placement_strategy"] = tfList
	}

	if v := apiObject.PlatformVersion; v != nil {
		tfMap["platform_version"] = aws.ToString(v)
	}

	tfMap[names.AttrPropagateTags] = apiObject.PropagateTags

	if v := apiObject.ReferenceId; v != nil {
		tfMap["reference_id"] = aws.ToString(v)
	}

	if v := apiObject.Tags; len(v) > 0 {
		tags := make(map[string]any)

		for _, v := range v {
			key := v[names.AttrKey]

			// The EventBridge Scheduler API documents raw maps instead of
			// the key-value structure expected by the RunTask API.
			if key == "" {
				continue
			}

			tags[key] = v[names.AttrValue]
		}

		tfMap[names.AttrTags] = tftags.New(ctx, tags).IgnoreAWS().Map()
	}

	if v := apiObject.TaskCount; v != nil {
		tfMap["task_count"] = int(aws.ToInt32(v))
	}

	if v := apiObject.TaskDefinitionArn; v != nil {
		tfMap["task_definition_arn"] = aws.ToString(v)
	}

	return tfMap
}

func expandEventBridgeParameters(tfMap map[string]any) *awstypes.EventBridgeParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.EventBridgeParameters{}

	if v, ok := tfMap["detail_type"].(string); ok && v != "" {
		apiObject.DetailType = aws.String(v)
	}

	if v, ok := tfMap[names.AttrSource].(string); ok && v != "" {
		apiObject.Source = aws.String(v)
	}

	return apiObject
}

func flattenEventBridgeParameters(apiObject *awstypes.EventBridgeParameters) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.DetailType; v != nil {
		tfMap["detail_type"] = aws.ToString(v)
	}

	if v := apiObject.Source; v != nil {
		tfMap[names.AttrSource] = aws.ToString(v)
	}

	return tfMap
}

func expandFlexibleTimeWindow(tfMap map[string]any) *awstypes.FlexibleTimeWindow {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FlexibleTimeWindow{}

	if v, ok := tfMap["maximum_window_in_minutes"].(int); ok && v != 0 {
		apiObject.MaximumWindowInMinutes = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMode].(string); ok && v != "" {
		apiObject.Mode = awstypes.FlexibleTimeWindowMode(v)
	}

	return apiObject
}

func flattenFlexibleTimeWindow(apiObject *awstypes.FlexibleTimeWindow) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.MaximumWindowInMinutes; v != nil {
		tfMap["maximum_window_in_minutes"] = int(aws.ToInt32(v))
	}

	tfMap[names.AttrMode] = apiObject.Mode

	return tfMap
}

func expandKinesisParameters(tfMap map[string]any) *awstypes.KinesisParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.KinesisParameters{}

	if v, ok := tfMap["partition_key"].(string); ok && v != "" {
		apiObject.PartitionKey = aws.String(v)
	}

	return apiObject
}

func flattenKinesisParameters(apiObject *awstypes.KinesisParameters) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.PartitionKey; v != nil {
		tfMap["partition_key"] = aws.ToString(v)
	}

	return tfMap
}

func expandNetworkConfiguration(tfMap map[string]any) *awstypes.NetworkConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AwsVpcConfiguration{}

	if v, ok := tfMap["assign_public_ip"].(bool); ok {
		if v {
			apiObject.AssignPublicIp = awstypes.AssignPublicIpEnabled
		} else {
			apiObject.AssignPublicIp = awstypes.AssignPublicIpDisabled
		}
	}

	if v, ok := tfMap[names.AttrSecurityGroups].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroups = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSubnets].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Subnets = flex.ExpandStringValueSet(v)
	}

	return &awstypes.NetworkConfiguration{
		AwsvpcConfiguration: apiObject,
	}
}

func flattenNetworkConfiguration(apiObject *awstypes.NetworkConfiguration) map[string]any {
	if apiObject == nil || apiObject.AwsvpcConfiguration == nil {
		return nil
	}

	tfMap := map[string]any{}

	// Follow the example of EventBridge targets by flattening out
	// the AWS VPC configuration.

	if v := apiObject.AwsvpcConfiguration.AssignPublicIp; v != "" {
		tfMap["assign_public_ip"] = v == awstypes.AssignPublicIpEnabled
	}

	tfMap[names.AttrSecurityGroups] = apiObject.AwsvpcConfiguration.SecurityGroups
	tfMap[names.AttrSubnets] = apiObject.AwsvpcConfiguration.Subnets

	return tfMap
}

func expandPlacementConstraint(tfMap map[string]any) awstypes.PlacementConstraint {
	if tfMap == nil {
		return awstypes.PlacementConstraint{}
	}

	apiObject := awstypes.PlacementConstraint{}

	if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
		apiObject.Expression = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.PlacementConstraintType(v)
	}

	return apiObject
}

func flattenPlacementConstraint(apiObject awstypes.PlacementConstraint) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Expression; v != nil {
		tfMap[names.AttrExpression] = aws.ToString(v)
	}

	tfMap[names.AttrType] = apiObject.Type

	return tfMap
}

func expandPlacementStrategy(tfMap map[string]any) awstypes.PlacementStrategy {
	if tfMap == nil {
		return awstypes.PlacementStrategy{}
	}

	apiObject := awstypes.PlacementStrategy{}

	if v, ok := tfMap[names.AttrField].(string); ok && v != "" {
		apiObject.Field = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.PlacementStrategyType(v)
	}

	return apiObject
}

func flattenPlacementStrategy(apiObject awstypes.PlacementStrategy) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Field; v != nil {
		tfMap[names.AttrField] = aws.ToString(v)
	}

	tfMap[names.AttrType] = apiObject.Type

	return tfMap
}

func expandRetryPolicy(tfMap map[string]any) *awstypes.RetryPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.RetryPolicy{}

	if v, ok := tfMap["maximum_event_age_in_seconds"].(int); ok {
		apiObject.MaximumEventAgeInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["maximum_retry_attempts"].(int); ok {
		apiObject.MaximumRetryAttempts = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenRetryPolicy(apiObject *awstypes.RetryPolicy) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.MaximumEventAgeInSeconds; v != nil {
		tfMap["maximum_event_age_in_seconds"] = int(aws.ToInt32(v))
	}

	if v := apiObject.MaximumRetryAttempts; v != nil {
		tfMap["maximum_retry_attempts"] = int(aws.ToInt32(v))
	}

	return tfMap
}

func expandSageMakerPipelineParameter(tfMap map[string]any) awstypes.SageMakerPipelineParameter {
	if tfMap == nil {
		return awstypes.SageMakerPipelineParameter{}
	}

	apiObject := awstypes.SageMakerPipelineParameter{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func flattenSageMakerPipelineParameter(apiObject awstypes.SageMakerPipelineParameter) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return tfMap
}

func expandSageMakerPipelineParameters(tfMap map[string]any) *awstypes.SageMakerPipelineParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SageMakerPipelineParameters{}

	if v, ok := tfMap["pipeline_parameter"].(*schema.Set); ok && v.Len() > 0 {
		for _, v := range v.List() {
			apiObject.PipelineParameterList = append(apiObject.PipelineParameterList, expandSageMakerPipelineParameter(v.(map[string]any)))
		}
	}

	return apiObject
}

func flattenSageMakerPipelineParameters(apiObject *awstypes.SageMakerPipelineParameters) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.PipelineParameterList; v != nil {
		var tfList []any

		for _, v := range v {
			tfList = append(tfList, flattenSageMakerPipelineParameter(v))
		}

		tfMap["pipeline_parameter"] = tfList
	}

	return tfMap
}

func expandSQSParameters(tfMap map[string]any) *awstypes.SqsParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SqsParameters{}

	if v, ok := tfMap["message_group_id"].(string); ok && v != "" {
		apiObject.MessageGroupId = aws.String(v)
	}

	return apiObject
}

func flattenSQSParameters(apiObject *awstypes.SqsParameters) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.MessageGroupId; v != nil {
		tfMap["message_group_id"] = aws.ToString(v)
	}

	return tfMap
}

func expandTarget(ctx context.Context, tfMap map[string]any) *awstypes.Target {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Target{}

	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		apiObject.Arn = aws.String(v)
	}

	if v, ok := tfMap["dead_letter_config"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.DeadLetterConfig = expandDeadLetterConfig(v[0].(map[string]any))
	}

	if v, ok := tfMap["ecs_parameters"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.EcsParameters = expandECSParameters(ctx, v[0].(map[string]any))
	}

	if v, ok := tfMap["eventbridge_parameters"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.EventBridgeParameters = expandEventBridgeParameters(v[0].(map[string]any))
	}

	if v, ok := tfMap["input"].(string); ok && v != "" {
		apiObject.Input = aws.String(v)
	}

	if v, ok := tfMap["kinesis_parameters"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.KinesisParameters = expandKinesisParameters(v[0].(map[string]any))
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["retry_policy"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.RetryPolicy = expandRetryPolicy(v[0].(map[string]any))
	}

	if v, ok := tfMap["sagemaker_pipeline_parameters"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.SageMakerPipelineParameters = expandSageMakerPipelineParameters(v[0].(map[string]any))
	}

	if v, ok := tfMap["sqs_parameters"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.SqsParameters = expandSQSParameters(v[0].(map[string]any))
	}

	return apiObject
}

func flattenTarget(ctx context.Context, apiObject *awstypes.Target) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Arn; v != nil {
		tfMap[names.AttrARN] = aws.ToString(v)
	}

	if v := apiObject.DeadLetterConfig; v != nil {
		tfMap["dead_letter_config"] = []any{flattenDeadLetterConfig(v)}
	}

	if v := apiObject.EcsParameters; v != nil {
		tfMap["ecs_parameters"] = []any{flattenECSParameters(ctx, v)}
	}

	if v := apiObject.EventBridgeParameters; v != nil {
		tfMap["eventbridge_parameters"] = []any{flattenEventBridgeParameters(v)}
	}

	if v := apiObject.Input; v != nil {
		tfMap["input"] = aws.ToString(v)
	}

	if v := apiObject.KinesisParameters; v != nil {
		tfMap["kinesis_parameters"] = []any{flattenKinesisParameters(v)}
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	if v := apiObject.RetryPolicy; v != nil {
		tfMap["retry_policy"] = []any{flattenRetryPolicy(v)}
	}

	if v := apiObject.SageMakerPipelineParameters; v != nil {
		tfMap["sagemaker_pipeline_parameters"] = []any{flattenSageMakerPipelineParameters(v)}
	}

	if v := apiObject.SqsParameters; v != nil {
		tfMap["sqs_parameters"] = []any{flattenSQSParameters(v)}
	}

	return tfMap
}
