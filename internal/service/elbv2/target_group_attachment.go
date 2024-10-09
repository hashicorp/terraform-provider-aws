// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_alb_target_group_attachment", name="Target Group Attachment")
// @SDKResource("aws_lb_target_group_attachment", name="Target Group Attachment")
func resourceTargetGroupAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAttachmentCreate,
		ReadWithoutTimeout:   resourceAttachmentRead,
		DeleteWithoutTimeout: resourceAttachmentDelete,

		Schema: map[string]*schema.Schema{
			names.AttrAvailabilityZone: {
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
			names.AttrPort: {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
			},
		},
	}
}

func resourceAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	targetGroupARN := d.Get("target_group_arn").(string)
	input := &elasticloadbalancingv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(targetGroupARN),
		Targets: []awstypes.TargetDescription{{
			Id: aws.String(d.Get("target_id").(string)),
		}},
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
		input.Targets[0].AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPort); ok {
		input.Targets[0].Port = aws.Int32(int32(v.(int)))
	}

	const (
		timeout = 10 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[*awstypes.InvalidTargetException](ctx, timeout, func() (interface{}, error) {
		return conn.RegisterTargets(ctx, input)
	})

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
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	targetGroupARN := d.Get("target_group_arn").(string)
	input := &elasticloadbalancingv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(targetGroupARN),
		Targets: []awstypes.TargetDescription{{
			Id: aws.String(d.Get("target_id").(string)),
		}},
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
		input.Targets[0].AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPort); ok {
		input.Targets[0].Port = aws.Int32(int32(v.(int)))
	}

	_, err := findTargetHealthDescription(ctx, conn, input)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELBv2 Target Group Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Target Group Attachment (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	targetGroupARN := d.Get("target_group_arn").(string)
	input := &elasticloadbalancingv2.DeregisterTargetsInput{
		TargetGroupArn: aws.String(targetGroupARN),
		Targets: []awstypes.TargetDescription{{
			Id: aws.String(d.Get("target_id").(string)),
		}},
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
		input.Targets[0].AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPort); ok {
		input.Targets[0].Port = aws.Int32(int32(v.(int)))
	}

	log.Printf("[DEBUG] Deleting ELBv2 Target Group Attachment: %s", d.Id())
	_, err := conn.DeregisterTargets(ctx, input)

	if errs.IsA[*awstypes.LoadBalancerNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deregistering ELBv2 Target Group (%s) target: %s", targetGroupARN, err)
	}

	return diags
}

func findTargetHealthDescription(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeTargetHealthInput) (*awstypes.TargetHealthDescription, error) {
	output, err := findTargetHealthDescriptions(ctx, conn, input, func(v *awstypes.TargetHealthDescription) bool {
		// This will catch targets being removed by hand (draining as we plan) or that have been removed for a while
		// without trying to re-create ones that are just not in use. For example, a target can be `unused` if the
		// target group isnt assigned to anything, a scenario where we don't want to continuously recreate the resource.
		if v := v.TargetHealth; v != nil {
			switch v.Reason {
			case awstypes.TargetHealthReasonEnumDeregistrationInProgress, awstypes.TargetHealthReasonEnumNotRegistered:
				return false
			default:
				return true
			}
		}

		return false
	})

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTargetHealthDescriptions(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeTargetHealthInput, filter tfslices.Predicate[*awstypes.TargetHealthDescription]) ([]awstypes.TargetHealthDescription, error) {
	var targetHealthDescriptions []awstypes.TargetHealthDescription

	output, err := conn.DescribeTargetHealth(ctx, input)

	if errs.IsA[*awstypes.InvalidTargetException](err) || errs.IsA[*awstypes.TargetGroupNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, v := range output.TargetHealthDescriptions {
		if filter(&v) {
			targetHealthDescriptions = append(targetHealthDescriptions, v)
		}
	}

	return targetHealthDescriptions, nil
}
