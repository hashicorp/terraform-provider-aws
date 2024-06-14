// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_timestreaminfluxdb_db_instance", &resource.Sweeper{
		Name: "aws_timestreaminfluxdb_db_instance",
		F:    sweepDBInstances,
	})
}

func sweepDBInstances(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &timestreaminfluxdb.ListDbInstancesInput{}
	conn := client.TimestreamInfluxDBClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := timestreaminfluxdb.NewListDbInstancesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping TimestreamInfluxDB DB instance sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing TimestreamInfluxDB DB instances (%s): %w", region, err)
		}

		for _, v := range page.Items {
			id := aws.ToString(v.Id)
			log.Printf("[INFO] Deleting TimestreamInfluxDB DB instance: %s", id)

			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceDBInstance, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping TimestreamInfluxDB DB instances (%s): %w", region, err)
	}

	return nil
}
