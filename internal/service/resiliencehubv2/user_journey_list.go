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
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkListResource("aws_resiliencehubv2_user_journey")
func newResourceUserJourneyAsListResource() list.ListResourceWithConfigure {
	return &userJourneyListResource{}
}

var _ list.ListResource = &userJourneyListResource{}

type userJourneyListResource struct {
	resourceUserJourney
	framework.WithList
}

func (l *userJourneyListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"system_arn": listschema.StringAttribute{
				Required:    true,
				Description: "ARN of the system to list user journeys from.",
			},
		},
	}
}

func (l *userJourneyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ResilienceHubV2Client(ctx)

	var query listUserJourneyModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	systemArn := query.SystemArn.ValueString()
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("system_arn"), systemArn)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := resiliencehubv2.ListUserJourneysInput{
			SystemArn: aws.String(systemArn),
		}
		for item, err := range listUserJourneys(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			output, err := findUserJourneyByID(ctx, conn, systemArn, aws.ToString(item.UserJourneyId))
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data resourceUserJourneyModel
			data.SystemArn = types.StringValue(systemArn)
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data))
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

type listUserJourneyModel struct {
	framework.WithRegionModel
	SystemArn types.String `tfsdk:"system_arn"`
}

func listUserJourneys(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListUserJourneysInput) iter.Seq2[awstypes.UserJourneySummary, error] {
	return func(yield func(awstypes.UserJourneySummary, error) bool) {
		pages := resiliencehubv2.NewListUserJourneysPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.UserJourneySummary{}, fmt.Errorf("listing Resilience Hub V2 User Journey resources: %w", err))
				return
			}

			for _, item := range page.UserJourneySummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
