//go:build sweep
// +build sweep

package emr

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_emr_cluster", &resource.Sweeper{
		Name: "aws_emr_cluster",
		F:    sweepClusters,
	})
}

func sweepClusters(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).EMRConn

	input := &emr.ListClustersInput{
		ClusterStates: []*string{
			aws.String(emr.ClusterStateBootstrapping),
			aws.String(emr.ClusterStateRunning),
			aws.String(emr.ClusterStateStarting),
			aws.String(emr.ClusterStateWaiting),
		},
	}
	err = conn.ListClustersPages(input, func(page *emr.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.Clusters {
			describeClusterInput := &emr.DescribeClusterInput{
				ClusterId: cluster.Id,
			}
			terminateJobFlowsInput := &emr.TerminateJobFlowsInput{
				JobFlowIds: []*string{cluster.Id},
			}
			id := aws.StringValue(cluster.Id)

			log.Printf("[INFO] Deleting EMR Cluster: %s", id)
			_, err = conn.TerminateJobFlows(terminateJobFlowsInput)

			if err != nil {
				log.Printf("[ERROR] Error terminating EMR Cluster (%s): %s", id, err)
			}

			if err := conn.WaitUntilClusterTerminated(describeClusterInput); err != nil {
				log.Printf("[ERROR] Error waiting for EMR Cluster (%s) termination: %s", id, err)
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EMR Cluster sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving EMR Clusters: %w", err)
	}

	return nil
}
