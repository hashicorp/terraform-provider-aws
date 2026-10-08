// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_flow_log")
func newFlowLogResourceAsListResource() inttypes.ListResourceForSDK {
	l := flowLogListResource{}
	l.SetResourceSchema(resourceFlowLog())
	return &l
}

var _ list.ListResource = &flowLogListResource{}

type flowLogListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *flowLogListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &ec2.DescribeFlowLogsInput{}
		for item, err := range listFlowLogsPages(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			flowLogID := aws.ToString(item.FlowLogId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), flowLogID)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(flowLogID)

			if request.IncludeResource {
				if err := resourceFlowLogFlatten(ctx, l.Meta(), &item, rd); err != nil {
					tflog.Error(ctx, "Reading Flow Log", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = flowLogID

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

func listFlowLogsPages(ctx context.Context, conn *ec2.Client, input *ec2.DescribeFlowLogsInput) iter.Seq2[awstypes.FlowLog, error] {
	return func(yield func(awstypes.FlowLog, error) bool) {
		pages := ec2.NewDescribeFlowLogsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.FlowLog{}, fmt.Errorf("listing Flow Logs: %w", err))
				return
			}

			for _, item := range page.FlowLogs {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
