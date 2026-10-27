// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
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
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_cloudwatchomni_alert")
func newAlertResourceAsListResource() list.ListResourceWithConfigure {
	return &alertListResource{}
}

var _ list.ListResource = &alertListResource{}

type alertListResource struct {
	alertResource
	framework.WithList
}

func (l *alertListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"space_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the CloudWatch Omni space to list alerts from.",
			},
		},
	}
}

func (l *alertListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().CloudWatchOmniClient(ctx)

	var query alertListModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	spaceID := fwflex.StringValueFromFramework(ctx, query.SpaceID)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := cloudwatchomni.ListAlertsInput{
			SpaceId: aws.String(spaceID),
		}
		for item, err := range listAlerts(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			arn := aws.ToString(item.AlertArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)
			id := aws.ToString(item.AlertId)

			var output *awstypes.Alert
			if request.IncludeResource {
				var err error
				output, err = findAlertByTwoPartKey(ctx, conn, spaceID, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data alertResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.AlertID = fwflex.StringValueToFramework(ctx, id)
				data.SpaceID = fwflex.StringValueToFramework(ctx, spaceID)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data))
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.Name)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type alertListModel struct {
	framework.WithRegionModel
	SpaceID types.String `tfsdk:"space_id"`
}

func listAlerts(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.ListAlertsInput, optFns ...func(*cloudwatchomni.Options)) iter.Seq2[awstypes.AlertSummary, error] {
	return tfiter.ConcatValuesWithError(listAlertPages(ctx, conn, input, optFns...))
}

func listAlertPages(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.ListAlertsInput, optFns ...func(*cloudwatchomni.Options)) iter.Seq2[[]awstypes.AlertSummary, error] {
	return func(yield func([]awstypes.AlertSummary, error) bool) {
		pages := cloudwatchomni.NewListAlertsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing CloudWatch Omni Alerts: %w", err))
				return
			}

			if !yield(page.Items, nil) {
				return
			}
		}
	}
}
