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
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_eventbridgev2_event_bus")
func newEventBusResourceAsListResource() list.ListResourceWithConfigure {
	return &eventBusListResource{}
}

var _ list.ListResource = &eventBusListResource{}

type eventBusListResource struct {
	eventBusResource
	framework.WithList
}

func (l *eventBusListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrNamePrefix: listschema.StringAttribute{
				Optional:    true,
				Description: "Filter to list only event buses whose name begins with the specified prefix.",
			},
		},
	}
}

func (l *eventBusListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EventBridgeV2Client(ctx)

	var query listEventBusModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	// Only list event buses owned by the caller; buses shared from other accounts cannot be managed.
	input := eventbridgev2.ListEventBusesInput{
		EventBusAccountId: aws.String(l.Meta().AccountID(ctx)),
		NamePrefix:        fwflex.StringFromFramework(ctx, query.NamePrefix),
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listEventBuses(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			arn := aws.ToString(item.EventBusArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var output *eventbridgev2.DescribeEventBusOutput
			if request.IncludeResource {
				output, err = findEventBusByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data eventBusResourceModel
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

type listEventBusModel struct {
	framework.WithRegionModel
	NamePrefix types.String `tfsdk:"name_prefix"`
}

func listEventBuses(ctx context.Context, conn *eventbridgev2.Client, input *eventbridgev2.ListEventBusesInput) iter.Seq2[awstypes.EventBusSummary, error] {
	return tfiter.ConcatValuesWithError(listEventBusPages(ctx, conn, input))
}

func listEventBusPages(ctx context.Context, conn *eventbridgev2.Client, input *eventbridgev2.ListEventBusesInput) iter.Seq2[[]awstypes.EventBusSummary, error] {
	return func(yield func([]awstypes.EventBusSummary, error) bool) {
		pages := eventbridgev2.NewListEventBusesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(nil, fmt.Errorf("listing EventBridge V2 Event Buses: %w", err))
				return
			}

			if !yield(page.EventBuses, nil) {
				return
			}
		}
	}
}
