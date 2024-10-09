// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package athena

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_athena_database", &resource.Sweeper{
		Name: "aws_athena_database",
		F:    sweepDatabases,
	})
}

func sweepDatabases(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.AthenaClient(ctx)
	input := &athena.ListDatabasesInput{
		CatalogName: aws.String("AwsDataCatalog"),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := athena.NewListDatabasesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Athena Database sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Athena Databases (%s): %w", region, err)
		}

		for _, v := range page.DatabaseList {
			name := aws.ToString(v.Name)
			if name == "default" {
				continue
			}
			r := resourceDatabase()
			d := r.Data(nil)
			d.SetId(name)
			d.Set(names.AttrForceDestroy, true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Athena Databases (%s): %w", region, err)
	}

	return nil
}
