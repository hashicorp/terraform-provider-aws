// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudtrail

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_cloudtrail", &resource.Sweeper{
		Name: "aws_cloudtrail",
		F:    sweepTrails,
	})

	resource.AddTestSweepers("aws_cloudtrail_event_data_store", &resource.Sweeper{
		Name: "aws_cloudtrail_event_data_store",
		F:    sweepEventDataStores,
	})
}

func sweepTrails(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CloudTrailClient(ctx)
	input := &cloudtrail.ListTrailsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudtrail.NewListTrailsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudTrail Trail sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CloudTrail Trails (%s): %w", region, err)
		}

		for _, v := range page.Trails {
			arn := aws.ToString(v.TrailARN)

			if name := aws.ToString(v.Name); name == "AWSMacieTrail-DO-NOT-EDIT" {
				log.Printf("[INFO] Skipping CloudTrail Trail %s", arn)
				continue
			}

			trail, err := findTrailByARN(ctx, conn, arn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error reading CloudTrail Trail (%s): %w", arn, err)
			}

			if aws.ToBool(trail.IsOrganizationTrail) {
				log.Printf("[INFO] Skipping CloudTrail Trail %s: IsOrganizationTrail", arn)
				continue
			}

			r := resourceTrail()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudTrail Trails (%s): %w", region, err)
	}

	return nil
}

func sweepEventDataStores(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CloudTrailClient(ctx)
	input := &cloudtrail.ListEventDataStoresInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudtrail.NewListEventDataStoresPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudTrail Event Data Store sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CloudTrail Event Data Stores (%s): %w", region, err)
		}

		for _, v := range page.EventDataStores {
			r := resourceEventDataStore()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.EventDataStoreArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudTrail Event Data Stores (%s): %w", region, err)
	}

	return nil
}
