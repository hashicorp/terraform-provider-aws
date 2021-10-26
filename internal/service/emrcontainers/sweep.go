//go:build sweep
// +build sweep

package emrcontainers

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_emrcontainers_virtual_cluster", &resource.Sweeper{
		Name: "aws_emrcontainers_virtual_cluster",
		F:    sweepVirtualCluster,
	})
}

func sweepVirtualCluster(region string) error {
	client, err := sweep.SharedRegionalSweeperClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).emrcontainersconn

	input := &emrcontainers.ListVirtualClustersInput{}
	err = conn.ListVirtualClustersPages(input, func(page *emrcontainers.ListVirtualClustersOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, vc := range page.VirtualClusters {
			log.Printf("[INFO] EMR containers virtual cluster: %s", aws.StringValue(vc.Id))
			_, err = conn.DeleteVirtualCluster(&emrcontainers.DeleteVirtualClusterInput{
				Id: vc.Id,
			})

			if err != nil {
				log.Printf("[ERROR] Error deleting containers virtual cluster (%s): %s", aws.StringValue(vc.Id), err)
			}
		}

		return !isLast
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EMR containers virtual cluster sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving EMR containers virtual cluster: %s", err)
	}

	return nil
}
