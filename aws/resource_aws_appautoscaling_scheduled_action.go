package aws

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

const awsAppautoscalingScheduleTimeLayout = "2006-01-02T15:04:05Z"

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
							// Use TypeString to allow an "unspecified" value,
							// since TypeInt only has allows numbers with 0 as default.
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     false,
							ValidateFunc: validateTypeStringNullableInteger,
						},
						"min_capacity": {
							// Use TypeString to allow an "unspecified" value,
							// since TypeInt only has allows numbers with 0 as default.
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     false,
							ValidateFunc: validateTypeStringNullableInteger,
						},
					},
				},
			},
			"schedule": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"start_time": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"end_time": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
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
		t, err := time.Parse(awsAppautoscalingScheduleTimeLayout, v.(string))
		if err != nil {
			return fmt.Errorf("Error Parsing Appautoscaling Scheduled Action Start Time: %s", err.Error())
		}
		input.StartTime = aws.Time(t)
	}
	if v, ok := d.GetOk("end_time"); ok {
		t, err := time.Parse(awsAppautoscalingScheduleTimeLayout, v.(string))
		if err != nil {
			return fmt.Errorf("Error Parsing Appautoscaling Scheduled Action End Time: %s", err.Error())
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

	if err != nil {
		return err
	}

	d.SetId(d.Get("name").(string) + "-" + d.Get("service_namespace").(string) + "-" + d.Get("resource_id").(string))
	return resourceAwsAppautoscalingScheduledActionRead(d, meta)
}

func resourceAwsAppautoscalingScheduledActionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appautoscalingconn

	saName := d.Get("name").(string)
	input := &applicationautoscaling.DescribeScheduledActionsInput{
		ScheduledActionNames: []*string{aws.String(saName)},
		ServiceNamespace:     aws.String(d.Get("service_namespace").(string)),
		ResourceId:           aws.String(d.Get("resource_id").(string)),
	}
	resp, err := conn.DescribeScheduledActions(input)
	if err != nil {
		return err
	}

	if len(resp.ScheduledActions) < 1 {
		log.Printf("[WARN] Application Autoscaling Scheduled Action (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if len(resp.ScheduledActions) != 1 {
		return fmt.Errorf("Expected 1 scheduled action under %s, found %d", saName, len(resp.ScheduledActions))
	}
	sa := resp.ScheduledActions[0]
	if *sa.ScheduledActionName != saName {
		return fmt.Errorf("Scheduled Action (%s) not found", saName)
	}

	if err := d.Set("scalable_target_action", flattenScalableTargetActionConfiguration(sa.ScalableTargetAction)); err != nil {
		return fmt.Errorf("error setting scalable_target_action: %s", err)
	}
	d.Set("schedule", sa.Schedule)
	d.Set("start_time", sa.StartTime)
	d.Set("end_time", sa.EndTime)
	d.Set("arn", sa.ScheduledActionARN)
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
		m["max_capacity"] = strconv.FormatInt(aws.Int64Value(cfg.MaxCapacity), 10)
	}
	if cfg.MinCapacity != nil {
		m["min_capacity"] = strconv.FormatInt(aws.Int64Value(cfg.MinCapacity), 10)
	}

	return []interface{}{m}
}
