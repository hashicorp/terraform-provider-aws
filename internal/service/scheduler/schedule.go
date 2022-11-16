package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(0, 512)),
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
							Type:     schema.TypeString,
							Required: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice(
									slices.ApplyToAll(
										types.FlexibleTimeWindowMode("").Values(),
										func(v types.FlexibleTimeWindowMode) string {
											return string(v)
										},
									),
									false),
							),
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
						"role_arn": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
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
	conn := meta.(*conns.AWSClient).SchedulerClient

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))

	in := &scheduler.CreateScheduleInput{
		Name:               aws.String(name),
		ScheduleExpression: aws.String(d.Get("schedule_expression").(string)),
	}

	if v, ok := d.GetOk("description"); ok && v.(string) != "" {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("flexible_time_window"); ok && len(v.([]interface{})) > 0 {
		in.FlexibleTimeWindow = expandFlexibleTimeWindow(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("group_name"); ok {
		in.GroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("target"); ok && len(v.([]interface{})) > 0 {
		in.Target = expandTarget(v.([]interface{})[0].(map[string]interface{}))
	}

	out, err := conn.CreateSchedule(ctx, in)

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

func resourceScheduleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchedulerClient

	groupName, scheduleName, err := ResourceScheduleParseID(d.Id())

	if err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionReading, ResNameSchedule, d.Id(), fmt.Errorf("invalid resource id: %w", err))
	}

	out, err := findScheduleByGroupAndName(ctx, conn, groupName, scheduleName)

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
	d.Set("group_name", out.GroupName)
	d.Set("name", out.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.ToString(out.Name)))
	d.Set("schedule_expression", out.ScheduleExpression)

	if err := d.Set("flexible_time_window", []interface{}{flattenFlexibleTimeWindow(out.FlexibleTimeWindow)}); err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionSetting, ResNameSchedule, d.Id(), err)
	}

	if err := d.Set("target", []interface{}{flattenTarget(out.Target)}); err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionSetting, ResNameSchedule, d.Id(), err)
	}

	return nil
}

func resourceScheduleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchedulerClient

	in := &scheduler.UpdateScheduleInput{
		Description:        aws.String(d.Get("description").(string)),
		FlexibleTimeWindow: expandFlexibleTimeWindow(d.Get("flexible_time_window").([]interface{})[0].(map[string]interface{})),
		GroupName:          aws.String(d.Get("group_name").(string)),
		Name:               aws.String(d.Get("name").(string)),
		ScheduleExpression: aws.String(d.Get("schedule_expression").(string)),
		Target:             expandTarget(d.Get("target").([]interface{})[0].(map[string]interface{})),
	}

	log.Printf("[DEBUG] Updating EventBridge Scheduler Schedule (%s): %#v", d.Id(), in)
	_, err := conn.UpdateSchedule(ctx, in)
	if err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionUpdating, ResNameSchedule, d.Id(), err)
	}

	return resourceScheduleRead(ctx, d, meta)
}

func resourceScheduleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchedulerClient

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
