// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"fmt"
	"iter"
	"strings"

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

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_bedrockagentcore_evaluator")
func newEvaluatorResourceAsListResource() list.ListResourceWithConfigure {
	return &evaluatorListResource{}
}

var _ list.ListResource = &evaluatorListResource{}

type evaluatorListResource struct {
	evaluatorResource
	framework.WithList
}

func (l *evaluatorListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BedrockAgentCoreClient(ctx)

	var query listEvaluatorModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input bedrockagentcorecontrol.ListEvaluatorsInput

		for item, err := range listEvaluators(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			evaluatorID := aws.ToString(item.EvaluatorId)
			if strings.HasPrefix(evaluatorID, "Builtin.") {
				// Skip Built-in evaluators
				continue
			}

			arn := aws.ToString(item.EvaluatorArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)

			var data evaluatorResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				if request.IncludeResource {
					output, err := findEvaluatorByID(ctx, conn, evaluatorID)
					if err != nil {
						smerr.AddError(ctx, &result.Diagnostics, err, smerr.ID, evaluatorID)
						return
					}

					smerr.AddEnrich(ctx, &result.Diagnostics, fwflex.Flatten(ctx, output, &data))
				} else {
					smerr.AddEnrich(ctx, &result.Diagnostics, fwflex.Flatten(ctx, &item, &data))
				}
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = aws.ToString(item.EvaluatorName)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listEvaluatorModel struct {
	framework.WithRegionModel
}

func listEvaluators(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.ListEvaluatorsInput) iter.Seq2[awstypes.EvaluatorSummary, error] {
	return func(yield func(awstypes.EvaluatorSummary, error) bool) {
		pages := bedrockagentcorecontrol.NewListEvaluatorsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.EvaluatorSummary](), fmt.Errorf("listing Bedrock AgentCore Evaluators: %w", err))
				return
			}

			for _, item := range page.Evaluators {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
