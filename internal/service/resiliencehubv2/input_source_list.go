// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
)

// @FrameworkListResource("aws_resiliencehubv2_input_source")
func newResourceInputSourceAsListResource() list.ListResourceWithConfigure {
	return &inputSourceListResource{}
}

var _ list.ListResource = &inputSourceListResource{}

type inputSourceListResource struct {
	resourceInputSource
	framework.WithList
}

func (l *inputSourceListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"service_arn": listschema.StringAttribute{
				Required:    true,
				Description: "ARN of the service to list input sources from.",
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

	serviceArn := query.ServiceArn.ValueString()
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("service_arn"), serviceArn)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := resiliencehubv2.ListInputSourcesInput{
			ServiceArn: aws.String(serviceArn),
		}
		for item, err := range listInputSources(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data resourceInputSourceModel
			data.ServiceArn = types.StringValue(serviceArn)
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				l.flatten(&item, &data)

				data.ID = types.StringValue(serviceArn + "," + data.InputSourceId.ValueString())
				result.DisplayName = data.InputSourceId.ValueString()
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listInputSourceModel struct {
	framework.WithRegionModel
	ServiceArn types.String `tfsdk:"service_arn"`
}

func listInputSources(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListInputSourcesInput) iter.Seq2[awstypes.InputSourceSummary, error] {
	return func(yield func(awstypes.InputSourceSummary, error) bool) {
		pages := resiliencehubv2.NewListInputSourcesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.InputSourceSummary{}, fmt.Errorf("listing Resilience Hub V2 Input Source resources: %w", err))
				return
			}

			for _, item := range page.InputSourceSummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
