//go:build sweep
// +build sweep

package docdb

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_docdb_global_cluster", &resource.Sweeper{
		Name:         "aws_docdb_global_cluster",
		F:            sweepGlobalClusters,
		Dependencies: []string{
			// "aws_docdb_cluster",
		},
	})
}

func sweepGlobalClusters(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DocDBConn
	input := &docdb.DescribeGlobalClustersInput{}

	err = conn.DescribeGlobalClustersPages(input, func(out *docdb.DescribeGlobalClustersOutput, lastPage bool) bool {
		for _, globalCluster := range out.GlobalClusters {
			id := aws.StringValue(globalCluster.GlobalClusterIdentifier)
			input := &docdb.DeleteGlobalClusterInput{
				GlobalClusterIdentifier: globalCluster.GlobalClusterIdentifier,
			}

			log.Printf("[INFO] Deleting DocDB Global Cluster: %s", id)

			_, err := conn.DeleteGlobalCluster(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DocDB Global Cluster (%s): %s", id, err)
				continue
			}

			if err := WaitForGlobalClusterDeletion(context.TODO(), conn, id, GlobalClusterDeleteTimeout); err != nil {
				log.Printf("[ERROR] Failure while waiting for DocDB Global Cluster (%s) to be deleted: %s", id, err)
			}
		}
		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DocDB Global Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving DocDB Global Clusters: %w", err)
	}

	return nil
}
