package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsAutoscalingInstanceRefresh() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAutoscalingInstanceRefreshCreate,
		Read:   resourceAwsAutoscalingInstanceRefreshRead,
		Update: resourceAwsAutoscalingInstanceRefreshUpdate,
		Delete: resourceAwsAutoscalingInstanceRefreshDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				fields := strings.Split(d.Id(), "/")
				if len(fields) != 2 || fields[0] == "" || fields[1] == "" {
					return nil, fmt.Errorf("invalid id %s: expected asg-name/instance-refresh-id", d.Id())
				}

				d.Set("autoscaling_group_name", fields[0])
				d.Set("instance_refresh_id", fields[1])
				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"autoscaling_group_name": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(1, 255),
				Required:     true,
				ForceNew:     true,
			},
			"instance_refresh_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_warmup_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      -1, // default to ASG health check grace period
				ValidateFunc: validation.IntAtLeast(-1),
			},
			"min_healthy_percentage": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      90,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			"strategy": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice(
					[]string{autoscaling.RefreshStrategyRolling},
					false),
			},
			"triggers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"wait_for_completion": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceAwsAutoscalingInstanceRefreshCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn

	asgName := d.Get("autoscaling_group_name").(string)
	instanceWarmupSeconds := d.Get("instance_warmup_seconds").(int)
	minHealthyPercentage := d.Get("min_healthy_percentage").(int)
	strategy := d.Get("strategy").(string)
	waitForCompletion := d.Get("wait_for_completion").(bool)

	input := autoscaling.StartInstanceRefreshInput{
		AutoScalingGroupName: aws.String(asgName),
		Preferences: &autoscaling.RefreshPreferences{
			MinHealthyPercentage: aws.Int64(int64(minHealthyPercentage)),
		},
		Strategy: aws.String(strategy),
	}

	if instanceWarmupSeconds > -1 {
		input.Preferences.InstanceWarmup = aws.Int64(int64(instanceWarmupSeconds))
	}

	output, err := conn.StartInstanceRefresh(&input)
	if err != nil {
		return fmt.Errorf("start instance refresh of %s: %s", asgName, err)
	}

	refreshId := aws.StringValue(output.InstanceRefreshId)
	d.Set("instance_refresh_id", refreshId)
	d.SetId(asgName + "/" + refreshId)

	log.Printf("[DEBUG] started instance refresh %s", d.Id())

	if waitForCompletion {
		switch err := waitUntilAutoscalingGroupInstanceRefreshTerminal(
			conn, d.Timeout(schema.TimeoutCreate), asgName, refreshId); {
		case isResourceTimeoutError(err):
			log.Printf("[WARN] instance refresh %s timed out; cancelling...", d.Id())

			err := cancelAutoscalingInstanceRefresh(conn, asgName)
			if err != nil {
				return fmt.Errorf("cancel instance refresh %s: %w", d.Id(), err)
			}

			err = waitUntilAutoscalingGroupInstanceRefreshTerminal(
				conn, d.Timeout(schema.TimeoutDelete), asgName, "")
			if err != nil {
				return fmt.Errorf("wait for cancellation of instance refresh %s: %w", d.Id(), err)
			}

			return fmt.Errorf("instance refresh %s cancelled: create timed out; consider increasing the timeout", d.Id())

		case err != nil:
			return fmt.Errorf("instance refresh %s failed: %s", d.Id(), err)
		}
	}

	return resourceAwsAutoscalingInstanceRefreshRead(d, meta)
}

func resourceAwsAutoscalingInstanceRefreshRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn
	asgName := d.Get("autoscaling_group_name").(string)
	refreshId := d.Get("instance_refresh_id").(string)

	_, err := describeAutoscalingInstanceRefresh(conn, asgName, refreshId)
	switch {
	case isResourceNotFoundError(err):
		log.Printf("[WARN] removing instance refresh %s/%s from state: not found", asgName, refreshId)
		d.SetId("")
		return nil
	case err != nil:
		return fmt.Errorf("describe instance refresh %s/%s: %s", asgName, refreshId, err)
	}

	// DescribeInstanceRefreshes doesn't provide information about the original
	// parameters used to trigger the refresh, so keep whatever was in state.

	return nil
}

func resourceAwsAutoscalingInstanceRefreshUpdate(d *schema.ResourceData, meta interface{}) error {
	// An instance refresh cannot be updated. Changes will take effect when
	// the next refresh is created by way of ForceNew.
	return nil
}

func resourceAwsAutoscalingInstanceRefreshDelete(d *schema.ResourceData, meta interface{}) error {
	// An instance refresh cannot be deleted.
	d.SetId("")
	return nil
}

func describeAutoscalingInstanceRefresh(
	conn *autoscaling.AutoScaling,
	asgName, refreshId string,
) (*autoscaling.InstanceRefresh, error) {
	input := autoscaling.DescribeInstanceRefreshesInput{
		AutoScalingGroupName: aws.String(asgName),
		MaxRecords:           aws.Int64(1),
	}

	if refreshId != "" {
		input.InstanceRefreshIds = aws.StringSlice([]string{refreshId})
	}

	output, err := conn.DescribeInstanceRefreshes(&input)
	switch {
	case err != nil:
		return nil, fmt.Errorf("describe instance refreshes: %s", err)
	case len(output.InstanceRefreshes) != 1:
		return nil, &resource.NotFoundError{
			Message: fmt.Sprintf("instance refresh %s/%s", asgName, refreshId),
		}
	}

	return output.InstanceRefreshes[0], nil
}

func waitUntilAutoscalingGroupInstanceRefreshTerminal(
	conn *autoscaling.AutoScaling,
	timeout time.Duration,
	asgName, refreshId string,
) error {
	log.Printf("[DEBUG] waiting for terminal state of instance refresh %s/%s...", asgName, refreshId)

	errNotTerminal := fmt.Errorf("refresh status is not terminal")
	err := resource.Retry(timeout, func() *resource.RetryError {
		instanceRefresh, err := describeAutoscalingInstanceRefresh(conn, asgName, refreshId)
		switch {
		case isResourceNotFoundError(err):
			return resource.NonRetryableError(fmt.Errorf("instance refresh %s/%s not found", asgName, refreshId))
		case err != nil:
			return resource.NonRetryableError(err)
		}

		log.Printf(
			"[DEBUG] instance refresh %s/%s state is %s",
			asgName, refreshId, aws.StringValue(instanceRefresh.Status))

		switch status := aws.StringValue(instanceRefresh.Status); status {
		case
			autoscaling.InstanceRefreshStatusCancelled,
			autoscaling.InstanceRefreshStatusFailed,
			autoscaling.InstanceRefreshStatusSuccessful:
			return nil
		default:
			return resource.RetryableError(errNotTerminal)
		}
	})

	switch {
	case err == errNotTerminal:
		// work around the .LastError == nil condition in isResourceTimeoutError
		return &resource.TimeoutError{}
	case err != nil:
		return err
	default:
		return nil
	}
}

func cancelAutoscalingInstanceRefresh(
	conn *autoscaling.AutoScaling,
	asgName string,
) error {
	log.Printf("[DEBUG] cancelling refresh of %s...", asgName)

	input := autoscaling.CancelInstanceRefreshInput{
		AutoScalingGroupName: aws.String(asgName),
	}

	_, err := conn.CancelInstanceRefresh(&input)
	switch {
	case isAWSErr(err, autoscaling.ErrCodeActiveInstanceRefreshNotFoundFault, ""):
		log.Printf("[DEBUG] No active Instance Refresh in ASG %s", asgName)
		return nil
	case err != nil:
		return err
	}

	return nil
}
