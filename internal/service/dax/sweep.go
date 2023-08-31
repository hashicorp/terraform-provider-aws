// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package dax

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_dax_cluster", &resource.Sweeper{
		Name: "aws_dax_cluster",
		F:    sweepClusters,
	})
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.DAXConn(ctx)

	resp, err := conn.DescribeClustersWithContext(ctx, &dax.DescribeClustersInput{})
	if err != nil {
		// GovCloud (with no DAX support) has an endpoint that responds with:
		// InvalidParameterValueException: Access Denied to API Version: DAX_V3
		if sweep.SkipSweepError(err) || tfawserr.ErrMessageContains(err, "InvalidParameterValueException", "Access Denied to API Version: DAX_V3") {
			log.Printf("[WARN] Skipping DAX Cluster sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving DAX clusters: %s", err)
	}

	if len(resp.Clusters) == 0 {
		log.Print("[DEBUG] No DAX clusters to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d DAX clusters", len(resp.Clusters))

	for _, cluster := range resp.Clusters {
		log.Printf("[INFO] Deleting DAX cluster %s", *cluster.ClusterName)
		_, err := conn.DeleteClusterWithContext(ctx, &dax.DeleteClusterInput{
			ClusterName: cluster.ClusterName,
		})
		if err != nil {
			return fmt.Errorf("Error deleting DAX cluster %s: %s", *cluster.ClusterName, err)
		}
	}

	return nil
}
