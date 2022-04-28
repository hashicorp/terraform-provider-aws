//go:build sweep
// +build sweep

package emrcontainers

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_emrcontainers_virtual_cluster", &resource.Sweeper{
		Name: "aws_emrcontainers_virtual_cluster",
		F:    sweepVirtualClusters,
	})
}

func sweepVirtualClusters(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EMRContainersConn
	input := &emrcontainers.ListVirtualClustersInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListVirtualClustersPages(input, func(page *emrcontainers.ListVirtualClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VirtualClusters {
			if aws.StringValue(v.State) == emrcontainers.VirtualClusterStateTerminated {
				continue
			}

			r := ResourceVirtualCluster()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EMR Containers Virtual Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EMR Containers Virtual Clusters (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EMR Containers Virtual Clusters (%s): %w", region, err)
	}

	return nil
}
