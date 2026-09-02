// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
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

// @FrameworkListResource("aws_accountaccess_application")
func newApplicationResourceAsListResource() list.ListResourceWithConfigure {
	return &applicationListResource{}
}

var _ list.ListResource = &applicationListResource{}

type applicationListResource struct {
	applicationResource
	framework.WithList
}

func (l *applicationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().AccountAccessClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input accountaccess.ListApplicationsInput
		for item, err := range listApplications(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.ApplicationArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var app *accountaccess.GetApplicationOutput
			if request.IncludeResource {
				app, err = findApplicationByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)
			var data applicationResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ApplicationARN = fwflex.StringValueToFramework(ctx, arn)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, app, &data))
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = arn
			})

			if !yield(result) {
				return
			}
		}
	}
}

func listApplications(ctx context.Context, conn *accountaccess.Client, input *accountaccess.ListApplicationsInput, optFns ...func(*accountaccess.Options)) iter.Seq2[awstypes.ApplicationSummary, error] {
	return tfiter.ConcatValuesWithError(listApplicationPages(ctx, conn, input, optFns...))
}

func listApplicationPages(ctx context.Context, conn *accountaccess.Client, input *accountaccess.ListApplicationsInput, optFns ...func(*accountaccess.Options)) iter.Seq2[[]awstypes.ApplicationSummary, error] {
	return func(yield func([]awstypes.ApplicationSummary, error) bool) {
		pages := accountaccess.NewListApplicationsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Account Access Applications: %w", err))
				return
			}

			if !yield(page.Applications, nil) {
				return
			}
		}
	}
}
