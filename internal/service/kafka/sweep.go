//go:build sweep
// +build sweep

package kafka

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_msk_cluster", &resource.Sweeper{
		Name: "aws_msk_cluster",
		F:    sweepClusters,
		Dependencies: []string{
			"aws_mskconnect_connector",
		},
	})

	resource.AddTestSweepers("aws_msk_configuration", &resource.Sweeper{
		Name: "aws_msk_configuration",
		F:    sweepConfigurations,
		Dependencies: []string{
			"aws_msk_cluster",
		},
	})
}

func sweepClusters(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).KafkaConn
	input := &kafka.ListClustersV2Input{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListClustersV2Pages(input, func(page *kafka.ListClustersV2Output, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ClusterInfoList {
			r := ResourceCluster()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.ClusterArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MSK Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing MSK Clusters (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MSK Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepConfigurations(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).KafkaConn
	var sweeperErrs *multierror.Error

	input := &kafka.ListConfigurationsInput{}

	err = conn.ListConfigurationsPages(input, func(page *kafka.ListConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, configuration := range page.Configurations {
			if configuration == nil {
				continue
			}

			arn := aws.StringValue(configuration.Arn)
			log.Printf("[INFO] Deleting MSK Configuration: %s", arn)

			r := ResourceConfiguration()
			d := r.Data(nil)
			d.SetId(arn)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MSK Configurations sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving MSK Configurations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
