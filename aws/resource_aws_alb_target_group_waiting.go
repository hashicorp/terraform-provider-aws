package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

// waitForTargetGroupCapacity gathers the current numbers of healthy instances
// in the target group and yields these numbers to a
// capacitySatifiedFunction. Loops for up to wait_for_capacity_timeout until
// the capacitySatisfiedFunc returns true.
//
// See "Waiting for Capacity" in docs for more discussion of the feature.
func waitForListenerTargetGroupCapacity(
	d *schema.ResourceData, // this is the resource data corresponding to the listener
	meta interface{},
	satisfiedFunc capacitySatisfiedFunc) error {

	// wait, err := time.ParseDuration(d.Get("wait_for_capacity_timeout").(string))
	// if err != nil {
	// 	return err
	// }
	defaultActionMap := d.Get("default_action").([]interface{})[0].(map[string]interface{})
	targetGroupArn := defaultActionMap["target_group_arn"].(string)

	// TODO(mdeboer): Start with the default for now -->
	// ideally, we'll try to pull in the ResourceData from the associated aws_autoscaling_group
	// object, assuming we can make the association (although it's a long chain)
	wait := time.Minute * 10
	targetCount := 2 // => need to read this from the ASG directly

	// if wait == 0 {
	// 	log.Printf("[DEBUG] Capacity timeout set to 0, skipping capacity waiting.")
	// 	return nil
	// }

	log.Printf("[DEBUG] Waiting on %s for target group capacity...", d.Id())

	err := resource.Retry(wait, func() *resource.RetryError {
		// g, err := getAwsAutoscalingGroup(d.Id(), meta.(*AWSClient).autoscalingconn)
		// if err != nil {
		// 	return resource.NonRetryableError(err)
		// }
		// if g == nil {
		// 	log.Printf("[INFO] Autoscaling Group %q not found", d.Id())
		// 	d.SetId("")
		// 	return nil
		// }

		states, err := _getTargetGroupInstanceStates(&targetGroupArn, meta)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		healthyCount := 0
		for _, state := range states {
			if strings.EqualFold(state, "healthy") {
				healthyCount++
			}
		}

		satisfied, reason := satisfiedFunc(d, healthyCount, targetCount)

		log.Printf("[DEBUG] %q Capacity: %d target group, %d healthy, satisfied: %t, reason: %q",
			d.Id(), healthyCount, targetCount, satisfied, reason)

		if satisfied {
			return nil
		}

		return resource.RetryableError(
			fmt.Errorf("%q: Waiting up to %s: %s", d.Id(), wait, reason))
	})

	if err == nil {
		return nil
	}

	recentStatus := ""

	conn := meta.(*AWSClient).autoscalingconn
	resp, aErr := conn.DescribeScalingActivities(&autoscaling.DescribeScalingActivitiesInput{
		AutoScalingGroupName: aws.String(d.Id()),
		MaxRecords:           aws.Int64(1),
	})
	if aErr == nil {
		if len(resp.Activities) > 0 {
			recentStatus = fmt.Sprintf("%s", resp.Activities[0])
		} else {
			recentStatus = "(0 activities found)"
		}
	} else {
		recentStatus = fmt.Sprintf("(Failed to describe scaling activities: %s)", aErr)
	}

	msg := fmt.Sprintf("{{err}}. Most recent activity: %s", recentStatus)
	return errwrap.Wrapf(msg, err)
}

// type capacitySatisfiedFunc func(*schema.ResourceData, int, int) (bool, string)

// // capacitySatisfiedCreate treats all targets as minimums
// func capacitySatisfiedCreate(d *schema.ResourceData, haveASG, haveELB int) (bool, string) {
// 	minASG := d.Get("min_size").(int)
// 	if wantASG := d.Get("desired_capacity").(int); wantASG > 0 {
// 		minASG = wantASG
// 	}
// 	if haveASG < minASG {
// 		return false, fmt.Sprintf(
// 			"Need at least %d healthy instances in ASG, have %d", minASG, haveASG)
// 	}
// 	minELB := d.Get("min_elb_capacity").(int)
// 	if wantELB := d.Get("wait_for_elb_capacity").(int); wantELB > 0 {
// 		minELB = wantELB
// 	}
// 	if haveELB < minELB {
// 		return false, fmt.Sprintf(
// 			"Need at least %d healthy instances in ELB, have %d", minELB, haveELB)
// 	}
// 	return true, ""
// }

// capacitySatisfiedUpdate only cares about specific targets
// func capacitySatisfiedUpdate(d *schema.ResourceData, haveASG, haveELB int) (bool, string) {
// 	if wantASG := d.Get("desired_capacity").(int); wantASG > 0 {
// 		if haveASG != wantASG {
// 			return false, fmt.Sprintf(
// 				"Need exactly %d healthy instances in ASG, have %d", wantASG, haveASG)
// 		}
// 	}
// 	if wantELB := d.Get("wait_for_elb_capacity").(int); wantELB > 0 {
// 		if haveELB != wantELB {
// 			return false, fmt.Sprintf(
// 				"Need exactly %d healthy instances in ELB, have %d", wantELB, haveELB)
// 		}
// 	}
// 	return true, ""
// }

// getTargetGroupInstanceStates returns a mapping of the instance states of
// all the ALB target groups attached to the provided ASG.
//
// Note that this is the instance state function for Application Load
// Balancing (aka ELBv2).
//
// Nested like: targetGroupARN -> instanceId -> instanceState
func _getTargetGroupInstanceStates(targetGroupARN *string, meta interface{}) (map[string]string, error) {
	targetInstanceStates := make(map[string]string)

	elbv2conn := meta.(*AWSClient).elbv2conn
	opts := &elbv2.DescribeTargetHealthInput{TargetGroupArn: targetGroupARN}
	r, err := elbv2conn.DescribeTargetHealth(opts)
	if err != nil {
		return nil, err
	}
	for _, desc := range r.TargetHealthDescriptions {
		if desc.Target == nil || desc.Target.Id == nil || desc.TargetHealth == nil || desc.TargetHealth.State == nil {
			continue
		}
		targetInstanceStates[*desc.Target.Id] = *desc.TargetHealth.State
	}

	return targetInstanceStates, nil
}
