// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_waf_byte_match_set", &resource.Sweeper{
		Name: "aws_waf_byte_match_set",
		F:    sweepByteMatchSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})

	resource.AddTestSweepers("aws_waf_geo_match_set", &resource.Sweeper{
		Name: "aws_waf_geo_match_set",
		F:    sweepGeoMatchSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})

	resource.AddTestSweepers("aws_waf_ipset", &resource.Sweeper{
		Name: "aws_waf_ipset",
		F:    sweepIPSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})

	resource.AddTestSweepers("aws_waf_rate_based_rule", &resource.Sweeper{
		Name: "aws_waf_rate_based_rule",
		F:    sweepRateBasedRules,
		Dependencies: []string{
			"aws_waf_rule_group",
			"aws_waf_web_acl",
		},
	})

	resource.AddTestSweepers("aws_waf_regex_match_set", &resource.Sweeper{
		Name: "aws_waf_regex_match_set",
		F:    sweepRegexMatchSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})

	resource.AddTestSweepers("aws_waf_regex_pattern_set", &resource.Sweeper{
		Name: "aws_waf_regex_pattern_set",
		F:    sweepRegexPatternSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})

	resource.AddTestSweepers("aws_waf_rule_group", &resource.Sweeper{
		Name: "aws_waf_rule_group",
		F:    sweepRuleGroups,
		Dependencies: []string{
			"aws_waf_web_acl",
		},
	})

	resource.AddTestSweepers("aws_waf_rule", &resource.Sweeper{
		Name: "aws_waf_rule",
		F:    sweepRules,
		Dependencies: []string{
			"aws_waf_rule_group",
			"aws_waf_web_acl",
		},
	})

	resource.AddTestSweepers("aws_waf_size_constraint_set", &resource.Sweeper{
		Name: "aws_waf_size_constraint_set",
		F:    sweepSizeConstraintSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})

	resource.AddTestSweepers("aws_waf_sql_injection_match_set", &resource.Sweeper{
		Name: "aws_waf_sql_injection_match_set",
		F:    sweepSQLInjectionMatchSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})

	resource.AddTestSweepers("aws_waf_web_acl", &resource.Sweeper{
		Name: "aws_waf_web_acl",
		F:    sweepWebACLs,
	})

	resource.AddTestSweepers("aws_waf_xss_match_set", &resource.Sweeper{
		Name: "aws_waf_xss_match_set",
		F:    sweepXSSMatchSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})
}

func sweepByteMatchSet(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFClient(ctx)
	input := &waf.ListByteMatchSetsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listByteMatchSetsPages(ctx, conn, input, func(page *waf.ListByteMatchSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ByteMatchSets {
			id := aws.ToString(v.ByteMatchSetId)
			r := resourceByteMatchSet()
			d := r.Data(nil)
			d.SetId(id)
			// Refresh.
			if err := sdk.ReadResource(ctx, r, d, client); err != nil {
				log.Printf("[WARN] Skipping WAF Byte Match Set %s: %s", id, err)
				continue
			}
			if d.Id() == "" {
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Byte Match Set sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAF Byte Match Sets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAF Byte Match Sets (%s): %w", region, err)
	}

	return nil
}

func sweepGeoMatchSet(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFClient(ctx)
	input := &waf.ListGeoMatchSetsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listGeoMatchSetsPages(ctx, conn, input, func(page *waf.ListGeoMatchSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.GeoMatchSets {
			id := aws.ToString(v.GeoMatchSetId)
			r := resourceGeoMatchSet()
			d := r.Data(nil)
			d.SetId(id)
			// Refresh.
			if err := sdk.ReadResource(ctx, r, d, client); err != nil {
				log.Printf("[WARN] Skipping WAF Geo Match Set %s: %s", id, err)
				continue
			}
			if d.Id() == "" {
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Geo Match Set sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAF Geo Match Sets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAF Geo Match Sets (%s): %w", region, err)
	}

	return nil
}

func sweepIPSet(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFClient(ctx)
	input := &waf.ListIPSetsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listIPSetsPages(ctx, conn, input, func(page *waf.ListIPSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.IPSets {
			id := aws.ToString(v.IPSetId)
			r := resourceIPSet()
			d := r.Data(nil)
			d.SetId(id)
			// Refresh.
			if err := sdk.ReadResource(ctx, r, d, client); err != nil {
				log.Printf("[WARN] Skipping WAF IP Set %s: %s", id, err)
				continue
			}
			if d.Id() == "" {
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF IP Set sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAF IP Sets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAF IP Sets (%s): %w", region, err)
	}

	return nil
}

func sweepRateBasedRules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFClient(ctx)
	input := &waf.ListRateBasedRulesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listRateBasedRulesPages(ctx, conn, input, func(page *waf.ListRateBasedRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Rules {
			id := aws.ToString(v.RuleId)
			r := resourceRateBasedRule()
			d := r.Data(nil)
			d.SetId(id)
			// Refresh.
			if err := sdk.ReadResource(ctx, r, d, client); err != nil {
				log.Printf("[WARN] Skipping WAF Rate Based Rule %s: %s", id, err)
				continue
			}
			if d.Id() == "" {
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Rate Based Rule sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAF Rate Based Rules (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAF Rate Based Rules (%s): %w", region, err)
	}

	return nil
}

func sweepRegexMatchSet(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFClient(ctx)
	input := &waf.ListRegexMatchSetsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listRegexMatchSetsPages(ctx, conn, input, func(page *waf.ListRegexMatchSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RegexMatchSets {
			id := aws.ToString(v.RegexMatchSetId)
			r := resourceRegexMatchSet()
			d := r.Data(nil)
			d.SetId(id)
			// Refresh.
			if err := sdk.ReadResource(ctx, r, d, client); err != nil {
				log.Printf("[WARN] Skipping WAF Regex Match Set %s: %s", id, err)
				continue
			}
			if d.Id() == "" {
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Regex Match Set sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAF Regex Match Sets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAF Regex Match Sets (%s): %w", region, err)
	}

	return nil
}

func sweepRegexPatternSet(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFClient(ctx)
	input := &waf.ListRegexPatternSetsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listRegexPatternSetsPages(ctx, conn, input, func(page *waf.ListRegexPatternSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RegexPatternSets {
			id := aws.ToString(v.RegexPatternSetId)
			r := resourceRegexPatternSet()
			d := r.Data(nil)
			d.SetId(id)
			// Refresh.
			if err := sdk.ReadResource(ctx, r, d, client); err != nil {
				log.Printf("[WARN] Skipping WAF Regex Pattern Set %s: %s", id, err)
				continue
			}
			if d.Id() == "" {
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Regex Pattern Set sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAF Regex Pattern Sets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAF Regex Pattern Sets (%s): %w", region, err)
	}

	return nil
}

func sweepRuleGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFClient(ctx)
	input := &waf.ListRuleGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listRuleGroupsPages(ctx, conn, input, func(page *waf.ListRuleGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RuleGroups {
			id := aws.ToString(v.RuleGroupId)
			r := resourceRuleGroup()
			d := r.Data(nil)
			d.SetId(id)
			// Refresh.
			if err := sdk.ReadResource(ctx, r, d, client); err != nil {
				log.Printf("[WARN] Skipping WAF Rule Group %s: %s", id, err)
				continue
			}
			if d.Id() == "" {
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Rule Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAF Rule Groups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAF Rule Groups (%s): %w", region, err)
	}

	return nil
}

func sweepRules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFClient(ctx)
	input := &waf.ListRulesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listRulesPages(ctx, conn, input, func(page *waf.ListRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Rules {
			id := aws.ToString(v.RuleId)
			r := resourceRule()
			d := r.Data(nil)
			d.SetId(id)
			// Refresh.
			if err := sdk.ReadResource(ctx, r, d, client); err != nil {
				log.Printf("[WARN] Skipping WAF Rule %s: %s", id, err)
				continue
			}
			if d.Id() == "" {
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Rule sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAF Rules (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAF Rules (%s): %w", region, err)
	}

	return nil
}

func sweepSizeConstraintSet(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFClient(ctx)
	input := &waf.ListSizeConstraintSetsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listSizeConstraintSetsPages(ctx, conn, input, func(page *waf.ListSizeConstraintSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.SizeConstraintSets {
			id := aws.ToString(v.SizeConstraintSetId)
			r := resourceSizeConstraintSet()
			d := r.Data(nil)
			d.SetId(id)
			// Refresh.
			if err := sdk.ReadResource(ctx, r, d, client); err != nil {
				log.Printf("[WARN] Skipping WAF Size Constraint Set %s: %s", id, err)
				continue
			}
			if d.Id() == "" {
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Size Constraint Set sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAF Size Constraint Sets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAF Size Constraint Sets (%s): %w", region, err)
	}

	return nil
}

func sweepSQLInjectionMatchSet(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFClient(ctx)
	input := &waf.ListSqlInjectionMatchSetsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listSQLInjectionMatchSetsPages(ctx, conn, input, func(page *waf.ListSqlInjectionMatchSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.SqlInjectionMatchSets {
			id := aws.ToString(v.SqlInjectionMatchSetId)
			r := resourceSQLInjectionMatchSet()
			d := r.Data(nil)
			d.SetId(id)
			// Refresh.
			if err := sdk.ReadResource(ctx, r, d, client); err != nil {
				log.Printf("[WARN] Skipping WAF SQL Injection Match Set %s: %s", id, err)
				continue
			}
			if d.Id() == "" {
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF SQL Injection Match Set sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAF SQL Injection Match Sets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAF SQL Injection Match Sets (%s): %w", region, err)
	}

	return nil
}

func sweepWebACLs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.WAFClient(ctx)
	input := &waf.ListWebACLsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listWebACLsPages(ctx, conn, input, func(page *waf.ListWebACLsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.WebACLs {
			id := aws.ToString(v.WebACLId)
			r := resourceWebACL()
			d := r.Data(nil)
			d.SetId(id)
			// Refresh.
			if err := sdk.ReadResource(ctx, r, d, client); err != nil {
				log.Printf("[WARN] Skipping WAF Web ACL %s: %s", id, err)
				continue
			}
			if d.Id() == "" {
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Web ACL sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAF Web ACLs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAFWeb ACLs (%s): %w", region, err)
	}

	return nil
}

func sweepXSSMatchSet(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFClient(ctx)
	input := &waf.ListXssMatchSetsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listXSSMatchSetsPages(ctx, conn, input, func(page *waf.ListXssMatchSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.XssMatchSets {
			id := aws.ToString(v.XssMatchSetId)
			r := resourceXSSMatchSet()
			d := r.Data(nil)
			d.SetId(id)
			// Refresh.
			if err := sdk.ReadResource(ctx, r, d, client); err != nil {
				log.Printf("[WARN] Skipping WAF XSS Match Set %s: %s", id, err)
				continue
			}
			if d.Id() == "" {
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF XSS Match Set sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAF XSS Match Sets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAF XSS Match Sets (%s): %w", region, err)
	}

	return nil
}
