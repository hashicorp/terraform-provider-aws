// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeartifact

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_codeartifact_domain", &resource.Sweeper{
		Name: "aws_codeartifact_domain",
		F:    sweepDomains,
	})

	resource.AddTestSweepers("aws_codeartifact_repository", &resource.Sweeper{
		Name: "aws_codeartifact_repository",
		F:    sweepRepositories,
	})
}

func sweepDomains(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CodeArtifactClient(ctx)
	input := &codeartifact.ListDomainsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := codeartifact.NewListDomainsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CodeArtifact Domain sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CodeArtifact Domains (%s): %w", region, err)
		}

		for _, v := range page.Domains {
			r := resourceDomain()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CodeArtifact Domains (%s): %w", region, err)
	}

	return nil
}

func sweepRepositories(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CodeArtifactClient(ctx)
	input := &codeartifact.ListRepositoriesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := codeartifact.NewListRepositoriesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CodeArtifact Repository sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CodeArtifact Repositories (%s): %w", region, err)
		}

		for _, v := range page.Repositories {
			r := resourceRepository()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CodeArtifact Repositories (%s): %w", region, err)
	}

	return nil
}
