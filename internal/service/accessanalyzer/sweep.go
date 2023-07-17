// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package accessanalyzer

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func init() {
	resource.AddTestSweepers("aws_accessanalyzer_analyzer", &resource.Sweeper{
		Name: "aws_accessanalyzer_analyzer",
		F:    sweepAnalyzers,
	})
}

func sweepAnalyzers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.AccessAnalyzerClient(ctx)
	input := &accessanalyzer.ListAnalyzersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := accessanalyzer.NewListAnalyzersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IAM Access Analyzer Analyzer sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing IAM Access Analyzer Analyzers (%s): %w", region, err)
		}

		for _, v := range page.Analyzers {
			r := resourceAnalyzer()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping IAM Access Analyzer Analyzers (%s): %w", region, err)
	}

	return nil
}
