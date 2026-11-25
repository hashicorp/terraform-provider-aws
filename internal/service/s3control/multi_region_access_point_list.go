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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_s3control_multi_region_access_point")
func newMultiRegionAccessPointResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceMultiRegionAccessPoint(), framework.SDKListSpec[multiRegionAccessPointItem, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[multiRegionAccessPointItem, error], diag.Diagnostics) {
			accountID := c.AccountID(ctx)
			input := s3control.ListMultiRegionAccessPointsInput{
				AccountId: aws.String(accountID),
			}

			reports := listMultiRegionAccessPointsIterator(ctx, c.S3ControlClient(ctx), &input)

			return func(yield func(multiRegionAccessPointItem, error) bool) {
				for report, err := range reports {
					if !yield(multiRegionAccessPointItem{report: report, accountID: accountID}, err) {
						return
					}
				}
			}, nil
		},
		ID: func(item multiRegionAccessPointItem, _ framework.RegionQuery) string {
			return multiRegionAccessPointCreateResourceID(item.accountID, aws.ToString(item.report.Name))
		},
		Flatten: func(ctx context.Context, c *conns.AWSClient, item *multiRegionAccessPointItem, rd *schema.ResourceData) error {
			return resourceMultiRegionAccessPointFlatten(ctx, c, &item.report, rd)
		},
		DisplayName: func(item multiRegionAccessPointItem) string { return aws.ToString(item.report.Name) },
	})
}

// The resource ID embeds the account ID, which is listing scope rather than
// per-item, so the iterator pairs it with each report.
type multiRegionAccessPointItem struct {
	report    awstypes.MultiRegionAccessPointReport
	accountID string
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
