// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_s3control_multi_region_access_point")
func newMultiRegionAccessPointResourceAsListResource() inttypes.ListResourceForSDK {
	l := multiRegionAccessPointListResource{}
	l.SetResourceSchema(resourceMultiRegionAccessPoint())
	return &l
}

var _ list.ListResource = &multiRegionAccessPointListResource{}

type multiRegionAccessPointListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type multiRegionAccessPointListResourceModel struct {
	framework.WithRegionModel
}

func (l *multiRegionAccessPointListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query multiRegionAccessPointListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	awsClient := l.Meta()
	conn := awsClient.S3ControlClient(ctx)
	accountID := awsClient.AccountID(ctx)

	tflog.Info(ctx, "Listing S3 Multi-Region Access Points")

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
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			id := multiRegionAccessPointCreateResourceID(accountID, name)
			rd.SetId(id)
			// Set the identity attribute.
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				if err := resourceMultiRegionAccessPointFlatten(ctx, awsClient, accountID, &item, rd); err != nil {
					tflog.Error(ctx, "Flattening S3 Multi-Region Access Point", map[string]any{
						"error":      err.Error(),
						names.AttrID: id,
					})
					continue
				}
			}

			result.DisplayName = name

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
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

func listMultiRegionAccessPointsIterator(ctx context.Context, conn *s3control.Client, input *s3control.ListMultiRegionAccessPointsInput) iter.Seq2[awstypes.MultiRegionAccessPointReport, error] {
	return func(yield func(awstypes.MultiRegionAccessPointReport, error) bool) {
		pages := s3control.NewListMultiRegionAccessPointsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, func(o *s3control.Options) {
				// All Multi-Region Access Point actions are routed to the US West (Oregon) Region.
				o.Region = endpoints.UsWest2RegionID
			})
			if err != nil {
				yield(awstypes.MultiRegionAccessPointReport{}, fmt.Errorf("listing S3 Multi-Region Access Points: %w", err))
				return
			}

			for _, item := range page.AccessPoints {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
