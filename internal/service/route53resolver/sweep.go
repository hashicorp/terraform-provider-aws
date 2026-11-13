// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_route53_resolver_dnssec_config", sweepDNSSECConfig)
	awsv2.Register("aws_route53_resolver_endpoint", sweepEndpoints, "aws_route53_resolver_rule")
	awsv2.Register("aws_route53_resolver_firewall_config", sweepFirewallConfigs)
	awsv2.Register("aws_route53_resolver_firewall_domain_list", sweepFirewallDomainLists, "aws_route53_resolver_firewall_rule")
	awsv2.Register("aws_route53_resolver_firewall_rule_group_association", sweepFirewallRuleGroupAssociations)
	awsv2.Register("aws_route53_resolver_firewall_rule_group", sweepFirewallRuleGroups,
		"aws_route53_resolver_firewall_rule",
		"aws_route53_resolver_firewall_rule_group_association",
	)
	awsv2.Register("aws_route53_resolver_firewall_rule", sweepFirewallRules, "aws_route53_resolver_firewall_rule_group_association")
	awsv2.Register("aws_route53_resolver_query_log_config_association", sweepQueryLogConfigAssociations)
	awsv2.Register("aws_route53_resolver_query_log_config", sweepQueryLogsConfig, "aws_route53_resolver_query_log_config_association")
	awsv2.Register("aws_route53_resolver_rule_association", sweepRuleAssociations)
	awsv2.Register("aws_route53_resolver_rule", sweepRules, "aws_route53_resolver_rule_association")
}

func sweepDNSSECConfig(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListResolverDnssecConfigsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListResolverDnssecConfigsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.ResolverDnssecConfigs {
			r := resourceDNSSECConfig()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))
			d.Set(names.AttrResourceID, v.ResourceId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepEndpoints(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListResolverEndpointsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListResolverEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.ResolverEndpoints {
			r := resourceEndpoint()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepFirewallConfigs(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListFirewallConfigsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListFirewallConfigsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.FirewallConfigs {
			r := resourceFirewallConfig()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))
			d.Set(names.AttrResourceID, v.ResourceId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepFirewallDomainLists(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListFirewallDomainListsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListFirewallDomainListsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.FirewallDomainLists {
			r := resourceFirewallDomainList()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepFirewallRuleGroupAssociations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListFirewallRuleGroupAssociationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListFirewallRuleGroupAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.FirewallRuleGroupAssociations {
			r := resourceFirewallRuleGroupAssociation()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepFirewallRuleGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListFirewallRuleGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListFirewallRuleGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.FirewallRuleGroups {
			id := aws.ToString(v.Id)

			if shareStatus := v.ShareStatus; shareStatus == awstypes.ShareStatusSharedWithMe {
				log.Printf("[INFO] Skipping Route53 Resolver Firewall Rule Group %s: ShareStatus=%s", id, shareStatus)
				continue
			}

			r := resourceFirewallRuleGroup()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepFirewallRules(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListFirewallRuleGroupsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListFirewallRuleGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			return nil, err
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, err)
			continue
		}

		for _, v := range page.FirewallRuleGroups {
			id := aws.ToString(v.Id)

			if shareStatus := v.ShareStatus; shareStatus == awstypes.ShareStatusSharedWithMe {
				log.Printf("[INFO] Skipping Route53 Resolver Firewall Rule Group %s: ShareStatus=%s", id, shareStatus)
				continue
			}

			input := &route53resolver.ListFirewallRulesInput{
				FirewallRuleGroupId: aws.String(id),
			}

			pages := route53resolver.NewListFirewallRulesPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, err)
					continue
				}

				for _, v := range page.FirewallRules {
					r := resourceFirewallRule()
					d := r.Data(nil)
					d.SetId(firewallRuleCreateResourceID(aws.ToString(v.FirewallRuleGroupId), aws.ToString(v.FirewallDomainListId)))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	return sweepResources, sweeperErrs.ErrorOrNil()
}

func sweepQueryLogConfigAssociations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListResolverQueryLogConfigAssociationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListResolverQueryLogConfigAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.ResolverQueryLogConfigAssociations {
			r := resourceQueryLogConfigAssociation()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))
			d.Set("resolver_query_log_config_id", v.ResolverQueryLogConfigId)
			d.Set(names.AttrResourceID, v.ResourceId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepQueryLogsConfig(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListResolverQueryLogConfigsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListResolverQueryLogConfigsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.ResolverQueryLogConfigs {
			r := resourceQueryLogConfig()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepRuleAssociations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListResolverRuleAssociationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListResolverRuleAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.ResolverRuleAssociations {
			id := aws.ToString(v.Id)
			// Cannot associate or disassociate system defined rules
			if strings.Contains(strings.ToLower(id), "autodefined") {
				log.Printf("[INFO] Skipping Route53 Resolver Rule Associations %s: System Defined", id)
				continue
			}

			r := resourceRuleAssociation()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("resolver_rule_id", v.ResolverRuleId)
			d.Set(names.AttrVPCID, v.VPCId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepRules(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListResolverRulesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListResolverRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.ResolverRules {
			if aws.ToString(v.OwnerId) != client.AccountID(ctx) {
				continue
			}

			r := resourceRule()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
