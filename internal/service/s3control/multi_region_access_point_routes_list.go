// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// @FrameworkListResource("aws_s3control_multi_region_access_point_routes")
func newMultiRegionAccessPointRoutesResourceAsListResource() list.ListResourceWithConfigure {
	return &multiRegionAccessPointRoutesListResource{}
}

var _ list.ListResource = &multiRegionAccessPointRoutesListResource{}

type multiRegionAccessPointRoutesListResource struct {
	multiRegionAccessPointRoutesResource
	framework.WithList
}

type multiRegionAccessPointRoutesListModel struct {
	framework.WithRegionModel
}

func (l *multiRegionAccessPointRoutesListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query multiRegionAccessPointRoutesListModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	awsClient := l.Meta()
	conn := awsClient.S3ControlClient(ctx)
	accountID := awsClient.AccountID(ctx)

	tflog.Info(ctx, "Listing S3 Multi-Region Access Point Routes")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := s3control.ListMultiRegionAccessPointsInput{
			AccountId: aws.String(accountID),
		}

		for item, err := range listMultiRegionAccessPointsIterator(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(item.Name)
			alias := aws.ToString(item.Alias)
			mrapARN := multiRegionAccessPointARN(ctx, awsClient, alias)

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("mrap"), mrapARN)

			result := request.NewListResult(ctx)

			var data multiRegionAccessPointRoutesResourceModel
			l.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				data.AccountID = fwflex.StringValueToFramework(ctx, accountID)
				data.Mrap = fwflex.StringValueToFramework(ctx, mrapARN)

				if request.IncludeResource {
					output, err := findMultiRegionAccessPointRoutesByTwoPartKey(ctx, conn, accountID, mrapARN)
					if retry.NotFound(err) {
						tflog.Warn(ctx, "S3 Multi-Region Access Point Routes not found", map[string]any{
							logging.ResourceAttributeKey("mrap"): mrapARN,
						})
						return
					}
					if err != nil {
						result.Diagnostics.AddError(fmt.Sprintf("reading S3 Multi-Region Access Point Routes (%s)", mrapARN), err.Error())
						return
					}

					result.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}
			})

			result.DisplayName = name

			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}
