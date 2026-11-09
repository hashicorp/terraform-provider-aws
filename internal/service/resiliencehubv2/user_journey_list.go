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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkListResource("aws_resiliencehubv2_user_journey")
func newUserJourneyResourceAsListResource() list.ListResourceWithConfigure {
	return &userJourneyListResource{}
}

var _ list.ListResource = &userJourneyListResource{}

type userJourneyListResource struct {
	userJourneyResource
	framework.WithList
}

func (l *userJourneyListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"system_arn": listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
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

	systemARN := fwflex.StringValueFromFramework(ctx, query.SystemARN)
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("system_arn"), systemARN)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := resiliencehubv2.ListUserJourneysInput{
			SystemArn: aws.String(systemARN),
		}
		for item, err := range listUserJourneys(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			userJourneyID := aws.ToString(item.UserJourneyId)
			var output *awstypes.UserJourney
			if request.IncludeResource {
				output, err = findUserJourneyByTwoPartKey(ctx, conn, systemARN, userJourneyID)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					result := fwdiag.NewListResultErrorDiagnostic(err)
					yield(result)
					return
				}
			}

			result := request.NewListResult(ctx)

			var data userJourneyResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.SystemARN = fwtypes.ARNValue(systemARN)
				data.UserJourneyID = fwflex.StringValueToFramework(ctx, userJourneyID)

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

type listUserJourneyModel struct {
	framework.WithRegionModel
	SystemARN fwtypes.ARN `tfsdk:"system_arn"`
}

func listUserJourneys(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListUserJourneysInput, optFns ...func(*resiliencehubv2.Options)) iter.Seq2[awstypes.UserJourneySummary, error] {
	return tfiter.ConcatValuesWithError(listUserJourneyPages(ctx, conn, input, optFns...))
}

func listUserJourneyPages(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListUserJourneysInput, optFns ...func(*resiliencehubv2.Options)) iter.Seq2[[]awstypes.UserJourneySummary, error] {
	return func(yield func([]awstypes.UserJourneySummary, error) bool) {
		pages := resiliencehubv2.NewListUserJourneysPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Resilience Hub V2 User Journeys: %w", err))
				return
			}

			if !yield(page.UserJourneySummaries, nil) {
				return
			}
		}
	}
}
