// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appautoscaling

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
			names.AttrResourceID: {
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
						names.AttrMaxCapacity: {
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
			names.AttrSchedule: {
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
			names.AttrStartTime: {
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
	conn := meta.(*conns.AWSClient).AppAutoScalingClient(ctx)

	name, serviceNamespace, resourceID := d.Get(names.AttrName).(string), d.Get("service_namespace").(string), d.Get(names.AttrResourceID).(string)
	id := strings.Join([]string{name, serviceNamespace, resourceID}, "-")
	input := &applicationautoscaling.PutScheduledActionInput{
		ResourceId:          aws.String(resourceID),
		ScalableDimension:   awstypes.ScalableDimension(d.Get("scalable_dimension").(string)),
		ScheduledActionName: aws.String(name),
		ServiceNamespace:    awstypes.ServiceNamespace(serviceNamespace),
	}

	if d.IsNewResource() {
		if v, ok := d.GetOk("end_time"); ok {
			t, _ := time.Parse(time.RFC3339, v.(string))
			input.EndTime = aws.Time(t)
		}
		input.ScalableTargetAction = expandScalableTargetAction(d.Get("scalable_target_action").([]interface{}))
		input.Schedule = aws.String(d.Get(names.AttrSchedule).(string))
		if v, ok := d.GetOk(names.AttrStartTime); ok {
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
		if d.HasChange(names.AttrSchedule) {
			input.Schedule = aws.String(d.Get(names.AttrSchedule).(string))
		}
		if v, ok := d.GetOk(names.AttrStartTime); ok {
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
	_, err := tfresource.RetryWhenIsA[*awstypes.ObjectNotFoundException](ctx, timeout, func() (interface{}, error) {
		return conn.PutScheduledAction(ctx, input)
	})

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
	conn := meta.(*conns.AWSClient).AppAutoScalingClient(ctx)

	scheduledAction, err := findScheduledActionByFourPartKey(ctx, conn, d.Get(names.AttrName).(string), d.Get("service_namespace").(string), d.Get(names.AttrResourceID).(string), d.Get("scalable_dimension").(string))

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
	d.Set(names.AttrSchedule, scheduledAction.Schedule)
	if scheduledAction.StartTime != nil {
		d.Set(names.AttrStartTime, scheduledAction.StartTime.Format(time.RFC3339))
	}
	d.Set("timezone", scheduledAction.Timezone)

	return diags
}

func resourceScheduledActionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppAutoScalingClient(ctx)

	log.Printf("[DEBUG] Deleting Application Auto Scaling Scheduled Action: %s", d.Id())
	_, err := conn.DeleteScheduledAction(ctx, &applicationautoscaling.DeleteScheduledActionInput{
		ResourceId:          aws.String(d.Get(names.AttrResourceID).(string)),
		ScalableDimension:   awstypes.ScalableDimension(d.Get("scalable_dimension").(string)),
		ScheduledActionName: aws.String(d.Get(names.AttrName).(string)),
		ServiceNamespace:    awstypes.ServiceNamespace(d.Get("service_namespace").(string)),
	})

	if errs.IsA[*awstypes.ObjectNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Application Auto Scaling Scheduled Action (%s): %s", d.Id(), err)
	}

	return diags
}

func findScheduledActionByFourPartKey(ctx context.Context, conn *applicationautoscaling.Client, name, serviceNamespace, resourceID, scalableDimension string) (*awstypes.ScheduledAction, error) {
	input := &applicationautoscaling.DescribeScheduledActionsInput{
		ResourceId:           aws.String(resourceID),
		ScalableDimension:    awstypes.ScalableDimension(scalableDimension),
		ScheduledActionNames: []string{name},
		ServiceNamespace:     awstypes.ServiceNamespace(serviceNamespace),
	}

	return findScheduledAction(ctx, conn, input, func(v awstypes.ScheduledAction) bool {
		return aws.ToString(v.ScheduledActionName) == name && string(v.ScalableDimension) == scalableDimension
	})
}

func findScheduledAction(ctx context.Context, conn *applicationautoscaling.Client, input *applicationautoscaling.DescribeScheduledActionsInput, filter tfslices.Predicate[awstypes.ScheduledAction]) (*awstypes.ScheduledAction, error) {
	output, err := findScheduledActions(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findScheduledActions(ctx context.Context, conn *applicationautoscaling.Client, input *applicationautoscaling.DescribeScheduledActionsInput, filter tfslices.Predicate[awstypes.ScheduledAction]) ([]awstypes.ScheduledAction, error) {
	var output []awstypes.ScheduledAction

	pages := applicationautoscaling.NewDescribeScheduledActionsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ScheduledActions {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func expandScalableTargetAction(l []interface{}) *awstypes.ScalableTargetAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	result := &awstypes.ScalableTargetAction{}

	if v, ok := m[names.AttrMaxCapacity]; ok {
		if v, null, _ := nullable.Int(v.(string)).ValueInt32(); !null {
			result.MaxCapacity = aws.Int32(v)
		}
	}
	if v, ok := m["min_capacity"]; ok {
		if v, null, _ := nullable.Int(v.(string)).ValueInt32(); !null {
			result.MinCapacity = aws.Int32(v)
		}
	}

	return result
}

func flattenScalableTargetAction(cfg *awstypes.ScalableTargetAction) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if cfg.MaxCapacity != nil {
		m[names.AttrMaxCapacity] = strconv.FormatInt(int64(aws.ToInt32(cfg.MaxCapacity)), 10)
	}
	if cfg.MinCapacity != nil {
		m["min_capacity"] = strconv.FormatInt(int64(aws.ToInt32(cfg.MinCapacity)), 10)
	}

	return []interface{}{m}
}
