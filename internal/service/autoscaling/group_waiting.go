package autoscaling

import ( // nosemgrep: aws-sdk-go-multiple-service-imports
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// waitForASGCapacityTimeout gathers the current numbers of healthy instances
// in the ASG and its attached ELBs and yields these numbers to a
// capacitySatifiedFunction. Loops for up to wait_for_capacity_timeout until
// the capacitySatisfiedFunc returns true.
//
// See "Waiting for Capacity" in docs for more discussion of the feature.
func waitForASGCapacity(
	d *schema.ResourceData,
	meta interface{},
	satisfiedFunc capacitySatisfiedFunc) error {
	wait, err := time.ParseDuration(d.Get("wait_for_capacity_timeout").(string))
	if err != nil {
		return err
	}

	if wait == 0 {
		log.Printf("[DEBUG] Capacity timeout set to 0, skipping capacity waiting.")
		return nil
	}

	log.Printf("[DEBUG] Waiting on %s for capacity...", d.Id())

	err = resource.Retry(wait, func() *resource.RetryError {
		g, err := getGroup(d.Id(), meta.(*conns.AWSClient).AutoScalingConn)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		if g == nil && !d.IsNewResource() {
			log.Printf("[WARN] Auto Scaling Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		satisfied, reason := isELBCapacitySatisfied(d, meta, g, satisfiedFunc)
		if satisfied {
			return nil
		}

		return resource.RetryableError(fmt.Errorf("%q: Waiting up to %s: %s", d.Id(), wait, reason))
	})
	if tfresource.TimedOut(err) {
		g, err := getGroup(d.Id(), meta.(*conns.AWSClient).AutoScalingConn)

		if err != nil {
			return fmt.Errorf("Error getting Auto Scaling Group info: %s", err)
		}

		if g == nil && !d.IsNewResource() {
			log.Printf("[WARN] Auto Scaling Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		satisfied, _ := isELBCapacitySatisfied(d, meta, g, satisfiedFunc)
		if satisfied {
			return nil
		}
	}

	if err == nil {
		return nil
	}

	recentStatus := ""

	conn := meta.(*conns.AWSClient).AutoScalingConn
	resp, aErr := conn.DescribeScalingActivities(&autoscaling.DescribeScalingActivitiesInput{
		AutoScalingGroupName: aws.String(d.Id()),
		MaxRecords:           aws.Int64(1),
	})
	if aErr == nil {
		if len(resp.Activities) > 0 {
			recentStatus = (resp.Activities[0].String())
		} else {
			recentStatus = "(0 activities found)"
		}
	} else {
		recentStatus = fmt.Sprintf("(Failed to describe scaling activities: %s)", aErr)
	}

	return fmt.Errorf("%s. Most recent activity: %s", err, recentStatus)
}

func isELBCapacitySatisfied(d *schema.ResourceData, meta interface{}, g *autoscaling.Group, satisfiedFunc capacitySatisfiedFunc) (bool, string) {
	elbis, err := getELBInstanceStates(g, meta)
	if err != nil {
		return false, fmt.Sprintf("Error getting ELB instance states: %s", err)
	}
	albis, err := getTargetGroupInstanceStates(g, meta)
	if err != nil {
		return false, fmt.Sprintf("Error getting target group instance states: %s", err)
	}

	haveASG := 0
	haveELB := 0

	for _, i := range g.Instances {
		if i.HealthStatus == nil || i.InstanceId == nil || i.LifecycleState == nil {
			continue
		}

		if !strings.EqualFold(*i.HealthStatus, "Healthy") {
			continue
		}

		if !strings.EqualFold(*i.LifecycleState, "InService") {
			continue
		}

		capacity := 1
		if i.WeightedCapacity != nil {
			capacity, err = strconv.Atoi(*i.WeightedCapacity)
			if err != nil {
				capacity = 1
			}
		}

		haveASG += capacity

		inAllLbs := true
		for _, states := range elbis {
			state, ok := states[*i.InstanceId]
			if !ok || !strings.EqualFold(state, "InService") {
				inAllLbs = false
			}
		}
		for _, states := range albis {
			state, ok := states[*i.InstanceId]
			if !ok || !strings.EqualFold(state, "healthy") {
				inAllLbs = false
			}
		}
		if inAllLbs {
			haveELB += capacity
		}
	}

	satisfied, reason := satisfiedFunc(d, haveASG, haveELB)

	log.Printf("[DEBUG] %q Capacity: %d ASG, %d ELB/ALB, satisfied: %t, reason: %q",
		d.Id(), haveASG, haveELB, satisfied, reason)

	return satisfied, reason
}

type capacitySatisfiedFunc func(*schema.ResourceData, int, int) (bool, string)

// CapacitySatisfiedCreate treats all targets as minimums
func CapacitySatisfiedCreate(d *schema.ResourceData, haveASG, haveELB int) (bool, string) {
	minASG := d.Get("min_size").(int)
	if wantASG := d.Get("desired_capacity").(int); wantASG > 0 {
		minASG = wantASG
	}
	if haveASG < minASG {
		return false, fmt.Sprintf(
			"Need at least %d healthy instances in ASG, have %d", minASG, haveASG)
	}
	minELB := d.Get("min_elb_capacity").(int)
	if wantELB := d.Get("wait_for_elb_capacity").(int); wantELB > 0 {
		minELB = wantELB
	}
	if haveELB < minELB {
		return false, fmt.Sprintf(
			"Need at least %d healthy instances in ELB, have %d", minELB, haveELB)
	}
	return true, ""
}

// CapacitySatisfiedUpdate only cares about specific targets
func CapacitySatisfiedUpdate(d *schema.ResourceData, haveASG, haveELB int) (bool, string) {
	minASG := d.Get("min_size").(int)
	if wantASG := d.Get("desired_capacity").(int); wantASG > minASG {
		minASG = wantASG
	}
	if haveASG != minASG {
		return false, fmt.Sprintf(
			"Need exactly %d healthy instances in ASG, have %d", minASG, haveASG)
	}
	if wantELB := d.Get("wait_for_elb_capacity").(int); wantELB > 0 {
		if haveELB != wantELB {
			return false, fmt.Sprintf(
				"Need exactly %d healthy instances in ELB, have %d", wantELB, haveELB)
		}
	}
	return true, ""
}

// TODO: make this a finder
// TODO: this should return a NotFoundError if not found
func getGroup(asgName string, conn *autoscaling.AutoScaling) (*autoscaling.Group, error) {
	describeOpts := autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(asgName)},
	}

	log.Printf("[DEBUG] Auto Scaling Group describe configuration: %#v", describeOpts)
	describeGroups, err := conn.DescribeAutoScalingGroups(&describeOpts)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, "InvalidGroup.NotFound") {
			return nil, nil
		}

		return nil, fmt.Errorf("Error retrieving Auto Scaling Groups: %s", err)
	}

	// Search for the Auto Scaling Group
	for idx, asc := range describeGroups.AutoScalingGroups {
		if asc == nil {
			continue
		}

		if aws.StringValue(asc.AutoScalingGroupName) == asgName {
			return describeGroups.AutoScalingGroups[idx], nil
		}
	}

	return nil, nil
}

// getELBInstanceStates returns a mapping of the instance states of all the ELBs attached to the
// provided ASG.
//
// Note that this is the instance state function for ELB Classic.
//
// Nested like: lbName -> instanceId -> instanceState
func getELBInstanceStates(g *autoscaling.Group, meta interface{}) (map[string]map[string]string, error) {
	lbInstanceStates := make(map[string]map[string]string)
	conn := meta.(*conns.AWSClient).ELBConn

	for _, lbName := range g.LoadBalancerNames {
		lbInstanceStates[aws.StringValue(lbName)] = make(map[string]string)
		opts := &elb.DescribeInstanceHealthInput{LoadBalancerName: lbName}
		r, err := conn.DescribeInstanceHealth(opts)
		if err != nil {
			return nil, err
		}
		for _, is := range r.InstanceStates {
			if is == nil || is.InstanceId == nil || is.State == nil {
				continue
			}
			lbInstanceStates[aws.StringValue(lbName)][aws.StringValue(is.InstanceId)] = aws.StringValue(is.State)
		}
	}

	return lbInstanceStates, nil
}

// getTargetGroupInstanceStates returns a mapping of the instance states of
// all the ALB target groups attached to the provided ASG.
//
// Note that this is the instance state function for Application Load
// Balancing (aka ELBv2).
//
// Nested like: targetGroupARN -> instanceId -> instanceState
func getTargetGroupInstanceStates(g *autoscaling.Group, meta interface{}) (map[string]map[string]string, error) {
	targetInstanceStates := make(map[string]map[string]string)
	conn := meta.(*conns.AWSClient).ELBV2Conn

	for _, targetGroupARN := range g.TargetGroupARNs {
		targetInstanceStates[aws.StringValue(targetGroupARN)] = make(map[string]string)
		opts := &elbv2.DescribeTargetHealthInput{TargetGroupArn: targetGroupARN}
		r, err := conn.DescribeTargetHealth(opts)
		if err != nil {
			return nil, err
		}
		for _, desc := range r.TargetHealthDescriptions {
			if desc == nil || desc.Target == nil || desc.Target.Id == nil || desc.TargetHealth == nil || desc.TargetHealth.State == nil {
				continue
			}
			targetInstanceStates[aws.StringValue(targetGroupARN)][aws.StringValue(desc.Target.Id)] = aws.StringValue(desc.TargetHealth.State)
		}
	}

	return targetInstanceStates, nil
}
