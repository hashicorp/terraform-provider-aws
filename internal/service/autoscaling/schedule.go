// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const ScheduleTimeLayout = "2006-01-02T15:04:05Z"

// @SDKResource("aws_autoscaling_schedule", name="Scheduled Action")
func resourceSchedule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSchedulePut,
		ReadWithoutTimeout:   resourceScheduleRead,
		UpdateWithoutTimeout: resourceSchedulePut,
		DeleteWithoutTimeout: resourceScheduleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceScheduleImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
			names.AttrStartTime: {
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
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

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

	if v, ok := d.GetOk(names.AttrStartTime); ok {
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
	minSize := int32(d.Get("min_size").(int))
	maxSize := int32(d.Get("max_size").(int))
	desiredCapacity := int32(d.Get("desired_capacity").(int))
	if minSize != -1 {
		input.MinSize = aws.Int32(minSize)
	}
	if maxSize != -1 {
		input.MaxSize = aws.Int32(maxSize)
	}
	if desiredCapacity != -1 {
		input.DesiredCapacity = aws.Int32(desiredCapacity)
	}

	_, err := conn.PutScheduledUpdateGroupAction(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting Auto Scaling Scheduled Action (%s): %s", name, err)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	return append(diags, resourceScheduleRead(ctx, d, meta)...)
}

func resourceScheduleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	sa, err := findScheduleByTwoPartKey(ctx, conn, d.Get("autoscaling_group_name").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Auto Scaling Scheduled Action %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Scheduled Action (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, sa.ScheduledActionARN)
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
		d.Set(names.AttrStartTime, sa.StartTime.Format(ScheduleTimeLayout))
	}
	d.Set("time_zone", sa.TimeZone)

	return diags
}

func resourceScheduleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	log.Printf("[INFO] Deleting Auto Scaling Scheduled Action: %s", d.Id())
	_, err := conn.DeleteScheduledAction(ctx, &autoscaling.DeleteScheduledActionInput{
		AutoScalingGroupName: aws.String(d.Get("autoscaling_group_name").(string)),
		ScheduledActionName:  aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, errCodeValidationError, "not found") {
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

func findSchedule(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribeScheduledActionsInput) (*awstypes.ScheduledUpdateGroupAction, error) {
	output, err := findSchedules(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSchedules(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribeScheduledActionsInput) ([]awstypes.ScheduledUpdateGroupAction, error) {
	var output []awstypes.ScheduledUpdateGroupAction

	pages := autoscaling.NewDescribeScheduledActionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrMessageContains(err, errCodeValidationError, "not found") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ScheduledUpdateGroupActions...)
	}

	return output, nil
}

func findScheduleByTwoPartKey(ctx context.Context, conn *autoscaling.Client, asgName, actionName string) (*awstypes.ScheduledUpdateGroupAction, error) {
	input := &autoscaling.DescribeScheduledActionsInput{
		AutoScalingGroupName: aws.String(asgName),
		ScheduledActionNames: []string{actionName},
	}

	return findSchedule(ctx, conn, input)
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
