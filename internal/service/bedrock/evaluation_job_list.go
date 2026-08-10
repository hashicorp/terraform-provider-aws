// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_bedrock_evaluation_job")
func newEvaluationJobResourceAsListResource() list.ListResourceWithConfigure {
	return &evaluationJobListResource{}
}

var _ list.ListResource = &evaluationJobListResource{}

type evaluationJobListResource struct {
	evaluationJobResource
	framework.WithList
}

func (l *evaluationJobListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BedrockClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := bedrock.ListEvaluationJobsInput{}
		for item, err := range listEvaluationJobs(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.JobArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)

			var data evaluationJobResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.JobARN = fwflex.StringToFramework(ctx, item.JobArn)
				if request.IncludeResource {
					job, err := findEvaluationJobByARN(ctx, conn, arn)
					if err != nil {
						result.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
						return
					}
					result.Diagnostics.Append(fwflex.Flatten(ctx, job, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.JobName)
			})

			if !yield(result) {
				return
			}
		}
	}
}

func listEvaluationJobs(ctx context.Context, conn *bedrock.Client, input *bedrock.ListEvaluationJobsInput) iter.Seq2[awstypes.EvaluationSummary, error] {
	return func(yield func(awstypes.EvaluationSummary, error) bool) {
		pages := bedrock.NewListEvaluationJobsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.EvaluationSummary{}, fmt.Errorf("listing Bedrock Evaluation Job resources: %w", err))
				return
			}

			for _, item := range page.JobSummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
