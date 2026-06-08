// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_resiliencehubv2_policy", sweepPolicies)
}

func sweepPolicies(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubV2Client(ctx)

	var sweepResources []sweep.Sweepable

	pages := resiliencehubv2.NewListPoliciesPaginator(conn, &resiliencehubv2.ListPoliciesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, policy := range page.PolicySummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourcePolicy, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(policy.PolicyArn)),
			))
		}
	}

	return sweepResources, nil
}
