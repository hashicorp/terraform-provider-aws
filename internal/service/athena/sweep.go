// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package athena

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	conn := client.AthenaConn(ctx)
	input := &athena.ListDatabasesInput{
		CatalogName: aws.String("AwsDataCatalog"),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListDatabasesPagesWithContext(ctx, input, func(page *athena.ListDatabasesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DatabaseList {
			name := aws.StringValue(v.Name)
			if name == "default" {
				continue
			}
			r := ResourceDatabase()
			d := r.Data(nil)
			d.SetId(name)
			d.Set("force_destroy", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Athena Database sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Athena Databases (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Athena Databases (%s): %w", region, err)
	}

	return nil
}
