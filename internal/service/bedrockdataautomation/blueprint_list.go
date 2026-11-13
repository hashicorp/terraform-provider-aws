// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockdataautomation

import (
	"context"
	"iter"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockdataautomation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockdataautomation/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_bedrockdataautomation_blueprint")
func newBlueprintResourceAsListResource() list.ListResourceWithConfigure {
	return &blueprintListResource{}
}

var _ list.ListResource = &blueprintListResource{}

type blueprintListResource struct {
	blueprintResource
	framework.WithList
}

func (l *blueprintListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BedrockDataAutomationClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		// Service-owned blueprints can't be managed, so list only the account's own.
		input := bedrockdataautomation.ListBlueprintsInput{
			ResourceOwner: awstypes.ResourceOwnerAccount,
		}
		for item, err := range listBlueprints(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			arn := aws.ToString(item.BlueprintArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var blueprint *awstypes.Blueprint
			if request.IncludeResource {
				var err error
				blueprint, err = findBlueprintByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(smerr.NewListResultError(ctx, err, smerr.ID, arn))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data blueprintResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ARN = fwflex.StringValueToFramework(ctx, arn)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, blueprint, &data), smerr.ID, arn)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.BlueprintName)
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

func listBlueprints(ctx context.Context, conn *bedrockdataautomation.Client, input *bedrockdataautomation.ListBlueprintsInput) iter.Seq2[awstypes.BlueprintSummary, error] {
	return tfiter.ConcatValuesWithError(listBlueprintPages(ctx, conn, input))
}

func listBlueprintPages(ctx context.Context, conn *bedrockdataautomation.Client, input *bedrockdataautomation.ListBlueprintsInput) iter.Seq2[[]awstypes.BlueprintSummary, error] {
	return func(yield func([]awstypes.BlueprintSummary, error) bool) {
		pages := bedrockdataautomation.NewListBlueprintsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(nil, smarterr.NewError(err))
				return
			}

			if !yield(page.Blueprints, nil) {
				return
			}
		}
	}
}
