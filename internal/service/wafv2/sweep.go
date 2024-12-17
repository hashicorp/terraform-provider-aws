// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_wafv2_ip_set", &resource.Sweeper{
		Name: "aws_wafv2_ip_set",
		F:    sweepIPSets,
		Dependencies: []string{
			"aws_wafv2_rule_group",
			"aws_wafv2_web_acl",
		},
	})

	resource.AddTestSweepers("aws_wafv2_regex_pattern_set", &resource.Sweeper{
		Name: "aws_wafv2_regex_pattern_set",
		F:    sweepRegexPatternSets,
		Dependencies: []string{
			"aws_wafv2_rule_group",
			"aws_wafv2_web_acl",
		},
	})

	resource.AddTestSweepers("aws_wafv2_rule_group", &resource.Sweeper{
		Name: "aws_wafv2_rule_group",
		F:    sweepRuleGroups,
		Dependencies: []string{
			"aws_wafv2_web_acl",
		},
	})

	resource.AddTestSweepers("aws_wafv2_web_acl", &resource.Sweeper{
		Name: "aws_wafv2_web_acl",
		F:    sweepWebACLs,
	})
}

func sweepIPSets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFV2Client(ctx)
	input := &wafv2.ListIPSetsInput{
		Scope: awstypes.ScopeRegional,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listIPSetsPages(ctx, conn, input, func(page *wafv2.ListIPSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.IPSets {
			r := resourceIPSet()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))
			d.Set("lock_token", v.LockToken)
			d.Set(names.AttrName, v.Name)
			d.Set(names.AttrScope, input.Scope)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAFv2 IPSet sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAFv2 IPSets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAFv2 IPSets (%s): %w", region, err)
	}

	return nil
}

func sweepRegexPatternSets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFV2Client(ctx)
	input := &wafv2.ListRegexPatternSetsInput{
		Scope: awstypes.ScopeRegional,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listRegexPatternSetsPages(ctx, conn, input, func(page *wafv2.ListRegexPatternSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RegexPatternSets {
			r := resourceRegexPatternSet()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))
			d.Set("lock_token", v.LockToken)
			d.Set(names.AttrName, v.Name)
			d.Set(names.AttrScope, input.Scope)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAFv2 RegexPatternSet sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAFv2 RegexPatternSets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAFv2 RegexPatternSets (%s): %w", region, err)
	}

	return nil
}

func sweepRuleGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFV2Client(ctx)
	input := &wafv2.ListRuleGroupsInput{
		Scope: awstypes.ScopeRegional,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listRuleGroupsPages(ctx, conn, input, func(page *wafv2.ListRuleGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RuleGroups {
			r := resourceRuleGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))
			d.Set("lock_token", v.LockToken)
			d.Set(names.AttrName, v.Name)
			d.Set(names.AttrScope, input.Scope)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAFv2 RuleGroup sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAFv2 RuleGroups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAFv2 RuleGroups (%s): %w", region, err)
	}

	return nil
}

func sweepWebACLs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFV2Client(ctx)
	input := &wafv2.ListWebACLsInput{
		Scope: awstypes.ScopeRegional,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listWebACLsPages(ctx, conn, input, func(page *wafv2.ListWebACLsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.WebACLs {
			name := aws.ToString(v.Name)

			// Exclude WebACLs managed by Firewall Manager as deletion returns AccessDeniedException.
			// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19149
			// Prefix Reference: https://docs.aws.amazon.com/waf/latest/developerguide/get-started-fms-create-security-policy.html
			if strings.HasPrefix(name, "FMManagedWebACLV2") {
				log.Printf("[WARN] Skipping WAFv2 Web ACL: %s", name)
				continue
			}

			r := resourceWebACL()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))
			d.Set("lock_token", v.LockToken)
			d.Set(names.AttrName, name)
			d.Set(names.AttrScope, input.Scope)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAFv2 WebACL sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAFv2 WebACLs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAFv2 WebACLs (%s): %w", region, err)
	}

	return nil
}
