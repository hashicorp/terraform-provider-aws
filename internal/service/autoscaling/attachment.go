// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_autoscaling_attachment")
func ResourceAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAttachmentCreate,
		ReadWithoutTimeout:   resourceAttachmentRead,
		DeleteWithoutTimeout: resourceAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"autoscaling_group_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"elb": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ExactlyOneOf: []string{"elb", "lb_target_group_arn"},
			},
			"lb_target_group_arn": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ExactlyOneOf: []string{"elb", "lb_target_group_arn"},
			},
		},
	}
}

func resourceAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn(ctx)
	asgName := d.Get("autoscaling_group_name").(string)

	if v, ok := d.GetOk("elb"); ok {
		lbName := v.(string)
		input := &autoscaling.AttachLoadBalancersInput{
			AutoScalingGroupName: aws.String(asgName),
			LoadBalancerNames:    aws.StringSlice([]string{lbName}),
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, d.Timeout(schema.TimeoutCreate),
			func() (interface{}, error) {
				return conn.AttachLoadBalancersWithContext(ctx, input)
			},
			// ValidationError: Trying to update too many Load Balancers/Target Groups at once. The limit is 10
			ErrCodeValidationError, "update too many")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "attaching Auto Scaling Group (%s) load balancer (%s): %s", asgName, lbName, err)
		}
	} else {
		input := &autoscaling.AttachLoadBalancerTargetGroupsInput{
			AutoScalingGroupName: aws.String(asgName),
			TargetGroupARNs:      aws.StringSlice([]string{d.Get("lb_target_group_arn").(string)}),
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, d.Timeout(schema.TimeoutCreate),
			func() (interface{}, error) {
				return conn.AttachLoadBalancerTargetGroupsWithContext(ctx, input)
			},
			ErrCodeValidationError, "update too many")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "attaching Auto Scaling Group (%s) target group (%s): %s", asgName, d.Get("lb_target_group_arn").(string), err)
		}
	}

	//lintignore:R016 // Allow legacy unstable ID usage in managed resource
	d.SetId(id.PrefixedUniqueId(fmt.Sprintf("%s-", asgName)))

	return append(diags, resourceAttachmentRead(ctx, d, meta)...)
}

func resourceAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn(ctx)
	asgName := d.Get("autoscaling_group_name").(string)

	var err error

	if v, ok := d.GetOk("elb"); ok {
		err = FindAttachmentByLoadBalancerName(ctx, conn, asgName, v.(string))
	} else {
		err = FindAttachmentByTargetGroupARN(ctx, conn, asgName, d.Get("lb_target_group_arn").(string))
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Auto Scaling Group Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Group Attachment (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn(ctx)
	asgName := d.Get("autoscaling_group_name").(string)

	if v, ok := d.GetOk("elb"); ok {
		lbName := v.(string)
		input := &autoscaling.DetachLoadBalancersInput{
			AutoScalingGroupName: aws.String(asgName),
			LoadBalancerNames:    aws.StringSlice([]string{lbName}),
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, d.Timeout(schema.TimeoutCreate),
			func() (interface{}, error) {
				return conn.DetachLoadBalancersWithContext(ctx, input)
			},
			ErrCodeValidationError, "update too many")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "detaching Auto Scaling Group (%s) load balancer (%s): %s", asgName, lbName, err)
		}
	} else {
		input := &autoscaling.DetachLoadBalancerTargetGroupsInput{
			AutoScalingGroupName: aws.String(asgName),
			TargetGroupARNs:      aws.StringSlice([]string{d.Get("lb_target_group_arn").(string)}),
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, d.Timeout(schema.TimeoutCreate),
			func() (interface{}, error) {
				return conn.DetachLoadBalancerTargetGroupsWithContext(ctx, input)
			},
			ErrCodeValidationError, "update too many")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "detaching Auto Scaling Group (%s) target group (%s): %s", asgName, d.Get("lb_target_group_arn").(string), err)
		}
	}

	return diags
}

func FindAttachmentByLoadBalancerName(ctx context.Context, conn *autoscaling.AutoScaling, asgName, loadBalancerName string) error {
	asg, err := FindGroupByName(ctx, conn, asgName)

	if err != nil {
		return err
	}

	for _, v := range asg.LoadBalancerNames {
		if aws.StringValue(v) == loadBalancerName {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastError: fmt.Errorf("Auto Scaling Group (%s) load balancer (%s) attachment not found", asgName, loadBalancerName),
	}
}

func FindAttachmentByTargetGroupARN(ctx context.Context, conn *autoscaling.AutoScaling, asgName, targetGroupARN string) error {
	asg, err := FindGroupByName(ctx, conn, asgName)

	if err != nil {
		return err
	}

	for _, v := range asg.TargetGroupARNs {
		if aws.StringValue(v) == targetGroupARN {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastError: fmt.Errorf("Auto Scaling Group (%s) target group (%s) attachment not found", asgName, targetGroupARN),
	}
}
