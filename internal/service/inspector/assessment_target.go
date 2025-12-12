// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package inspector

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_inspector_assessment_target", name="Assessment Target")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/inspector/types;types.AssessmentTarget")
// @Testing(preIdentityVersion="v6.4.0")
func ResourceAssessmentTarget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAssessmentTargetCreate,
		ReadWithoutTimeout:   resourceAssessmentTargetRead,
		UpdateWithoutTimeout: resourceAssessmentTargetUpdate,
		DeleteWithoutTimeout: resourceAssessmentTargetDelete,

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			names.AttrARN: {
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

const (
	ResNameAssessmentTarget = "Assessment Target"
)

func resourceAssessmentTargetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	input := &inspector.CreateAssessmentTargetInput{
		AssessmentTargetName: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("resource_group_arn"); ok {
		input.ResourceGroupArn = aws.String(v.(string))
	}

	resp, err := conn.CreateAssessmentTarget(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Inspector Classic Assessment Target: %s", err)
	}

	d.SetId(aws.ToString(resp.AssessmentTargetArn))

	return append(diags, resourceAssessmentTargetRead(ctx, d, meta)...)
}

func resourceAssessmentTargetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	assessmentTarget, err := FindAssessmentTargetByID(ctx, conn, d.Id())
	if errs.IsA[*sdkretry.NotFoundError](err) {
		log.Printf("[WARN] Inspector Classic Assessment Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Inspector Classic Assessment Target (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, assessmentTarget.Arn)
	d.Set(names.AttrName, assessmentTarget.Name)
	d.Set("resource_group_arn", assessmentTarget.ResourceGroupArn)

	return diags
}

func resourceAssessmentTargetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	input := inspector.UpdateAssessmentTargetInput{
		AssessmentTargetArn:  aws.String(d.Id()),
		AssessmentTargetName: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("resource_group_arn"); ok {
		input.ResourceGroupArn = aws.String(v.(string))
	}

	_, err := conn.UpdateAssessmentTarget(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Inspector Classic Assessment Target (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAssessmentTargetRead(ctx, d, meta)...)
}

func resourceAssessmentTargetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)
	input := &inspector.DeleteAssessmentTargetInput{
		AssessmentTargetArn: aws.String(d.Id()),
	}
	err := tfresource.Retry(ctx, 60*time.Minute, func(ctx context.Context) *tfresource.RetryError {
		_, err := conn.DeleteAssessmentTarget(ctx, input)

		if errs.IsA[*awstypes.AssessmentRunInProgressException](err) {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Inspector Classic Assessment Target: %s", err)
	}
	return diags
}

func FindAssessmentTargetByID(ctx context.Context, conn *inspector.Client, arn string) (*awstypes.AssessmentTarget, error) {
	input := &inspector.DescribeAssessmentTargetsInput{
		AssessmentTargetArns: []string{arn},
	}

	output, err := conn.DescribeAssessmentTargets(ctx, input)

	if errs.IsA[*awstypes.InvalidInputException](err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	for _, target := range output.AssessmentTargets {
		if aws.ToString(target.Arn) == arn {
			return &target, nil
		}
	}

	return nil, &sdkretry.NotFoundError{
		LastRequest: input,
	}
}
