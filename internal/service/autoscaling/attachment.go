package autoscaling

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAttachmentCreate,
		Read:   resourceAttachmentRead,
		Delete: resourceAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"alb_target_group_arn": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				Deprecated:    "Use lb_target_group_arn instead",
				ConflictsWith: []string{"lb_target_group_arn"},
			},
			"autoscaling_group_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"elb": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"lb_target_group_arn": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{"alb_target_group_arn"},
			},
		},
	}
}

func resourceAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn
	asgName := d.Get("autoscaling_group_name").(string)

	if v, ok := d.GetOk("elb"); ok {
		lbName := v.(string)
		input := &autoscaling.AttachLoadBalancersInput{
			AutoScalingGroupName: aws.String(asgName),
			LoadBalancerNames:    aws.StringSlice([]string{lbName}),
		}

		if _, err := conn.AttachLoadBalancers(input); err != nil {
			return fmt.Errorf("attaching Auto Scaling Group (%s) load balancer (%s): %w", asgName, lbName, err)
		}
	}

	var targetGroupARN string
	if v, ok := d.GetOk("alb_target_group_arn"); ok {
		targetGroupARN = v.(string)
	} else if v, ok := d.GetOk("lb_target_group_arn"); ok {
		targetGroupARN = v.(string)
	}

	if targetGroupARN != "" {
		input := &autoscaling.AttachLoadBalancerTargetGroupsInput{
			AutoScalingGroupName: aws.String(asgName),
			TargetGroupARNs:      aws.StringSlice([]string{targetGroupARN}),
		}

		if _, err := conn.AttachLoadBalancerTargetGroups(input); err != nil {
			return fmt.Errorf("attaching Auto Scaling Group (%s) target group (%s): %w", asgName, targetGroupARN, err)
		}
	}

	//lintignore:R016 // Allow legacy unstable ID usage in managed resource
	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-", asgName)))

	return resourceAttachmentRead(d, meta)
}

func resourceAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn
	asgName := d.Get("autoscaling_group_name").(string)

	asg, err := FindGroupByName(conn, asgName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Auto Scaling Group %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Auto Scaling Group (%s): %w", asgName, err)
	}

	if v, ok := d.GetOk("elb"); ok {
		found := false
		lbName := v.(string)

		for _, v := range asg.LoadBalancerNames {
			if aws.StringValue(v) == lbName {
				d.Set("elb", lbName)
				found = true
				break
			}
		}

		if !found {
			log.Printf("[WARN] Association for %s was not found in ASG association", lbName)
			d.SetId("")
		}
	}

	var key, targetGroupARN string
	if v, ok := d.GetOk("alb_target_group_arn"); ok {
		key = "alb_target_group_arn"
		targetGroupARN = v.(string)
	} else if v, ok := d.GetOk("lb_target_group_arn"); ok {
		key = "lb_target_group_arn"
		targetGroupARN = v.(string)
	}

	if targetGroupARN != "" {
		found := false

		for _, v := range asg.TargetGroupARNs {
			if aws.StringValue(v) == targetGroupARN {
				d.Set(key, targetGroupARN)
				found = true
				break
			}
		}

		if !found {
			log.Printf("[WARN] Association for %s was not found in ASG association", targetGroupARN)
			d.SetId("")
		}
	}

	return nil
}

func resourceAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn
	asgName := d.Get("autoscaling_group_name").(string)

	if v, ok := d.GetOk("elb"); ok {
		lbName := v.(string)
		input := &autoscaling.DetachLoadBalancersInput{
			AutoScalingGroupName: aws.String(asgName),
			LoadBalancerNames:    aws.StringSlice([]string{lbName}),
		}

		if _, err := conn.DetachLoadBalancers(input); err != nil {
			return fmt.Errorf("detaching Auto Scaling Group (%s) load balancer (%s): %w", asgName, lbName, err)
		}
	}

	var targetGroupARN string
	if v, ok := d.GetOk("alb_target_group_arn"); ok {
		targetGroupARN = v.(string)
	} else if v, ok := d.GetOk("lb_target_group_arn"); ok {
		targetGroupARN = v.(string)
	}

	if targetGroupARN != "" {
		input := &autoscaling.DetachLoadBalancerTargetGroupsInput{
			AutoScalingGroupName: aws.String(asgName),
			TargetGroupARNs:      aws.StringSlice([]string{targetGroupARN}),
		}

		if _, err := conn.DetachLoadBalancerTargetGroups(input); err != nil {
			return fmt.Errorf("detaching Auto Scaling Group (%s) target group (%s): %w", asgName, targetGroupARN, err)
		}
	}

	return nil
}
