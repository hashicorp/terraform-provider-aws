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
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_bedrockagentcore_online_evaluation_config")
func newOnlineEvaluationConfigResourceAsListResource() list.ListResourceWithConfigure {
	return &onlineEvaluationConfigListResource{}
}

var _ list.ListResource = &onlineEvaluationConfigListResource{}

type onlineEvaluationConfigListResource struct {
	onlineEvaluationConfigResource
	framework.WithList
}

func (l *onlineEvaluationConfigListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BedrockAgentCoreClient(ctx)

	var query listOnlineEvaluationConfigModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input bedrockagentcorecontrol.ListOnlineEvaluationConfigsInput
		for item, err := range listOnlineEvaluationConfigs(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.OnlineEvaluationConfigArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			configID := aws.ToString(item.OnlineEvaluationConfigId)
			output, err := findOnlineEvaluationConfigByID(ctx, conn, configID)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data onlineEvaluationConfigResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, fwflex.Flatten(ctx, output, &data))
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = aws.ToString(item.OnlineEvaluationConfigName)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listOnlineEvaluationConfigModel struct {
	framework.WithRegionModel
}

func listOnlineEvaluationConfigs(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.ListOnlineEvaluationConfigsInput) iter.Seq2[awstypes.OnlineEvaluationConfigSummary, error] {
	return func(yield func(awstypes.OnlineEvaluationConfigSummary, error) bool) {
		pages := bedrockagentcorecontrol.NewListOnlineEvaluationConfigsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.OnlineEvaluationConfigSummary](), fmt.Errorf("listing Bedrock AgentCore Online Evaluation Config resources: %w", err))
				return
			}

			for _, item := range page.OnlineEvaluationConfigs {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
