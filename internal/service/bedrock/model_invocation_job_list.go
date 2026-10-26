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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_bedrock_model_invocation_job")
func newModelInvocationJobResourceAsListResource() list.ListResourceWithConfigure {
	return &modelInvocationJobListResource{}
}

var _ list.ListResource = &modelInvocationJobListResource{}

type modelInvocationJobListResource struct {
	modelInvocationJobResource
	framework.WithList
}

func (l *modelInvocationJobListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BedrockClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := bedrock.ListModelInvocationJobsInput{}
		for item, err := range listModelInvocationJobs(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.JobArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var job *bedrock.GetModelInvocationJobOutput
			if request.IncludeResource {
				var err error
				job, err = findModelInvocationJobByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data modelInvocationJobResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.JobARN = fwflex.StringToFramework(ctx, item.JobArn)
				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, job, &data)...)
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

func listModelInvocationJobs(ctx context.Context, conn *bedrock.Client, input *bedrock.ListModelInvocationJobsInput) iter.Seq2[awstypes.ModelInvocationJobSummary, error] {
	return func(yield func(awstypes.ModelInvocationJobSummary, error) bool) {
		pages := bedrock.NewListModelInvocationJobsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ModelInvocationJobSummary{}, fmt.Errorf("listing Bedrock Model Invocation Job resources: %w", err))
				return
			}

			for _, item := range page.InvocationJobSummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
