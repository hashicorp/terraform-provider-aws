//go:build sweep
// +build sweep

package wafv2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).WAFV2Conn()
	input := &wafv2.ListIPSetsInput{
		Scope: aws.String(wafv2.ScopeRegional),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listIPSetsPages(ctx, conn, input, func(page *wafv2.ListIPSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.IPSets {
			r := ResourceIPSet()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))
			d.Set("lock_token", v.LockToken)
			d.Set("name", v.Name)
			d.Set("scope", input.Scope)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAFv2 IPSet sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAFv2 IPSets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAFv2 IPSets (%s): %w", region, err)
	}

	return nil
}

func sweepRegexPatternSets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).WAFV2Conn()
	input := &wafv2.ListRegexPatternSetsInput{
		Scope: aws.String(wafv2.ScopeRegional),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listRegexPatternSetsPages(ctx, conn, input, func(page *wafv2.ListRegexPatternSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RegexPatternSets {
			r := ResourceRegexPatternSet()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))
			d.Set("lock_token", v.LockToken)
			d.Set("name", v.Name)
			d.Set("scope", input.Scope)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAFv2 RegexPatternSet sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAFv2 RegexPatternSets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAFv2 RegexPatternSets (%s): %w", region, err)
	}

	return nil
}

func sweepRuleGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).WAFV2Conn()
	input := &wafv2.ListRuleGroupsInput{
		Scope: aws.String(wafv2.ScopeRegional),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listRuleGroupsPages(ctx, conn, input, func(page *wafv2.ListRuleGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RuleGroups {
			r := ResourceRuleGroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))
			d.Set("lock_token", v.LockToken)
			d.Set("name", v.Name)
			d.Set("scope", input.Scope)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAFv2 RuleGroup sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAFv2 RuleGroups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAFv2 RuleGroups (%s): %w", region, err)
	}

	return nil
}

func sweepWebACLs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).WAFV2Conn()
	input := &wafv2.ListWebACLsInput{
		Scope: aws.String(wafv2.ScopeRegional),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listWebACLsPages(ctx, conn, input, func(page *wafv2.ListWebACLsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.WebACLs {
			name := aws.StringValue(v.Name)

			// Exclude WebACLs managed by Firewall Manager as deletion returns AccessDeniedException.
			// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19149
			// Prefix Reference: https://docs.aws.amazon.com/waf/latest/developerguide/get-started-fms-create-security-policy.html
			if strings.HasPrefix(name, "FMManagedWebACLV2") {
				log.Printf("[WARN] Skipping WAFv2 Web ACL: %s", name)
				continue
			}

			r := ResourceWebACL()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))
			d.Set("lock_token", v.LockToken)
			d.Set("name", name)
			d.Set("scope", input.Scope)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAFv2 WebACL sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAFv2 WebACLs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAFv2 WebACLs (%s): %w", region, err)
	}

	return nil
}
