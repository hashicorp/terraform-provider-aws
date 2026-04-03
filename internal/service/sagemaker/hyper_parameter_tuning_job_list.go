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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_sagemaker_hyper_parameter_tuning_job")
func newHyperParameterTuningJobResourceAsListResource() list.ListResourceWithConfigure {
	return &hyperParameterTuningJobListResource{}
}

var _ list.ListResource = &hyperParameterTuningJobListResource{}

type hyperParameterTuningJobListResource struct {
	hyperParameterTuningJobResource
	framework.WithList
}

func (l *hyperParameterTuningJobListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.SageMakerClient(ctx)

	var query listHyperParameterTuningJobModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing SageMaker Hyper Parameter Tuning Job resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input sagemaker.ListHyperParameterTuningJobsInput

		for item, err := range listHyperParameterTuningJobs(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			hyperParameterTuningJobName := aws.ToString(item.HyperParameterTuningJobName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("name"), hyperParameterTuningJobName)

			result := request.NewListResult(ctx)

			var data hyperParameterTuningJobResourceModel
			data.Name = fwflex.StringValueToFramework(ctx, hyperParameterTuningJobName)

			l.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				if request.IncludeResource {
					output, err := findHyperParameterTuningJobByName(ctx, conn, hyperParameterTuningJobName)
					if retry.NotFound(err) {
						tflog.Warn(ctx, "Resource disappeared during listing, skipping")
						return
					}
					if err != nil {
						result.Diagnostics.Append(diag.NewErrorDiagnostic("Reading SageMaker Hyper Parameter Tuning Job", err.Error()))
						return
					}

					l.flatten(ctx, output, &data, &result.Diagnostics)
				}

				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = hyperParameterTuningJobName
			})

			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				break
			}
		}
	}
}

type listHyperParameterTuningJobModel struct {
	framework.WithRegionModel
}

func listHyperParameterTuningJobs(ctx context.Context, conn *sagemaker.Client, input *sagemaker.ListHyperParameterTuningJobsInput) iter.Seq2[awstypes.HyperParameterTuningJobSummary, error] {
	return func(yield func(awstypes.HyperParameterTuningJobSummary, error) bool) {
		pages := sagemaker.NewListHyperParameterTuningJobsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.HyperParameterTuningJobSummary{}, fmt.Errorf("listing SageMaker Hyper Parameter Tuning Job resources: %w", err))
				return
			}

			for _, item := range page.HyperParameterTuningJobSummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
