// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package billing

import (
	"context"
	"fmt"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/billing"
	awstypes "github.com/aws/aws-sdk-go-v2/service/billing/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_billing_view", sweepViews)
}

func sweepViews(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.BillingClient(ctx)

	var sweepResources []sweep.Sweepable

	input := billing.ListBillingViewsInput{}
	pages := billing.NewListBillingViewsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.BillingViews {
			if v.BillingViewType != awstypes.BillingViewTypeCustom {
				tflog.Info(ctx, "Skipping resource", map[string]any{
					names.AttrName: aws.ToString(v.Name),
					"skip_reason":  fmt.Sprintf("View Type is %q", v.BillingViewType),
				})
				continue
			}
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceView, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.Arn))),
			)
		}
	}

	return sweepResources, nil
}
