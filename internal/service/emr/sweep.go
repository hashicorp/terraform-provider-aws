// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_emr_cluster", &resource.Sweeper{
		Name: "aws_emr_cluster",
		F:    sweepClusters,
	})

	resource.AddTestSweepers("aws_emr_studio", &resource.Sweeper{
		Name: "aws_emr_studio",
		F:    sweepStudios,
	})
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EMRClient(ctx)
	input := &emr.ListClustersInput{
		ClusterStates: []awstypes.ClusterState{awstypes.ClusterStateBootstrapping, awstypes.ClusterStateRunning, awstypes.ClusterStateStarting, awstypes.ClusterStateWaiting},
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := emr.NewListClustersPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EMR Clusters sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EMR Clusters (%s): %w", region, err)
		}

		for _, v := range page.Clusters {
			id := aws.ToString(v.Id)

			_, err := conn.SetTerminationProtection(ctx, &emr.SetTerminationProtectionInput{
				JobFlowIds:           []string{id},
				TerminationProtected: aws.Bool(false),
			})

			if err != nil {
				log.Printf("[ERROR] unsetting EMR Cluster (%s) termination protection: %s", id, err)
			}

			r := resourceCluster()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EMR Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepStudios(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.EMRClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error
	input := &emr.ListStudiosInput{}

	pages := emr.NewListStudiosPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EMR Studios sweep for %s: %s", region, sweeperErrs)
			return nil
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EMR Studios for %s: %w", region, err))
		}

		for _, studio := range page.Studios {
			r := resourceStudio()
			d := r.Data(nil)
			d.SetId(aws.ToString(studio.StudioId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EMR Studios for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
