// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_networkfirewall_proxy", &resource.Sweeper{
		Name: "aws_networkfirewall_proxy",
		F:    sweepProxies,
		Dependencies: []string{
			"aws_networkfirewall_proxy_configuration",
		},
	})

	resource.AddTestSweepers("aws_networkfirewall_proxy_configuration", &resource.Sweeper{
		Name: "aws_networkfirewall_proxy_configuration",
		F:    sweepProxyConfigurations,
		Dependencies: []string{
			"aws_networkfirewall_proxy_rule_group",
		},
	})

	resource.AddTestSweepers("aws_networkfirewall_proxy_rule_group", &resource.Sweeper{
		Name: "aws_networkfirewall_proxy_rule_group",
		F:    sweepProxyRuleGroups,
	})

	resource.AddTestSweepers("aws_networkfirewall_firewall_policy", &resource.Sweeper{
		Name: "aws_networkfirewall_firewall_policy",
		F:    sweepFirewallPolicies,
		Dependencies: []string{
			"aws_networkfirewall_firewall",
		},
	})

	resource.AddTestSweepers("aws_networkfirewall_firewall", &resource.Sweeper{
		Name: "aws_networkfirewall_firewall",
		F:    sweepFirewalls,
		Dependencies: []string{
			"aws_networkfirewall_logging_configuration",
		},
	})

	resource.AddTestSweepers("aws_networkfirewall_logging_configuration", &resource.Sweeper{
		Name: "aws_networkfirewall_logging_configuration",
		F:    sweepLoggingConfigurations,
	})

	resource.AddTestSweepers("aws_networkfirewall_rule_group", &resource.Sweeper{
		Name: "aws_networkfirewall_rule_group",
		F:    sweepRuleGroups,
		Dependencies: []string{
			"aws_networkfirewall_firewall_policy",
		},
	})
}

func sweepProxies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.NetworkFirewallClient(ctx)
	input := &networkfirewall.ListProxiesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkfirewall.NewListProxiesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping NetworkFirewall Proxy sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing NetworkFirewall Proxies (%s): %w", region, err)
		}

		for _, v := range page.Proxies {
			id := aws.ToString(v.Arn)
			log.Printf("[INFO] Deleting NetworkFirewall Proxy: %s", id)
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceProxy, client,
				sweepfw.NewAttribute(names.AttrID, id),
			))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping NetworkFirewall Proxies (%s): %w", region, err)
	}

	return nil
}

func sweepProxyConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.NetworkFirewallClient(ctx)
	input := &networkfirewall.ListProxyConfigurationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkfirewall.NewListProxyConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping NetworkFirewall Proxy Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing NetworkFirewall Proxy Configurations (%s): %w", region, err)
		}

		for _, v := range page.ProxyConfigurations {
			id := aws.ToString(v.Arn)
			log.Printf("[INFO] Deleting NetworkFirewall Proxy Configuration: %s", id)
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceProxyConfiguration, client,
				sweepfw.NewAttribute(names.AttrID, id),
			))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping NetworkFirewall Proxy Configurations (%s): %w", region, err)
	}

	return nil
}

func sweepProxyRuleGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.NetworkFirewallClient(ctx)
	input := &networkfirewall.ListProxyRuleGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkfirewall.NewListProxyRuleGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping NetworkFirewall Proxy Rule Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing NetworkFirewall Proxy Rule Groups (%s): %w", region, err)
		}

		for _, v := range page.ProxyRuleGroups {
			id := aws.ToString(v.Arn)
			log.Printf("[INFO] Deleting NetworkFirewall Proxy Rule Group: %s", id)
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceProxyRuleGroup, client,
				sweepfw.NewAttribute(names.AttrID, id),
			))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping NetworkFirewall Proxy Rule Groups (%s): %w", region, err)
	}

	return nil
}

func sweepFirewallPolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.NetworkFirewallClient(ctx)
	input := &networkfirewall.ListFirewallPoliciesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkfirewall.NewListFirewallPoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping NetworkFirewall Firewall Policy sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing NetworkFirewall Firewall Policies (%s): %w", region, err)
		}

		for _, v := range page.FirewallPolicies {
			r := resourceFirewallPolicy()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping NetworkFirewall Firewall Policies (%s): %w", region, err)
	}

	return nil
}

func sweepFirewalls(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.NetworkFirewallClient(ctx)
	input := &networkfirewall.ListFirewallsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkfirewall.NewListFirewallsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping NetworkFirewall Firewall sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing NetworkFirewall Firewalls (%s): %w", region, err)
		}

		for _, v := range page.Firewalls {
			r := resourceFirewall()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.FirewallArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping NetworkFirewall Firewalls (%s): %w", region, err)
	}

	return nil
}

func sweepLoggingConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.NetworkFirewallClient(ctx)
	input := &networkfirewall.ListFirewallsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkfirewall.NewListFirewallsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping NetworkFirewall Logging Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing NetworkFirewall Firewalls (%s): %w", region, err)
		}

		for _, v := range page.Firewalls {
			r := resourceLoggingConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.FirewallArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping NetworkFirewall Logging Configurations (%s): %w", region, err)
	}

	return nil
}

func sweepRuleGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.NetworkFirewallClient(ctx)
	input := &networkfirewall.ListRuleGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkfirewall.NewListRuleGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping NetworkFirewall Rule Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing NetworkFirewall Rule Groups (%s): %w", region, err)
		}

		for _, v := range page.RuleGroups {
			r := resourceRuleGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping NetworkFirewall Rule Groups (%s): %w", region, err)
	}

	return nil
}
