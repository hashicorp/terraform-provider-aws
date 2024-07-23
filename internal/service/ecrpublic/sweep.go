// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecrpublic

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_ecrpublic_repository", &resource.Sweeper{
		Name: "aws_ecrpublic_repository",
		F:    sweepRepositories,
	})
}

func sweepRepositories(region string) error {
	ctx := sweep.Context(region)
	// "UnsupportedCommandException: DescribeRepositories command is only supported in us-east-1".
	if region != names.USEast1RegionID {
		log.Printf("[WARN] Skipping ECR Public Repository sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ECRPublicClient(ctx)
	input := &ecrpublic.DescribeRepositoriesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	paginator := ecrpublic.NewDescribeRepositoriesPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ECR Public Repository sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ECR Public Repositories (%s): %w", region, err)
		}

		for _, repository := range page.Repositories {
			r := ResourceRepository()
			d := r.Data(nil)
			d.SetId(aws.ToString(repository.RepositoryName))
			d.Set("registry_id", repository.RegistryId)
			d.Set(names.AttrForceDestroy, true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ECR Public Repositories (%s): %w", region, err)
	}

	return nil
}
