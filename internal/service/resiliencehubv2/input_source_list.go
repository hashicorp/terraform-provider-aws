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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_resiliencehubv2_input_source")
func newInputSourceResourceAsListResource() list.ListResourceWithConfigure {
	return &inputSourceListResource{}
}

var _ list.ListResource = &inputSourceListResource{}

type inputSourceListResource struct {
	inputSourceResource
	framework.WithList
}

func (l *inputSourceListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"service_arn": listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN of the service to list input sources from.",
			},
			names.AttrType: listschema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.InputSourceType](),
				Optional:    true,
				Description: "Filter input sources by type.",
			},
		},
	}
}

func (l *inputSourceListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ResilienceHubV2Client(ctx)

	var query listInputSourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	serviceARN := fwflex.StringValueFromFramework(ctx, query.ServiceARN)
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("service_arn"), serviceARN)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := resiliencehubv2.ListInputSourcesInput{
			ServiceArn: aws.String(serviceARN),
		}
		if !query.Type.IsNull() {
			input.Type = query.Type.ValueEnum()
		}

		for item, err := range listInputSources(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data inputSourceResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ServiceARN = fwtypes.ARNValue(serviceARN)

				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, &item, &data))
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = aws.ToString(item.InputSourceId)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listInputSourceModel struct {
	framework.WithRegionModel
	ServiceARN fwtypes.ARN                                  `tfsdk:"service_arn"`
	Type       fwtypes.StringEnum[awstypes.InputSourceType] `tfsdk:"type"`
}

func listInputSources(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListInputSourcesInput, optFns ...func(*resiliencehubv2.Options)) iter.Seq2[awstypes.InputSourceSummary, error] {
	return tfiter.ConcatValuesWithError(listInputSourcePages(ctx, conn, input, optFns...))
}
