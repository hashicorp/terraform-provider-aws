// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
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
	input := &ecr.DescribeRepositoriesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ecr.NewDescribeRepositoriesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ECR Repository sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ECR Repositories (%s): %w", region, err)
		}

		for _, v := range page.Repositories {
			r := resourceRepository()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.RepositoryName))
			d.Set(names.AttrForceDelete, true)
			d.Set("registry_id", v.RegistryId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ECR Repositories (%s): %w", region, err)
	}

	return nil
}
