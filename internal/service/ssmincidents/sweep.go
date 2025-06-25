// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmincidents

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_ssmincidents_replication_set", sweepReplicationSets)
}

func sweepReplicationSets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	if region := client.Region(ctx); region == endpoints.UsWest1RegionID {
		log.Printf("[WARN] Skipping SSMIncidents Replication Sets sweep for region: %s", region)
		return nil, nil
	}
	conn := client.SSMIncidentsClient(ctx)
	var input ssmincidents.ListReplicationSetsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ssmincidents.NewListReplicationSetsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ReplicationSetArns {
			r := resourceReplicationSet()
			d := r.Data(nil)
			d.SetId(v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
