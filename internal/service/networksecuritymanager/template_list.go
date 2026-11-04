// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"
	"iter"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
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

// @FrameworkListResource("aws_networksecuritymanager_template")
func newTemplateResourceAsListResource() list.ListResourceWithConfigure {
	return &templateListResource{}
}

var _ list.ListResource = &templateListResource{}

type templateListResource struct {
	templateResource
	framework.WithList
}

func (l *templateListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().NetworkSecurityManagerClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input networksecuritymanager.ListTemplatesInput
		for item, err := range listTemplates(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			// A draft is listed with its DRAFT-qualified ARN.
			arn := templateBaseARN(aws.ToString(item.TemplateArn))
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var template *networksecuritymanager.GetTemplateOutput
			if request.IncludeResource {
				var err error
				template, err = findTemplateByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(smerr.NewListResultError(ctx, err, smerr.ID, arn))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data templateResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ARN = fwflex.StringValueToFramework(ctx, arn)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, template, &data), smerr.ID, arn)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.TemplateName)
			})

			if result.Diagnostics.HasError() {
				// The template can be deleted after it is listed, which fails reading its tags.
				if request.IncludeResource {
					if _, err := findTemplateByARN(ctx, conn, arn); retry.NotFound(err) {
						continue
					}
				}

				yield(list.ListResult{Diagnostics: result.Diagnostics})
				return
			}
			if !yield(result) {
				return
			}
		}
	}
}

func listTemplates(ctx context.Context, conn *networksecuritymanager.Client, input *networksecuritymanager.ListTemplatesInput) iter.Seq2[awstypes.TemplateSummary, error] {
	return tfiter.ConcatValuesWithError(listTemplatePages(ctx, conn, input))
}

func listTemplatePages(ctx context.Context, conn *networksecuritymanager.Client, input *networksecuritymanager.ListTemplatesInput) iter.Seq2[[]awstypes.TemplateSummary, error] {
	return func(yield func([]awstypes.TemplateSummary, error) bool) {
		pages := networksecuritymanager.NewListTemplatesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(nil, smarterr.NewError(err))
				return
			}

			if !yield(page.Templates, nil) {
				return
			}
		}
	}
}
