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
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_sagemaker_training_job")
func newTrainingJobResourceAsListResource() list.ListResourceWithConfigure {
	return &trainingJobListResource{}
}

var _ list.ListResource = &trainingJobListResource{}

type trainingJobListResource struct {
	resourceTrainingJob
	framework.WithList
}

func (l *trainingJobListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SageMakerClient(ctx)

	var query listTrainingJobModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing SageMaker Training Jobs")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input sagemaker.ListTrainingJobsInput

		for item, err := range listTrainingJobs(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.TrainingJobArn)
			trainingJobName := aws.ToString(item.TrainingJobName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)
			trainingJob, err := findTrainingJobByName(ctx, conn, trainingJobName)
			if err != nil {
				result.Diagnostics.Append(diag.NewErrorDiagnostic("Reading SageMaker Training Job", err.Error()))
				yield(result)
				return
			}

			var data resourceTrainingJobModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, trainingJob, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = trainingJobName
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

type listTrainingJobModel struct {
	framework.WithRegionModel
}

func listTrainingJobs(ctx context.Context, conn *sagemaker.Client, input *sagemaker.ListTrainingJobsInput) iter.Seq2[awstypes.TrainingJobSummary, error] {
	return func(yield func(awstypes.TrainingJobSummary, error) bool) {
		pages := sagemaker.NewListTrainingJobsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.TrainingJobSummary{}, fmt.Errorf("listing SageMaker Training Job resources: %w", err))
				return
			}

			for _, item := range page.TrainingJobSummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
