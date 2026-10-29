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

// @FrameworkListResource("aws_bedrockdataautomation_project")
func newProjectResourceAsListResource() list.ListResourceWithConfigure {
	return &projectListResource{}
}

var _ list.ListResource = &projectListResource{}

type projectListResource struct {
	projectResource
	framework.WithList
}

func (l *projectListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BedrockDataAutomationClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		// Service-owned projects can't be managed, so list only the account's own.
		input := bedrockdataautomation.ListDataAutomationProjectsInput{
			ResourceOwner: awstypes.ResourceOwnerAccount,
		}
		for item, err := range listProjects(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			arn := aws.ToString(item.ProjectArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var project *awstypes.DataAutomationProject
			if request.IncludeResource {
				var err error
				project, err = findProjectByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(smerr.NewListResultError(ctx, err, smerr.ID, arn))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data projectResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ARN = fwflex.StringValueToFramework(ctx, arn)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, project, &data), smerr.ID, arn)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.ProjectName)
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

func listProjects(ctx context.Context, conn *bedrockdataautomation.Client, input *bedrockdataautomation.ListDataAutomationProjectsInput) iter.Seq2[awstypes.DataAutomationProjectSummary, error] {
	return tfiter.ConcatValuesWithError(listProjectPages(ctx, conn, input))
}

func listProjectPages(ctx context.Context, conn *bedrockdataautomation.Client, input *bedrockdataautomation.ListDataAutomationProjectsInput) iter.Seq2[[]awstypes.DataAutomationProjectSummary, error] {
	return func(yield func([]awstypes.DataAutomationProjectSummary, error) bool) {
		pages := bedrockdataautomation.NewListDataAutomationProjectsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(nil, smarterr.NewError(err))
				return
			}

			if !yield(page.Projects, nil) {
				return
			}
		}
	}
}
