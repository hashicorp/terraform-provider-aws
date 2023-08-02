// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package codeartifact

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	conn := client.CodeArtifactConn(ctx)
	input := &codeartifact.ListDomainsInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListDomainsPagesWithContext(ctx, input, func(page *codeartifact.ListDomainsOutput, lastPage bool) bool {
		for _, domainPtr := range page.Domains {
			if domainPtr == nil {
				continue
			}

			domain := aws.StringValue(domainPtr.Name)
			input := &codeartifact.DeleteDomainInput{
				Domain: domainPtr.Name,
			}

			log.Printf("[INFO] Deleting CodeArtifact Domain: %s", domain)

			_, err := conn.DeleteDomainWithContext(ctx, input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CodeArtifact Domain (%s): %w", domain, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeArtifact Domain sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CodeArtifact Domains: %w", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepRepositories(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CodeArtifactConn(ctx)
	input := &codeartifact.ListRepositoriesInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListRepositoriesPagesWithContext(ctx, input, func(page *codeartifact.ListRepositoriesOutput, lastPage bool) bool {
		for _, repositoryPtr := range page.Repositories {
			if repositoryPtr == nil {
				continue
			}

			repository := aws.StringValue(repositoryPtr.Name)
			input := &codeartifact.DeleteRepositoryInput{
				Repository:  repositoryPtr.Name,
				Domain:      repositoryPtr.DomainName,
				DomainOwner: repositoryPtr.DomainOwner,
			}

			log.Printf("[INFO] Deleting CodeArtifact Repository: %s", repository)

			_, err := conn.DeleteRepositoryWithContext(ctx, input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CodeArtifact Repository (%s): %w", repository, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeArtifact Repository sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CodeArtifact Repositories: %w", err)
	}

	return sweeperErrs.ErrorOrNil()
}
