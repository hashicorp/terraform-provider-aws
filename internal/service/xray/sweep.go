// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package xray

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/xray"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	sweep.Register("aws_xray_group", sweepGroups)

	sweep.Register("aws_xray_sampling_rule", sweepSamplingRules)
}

func sweepGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.XRayClient(ctx)

	var sweepResources []sweep.Sweepable
	r := resourceGroup()

	pages := xray.NewGetGroupsPaginator(conn, &xray.GetGroupsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			tflog.Warn(ctx, "Skipping sweeper", map[string]any{
				"error": err.Error(),
			})
			return nil, nil
		}
		if err != nil {
			return nil, err
		}

		for _, v := range page.Groups {
			if aws.ToString(v.GroupName) == "Default" {
				tflog.Debug(ctx, "Skipping resource", map[string]any{
					"skip_reason": `Cannot delete "Default"`,
					names.AttrARN: aws.ToString(v.GroupARN),
				})
				continue
			}
			d := r.Data(nil)
			d.SetId(aws.ToString(v.GroupARN))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepSamplingRules(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.XRayClient(ctx)

	var sweepResources []sweep.Sweepable
	r := resourceSamplingRule()

	pages := xray.NewGetSamplingRulesPaginator(conn, &xray.GetSamplingRulesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			tflog.Warn(ctx, "Skipping sweeper", map[string]any{
				"error": err.Error(),
			})
			return nil, nil
		}
		if err != nil {
			return nil, err
		}

		for _, v := range page.SamplingRuleRecords {
			if aws.ToString(v.SamplingRule.RuleName) == "Default" {
				tflog.Debug(ctx, "Skipping resource", map[string]any{
					"skip_reason": `Cannot delete "Default"`,
					names.AttrARN: aws.ToString(v.SamplingRule.RuleARN),
				})
				continue
			}
			d := r.Data(nil)
			d.SetId(aws.ToString(v.SamplingRule.RuleName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
