// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appautoscaling

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appautoscaling_scheduled_action", namae="Scheduled Action")
func resourceScheduledAction() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceScheduledActionPut,
		ReadWithoutTimeout:   resourceScheduledActionRead,
		UpdateWithoutTimeout: resourceScheduledActionPut,
		DeleteWithoutTimeout: resourceScheduledActionDelete,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"end_time": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.IsRFC3339Time,
				DiffSuppressFunc: suppressEquivalentTime,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scalable_dimension": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scalable_target_action": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_capacity": {
							Type:         nullable.TypeNullableInt,
							Optional:     true,
							ValidateFunc: nullable.ValidateTypeStringNullableIntAtLeast(0),
							AtLeastOneOf: []string{
								"scalable_target_action.0.max_capacity",
								"scalable_target_action.0.min_capacity",
							},
						},
						"min_capacity": {
							Type:         nullable.TypeNullableInt,
							Optional:     true,
							ValidateFunc: nullable.ValidateTypeStringNullableIntAtLeast(0),
							AtLeastOneOf: []string{
								"scalable_target_action.0.max_capacity",
								"scalable_target_action.0.min_capacity",
							},
						},
					},
				},
			},
			"schedule": {
				Type:     schema.TypeString,
				Required: true,
			},
			"service_namespace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// The AWS API normalizes start_time and end_time to UTC. Uses
			// suppressEquivalentTime to allow any timezone to be used.
			"start_time": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.IsRFC3339Time,
				DiffSuppressFunc: suppressEquivalentTime,
			},
			"timezone": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "UTC",
			},
		},
	}
}

func resourceScheduledActionPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppAutoScalingConn(ctx)

	name, serviceNamespace, resourceID := d.Get(names.AttrName).(string), d.Get("service_namespace").(string), d.Get("resource_id").(string)
	id := strings.Join([]string{name, serviceNamespace, resourceID}, "-")
	input := &applicationautoscaling.PutScheduledActionInput{
		ResourceId:          aws.String(resourceID),
		ScalableDimension:   aws.String(d.Get("scalable_dimension").(string)),
		ScheduledActionName: aws.String(name),
		ServiceNamespace:    aws.String(serviceNamespace),
	}

	if d.IsNewResource() {
		if v, ok := d.GetOk("end_time"); ok {
			t, _ := time.Parse(time.RFC3339, v.(string))
			input.EndTime = aws.Time(t)
		}
		input.ScalableTargetAction = expandScalableTargetAction(d.Get("scalable_target_action").([]interface{}))
		input.Schedule = aws.String(d.Get("schedule").(string))
		if v, ok := d.GetOk("start_time"); ok {
			t, _ := time.Parse(time.RFC3339, v.(string))
			input.StartTime = aws.Time(t)
		}
		input.Timezone = aws.String(d.Get("timezone").(string))
	} else {
		if v, ok := d.GetOk("end_time"); ok {
			t, _ := time.Parse(time.RFC3339, v.(string))
			input.EndTime = aws.Time(t)
		}
		if d.HasChange("scalable_target_action") {
			input.ScalableTargetAction = expandScalableTargetAction(d.Get("scalable_target_action").([]interface{}))
		}
		if d.HasChange("schedule") {
			input.Schedule = aws.String(d.Get("schedule").(string))
		}
		if v, ok := d.GetOk("start_time"); ok {
			t, _ := time.Parse(time.RFC3339, v.(string))
			input.StartTime = aws.Time(t)
		}
		if d.HasChange("timezone") {
			input.Timezone = aws.String(d.Get("timezone").(string))
		}
	}

	const (
		timeout = 5 * time.Minute
	)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout, func() (interface{}, error) {
		return conn.PutScheduledActionWithContext(ctx, input)
	}, applicationautoscaling.ErrCodeObjectNotFoundException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting Application Auto Scaling Scheduled Action (%s): %s", id, err)
	}

	if d.IsNewResource() {
		d.SetId(id)
	}

	return append(diags, resourceScheduledActionRead(ctx, d, meta)...)
}

func resourceScheduledActionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppAutoScalingConn(ctx)

	scheduledAction, err := findScheduledActionByFourPartKey(ctx, conn, d.Get(names.AttrName).(string), d.Get("service_namespace").(string), d.Get("resource_id").(string), d.Get("scalable_dimension").(string))

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] Application Auto Scaling Scheduled Action (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Application Auto Scaling Scheduled Action (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, scheduledAction.ScheduledActionARN)
	if scheduledAction.EndTime != nil {
		d.Set("end_time", scheduledAction.EndTime.Format(time.RFC3339))
	}
	if err := d.Set("scalable_target_action", flattenScalableTargetAction(scheduledAction.ScalableTargetAction)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting scalable_target_action: %s", err)
	}
	d.Set("schedule", scheduledAction.Schedule)
	if scheduledAction.StartTime != nil {
		d.Set("start_time", scheduledAction.StartTime.Format(time.RFC3339))
	}
	d.Set("timezone", scheduledAction.Timezone)

	return diags
}

func resourceScheduledActionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppAutoScalingConn(ctx)

	log.Printf("[DEBUG] Deleting Application Auto Scaling Scheduled Action: %s", d.Id())
	_, err := conn.DeleteScheduledActionWithContext(ctx, &applicationautoscaling.DeleteScheduledActionInput{
		ResourceId:          aws.String(d.Get("resource_id").(string)),
		ScalableDimension:   aws.String(d.Get("scalable_dimension").(string)),
		ScheduledActionName: aws.String(d.Get(names.AttrName).(string)),
		ServiceNamespace:    aws.String(d.Get("service_namespace").(string)),
	})

	if tfawserr.ErrCodeEquals(err, applicationautoscaling.ErrCodeObjectNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Application Auto Scaling Scheduled Action (%s): %s", d.Id(), err)
	}

	return diags
}

func findScheduledActionByFourPartKey(ctx context.Context, conn *applicationautoscaling.ApplicationAutoScaling, name, serviceNamespace, resourceID, scalableDimension string) (*applicationautoscaling.ScheduledAction, error) {
	input := &applicationautoscaling.DescribeScheduledActionsInput{
		ResourceId:           aws.String(resourceID),
		ScalableDimension:    aws.String(scalableDimension),
		ScheduledActionNames: aws.StringSlice([]string{name}),
		ServiceNamespace:     aws.String(serviceNamespace),
	}

	return findScheduledAction(ctx, conn, input, func(v *applicationautoscaling.ScheduledAction) bool {
		return aws.StringValue(v.ScheduledActionName) == name && aws.StringValue(v.ScalableDimension) == scalableDimension
	})
}

func findScheduledAction(ctx context.Context, conn *applicationautoscaling.ApplicationAutoScaling, input *applicationautoscaling.DescribeScheduledActionsInput, filter tfslices.Predicate[*applicationautoscaling.ScheduledAction]) (*applicationautoscaling.ScheduledAction, error) {
	output, err := findScheduledActions(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findScheduledActions(ctx context.Context, conn *applicationautoscaling.ApplicationAutoScaling, input *applicationautoscaling.DescribeScheduledActionsInput, filter tfslices.Predicate[*applicationautoscaling.ScheduledAction]) ([]*applicationautoscaling.ScheduledAction, error) {
	var output []*applicationautoscaling.ScheduledAction

	err := conn.DescribeScheduledActionsPagesWithContext(ctx, input, func(page *applicationautoscaling.DescribeScheduledActionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ScheduledActions {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func expandScalableTargetAction(l []interface{}) *applicationautoscaling.ScalableTargetAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	result := &applicationautoscaling.ScalableTargetAction{}

	if v, ok := m["max_capacity"]; ok {
		if v, null, _ := nullable.Int(v.(string)).ValueInt64(); !null {
			result.MaxCapacity = aws.Int64(v)
		}
	}
	if v, ok := m["min_capacity"]; ok {
		if v, null, _ := nullable.Int(v.(string)).ValueInt64(); !null {
			result.MinCapacity = aws.Int64(v)
		}
	}

	return result
}

func flattenScalableTargetAction(cfg *applicationautoscaling.ScalableTargetAction) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if cfg.MaxCapacity != nil {
		m["max_capacity"] = strconv.FormatInt(aws.Int64Value(cfg.MaxCapacity), 10)
	}
	if cfg.MinCapacity != nil {
		m["min_capacity"] = strconv.FormatInt(aws.Int64Value(cfg.MinCapacity), 10)
	}

	return []interface{}{m}
}
