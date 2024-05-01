// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmincidents

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_ssmincidents_replication_set", &resource.Sweeper{
		Name: "aws_ssmincidents_replication_set",
		F:    sweepReplicationSets,
	})
}

func sweepReplicationSets(region string) error {
	ctx := sweep.Context(region)
	if region == names.USWest1RegionID {
		log.Printf("[WARN] Skipping SSMIncidents Replication Sets sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.SSMIncidentsClient(ctx)
	input := &ssmincidents.ListReplicationSetsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ssmincidents.NewListReplicationSetsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SSMIncidents Replication Sets sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving SSMIncidents Replication Sets: %w", err)
		}

		for _, rs := range page.ReplicationSetArns {
			id := rs

			r := ResourceReplicationSet()
			d := r.Data(nil)
			d.SetId(id)

			log.Printf("[INFO] Deleting SSMIncidents Replication Set: %s", id)
			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping SSMIncidents Replication Sets for %s: %w", region, err)
	}

	return nil
}
