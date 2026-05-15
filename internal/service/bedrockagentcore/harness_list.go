// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_bedrockagentcore_harness")
func newHarnessResourceAsListResource() list.ListResourceWithConfigure {
	return &harnessListResource{}
}

var _ list.ListResource = &harnessListResource{}

type harnessListResource struct {
	harnessResource
	framework.WithList
}

func (l *harnessListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BedrockAgentCoreClient(ctx)

	var query listHarnessModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input bedrockagentcorecontrol.ListHarnessesInput
		for item, err := range listHarnesses(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.Arn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			harnessID := aws.ToString(item.HarnessId)
			output, err := findHarnessByID(ctx, conn, harnessID)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data harnessResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, output, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = aws.ToString(item.HarnessName)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listHarnessModel struct {
	framework.WithRegionModel
}

func listHarnesses(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.ListHarnessesInput) iter.Seq2[awstypes.HarnessSummary, error] {
	return func(yield func(awstypes.HarnessSummary, error) bool) {
		pages := bedrockagentcorecontrol.NewListHarnessesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.HarnessSummary](), fmt.Errorf("listing Bedrock AgentCore Harnesses: %w", err))
				return
			}

			for _, item := range page.Harnesses {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
