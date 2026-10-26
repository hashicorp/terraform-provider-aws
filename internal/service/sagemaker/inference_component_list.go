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
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_sagemaker_inference_component")
func newInferenceComponentResourceAsListResource() list.ListResourceWithConfigure {
	return &inferenceComponentListResource{}
}

var _ list.ListResource = &inferenceComponentListResource{}

type inferenceComponentListResource struct {
	inferenceComponentResource
	framework.WithList
}

func (l *inferenceComponentListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.SageMakerClient(ctx)

	tflog.Info(ctx, "Listing SageMaker Inference Component resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input sagemaker.ListInferenceComponentsInput

		for item, err := range listInferenceComponents(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			name := aws.ToString(item.InferenceComponentName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			// The summary carries the identity attribute, so the component is described
			// only when the full resource object is requested.
			var output *sagemaker.DescribeInferenceComponentOutput
			if request.IncludeResource {
				var err error
				output, err = findInferenceComponentByName(ctx, conn, name)
				if retry.NotFound(err) {
					tflog.Warn(ctx, "Resource disappeared during listing, skipping")
					continue
				}
				if err != nil {
					yield(smerr.NewListResultError(ctx, err, smerr.ID, name))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data inferenceComponentResourceModel
			l.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				data.InferenceComponentName = fwflex.StringValueToFramework(ctx, name)

				if request.IncludeResource {
					// There is no prior state to correlate against, so the omitted-block and
					// empty-environment preservation in flatten is inert here and the model is
					// populated entirely from the API, exactly as it is on import.
					l.flatten(ctx, output, &data, &result.Diagnostics)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = name
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

func listInferenceComponents(ctx context.Context, conn *sagemaker.Client, input *sagemaker.ListInferenceComponentsInput) iter.Seq2[awstypes.InferenceComponentSummary, error] {
	return func(yield func(awstypes.InferenceComponentSummary, error) bool) {
		pages := sagemaker.NewListInferenceComponentsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.InferenceComponentSummary{}, fmt.Errorf("listing SageMaker Inference Component resources: %w", err))
				return
			}

			for _, item := range page.InferenceComponents {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
