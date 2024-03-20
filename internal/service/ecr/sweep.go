// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_ecr_repository", &resource.Sweeper{
		Name: "aws_ecr_repository",
		F:    sweepRepositories,
	})
}

func sweepRepositories(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.ECRClient(ctx)

	var errors error
	err = describeRepositoriesPages(ctx, conn, &ecr.DescribeRepositoriesInput{}, func(page *ecr.DescribeRepositoriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, repository := range page.Repositories {
			repositoryName := aws.ToString(repository.RepositoryName)
			log.Printf("[INFO] Deleting ECR repository: %s", repositoryName)

			_, err = conn.DeleteRepository(ctx, &ecr.DeleteRepositoryInput{
				// We should probably sweep repositories even if there are images.
				Force:          true,
				RegistryId:     repository.RegistryId,
				RepositoryName: repository.RepositoryName,
			})
			if err != nil {
				if errs.IsA[*awstypes.RepositoryNotFoundException](err) {
					sweeperErr := fmt.Errorf("Error deleting ECR repository (%s): %w", repositoryName, err)
					log.Printf("[ERROR] %s", sweeperErr)
					errors = multierror.Append(errors, sweeperErr)
				}
				continue
			}
		}

		return !lastPage
	})
	if err != nil {
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ECR repository sweep for %s: %s", region, err)
			return nil
		}
		errors = multierror.Append(errors, fmt.Errorf("Error retrieving ECR repositories: %w", err))
	}

	return errors
}
