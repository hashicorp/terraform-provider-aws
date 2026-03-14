// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"
	"iter"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
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

	stream.Results = func(yield func(list.ListResult) bool) {
		var input sagemaker.ListTrainingJobsInput

		for item, err := range listTrainingJobs(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			trainingJobName := aws.ToString(item.TrainingJobName)

			result := request.NewListResult(ctx)

			var data resourceTrainingJobModel
			data.TrainingJobName = fwflex.StringValueToFramework(ctx, trainingJobName)
			data.TrainingJobName = fwflex.StringValueToFramework(ctx, trainingJobName)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				if request.IncludeResource {
					trainingJob, err := findTrainingJobByName(ctx, conn, trainingJobName)
					if err != nil {
						result.Diagnostics.Append(diag.NewErrorDiagnostic("Reading SageMaker Training Job", err.Error()))
						return
					}

					result.Diagnostics.Append(fwflex.Flatten(ctx, trainingJob, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.Diagnostics.Append(setZeroAttrValuesToNull(ctx, &data)...)
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

func setZeroAttrValuesToNull(ctx context.Context, target any) diag.Diagnostics {
	var diags diag.Diagnostics

	value := reflect.ValueOf(target)
	if !value.IsValid() || value.Kind() != reflect.Ptr || value.IsNil() {
		return diags
	}

	walkStructSetZeroAttrNull(ctx, value.Elem(), &diags)

	return diags
}

func walkStructSetZeroAttrNull(ctx context.Context, value reflect.Value, diags *diag.Diagnostics) {
	if diags.HasError() || !value.IsValid() || value.Kind() != reflect.Struct {
		return
	}

	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		if !field.CanSet() {
			continue
		}

		if field.Kind() != reflect.Struct {
			continue
		}

		if attrValue, ok := field.Interface().(attr.Value); ok {
			if field.IsZero() {
				nullValue, err := fwtypes.NullValueOf(ctx, attrValue)
				if err != nil {
					diags.AddError("Normalizing List Result", err.Error())
					return
				}

				if nullValue == nil {
					continue
				}

				nullValueReflect := reflect.ValueOf(nullValue)
				switch {
				case nullValueReflect.Type().AssignableTo(field.Type()):
					field.Set(nullValueReflect)
				case nullValueReflect.Type().ConvertibleTo(field.Type()):
					field.Set(nullValueReflect.Convert(field.Type()))
				default:
					diags.AddError("Normalizing List Result", fmt.Sprintf("cannot assign null value of type %T to field type %s", nullValue, field.Type()))
					return
				}
			}

			continue
		}

		walkStructSetZeroAttrNull(ctx, field, diags)
		if diags.HasError() {
			return
		}
	}
}
