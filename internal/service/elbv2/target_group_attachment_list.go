// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_lb_target_group_attachment", name="Target Group Attachment")
func newTargetGroupAttachmentResourceAsListResource() inttypes.ListResourceForSDK {
	l := targetGroupAttachmentListResource{}
	l.SetResourceSchema(resourceTargetGroupAttachment())
	return &l
}

var _ list.ListResource = &targetGroupAttachmentListResource{}

type targetGroupAttachmentListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type listTargetGroupAttachmentModel struct {
	framework.WithRegionModel
	TargetGroupARN types.String `tfsdk:"target_group_arn"`
}

func (l *targetGroupAttachmentListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"target_group_arn": listschema.StringAttribute{
				Required:    true,
				Description: "ARN of the Target Group to list Attachments from.",
			},
		},
	}
}

func (l *targetGroupAttachmentListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ELBV2Client(ctx)

	var query listTargetGroupAttachmentModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	targetGroupARN := query.TargetGroupARN.ValueString()
	input := elasticloadbalancingv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(targetGroupARN),
	}

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("target_group_arn"): targetGroupARN,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listTargetGroupAttachments(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			targetID := aws.ToString(item.Target.Id)
			if targetID == "" {
				continue
			}

			ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("target_id"), targetID)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.Set("target_group_arn", targetGroupARN)
			flattenTargetGroupAttachment(rd, item.Target)
			rd.SetId(targetGroupAttachmentImportID{}.Create(rd))

			if request.IncludeResource { //nolint:revive,staticcheck // Be explicit about IncludeResource handling.
				// No-op, all readable attributes are already populated above.
			}

			result.DisplayName = fmt.Sprintf("%s (%s)", targetID, targetGroupARN)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

func listTargetGroupAttachments(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeTargetHealthInput) iter.Seq2[awstypes.TargetHealthDescription, error] {
	return func(yield func(awstypes.TargetHealthDescription, error) bool) {
		output, err := conn.DescribeTargetHealth(ctx, input)
		if err != nil {
			yield(awstypes.TargetHealthDescription{}, fmt.Errorf("listing ELB Target Group Attachment resources for target group (%s): %w", aws.ToString(input.TargetGroupArn), err))
			return
		}

		for _, item := range output.TargetHealthDescriptions {
			if !yield(item, nil) {
				return
			}
		}
	}
}

func flattenTargetGroupAttachment(d *schema.ResourceData, target *awstypes.TargetDescription) {
	d.Set("target_id", target.Id)
	d.Set(names.AttrPort, target.Port)
	d.Set(names.AttrAvailabilityZone, target.AvailabilityZone)

	if v := aws.ToString(target.QuicServerId); v != "" {
		d.Set("quic_server_id", v)
	}
}
