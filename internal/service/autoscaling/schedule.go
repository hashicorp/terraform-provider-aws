package autoscaling

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const ScheduleTimeLayout = "2006-01-02T15:04:05Z"

func ResourceSchedule() *schema.Resource {
	return &schema.Resource{
		Create: resourceScheduleCreate,
		Read:   resourceScheduleRead,
		Update: resourceScheduleCreate,
		Delete: resourceScheduleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceScheduleImport,
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

func resourceScheduleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	splitId := strings.Split(d.Id(), "/")
	if len(splitId) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of resource: %s. Please follow 'asg-name/action-name'", d.Id())
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

func resourceScheduleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn
	params := &autoscaling.PutScheduledUpdateGroupActionInput{
		AutoScalingGroupName: aws.String(d.Get("autoscaling_group_name").(string)),
		ScheduledActionName:  aws.String(d.Get("scheduled_action_name").(string)),
	}

	if attr, ok := d.GetOk("start_time"); ok {
		t, err := time.Parse(ScheduleTimeLayout, attr.(string))
		if err != nil {
			return fmt.Errorf("Error Parsing AWS Autoscaling Group Schedule Start Time: %w", err)
		}
		params.StartTime = aws.Time(t)
	}

	if attr, ok := d.GetOk("end_time"); ok {
		t, err := time.Parse(ScheduleTimeLayout, attr.(string))
		if err != nil {
			return fmt.Errorf("Error Parsing AWS Autoscaling Group Schedule End Time: %w", err)
		}
		params.EndTime = aws.Time(t)
	}

	if attr, ok := d.GetOk("time_zone"); ok {
		params.TimeZone = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("recurrence"); ok {
		params.Recurrence = aws.String(attr.(string))
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
		params.MinSize = aws.Int64(minSize)
	}
	if maxSize != -1 {
		params.MaxSize = aws.Int64(maxSize)
	}
	if desiredCapacity != -1 {
		params.DesiredCapacity = aws.Int64(desiredCapacity)
	}

	log.Printf("[INFO] Creating Autoscaling Scheduled Action: %s", d.Get("scheduled_action_name").(string))
	_, err := conn.PutScheduledUpdateGroupAction(params)
	if err != nil {
		return fmt.Errorf("Error Creating Autoscaling Scheduled Action: %w", err)
	}

	d.SetId(d.Get("scheduled_action_name").(string))

	return resourceScheduleRead(d, meta)
}

func resourceScheduleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn

	sa, err := FindScheduledUpdateGroupAction(conn, d.Get("autoscaling_group_name").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Auto Scaling Scheduled Action %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Auto Scaling Scheduled Action (%s): %w", d.Id(), err)
	}

	d.Set("autoscaling_group_name", sa.AutoScalingGroupName)
	d.Set("arn", sa.ScheduledActionARN)

	if sa.MinSize == nil {
		d.Set("min_size", -1)
	} else {
		d.Set("min_size", sa.MinSize)
	}
	if sa.MaxSize == nil {
		d.Set("max_size", -1)
	} else {
		d.Set("max_size", sa.MaxSize)
	}
	if sa.DesiredCapacity == nil {
		d.Set("desired_capacity", -1)
	} else {
		d.Set("desired_capacity", sa.DesiredCapacity)
	}

	d.Set("recurrence", sa.Recurrence)

	if sa.StartTime != nil {
		d.Set("start_time", sa.StartTime.Format(ScheduleTimeLayout))
	}

	if sa.EndTime != nil {
		d.Set("end_time", sa.EndTime.Format(ScheduleTimeLayout))
	}

	if sa.TimeZone != nil {
		d.Set("time_zone", sa.TimeZone)
	}

	return nil
}

func resourceScheduleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn

	log.Printf("[INFO] Deleting Auto Scaling Scheduled Action: %s", d.Id())
	_, err := conn.DeleteScheduledAction(&autoscaling.DeleteScheduledActionInput{
		AutoScalingGroupName: aws.String(d.Get("autoscaling_group_name").(string)),
		ScheduledActionName:  aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Auto Scaling Scheduled Action (%s): %w", d.Id(), err)
	}

	return nil
}

func FindScheduledUpdateGroupAction(conn *autoscaling.AutoScaling, asgName, actionName string) (*autoscaling.ScheduledUpdateGroupAction, error) {
	input := &autoscaling.DescribeScheduledActionsInput{
		AutoScalingGroupName: aws.String(asgName),
		ScheduledActionNames: aws.StringSlice([]string{actionName}),
	}
	var output []*autoscaling.ScheduledUpdateGroupAction

	err := conn.DescribeScheduledActionsPages(input, func(page *autoscaling.DescribeScheduledActionsOutput, lastPage bool) bool {
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
