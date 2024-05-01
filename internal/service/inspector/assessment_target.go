// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_inspector_assessment_target")
func ResourceAssessmentTarget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAssessmentTargetCreate,
		ReadWithoutTimeout:   resourceAssessmentTargetRead,
		UpdateWithoutTimeout: resourceAssessmentTargetUpdate,
		DeleteWithoutTimeout: resourceAssessmentTargetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_group_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAssessmentTargetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorConn(ctx)

	input := &inspector.CreateAssessmentTargetInput{
		AssessmentTargetName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("resource_group_arn"); ok {
		input.ResourceGroupArn = aws.String(v.(string))
	}

	resp, err := conn.CreateAssessmentTargetWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Inspector Classic Assessment Target: %s", err)
	}

	d.SetId(aws.StringValue(resp.AssessmentTargetArn))

	return append(diags, resourceAssessmentTargetRead(ctx, d, meta)...)
}

func resourceAssessmentTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorConn(ctx)

	assessmentTarget, err := DescribeAssessmentTarget(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Inspector Classic Assessment Target (%s): %s", d.Id(), err)
	}

	if assessmentTarget == nil {
		log.Printf("[WARN] Inspector Classic Assessment Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("arn", assessmentTarget.Arn)
	d.Set("name", assessmentTarget.Name)
	d.Set("resource_group_arn", assessmentTarget.ResourceGroupArn)

	return diags
}

func resourceAssessmentTargetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorConn(ctx)

	input := inspector.UpdateAssessmentTargetInput{
		AssessmentTargetArn:  aws.String(d.Id()),
		AssessmentTargetName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("resource_group_arn"); ok {
		input.ResourceGroupArn = aws.String(v.(string))
	}

	_, err := conn.UpdateAssessmentTargetWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Inspector Classic Assessment Target (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAssessmentTargetRead(ctx, d, meta)...)
}

func resourceAssessmentTargetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorConn(ctx)
	input := &inspector.DeleteAssessmentTargetInput{
		AssessmentTargetArn: aws.String(d.Id()),
	}
	err := retry.RetryContext(ctx, 60*time.Minute, func() *retry.RetryError {
		_, err := conn.DeleteAssessmentTargetWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, inspector.ErrCodeAssessmentRunInProgressException) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteAssessmentTargetWithContext(ctx, input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Inspector Classic Assessment Target: %s", err)
	}
	return diags
}

func DescribeAssessmentTarget(ctx context.Context, conn *inspector.Inspector, arn string) (*inspector.AssessmentTarget, error) {
	input := &inspector.DescribeAssessmentTargetsInput{
		AssessmentTargetArns: []*string{aws.String(arn)},
	}

	output, err := conn.DescribeAssessmentTargetsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, inspector.ErrCodeInvalidInputException) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	var assessmentTarget *inspector.AssessmentTarget
	for _, target := range output.AssessmentTargets {
		if aws.StringValue(target.Arn) == arn {
			assessmentTarget = target
			break
		}
	}

	return assessmentTarget, nil
}
