// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package sqs

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_sqs_queue")
func queueResourceAsListResource() inttypes.ListResourceForSDK {
	l := queueListResource{}
	l.SetResourceSchema(resourceQueue())
	return &l
}

type queueListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type queueListResourceModel struct {
	framework.WithRegionModel
}

func (l *queueListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query queueListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	awsClient := l.Meta()
	conn := awsClient.SQSClient(ctx)

	tflog.Info(ctx, "Listing SQS queues")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input sqs.ListQueuesInput
		for queueUrl, err := range listQueues(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), queueUrl)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(queueUrl)

			diags := resourceQueueRead(ctx, rd, awsClient)
			if diags.HasError() || rd.Id() == "" {
				// Resource can't be read or is logically deleted.
				// Log and continue.
				tflog.Error(ctx, "Reading SQS queue", map[string]any{
					names.AttrID: queueUrl,
					"diags":      sdkdiag.DiagnosticsString(diags),
				})
				continue
			}

			result.DisplayName = queueUrl

			l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
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

func listQueues(ctx context.Context, conn *sqs.Client, input *sqs.ListQueuesInput) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		pages := sqs.NewListQueuesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield("", fmt.Errorf("listing SQS Queues: %w", err))
				return
			}

			for _, queueUrl := range page.QueueUrls {
				if !yield(queueUrl, nil) {
					return
				}
			}
		}
	}
}
