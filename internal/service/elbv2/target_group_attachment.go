// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_alb_target_group_attachment")
// @SDKResource("aws_lb_target_group_attachment")
func ResourceTargetGroupAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAttachmentCreate,
		ReadWithoutTimeout:   resourceAttachmentRead,
		DeleteWithoutTimeout: resourceAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"target_group_arn": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"target_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"port": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
			},
		},
	}
}

func resourceAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	targetGroupARN := d.Get("target_group_arn").(string)
	input := &elbv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(targetGroupARN),
		Targets: []*elbv2.TargetDescription{{
			Id: aws.String(d.Get("target_id").(string)),
		}},
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		input.Targets[0].AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("port"); ok {
		input.Targets[0].Port = aws.Int64(int64(v.(int)))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 10*time.Minute, func() (interface{}, error) {
		return conn.RegisterTargetsWithContext(ctx, input)
	}, elbv2.ErrCodeInvalidTargetException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering ELBv2 Target Group (%s) target: %s", targetGroupARN, err)
	}

	//lintignore:R016 // Allow legacy unstable ID usage in managed resource
	d.SetId(id.PrefixedUniqueId(targetGroupARN + "-"))

	return diags
}

// resourceAttachmentRead requires all of the fields in order to describe the correct
// target, so there is no work to do beyond ensuring that the target and group still exist.
func resourceAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	target := &elbv2.TargetDescription{
		Id: aws.String(d.Get("target_id").(string)),
	}

	if v, ok := d.GetOk("port"); ok {
		target.Port = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		target.AvailabilityZone = aws.String(v.(string))
	}

	resp, err := conn.DescribeTargetHealthWithContext(ctx, &elbv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(d.Get("target_group_arn").(string)),
		Targets:        []*elbv2.TargetDescription{target},
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeTargetGroupNotFoundException) {
			log.Printf("[WARN] Target group does not exist, removing target attachment %s", d.Id())
			d.SetId("")
			return diags
		}
		if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeInvalidTargetException) {
			log.Printf("[WARN] Target does not exist, removing target attachment %s", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Target Health: %s", err)
	}

	for _, targetDesc := range resp.TargetHealthDescriptions {
		if targetDesc == nil || targetDesc.Target == nil {
			continue
		}

		if aws.StringValue(targetDesc.Target.Id) == d.Get("target_id").(string) {
			// These will catch targets being removed by hand (draining as we plan) or that have been removed for a while
			// without trying to re-create ones that are just not in use. For example, a target can be `unused` if the
			// target group isnt assigned to anything, a scenario where we don't want to continuously recreate the resource.
			if targetDesc.TargetHealth == nil {
				continue
			}

			reason := aws.StringValue(targetDesc.TargetHealth.Reason)

			if reason == elbv2.TargetHealthReasonEnumTargetNotRegistered || reason == elbv2.TargetHealthReasonEnumTargetDeregistrationInProgress {
				log.Printf("[WARN] Target Attachment does not exist, recreating attachment")
				d.SetId("")
				return diags
			}
		}
	}

	if len(resp.TargetHealthDescriptions) != 1 {
		log.Printf("[WARN] Target does not exist, removing target attachment %s", d.Id())
		d.SetId("")
		return diags
	}

	return diags
}

func resourceAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	targetGroupARN := d.Get("target_group_arn").(string)
	input := &elbv2.DeregisterTargetsInput{
		TargetGroupArn: aws.String(targetGroupARN),
		Targets: []*elbv2.TargetDescription{{
			Id: aws.String(d.Get("target_id").(string)),
		}},
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		input.Targets[0].AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("port"); ok {
		input.Targets[0].Port = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Deleting ELBv2 Target Group Attachment: %s", d.Id())
	_, err := conn.DeregisterTargetsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeTargetGroupNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deregistering ELBv2 Target Group (%s) target: %s", targetGroupARN, err)
	}

	return diags
}
