package aws

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/experimental/nullable"
)

func resourceAwsAppautoscalingScheduledAction() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppautoscalingScheduledActionPut,
		Read:   resourceAwsAppautoscalingScheduledActionRead,
		Update: resourceAwsAppautoscalingScheduledActionPut,
		Delete: resourceAwsAppautoscalingScheduledActionDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_namespace": {
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
				Optional: true,
				ForceNew: true,
			},
			"scalable_target_action": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_capacity": {
							Type:         nullable.TypeNullableInt,
							Optional:     true,
							ForceNew:     false,
							ValidateFunc: nullable.ValidateTypeStringNullableIntAtLeast(0),
						},
						"min_capacity": {
							Type:         nullable.TypeNullableInt,
							Optional:     true,
							ForceNew:     false,
							ValidateFunc: nullable.ValidateTypeStringNullableIntAtLeast(0),
						},
					},
				},
			},
			"schedule": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			// The AWS API normalizes start_time and end_time to UTC. Uses
			// suppressEquivalentTime to allow any timezone to be used.
			"start_time": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         false,
				ValidateFunc:     validation.IsRFC3339Time,
				DiffSuppressFunc: suppressEquivalentTime,
			},
			"end_time": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         false,
				ValidateFunc:     validation.IsRFC3339Time,
				DiffSuppressFunc: suppressEquivalentTime,
			},
			"timezone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  "UTC",
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsAppautoscalingScheduledActionPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appautoscalingconn

	input := &applicationautoscaling.PutScheduledActionInput{
		ScheduledActionName: aws.String(d.Get("name").(string)),
		ServiceNamespace:    aws.String(d.Get("service_namespace").(string)),
		ResourceId:          aws.String(d.Get("resource_id").(string)),
		Timezone:            aws.String(d.Get("timezone").(string)),
	}
	if v, ok := d.GetOk("scalable_dimension"); ok {
		input.ScalableDimension = aws.String(v.(string))
	}
	if v, ok := d.GetOk("schedule"); ok {
		input.Schedule = aws.String(v.(string))
	}
	if v, ok := d.GetOk("scalable_target_action"); ok {
		sta := &applicationautoscaling.ScalableTargetAction{}
		raw := v.([]interface{})[0].(map[string]interface{})
		if max, ok := raw["max_capacity"]; ok && max.(string) != "" {
			maxInt, err := strconv.ParseInt(max.(string), 10, 64)
			if err != nil {
				return fmt.Errorf("error converting max_capacity %q from string to integer: %s", v.(string), err)
			}
			sta.MaxCapacity = aws.Int64(maxInt)
		}
		if min, ok := raw["min_capacity"]; ok && min.(string) != "" {
			minInt, err := strconv.ParseInt(min.(string), 10, 64)
			if err != nil {
				return fmt.Errorf("error converting min_capacity %q from string to integer: %s", v.(string), err)
			}
			sta.MinCapacity = aws.Int64(minInt)
		}
		input.ScalableTargetAction = sta
	}
	if v, ok := d.GetOk("start_time"); ok {
		t, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return fmt.Errorf("Error Parsing Appautoscaling Scheduled Action Start Time: %w", err)
		}
		input.StartTime = aws.Time(t)
	}
	if v, ok := d.GetOk("end_time"); ok {
		t, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return fmt.Errorf("Error Parsing Appautoscaling Scheduled Action End Time: %w", err)
		}
		input.EndTime = aws.Time(t)
	}

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.PutScheduledAction(input)
		if err != nil {
			if isAWSErr(err, applicationautoscaling.ErrCodeObjectNotFoundException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.PutScheduledAction(input)
	}

	if err != nil {
		return fmt.Errorf("error putting scheduled action: %w", err)
	}

	d.SetId(d.Get("name").(string) + "-" + d.Get("service_namespace").(string) + "-" + d.Get("resource_id").(string))
	return resourceAwsAppautoscalingScheduledActionRead(d, meta)
}

func resourceAwsAppautoscalingScheduledActionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appautoscalingconn

	saName := d.Get("name").(string)
	input := &applicationautoscaling.DescribeScheduledActionsInput{
		ResourceId:           aws.String(d.Get("resource_id").(string)),
		ScheduledActionNames: []*string{aws.String(saName)},
		ServiceNamespace:     aws.String(d.Get("service_namespace").(string)),
	}
	resp, err := conn.DescribeScheduledActions(input)
	if err != nil {
		return fmt.Errorf("error describing Application Auto Scaling Scheduled Action (%s): %w", d.Id(), err)
	}
	if resp == nil {
		return fmt.Errorf("error describing Application Auto Scaling Scheduled Action (%s): empty response", d.Id())
	}

	var scheduledAction *applicationautoscaling.ScheduledAction

	for _, sa := range resp.ScheduledActions {
		if sa == nil {
			continue
		}

		if aws.StringValue(sa.ScheduledActionName) == saName {
			scheduledAction = sa
			break
		}
	}

	if scheduledAction == nil {
		log.Printf("[WARN] Application Autoscaling Scheduled Action (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("scalable_target_action", flattenScalableTargetActionConfiguration(scheduledAction.ScalableTargetAction)); err != nil {
		return fmt.Errorf("error setting scalable_target_action: %w", err)
	}

	d.Set("schedule", scheduledAction.Schedule)
	if scheduledAction.StartTime != nil {
		d.Set("start_time", scheduledAction.StartTime.Format(time.RFC3339))
	}
	if scheduledAction.EndTime != nil {
		d.Set("end_time", scheduledAction.EndTime.Format(time.RFC3339))
	}
	d.Set("timezone", scheduledAction.Timezone)
	d.Set("arn", scheduledAction.ScheduledActionARN)

	return nil
}

func resourceAwsAppautoscalingScheduledActionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appautoscalingconn

	input := &applicationautoscaling.DeleteScheduledActionInput{
		ScheduledActionName: aws.String(d.Get("name").(string)),
		ServiceNamespace:    aws.String(d.Get("service_namespace").(string)),
		ResourceId:          aws.String(d.Get("resource_id").(string)),
	}
	if v, ok := d.GetOk("scalable_dimension"); ok {
		input.ScalableDimension = aws.String(v.(string))
	}
	_, err := conn.DeleteScheduledAction(input)
	if err != nil {
		if isAWSErr(err, applicationautoscaling.ErrCodeObjectNotFoundException, "") {
			log.Printf("[WARN] Application Autoscaling Scheduled Action (%s) already gone, removing from state", d.Id())
			return nil
		}
		return err
	}

	return nil
}

func flattenScalableTargetActionConfiguration(cfg *applicationautoscaling.ScalableTargetAction) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if cfg.MaxCapacity != nil {
		// TODO: add parser for nullable
		m["max_capacity"] = strconv.FormatInt(aws.Int64Value(cfg.MaxCapacity), 10)
	}
	if cfg.MinCapacity != nil {
		m["min_capacity"] = strconv.FormatInt(aws.Int64Value(cfg.MinCapacity), 10)
	}

	return []interface{}{m}
}
