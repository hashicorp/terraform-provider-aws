// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/location"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_location_geofence_collection", &resource.Sweeper{
		Name: "aws_location_geofence_collection",
		F:    sweepGeofenceCollections,
	})

	resource.AddTestSweepers("aws_location_map", &resource.Sweeper{
		Name: "aws_location_map",
		F:    sweepMaps,
	})

	resource.AddTestSweepers("aws_location_place_index", &resource.Sweeper{
		Name: "aws_location_place_index",
		F:    sweepPlaceIndexes,
	})

	resource.AddTestSweepers("aws_location_route_calculator", &resource.Sweeper{
		Name: "aws_location_route_calculator",
		F:    sweepRouteCalculators,
	})

	resource.AddTestSweepers("aws_location_tracker", &resource.Sweeper{
		Name: "aws_location_tracker",
		F:    sweepTrackers,
	})

	resource.AddTestSweepers("aws_location_tracker_association", &resource.Sweeper{
		Name: "aws_location_tracker_association",
		F:    sweepTrackerAssociations,
	})
}

func sweepGeofenceCollections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.LocationClient(ctx)
	input := &location.ListGeofenceCollectionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := location.NewListGeofenceCollectionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Location Service Geofence Collection sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Location Service Geofence Collection for %s: %w", region, err)
		}

		for _, entry := range page.Entries {
			r := ResourceGeofenceCollection()
			d := r.Data(nil)

			id := aws.ToString(entry.CollectionName)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Location Service Geofence Collection for %s: %w", region, err)
	}

	return nil
}

func sweepMaps(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.LocationClient(ctx)

	input := &location.ListMapsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := location.NewListMapsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Location Service Map sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Location Service Map for %s: %w", region, err)
		}

		for _, entry := range page.Entries {
			r := ResourceMap()
			d := r.Data(nil)

			id := aws.ToString(entry.MapName)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Location Service Map for %s: %w", region, err)
	}

	return nil
}

func sweepPlaceIndexes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.LocationClient(ctx)

	input := &location.ListPlaceIndexesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := location.NewListPlaceIndexesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Location Service Place Index sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Location Service Place Index for %s: %w", region, err)
		}

		for _, entry := range page.Entries {
			r := ResourcePlaceIndex()
			d := r.Data(nil)

			id := aws.ToString(entry.IndexName)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Location Service Place Index for %s: %w", region, err)
	}

	return nil
}

func sweepRouteCalculators(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.LocationClient(ctx)

	input := &location.ListRouteCalculatorsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := location.NewListRouteCalculatorsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Location Service Route Calculator sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Location Service Route Calculator for %s: %w", region, err)
		}

		for _, entry := range page.Entries {
			r := ResourceRouteCalculator()
			d := r.Data(nil)

			id := aws.ToString(entry.CalculatorName)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Location Service Route Calculator for %s: %w", region, err)
	}

	return nil
}

func sweepTrackers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.LocationClient(ctx)

	input := &location.ListTrackersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := location.NewListTrackersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Location Service Tracker sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Location Service Tracker for %s: %w", region, err)
		}

		for _, entry := range page.Entries {
			r := ResourceTracker()
			d := r.Data(nil)

			id := aws.ToString(entry.TrackerName)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Location Service Tracker for %s: %w", region, err)
	}

	return nil
}

func sweepTrackerAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.LocationClient(ctx)

	input := &location.ListTrackersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := location.NewListTrackersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Location Service Tracker Association sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Location Service Tracker for %s: %w", region, err)
		}

		for _, entry := range page.Entries {
			input := &location.ListTrackerConsumersInput{
				TrackerName: entry.TrackerName,
			}

			consumerPages := location.NewListTrackerConsumersPaginator(conn, input)

			for consumerPages.HasMorePages() {
				consumerPage, err := consumerPages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					log.Printf("[WARN] Skipping Location Service Tracker Association sweep for %s: %s", region, err)
					return nil
				}

				if err != nil {
					return fmt.Errorf("error listing Location Service Tracker Association for %s: %w", region, err)
				}

				for _, arn := range consumerPage.ConsumerArns {
					r := ResourceTrackerAssociation()
					d := r.Data(nil)

					d.SetId(fmt.Sprintf("%s|%s", aws.ToString(entry.TrackerName), arn))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Location Service Tracker Association for %s: %w", region, err)
	}

	return nil
}
