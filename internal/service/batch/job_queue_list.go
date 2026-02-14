// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/service/batch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// @FrameworkListResource("aws_batch_job_queue")
func newJobQueueResourceAsListResource() list.ListResourceWithConfigure {
	return &listResourceJobQueue{}
}

var _ list.ListResource = &listResourceJobQueue{}

type listResourceJobQueue struct {
	jobQueueResource
	framework.WithList
}

func (r *listResourceJobQueue) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query jobQueueListModel

	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	awsClient := r.Meta()
	conn := awsClient.BatchClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		var input batch.DescribeJobQueuesInput
		for jobQueue, err := range listJobQueues(ctx, conn, &input) {
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			var data jobQueueResourceModel
			r.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				if diags := fwflex.Flatten(ctx, jobQueue, &data, fwflex.WithFieldNamePrefix("JobQueue")); diags.HasError() {
					result.Diagnostics.Append(diags...)
					yield(result)
					return
				}

				setTagsOut(ctx, jobQueue.Tags)
				result.DisplayName = data.JobQueueName.ValueString()
			})

			if result.Diagnostics.HasError() {
				result = list.ListResult{Diagnostics: result.Diagnostics}
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type jobQueueListModel struct {
	framework.WithRegionModel
}

// DescribeJobQueues is an "All-Or-Some" call.
func listJobQueues(ctx context.Context, conn *batch.Client, input *batch.DescribeJobQueuesInput) iter.Seq2[awstypes.JobQueueDetail, error] {
	return func(yield func(awstypes.JobQueueDetail, error) bool) {
		pages := batch.NewDescribeJobQueuesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.JobQueueDetail{}, fmt.Errorf("listing Batch Job Queues: %w", err))
				return
			}

			for _, jobQueue := range page.JobQueues {
				if !yield(jobQueue, nil) {
					return
				}
			}
		}
	}
}
