package autoscaling

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const ScheduleTimeLayout = "2006-01-02T15:04:05Z"

func ResourceSchedule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSchedulePut,
		ReadWithoutTimeout:   resourceScheduleRead,
		UpdateWithoutTimeout: resourceSchedulePut,
		DeleteWithoutTimeout: resourceScheduleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceScheduleImport,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"autoscaling_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"desired_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"end_time": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validScheduleTimestamp,
			},
			"max_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"min_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"recurrence": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"scheduled_action_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"start_time": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validScheduleTimestamp,
			},
			"time_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceSchedulePut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn()

	name := d.Get("scheduled_action_name").(string)
	input := &autoscaling.PutScheduledUpdateGroupActionInput{
		AutoScalingGroupName: aws.String(d.Get("autoscaling_group_name").(string)),
		ScheduledActionName:  aws.String(name),
	}

	if v, ok := d.GetOk("end_time"); ok {
		v, _ := time.Parse(ScheduleTimeLayout, v.(string))

		input.EndTime = aws.Time(v)
	}

	if v, ok := d.GetOk("recurrence"); ok {
		input.Recurrence = aws.String(v.(string))
	}

	if v, ok := d.GetOk("start_time"); ok {
		v, _ := time.Parse(ScheduleTimeLayout, v.(string))

		input.StartTime = aws.Time(v)
	}

	if v, ok := d.GetOk("time_zone"); ok {
		input.TimeZone = aws.String(v.(string))
	}

	// Scheduled actions don't need to set all three size parameters. For example,
	// you may want to change the min or max without also forcing an immediate
	// resize by changing a desired_capacity that may have changed due to other
	// autoscaling rules. Since Terraform doesn't have a great pattern for
	// differentiating between 0 and unset fields, we accept "-1" to mean "don't
	// include this parameter in the action".
	minSize := int64(d.Get("min_size").(int))
	maxSize := int64(d.Get("max_size").(int))
	desiredCapacity := int64(d.Get("desired_capacity").(int))
	if minSize != -1 {
		input.MinSize = aws.Int64(minSize)
	}
	if maxSize != -1 {
		input.MaxSize = aws.Int64(maxSize)
	}
	if desiredCapacity != -1 {
		input.DesiredCapacity = aws.Int64(desiredCapacity)
	}

	log.Printf("[INFO] Putting Auto Scaling Scheduled Action: %s", input)
	_, err := conn.PutScheduledUpdateGroupActionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Auto Scaling Scheduled Action (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceScheduleRead(ctx, d, meta)...)
}

func resourceScheduleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn()

	sa, err := FindScheduledUpdateGroupAction(ctx, conn, d.Get("autoscaling_group_name").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Auto Scaling Scheduled Action %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Scheduled Action (%s): %s", d.Id(), err)
	}

	d.Set("arn", sa.ScheduledActionARN)
	d.Set("autoscaling_group_name", sa.AutoScalingGroupName)
	if sa.DesiredCapacity == nil {
		d.Set("desired_capacity", -1)
	} else {
		d.Set("desired_capacity", sa.DesiredCapacity)
	}
	if sa.EndTime != nil {
		d.Set("end_time", sa.EndTime.Format(ScheduleTimeLayout))
	}
	if sa.MaxSize == nil {
		d.Set("max_size", -1)
	} else {
		d.Set("max_size", sa.MaxSize)
	}
	if sa.MinSize == nil {
		d.Set("min_size", -1)
	} else {
		d.Set("min_size", sa.MinSize)
	}
	d.Set("recurrence", sa.Recurrence)
	if sa.StartTime != nil {
		d.Set("start_time", sa.StartTime.Format(ScheduleTimeLayout))
	}
	d.Set("time_zone", sa.TimeZone)

	return diags
}

func resourceScheduleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn()

	log.Printf("[INFO] Deleting Auto Scaling Scheduled Action: %s", d.Id())
	_, err := conn.DeleteScheduledActionWithContext(ctx, &autoscaling.DeleteScheduledActionInput{
		AutoScalingGroupName: aws.String(d.Get("autoscaling_group_name").(string)),
		ScheduledActionName:  aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Auto Scaling Scheduled Action (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceScheduleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	splitId := strings.Split(d.Id(), "/")
	if len(splitId) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'asg-name/action-name'", d.Id())
	}

	asgName := splitId[0]
	actionName := splitId[1]

	err := d.Set("autoscaling_group_name", asgName)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("failed to set autoscaling_group_name value")
	}
	err = d.Set("scheduled_action_name", actionName)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("failed to set scheduled_action_name value")
	}
	d.SetId(actionName)
	return []*schema.ResourceData{d}, nil
}

func FindScheduledUpdateGroupAction(ctx context.Context, conn *autoscaling.AutoScaling, asgName, actionName string) (*autoscaling.ScheduledUpdateGroupAction, error) {
	input := &autoscaling.DescribeScheduledActionsInput{
		AutoScalingGroupName: aws.String(asgName),
		ScheduledActionNames: aws.StringSlice([]string{actionName}),
	}
	var output []*autoscaling.ScheduledUpdateGroupAction

	err := conn.DescribeScheduledActionsPagesWithContext(ctx, input, func(page *autoscaling.DescribeScheduledActionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ScheduledUpdateGroupActions {
			if v == nil || aws.StringValue(v.ScheduledActionName) != actionName {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "not found") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func validScheduleTimestamp(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := time.Parse(ScheduleTimeLayout, value)
	if err != nil {
		errors = append(errors, fmt.Errorf(
			"%q cannot be parsed as iso8601 Timestamp Format", value))
	}

	return
}
