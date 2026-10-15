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

// @FrameworkListResource("aws_cloudwatchomni_dashboard")
func newDashboardResourceAsListResource() list.ListResourceWithConfigure {
	return &dashboardListResource{}
}

var _ list.ListResource = &dashboardListResource{}

type dashboardListResource struct {
	dashboardResource
	framework.WithList
}

func (l *dashboardListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"space_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the CloudWatch Omni space to list dashboards from.",
			},
		},
	}
}

func (l *dashboardListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().CloudWatchOmniClient(ctx)

	var query dashboardListModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	spaceID := fwflex.StringValueFromFramework(ctx, query.SpaceID)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := cloudwatchomni.ListOmniDashboardsInput{
			SpaceId: aws.String(spaceID),
		}
		for item, err := range listDashboards(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			arn := aws.ToString(item.Arn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)
			id := aws.ToString(item.DashboardId)

			var output *awstypes.OmniDashboard
			if request.IncludeResource {
				var err error
				output, err = findDashboardByTwoPartKey(ctx, conn, spaceID, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data dashboardResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.DashboardID = fwflex.StringValueToFramework(ctx, id)
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

type dashboardListModel struct {
	framework.WithRegionModel
	SpaceID types.String `tfsdk:"space_id"`
}

func listDashboards(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.ListOmniDashboardsInput, optFns ...func(*cloudwatchomni.Options)) iter.Seq2[awstypes.OmniDashboardSummary, error] {
	return tfiter.ConcatValuesWithError(listDashboardPages(ctx, conn, input, optFns...))
}

func listDashboardPages(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.ListOmniDashboardsInput, optFns ...func(*cloudwatchomni.Options)) iter.Seq2[[]awstypes.OmniDashboardSummary, error] {
	return func(yield func([]awstypes.OmniDashboardSummary, error) bool) {
		pages := cloudwatchomni.NewListOmniDashboardsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing CloudWatch Omni Dashboards: %w", err))
				return
			}

			if !yield(page.Items, nil) {
				return
			}
		}
	}
}
