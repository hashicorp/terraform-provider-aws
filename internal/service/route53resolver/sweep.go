// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_route53_resolver_dnssec_config", &resource.Sweeper{
		Name: "aws_route53_resolver_dnssec_config",
		F:    sweepDNSSECConfig,
	})

	resource.AddTestSweepers("aws_route53_resolver_endpoint", &resource.Sweeper{
		Name: "aws_route53_resolver_endpoint",
		F:    sweepEndpoints,
		Dependencies: []string{
			"aws_route53_resolver_rule",
		},
	})

	resource.AddTestSweepers("aws_route53_resolver_firewall_config", &resource.Sweeper{
		Name: "aws_route53_resolver_firewall_config",
		F:    sweepFirewallConfigs,
	})

	resource.AddTestSweepers("aws_route53_resolver_firewall_domain_list", &resource.Sweeper{
		Name: "aws_route53_resolver_firewall_domain_list",
		F:    sweepFirewallDomainLists,
		Dependencies: []string{
			"aws_route53_resolver_firewall_rule",
		},
	})

	resource.AddTestSweepers("aws_route53_resolver_firewall_rule_group_association", &resource.Sweeper{
		Name: "aws_route53_resolver_firewall_rule_group_association",
		F:    sweepFirewallRuleGroupAssociations,
	})

	resource.AddTestSweepers("aws_route53_resolver_firewall_rule_group", &resource.Sweeper{
		Name: "aws_route53_resolver_firewall_rule_group",
		F:    sweepFirewallRuleGroups,
		Dependencies: []string{
			"aws_route53_resolver_firewall_rule",
			"aws_route53_resolver_firewall_rule_group_association",
		},
	})

	resource.AddTestSweepers("aws_route53_resolver_firewall_rule", &resource.Sweeper{
		Name: "aws_route53_resolver_firewall_rule",
		F:    sweepFirewallRules,
		Dependencies: []string{
			"aws_route53_resolver_firewall_rule_group_association",
		},
	})

	resource.AddTestSweepers("aws_route53_resolver_query_log_config_association", &resource.Sweeper{
		Name: "aws_route53_resolver_query_log_config_association",
		F:    sweepQueryLogConfigAssociations,
	})

	resource.AddTestSweepers("aws_route53_resolver_query_log_config", &resource.Sweeper{
		Name: "aws_route53_resolver_query_log_config",
		F:    sweepQueryLogsConfig,
		Dependencies: []string{
			"aws_route53_resolver_query_log_config_association",
		},
	})

	resource.AddTestSweepers("aws_route53_resolver_rule_association", &resource.Sweeper{
		Name: "aws_route53_resolver_rule_association",
		F:    sweepRuleAssociations,
	})

	resource.AddTestSweepers("aws_route53_resolver_rule", &resource.Sweeper{
		Name: "aws_route53_resolver_rule",
		F:    sweepRules,
		Dependencies: []string{
			"aws_route53_resolver_rule_association",
		},
	})
}

func sweepDNSSECConfig(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListResolverDnssecConfigsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListResolverDnssecConfigsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Resolver DNSSEC Config sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Route53 Resolver DNSSEC Configs (%s): %w", region, err)
		}

		for _, v := range page.ResolverDnssecConfigs {
			r := resourceDNSSECConfig()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))
			d.Set(names.AttrResourceID, v.ResourceId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver DNSSEC Configs (%s): %w", region, err)
	}

	return nil
}

func sweepEndpoints(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListResolverEndpointsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListResolverEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Resolver Endpoint sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Route53 Resolver Endpoints (%s): %w", region, err)
		}

		for _, v := range page.ResolverEndpoints {
			r := resourceEndpoint()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Endpoints (%s): %w", region, err)
	}

	return nil
}

func sweepFirewallConfigs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListFirewallConfigsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListFirewallConfigsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Resolver Firewall Config sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Route53 Resolver Firewall Configs (%s): %w", region, err)
		}

		for _, v := range page.FirewallConfigs {
			r := resourceFirewallConfig()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))
			d.Set(names.AttrResourceID, v.ResourceId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Firewall Configs (%s): %w", region, err)
	}

	return nil
}

func sweepFirewallDomainLists(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListFirewallDomainListsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListFirewallDomainListsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Resolver Firewall Domain List sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Route53 Resolver Firewall Domain Lists (%s): %w", region, err)
		}

		for _, v := range page.FirewallDomainLists {
			r := resourceFirewallDomainList()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Firewall Domain Lists (%s): %w", region, err)
	}

	return nil
}

func sweepFirewallRuleGroupAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListFirewallRuleGroupAssociationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListFirewallRuleGroupAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Resolver Firewall Rule Group Association sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Route53 Resolver Firewall Rule Group Associations (%s): %w", region, err)
		}

		for _, v := range page.FirewallRuleGroupAssociations {
			r := resourceFirewallRuleGroupAssociation()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Firewall Rule Group Associations (%s): %w", region, err)
	}

	return nil
}

func sweepFirewallRuleGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListFirewallRuleGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListFirewallRuleGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Resolver Firewall Rule Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Route53 Resolver Firewall Rule Groups (%s): %w", region, err)
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

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Firewall Rule Groups (%s): %w", region, err)
	}

	return nil
}

func sweepFirewallRules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListFirewallRuleGroupsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListFirewallRuleGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Print(fmt.Errorf("[WARN] Skipping Route53 Resolver Firewall Rule sweep for %s: %w", region, err))
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Route53 Resolver Firewall Rule Groups (%s): %w", region, err))
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
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Route53 Resolver Firewall Rules (%s): %w", region, err))
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

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Route53 Resolver Firewall Rules (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepQueryLogConfigAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListResolverQueryLogConfigAssociationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListResolverQueryLogConfigAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Resolver Query Log Config Association sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Route53 Resolver Query Log Config Associations (%s): %w", region, err)
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

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Query Log Config Associations (%s): %w", region, err)
	}

	return nil
}

func sweepQueryLogsConfig(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListResolverQueryLogConfigsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListResolverQueryLogConfigsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Resolver Query Log Config sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Route53 Resolver Query Log Configs (%s): %w", region, err)
		}

		for _, v := range page.ResolverQueryLogConfigs {
			r := resourceQueryLogConfig()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Query Log Configs (%s): %w", region, err)
	}

	return nil
}

func sweepRuleAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListResolverRuleAssociationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListResolverRuleAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Resolver Rule Association sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Route53 Resolver Rule Associations (%s): %w", region, err)
		}

		for _, v := range page.ResolverRuleAssociations {
			r := resourceRuleAssociation()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))
			d.Set("resolver_rule_id", v.ResolverRuleId)
			d.Set(names.AttrVPCID, v.VPCId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Rule Associations (%s): %w", region, err)
	}

	return nil
}

func sweepRules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.Route53ResolverClient(ctx)
	input := &route53resolver.ListResolverRulesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53resolver.NewListResolverRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Resolver Rule sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Route53 Resolver Rules (%s): %w", region, err)
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

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Rules (%s): %w", region, err)
	}

	return nil
}
