// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_sagemaker_algorithm")
func newAlgorithmResourceAsListResource() list.ListResourceWithConfigure {
	return &algorithmListResource{}
}

var _ list.ListResource = &algorithmListResource{}

type algorithmListResource struct {
	algorithmResource
	framework.WithList
}

func (l *algorithmListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.SageMakerClient(ctx)

	var query listAlgorithmModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing SageMaker Algorithm resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input sagemaker.ListAlgorithmsInput

		for item, err := range listAlgorithms(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			algorithmName := aws.ToString(item.AlgorithmName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("algorithm_name"), algorithmName)

			result := request.NewListResult(ctx)

			output, err := findAlgorithmByName(ctx, conn, algorithmName)
			if retry.NotFound(err) {
				tflog.Warn(ctx, "Resource disappeared during listing, skipping")
				continue
			}
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			var data algorithmResourceModel
			l.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, output, &data, nil)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = algorithmName
			})

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

type listAlgorithmModel struct {
	framework.WithRegionModel
}

func listAlgorithms(ctx context.Context, conn *sagemaker.Client, input *sagemaker.ListAlgorithmsInput) iter.Seq2[awstypes.AlgorithmSummary, error] {
	return func(yield func(awstypes.AlgorithmSummary, error) bool) {
		pages := sagemaker.NewListAlgorithmsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.AlgorithmSummary{}, fmt.Errorf("listing SageMaker Algorithm resources: %w", err))
				return
			}

			for _, item := range page.AlgorithmSummaryList {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
