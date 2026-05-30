// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package inspector

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_inspector_assessment_target", name="Assessment Target")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/inspector/types;types.AssessmentTarget")
// @Testing(preIdentityVersion="v6.4.0")
// @Testing(preCheck="testAccPreCheck")
func resourceAssessmentTarget() *schema.Resource {
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceAssessmentTargetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := inspector.CreateAssessmentTargetInput{
		AssessmentTargetName: aws.String(name),
	}

	if v, ok := d.GetOk("resource_group_arn"); ok {
		input.ResourceGroupArn = aws.String(v.(string))
	}

	output, err := conn.CreateAssessmentTarget(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Inspector Classic Assessment Target (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.AssessmentTargetArn))

	return append(diags, resourceAssessmentTargetRead(ctx, d, meta)...)
}

func resourceAssessmentTargetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	assessmentTarget, err := findAssessmentTargetByARN(ctx, conn, d.Id())

	if retry.NotFound(err) {
		log.Printf("[WARN] Inspector Classic Assessment Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inspector Classic Assessment Target (%s): %s", d.Id(), err)
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

	const (
		timeout = 60 * time.Minute
	)
	input := inspector.DeleteAssessmentTargetInput{
		AssessmentTargetArn: aws.String(d.Id()),
	}
	_, err := tfresource.RetryWhenIsA[any, *awstypes.AssessmentRunInProgressException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.DeleteAssessmentTarget(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Inspector Classic Assessment Target (%s): %s", d.Id(), err)
	}

	return diags
}

func findAssessmentTargets(ctx context.Context, conn *inspector.Client, input *inspector.DescribeAssessmentTargetsInput) ([]awstypes.AssessmentTarget, error) {
	output, err := conn.DescribeAssessmentTargets(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	if err := failedItemsError(output.FailedItems); err != nil {
		return nil, err
	}

	return output.AssessmentTargets, nil
}

func findAssessmentTarget(ctx context.Context, conn *inspector.Client, input *inspector.DescribeAssessmentTargetsInput) (*awstypes.AssessmentTarget, error) {
	output, err := findAssessmentTargets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAssessmentTargetByARN(ctx context.Context, conn *inspector.Client, arn string) (*awstypes.AssessmentTarget, error) {
	input := inspector.DescribeAssessmentTargetsInput{
		AssessmentTargetArns: []string{arn},
	}

	output, err := findAssessmentTarget(ctx, conn, &input)

	if tfawserr.ErrMessageContains(err, string(awstypes.FailedItemErrorCodeItemDoesNotExist), arn) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
