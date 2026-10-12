// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"
	"strings"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_cloudwatchomni_domain", sweepDomains)
	awsv2.Register("aws_cloudwatchomni_view", sweepViews)
}

func sweepDomains(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudwatchomni.ListDomainsInput
	conn := client.CloudWatchOmniClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := cloudwatchomni.NewListDomainsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			id, name := aws.ToString(v.DomainId), aws.ToString(v.Name)

			// An account has at most one domain, so only sweep domains created by acceptance tests.
			if !strings.HasPrefix(name, sweep.ResourcePrefix+"-") {
				tflog.Info(ctx, "Skipping CloudWatch Omni Domain without an acceptance test name", map[string]any{
					"domain_id":    id,
					names.AttrName: name,
				})
				continue
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newDomainResource, client,
				framework.NewAttribute("domain_id", id)),
			)
		}
	}

	return sweepResources, nil
}

func sweepViews(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := cloudwatchomni.ListViewsInput{
		Type: awstypes.ViewTypeUser,
	}
	conn := client.CloudWatchOmniClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := cloudwatchomni.NewListViewsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			name := aws.ToString(v.Name)

			// View names carry the "view." prefix, so acceptance test views are "view.tf-acc-test-...".
			if !strings.HasPrefix(name, "view."+sweep.ResourcePrefix+"-") {
				tflog.Info(ctx, "Skipping CloudWatch Omni View without an acceptance test name", map[string]any{
					names.AttrName: name,
				})
				continue
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newViewResource, client,
				framework.NewAttribute(names.AttrName, name)),
			)
		}
	}

	return sweepResources, nil
}
