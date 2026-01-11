// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package scheduler

import (
	"bytes"
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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
										Set:      capacityProviderHash,
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
										Set:      placementConstraintHash,
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
										Set:      placementStrategyHash,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrField: {
													Type:     schema.TypeString,
													Optional: true,
													DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
														return strings.EqualFold(old, new)
													},
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
										Set:      sagemakerPipelineParameterHash,
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

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
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

func sagemakerPipelineParameterHash(v any) int {
	m := v.(map[string]any)
	return create.StringHashcode(fmt.Sprintf("%s-%s", m[names.AttrName].(string), m[names.AttrValue].(string)))
}

func capacityProviderHash(v any) int {
	var buf bytes.Buffer
	m := v.(map[string]any)

	if v, ok := m["base"].(int); ok {
		fmt.Fprintf(&buf, "%d-", v)
	}

	if v, ok := m["capacity_provider"].(string); ok {
		fmt.Fprintf(&buf, "%s-", v)
	}

	if v, ok := m[names.AttrWeight].(int); ok {
		fmt.Fprintf(&buf, "%d-", v)
	}

	return create.StringHashcode(buf.String())
}

func placementConstraintHash(v any) int {
	var buf bytes.Buffer
	m := v.(map[string]any)

	if v, ok := m[names.AttrExpression]; ok {
		fmt.Fprintf(&buf, "%s-", v)
	}

	if v, ok := m[names.AttrType]; ok {
		fmt.Fprintf(&buf, "%s-", v)
	}

	return create.StringHashcode(buf.String())
}

func placementStrategyHash(v any) int {
	var buf bytes.Buffer
	m := v.(map[string]any)

	if v, ok := m[names.AttrField]; ok {
		fmt.Fprintf(&buf, "%s-", v)
	}

	if v, ok := m[names.AttrType]; ok {
		fmt.Fprintf(&buf, "%s-", v)
	}

	return create.StringHashcode(buf.String())
}
