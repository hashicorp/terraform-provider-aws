package scheduler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
	"strings"
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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	_sp.registerSDKResourceFactory("aws_scheduler_schedule", resourceSchedule)
}

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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
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
						"mode": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.FlexibleTimeWindowMode](),
						},
					},
				},
			},
			"group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringLenBetween(1, 64),
				),
			},
			"kms_key_arn": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z-_.]+$`), `The name must consist of alphanumerics, hyphens, and underscores.`),
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
			"schedule_expression": {
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
			"state": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.ScheduleStateEnabled,
				ValidateDiagFunc: enum.Validate[types.ScheduleState](),
			},
			"target": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
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
									"arn": {
										Type:             schema.TypeString,
										Optional:         true,
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
									"capacity_provider_strategy": {
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
												"weight": {
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
										ValidateDiagFunc: enum.Validate[types.LaunchType](),
									},
									"network_configuration": {
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
												"security_groups": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"subnets": {
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
												"expression": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 2000)),
												},
												"type": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.PlacementConstraintType](),
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
												"field": {
													Type:     schema.TypeString,
													Optional: true,
													DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
														return strings.EqualFold(old, new)
													},
												},
												"type": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.PlacementStrategyType](),
												},
											},
										},
									},
									"platform_version": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"propagate_tags": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.PropagateTags](),
									},
									"reference_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"tags": tftags.TagsSchema(),
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
									"source": {
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
						"role_arn": {
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
												"name": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 256)),
												},
												"value": {
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

const (
	ResNameSchedule = "Schedule"
)

func resourceScheduleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchedulerClient()

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))

	in := &scheduler.CreateScheduleInput{
		Name:               aws.String(name),
		ScheduleExpression: aws.String(d.Get("schedule_expression").(string)),
	}

	if v, ok := d.Get("description").(string); ok && v != "" {
		in.Description = aws.String(v)
	}

	if v, ok := d.Get("end_date").(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)
		in.EndDate = aws.Time(v)
	}

	if v, ok := d.Get("flexible_time_window").([]interface{}); ok && len(v) > 0 {
		in.FlexibleTimeWindow = expandFlexibleTimeWindow(v[0].(map[string]interface{}))
	}

	if v, ok := d.Get("group_name").(string); ok && v != "" {
		in.GroupName = aws.String(v)
	}

	if v, ok := d.Get("kms_key_arn").(string); ok && v != "" {
		in.KmsKeyArn = aws.String(v)
	}

	if v, ok := d.Get("schedule_expression_timezone").(string); ok && v != "" {
		in.ScheduleExpressionTimezone = aws.String(v)
	}

	if v, ok := d.Get("start_date").(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)
		in.StartDate = aws.Time(v)
	}

	if v, ok := d.Get("state").(string); ok && v != "" {
		in.State = types.ScheduleState(v)
	}

	if v, ok := d.Get("target").([]interface{}); ok && len(v) > 0 {
		in.Target = expandTarget(v[0].(map[string]interface{}))
	}

	out, err := retryWhenIAMNotPropagated(ctx, func() (*scheduler.CreateScheduleOutput, error) {
		return conn.CreateSchedule(ctx, in)
	})

	if err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionCreating, ResNameSchedule, name, err)
	}

	if out == nil || out.ScheduleArn == nil {
		return create.DiagError(names.Scheduler, create.ErrActionCreating, ResNameSchedule, name, errors.New("empty output"))
	}

	// When the schedule is created without specifying a group, it is assigned
	// to the "default" schedule group. The group name isn't explicitly available
	// in the output from CreateSchedule.
	//
	// To prevent having this implicit knowledge in the provider, derive the
	// group name from the resource ARN.

	id, err := ResourceScheduleIDFromARN(aws.ToString(out.ScheduleArn))

	if err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionCreating, ResNameSchedule, name, fmt.Errorf("invalid resource id: %w", err))
	}

	d.SetId(id)

	return resourceScheduleRead(ctx, d, meta)
}

func resourceScheduleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { // nosemgrep:ci.scheduler-in-func-name
	conn := meta.(*conns.AWSClient).SchedulerClient()

	groupName, scheduleName, err := ResourceScheduleParseID(d.Id())

	if err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionReading, ResNameSchedule, d.Id(), fmt.Errorf("invalid resource id: %w", err))
	}

	out, err := findScheduleByTwoPartKey(ctx, conn, groupName, scheduleName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Scheduler Schedule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionReading, ResNameSchedule, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("description", out.Description)

	if out.EndDate != nil {
		d.Set("end_date", aws.ToTime(out.EndDate).Format(time.RFC3339))
	} else {
		d.Set("end_date", nil)
	}

	if err := d.Set("flexible_time_window", []interface{}{flattenFlexibleTimeWindow(out.FlexibleTimeWindow)}); err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionSetting, ResNameSchedule, d.Id(), err)
	}

	d.Set("group_name", out.GroupName)
	d.Set("kms_key_arn", out.KmsKeyArn)
	d.Set("name", out.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.ToString(out.Name)))
	d.Set("schedule_expression", out.ScheduleExpression)
	d.Set("schedule_expression_timezone", out.ScheduleExpressionTimezone)

	if out.StartDate != nil {
		d.Set("start_date", aws.ToTime(out.StartDate).Format(time.RFC3339))
	} else {
		d.Set("start_date", nil)
	}

	d.Set("state", string(out.State))

	if err := d.Set("target", []interface{}{flattenTarget(out.Target)}); err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionSetting, ResNameSchedule, d.Id(), err)
	}

	return nil
}

func resourceScheduleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchedulerClient()

	in := &scheduler.UpdateScheduleInput{
		FlexibleTimeWindow: expandFlexibleTimeWindow(d.Get("flexible_time_window").([]interface{})[0].(map[string]interface{})),
		GroupName:          aws.String(d.Get("group_name").(string)),
		Name:               aws.String(d.Get("name").(string)),
		ScheduleExpression: aws.String(d.Get("schedule_expression").(string)),
		Target:             expandTarget(d.Get("target").([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := d.Get("description").(string); ok && v != "" {
		in.Description = aws.String(v)
	}

	if v, ok := d.Get("end_date").(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)
		in.EndDate = aws.Time(v)
	}

	if v, ok := d.Get("kms_key_arn").(string); ok && v != "" {
		in.KmsKeyArn = aws.String(v)
	}

	if v, ok := d.Get("schedule_expression_timezone").(string); ok && v != "" {
		in.ScheduleExpressionTimezone = aws.String(v)
	}

	if v, ok := d.Get("start_date").(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)
		in.StartDate = aws.Time(v)
	}

	if v, ok := d.Get("state").(string); ok && v != "" {
		in.State = types.ScheduleState(v)
	}

	log.Printf("[DEBUG] Updating EventBridge Scheduler Schedule (%s): %#v", d.Id(), in)

	_, err := retryWhenIAMNotPropagated(ctx, func() (*scheduler.UpdateScheduleOutput, error) {
		return conn.UpdateSchedule(ctx, in)
	})

	if err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionUpdating, ResNameSchedule, d.Id(), err)
	}

	return resourceScheduleRead(ctx, d, meta)
}

func resourceScheduleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchedulerClient()

	groupName, scheduleName, err := ResourceScheduleParseID(d.Id())

	if err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionDeleting, ResNameSchedule, d.Id(), fmt.Errorf("invalid resource id: %w", err))
	}

	log.Printf("[INFO] Deleting EventBridge Scheduler Schedule %s", d.Id())

	_, err = conn.DeleteSchedule(ctx, &scheduler.DeleteScheduleInput{
		GroupName: aws.String(groupName),
		Name:      aws.String(scheduleName),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.Scheduler, create.ErrActionDeleting, ResNameSchedule, d.Id(), err)
	}

	return nil
}

func findScheduleByTwoPartKey(ctx context.Context, conn *scheduler.Client, groupName, scheduleName string) (*scheduler.GetScheduleOutput, error) {
	in := &scheduler.GetScheduleInput{
		GroupName: aws.String(groupName),
		Name:      aws.String(scheduleName),
	}

	out, err := conn.GetSchedule(ctx, in)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Arn == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

// ResourceScheduleIDFromARN constructs a string of the form "group_name/schedule_name"
// from the given Schedule ARN.
func ResourceScheduleIDFromARN(arn string) (id string, err error) {
	parts := strings.Split(arn, "/")

	if len(parts) != 3 || parts[1] == "" || parts[2] == "" {
		err = errors.New("expected an schedule arn")
		return
	}

	groupName := parts[1]
	scheduleName := parts[2]

	return fmt.Sprintf("%s/%s", groupName, scheduleName), nil
}

func ResourceScheduleParseID(id string) (groupName, scheduleName string, err error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		err = errors.New("expected a resource id in the form: schedule-group-id/schedule-id")
		return
	}

	return parts[0], parts[1], nil
}

func sagemakerPipelineParameterHash(v interface{}) int {
	m := v.(map[string]interface{})
	return create.StringHashcode(fmt.Sprintf("%s-%s", m["name"].(string), m["value"].(string)))
}

func capacityProviderHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["base"].(int); ok {
		buf.WriteString(fmt.Sprintf("%d-", v))
	}

	if v, ok := m["capacity_provider"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}

	if v, ok := m["weight"].(int); ok {
		buf.WriteString(fmt.Sprintf("%d-", v))
	}

	return create.StringHashcode(buf.String())
}

func placementConstraintHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["expression"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}

	if v, ok := m["type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}

	return create.StringHashcode(buf.String())
}

func placementStrategyHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["field"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}

	if v, ok := m["type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}

	return create.StringHashcode(buf.String())
}
