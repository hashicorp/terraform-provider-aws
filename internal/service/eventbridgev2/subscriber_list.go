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

// @FrameworkListResource("aws_eventbridgev2_subscriber")
func newSubscriberResourceAsListResource() list.ListResourceWithConfigure {
	return &subscriberListResource{}
}

var _ list.ListResource = &subscriberListResource{}

type subscriberListResource struct {
	subscriberResource
	framework.WithList
}

func (l *subscriberListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"event_bus_arn": listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Optional:    true,
				Description: "ARN of the event bus whose subscribers are listed.",
			},
			names.AttrNamePrefix: listschema.StringAttribute{
				Optional:    true,
				Description: "Filter to list only subscribers whose name begins with the specified prefix.",
			},
		},
	}
}

func (l *subscriberListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EventBridgeV2Client(ctx)

	var query listSubscriberModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input eventbridgev2.ListSubscribersInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	accountID := l.Meta().AccountID(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listSubscribers(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			// Subscribers created by other accounts on a shared bus cannot be managed by the bus owner.
			if v := aws.ToString(item.SubscriberAccountId); v != "" && v != accountID {
				continue
			}

			arn := aws.ToString(item.SubscriberArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var output *eventbridgev2.DescribeSubscriberOutput
			if request.IncludeResource {
				output, err = findSubscriberByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data subscriberResourceModel
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

type listSubscriberModel struct {
	framework.WithRegionModel
	EventBusARN fwtypes.ARN  `tfsdk:"event_bus_arn"`
	NamePrefix  types.String `tfsdk:"name_prefix"`
}

func listSubscribers(ctx context.Context, conn *eventbridgev2.Client, input *eventbridgev2.ListSubscribersInput) iter.Seq2[awstypes.SubscriberSummary, error] {
	return tfiter.ConcatValuesWithError(listSubscriberPages(ctx, conn, input))
}

func listSubscriberPages(ctx context.Context, conn *eventbridgev2.Client, input *eventbridgev2.ListSubscribersInput) iter.Seq2[[]awstypes.SubscriberSummary, error] {
	return func(yield func([]awstypes.SubscriberSummary, error) bool) {
		pages := eventbridgev2.NewListSubscribersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(nil, fmt.Errorf("listing EventBridge V2 Subscribers: %w", err))
				return
			}

			if !yield(page.Subscribers, nil) {
				return
			}
		}
	}
}
