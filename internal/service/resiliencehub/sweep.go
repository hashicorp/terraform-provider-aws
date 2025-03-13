// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resiliencehub

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehub"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_resiliencehub_resiliency_policy", sweepResiliencyPolicy)
}

func sweepResiliencyPolicy(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubClient(ctx)

	var sweepResources []sweep.Sweepable

	pages := resiliencehub.NewListResiliencyPoliciesPaginator(conn, &resiliencehub.ListResiliencyPoliciesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, policies := range page.ResiliencyPolicies {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceResiliencyPolicy, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(policies.PolicyArn)),
			))
		}
	}

	return sweepResources, nil
}
