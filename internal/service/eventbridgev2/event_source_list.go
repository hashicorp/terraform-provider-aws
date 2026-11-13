// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eventbridgev2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridgev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eventbridgev2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_eventbridgev2_event_source")
func newEventSourceResourceAsListResource() list.ListResourceWithConfigure {
	return &eventSourceListResource{}
}

var _ list.ListResource = &eventSourceListResource{}

type eventSourceListResource struct {
	eventSourceResource
	framework.WithList
}

func (l *eventSourceListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"event_bus_arn": listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Optional:    true,
				Description: "ARN of the event bus whose event sources are listed.",
			},
			names.AttrNamePrefix: listschema.StringAttribute{
				Optional:    true,
				Description: "Filter to list only event sources whose name begins with the specified prefix.",
			},
		},
	}
}

func (l *eventSourceListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EventBridgeV2Client(ctx)

	var query listEventSourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input eventbridgev2.ListEventSourcesInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	accountID := l.Meta().AccountID(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listEventSources(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			// Event sources created by other accounts on a shared bus cannot be managed by the bus owner.
			if v := aws.ToString(item.EventSourceAccountId); v != "" && v != accountID {
				continue
			}

			arn := aws.ToString(item.EventSourceArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var output *eventbridgev2.DescribeEventSourceOutput
			if request.IncludeResource {
				output, err = findEventSourceByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data eventSourceResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ARN = fwflex.StringValueToFramework(ctx, arn)

				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, output, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.Name)
			})

			if result.Diagnostics.HasError() {
				yield(list.ListResult{Diagnostics: result.Diagnostics})
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listEventSourceModel struct {
	framework.WithRegionModel
	EventBusARN fwtypes.ARN  `tfsdk:"event_bus_arn"`
	NamePrefix  types.String `tfsdk:"name_prefix"`
}

func listEventSources(ctx context.Context, conn *eventbridgev2.Client, input *eventbridgev2.ListEventSourcesInput) iter.Seq2[awstypes.EventSourceSummary, error] {
	return tfiter.ConcatValuesWithError(listEventSourcePages(ctx, conn, input))
}

func listEventSourcePages(ctx context.Context, conn *eventbridgev2.Client, input *eventbridgev2.ListEventSourcesInput) iter.Seq2[[]awstypes.EventSourceSummary, error] {
	return func(yield func([]awstypes.EventSourceSummary, error) bool) {
		pages := eventbridgev2.NewListEventSourcesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(nil, fmt.Errorf("listing EventBridge V2 Event Sources: %w", err))
				return
			}

			if !yield(page.EventSources, nil) {
				return
			}
		}
	}
}
