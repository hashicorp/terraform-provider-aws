// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_networksecuritymanager_rule", sweepRules)
}

func sweepRules(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input networksecuritymanager.ListRulesInput
	conn := client.NetworkSecurityManagerClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := networksecuritymanager.NewListRulesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Rules {
			sweepResources = append(sweepResources, framework.NewSweepResource(newRuleResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.RuleArn))),
			)
		}
	}

	return sweepResources, nil
}
