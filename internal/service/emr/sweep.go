// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
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
	conn := client.EMRConn(ctx)
	input := &emr.ListClustersInput{
		ClusterStates: aws.StringSlice([]string{emr.ClusterStateBootstrapping, emr.ClusterStateRunning, emr.ClusterStateStarting, emr.ClusterStateWaiting}),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListClustersPagesWithContext(ctx, input, func(page *emr.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Clusters {
			id := aws.StringValue(v.Id)

			_, err := conn.SetTerminationProtectionWithContext(ctx, &emr.SetTerminationProtectionInput{
				JobFlowIds:           aws.StringSlice([]string{id}),
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

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EMR Clusters sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EMR Clusters (%s): %w", region, err)
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

	conn := client.EMRConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error
	input := &emr.ListStudiosInput{}

	err = conn.ListStudiosPagesWithContext(ctx, input, func(page *emr.ListStudiosOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, studio := range page.Studios {
			r := resourceStudio()
			d := r.Data(nil)
			d.SetId(aws.StringValue(studio.StudioId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EMR Studios sweep for %s: %s", region, sweeperErrs)
		return nil
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EMR Studios for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EMR Studios for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
