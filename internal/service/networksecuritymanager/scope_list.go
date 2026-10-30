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

// @FrameworkListResource("aws_networksecuritymanager_scope")
func newScopeResourceAsListResource() list.ListResourceWithConfigure {
	return &scopeListResource{}
}

var _ list.ListResource = &scopeListResource{}

type scopeListResource struct {
	scopeResource
	framework.WithList
}

func (l *scopeListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().NetworkSecurityManagerClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input networksecuritymanager.ListScopesInput
		for item, err := range listScopes(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			// A draft is listed with the DRAFT-qualified ARN; the resource is
			// identified by the base ARN, as flattenScope and Read treat it.
			arn := scopeBaseARN(aws.ToString(item.ScopeArn))
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var output *networksecuritymanager.GetScopeOutput
			if request.IncludeResource {
				var err error
				output, err = findScopeByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(smerr.NewListResultError(ctx, err, smerr.ID, arn))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data scopeResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ARN = fwflex.StringValueToFramework(ctx, arn)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, flattenScope(ctx, output, &data), smerr.ID, arn)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.ScopeName)
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

func listScopes(ctx context.Context, conn *networksecuritymanager.Client, input *networksecuritymanager.ListScopesInput, optFns ...func(*networksecuritymanager.Options)) iter.Seq2[awstypes.ScopeSummary, error] {
	return tfiter.ConcatValuesWithError(listScopePages(ctx, conn, input, optFns...))
}

func listScopePages(ctx context.Context, conn *networksecuritymanager.Client, input *networksecuritymanager.ListScopesInput, optFns ...func(*networksecuritymanager.Options)) iter.Seq2[[]awstypes.ScopeSummary, error] {
	return func(yield func([]awstypes.ScopeSummary, error) bool) {
		pages := networksecuritymanager.NewListScopesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, smarterr.NewError(err))
				return
			}

			if !yield(page.Scopes, nil) {
				return
			}
		}
	}
}
