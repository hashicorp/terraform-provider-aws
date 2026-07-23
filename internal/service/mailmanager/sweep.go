// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mailmanager

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_mailmanager_traffic_policy", sweepTrafficPolicies)
}

func sweepTrafficPolicies(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.MailManagerClient(ctx)
	var input mailmanager.ListTrafficPoliciesInput
	var sweepResources []sweep.Sweepable

	pages := mailmanager.NewListTrafficPoliciesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.TrafficPolicies {
			sweepResources = append(sweepResources, framework.NewSweepResource(newTrafficPolicyResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.TrafficPolicyId)),
			))
		}
	}

	return sweepResources, nil
}
