// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkListResource("aws_resiliencehubv2_service_function")
func newServiceFunctionResourceAsListResource() list.ListResourceWithConfigure {
	return &serviceFunctionListResource{}
}

var _ list.ListResource = &serviceFunctionListResource{}

type serviceFunctionListResource struct {
	serviceFunctionResource
	framework.WithList
}

func (l *serviceFunctionListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"service_arn": listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN of the service to list service functions from.",
			},
		},
	}
}

func (l *serviceFunctionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ResilienceHubV2Client(ctx)

	var query listServiceFunctionModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	serviceARN := fwflex.StringValueFromFramework(ctx, query.ServiceARN)
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("service_arn"), serviceARN)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := resiliencehubv2.ListServiceFunctionsInput{
			ServiceArn: aws.String(serviceARN),
		}
		for item, err := range listServiceFunctions(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data serviceFunctionResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, &item, &data))
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = aws.ToString(item.Name)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listServiceFunctionModel struct {
	framework.WithRegionModel
	ServiceARN fwtypes.ARN `tfsdk:"service_arn"`
}

func listServiceFunctions(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListServiceFunctionsInput, optFns ...func(*resiliencehubv2.Options)) iter.Seq2[awstypes.ServiceFunction, error] {
	return tfiter.ConcatValuesWithError(listServiceFunctionPages(ctx, conn, input, optFns...))
}
