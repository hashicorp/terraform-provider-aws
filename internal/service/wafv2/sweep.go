// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_wafv2_api_key", sweepAPIKeys)
	awsv2.Register("aws_wafv2_ip_set", sweepIPSets, "aws_wafv2_rule_group", "aws_wafv2_web_acl")
	awsv2.Register("aws_wafv2_regex_pattern_set", sweepRegexPatternSets, "aws_wafv2_rule_group", "aws_wafv2_web_acl")
	awsv2.Register("aws_wafv2_rule_group", sweepRuleGroups, "aws_wafv2_web_acl")
	awsv2.Register("aws_wafv2_web_acl", sweepWebACLs)
}

func sweepAPIKeys(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.WAFV2Client(ctx)
	var input wafv2.ListAPIKeysInput
	input.Scope = awstypes.ScopeRegional
	sweepResources := make([]sweep.Sweepable, 0)

	err := listAPIKeysPages(ctx, conn, &input, func(page *wafv2.ListAPIKeysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.APIKeySummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newAPIKeyResource, client,
				framework.NewAttribute("api_key", aws.ToString(v.APIKey)),
				framework.NewAttribute(names.AttrScope, awstypes.ScopeRegional),
			))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepIPSets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.WAFV2Client(ctx)
	var input wafv2.ListIPSetsInput
	input.Scope = awstypes.ScopeRegional
	sweepResources := make([]sweep.Sweepable, 0)

	err := listIPSetsPages(ctx, conn, &input, func(page *wafv2.ListIPSetsOutput, lastPage bool) bool {
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

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepRegexPatternSets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.WAFV2Client(ctx)
	var input wafv2.ListRegexPatternSetsInput
	input.Scope = awstypes.ScopeRegional
	sweepResources := make([]sweep.Sweepable, 0)

	err := listRegexPatternSetsPages(ctx, conn, &input, func(page *wafv2.ListRegexPatternSetsOutput, lastPage bool) bool {
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

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepRuleGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.WAFV2Client(ctx)
	var input wafv2.ListRuleGroupsInput
	input.Scope = awstypes.ScopeRegional
	sweepResources := make([]sweep.Sweepable, 0)

	err := listRuleGroupsPages(ctx, conn, &input, func(page *wafv2.ListRuleGroupsOutput, lastPage bool) bool {
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

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepWebACLs(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.WAFV2Client(ctx)
	var input wafv2.ListWebACLsInput
	input.Scope = awstypes.ScopeRegional
	sweepResources := make([]sweep.Sweepable, 0)

	err := listWebACLsPages(ctx, conn, &input, func(page *wafv2.ListWebACLsOutput, lastPage bool) bool {
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

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}
