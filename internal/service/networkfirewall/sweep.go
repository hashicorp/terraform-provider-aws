// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func RegisterSweepers() {
	awsv2.Register("aws_networkfirewall_container_association", sweepContainerAssociations)
	awsv2.Register("aws_networkfirewall_firewall", sweepFirewalls, "aws_networkfirewall_logging_configuration")
	awsv2.Register("aws_networkfirewall_firewall_policy", sweepFirewallPolicies, "aws_networkfirewall_firewall")
	awsv2.Register("aws_networkfirewall_logging_configuration", sweepLoggingConfigurations)
	awsv2.Register("aws_networkfirewall_rule_group", sweepRuleGroups, "aws_networkfirewall_firewall_policy")
}

func sweepContainerAssociations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input networkfirewall.ListContainerAssociationsInput
	conn := client.NetworkFirewallClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := networkfirewall.NewListContainerAssociationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.ContainerAssociations {
			sweepResources = append(sweepResources, framework.NewSweepResource(newContainerAssociationResource, client,
				framework.NewAttribute("container_association_arn", aws.ToString(v.Arn))))
		}
	}

	return sweepResources, nil
}

func sweepFirewalls(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input networkfirewall.ListFirewallsInput
	conn := client.NetworkFirewallClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := networkfirewall.NewListFirewallsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Firewalls {
			r := resourceFirewall()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.FirewallArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepFirewallPolicies(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input networkfirewall.ListFirewallPoliciesInput
	conn := client.NetworkFirewallClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := networkfirewall.NewListFirewallPoliciesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.FirewallPolicies {
			r := resourceFirewallPolicy()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepLoggingConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input networkfirewall.ListFirewallsInput
	conn := client.NetworkFirewallClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := networkfirewall.NewListFirewallsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Firewalls {
			r := resourceLoggingConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.FirewallArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepRuleGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input networkfirewall.ListRuleGroupsInput
	conn := client.NetworkFirewallClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := networkfirewall.NewListRuleGroupsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.RuleGroups {
			r := resourceRuleGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
