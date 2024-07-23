// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dax

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dax"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dax/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_dax_cluster", &resource.Sweeper{
		Name: "aws_dax_cluster",
		F:    sweepClusters,
	})
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err.Error())
	}
	conn := client.DAXClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)

	err = describeClustersPages(ctx, conn, &dax.DescribeClustersInput{}, func(page *dax.DescribeClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.Clusters {
			r := ResourceCluster()
			d := r.Data(nil)
			d.SetId(aws.ToString(cluster.ClusterName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	// GovCloud (with no DAX support) has an endpoint that responds with:
	// InvalidParameterValueException: Access Denied to API Version: DAX_V3
	if awsv2.SkipSweepError(err) || errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "Access Denied to API Version: DAX_V3") {
		log.Printf("[WARN] Skipping DAX Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing DAX Clusters (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping DAX Clusters (%s): %w", region, err)
	}

	return nil
}
