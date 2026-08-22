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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
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

	tflog.Info(ctx, "Listing resources")

	input := accountaccess.ListApplicationsInput{}

	stream.Results = func(yield func(list.ListResult) bool) {
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
				app, err = FindApplicationByARN(ctx, conn, arn)
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
				data.ARN = types.StringValue(arn)
				data.ID = data.ARN

				if request.IncludeResource {
					flattenApplication(ctx, app, &data)
				}

				result.DisplayName = arn
			})

			if !yield(result) {
				return
			}
		}
	}
}

// listApplications returns an iterator over all ApplicationSummary items using
// the SDK paginator.
func listApplications(ctx context.Context, conn *accountaccess.Client, input *accountaccess.ListApplicationsInput) iter.Seq2[awstypes.ApplicationSummary, error] {
	return func(yield func(awstypes.ApplicationSummary, error) bool) {
		pages := accountaccess.NewListApplicationsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ApplicationSummary{}, fmt.Errorf("listing Account Access Application resources: %w", err))
				return
			}

			for _, item := range page.Applications {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
