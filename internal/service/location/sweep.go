//go:build sweep
// +build sweep

package location

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).LocationConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &locationservice.ListGeofenceCollectionsInput{}

	err = conn.ListGeofenceCollectionsPagesWithContext(ctx, input, func(page *locationservice.ListGeofenceCollectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, entry := range page.Entries {
			r := ResourceGeofenceCollection()
			d := r.Data(nil)

			id := aws.StringValue(entry.CollectionName)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Location Service Geofence Collection for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Location Service Geofence Collection for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Location Service Geofence Collection sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepMaps(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).LocationConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &locationservice.ListMapsInput{}

	err = conn.ListMapsPagesWithContext(ctx, input, func(page *locationservice.ListMapsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, entry := range page.Entries {
			r := ResourceMap()
			d := r.Data(nil)

			id := aws.StringValue(entry.MapName)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Location Service Map for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Location Service Map for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Location Service Map sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepPlaceIndexes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).LocationConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &locationservice.ListPlaceIndexesInput{}

	err = conn.ListPlaceIndexesPagesWithContext(ctx, input, func(page *locationservice.ListPlaceIndexesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, entry := range page.Entries {
			r := ResourcePlaceIndex()
			d := r.Data(nil)

			id := aws.StringValue(entry.IndexName)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Location Service Place Index for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Location Service Place Index for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Location Service Place Index sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepRouteCalculators(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).LocationConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &locationservice.ListRouteCalculatorsInput{}

	err = conn.ListRouteCalculatorsPagesWithContext(ctx, input, func(page *locationservice.ListRouteCalculatorsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, entry := range page.Entries {
			r := ResourceRouteCalculator()
			d := r.Data(nil)

			id := aws.StringValue(entry.CalculatorName)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Location Service Route Calculator for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Location Service Route Calculator for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Location Service Route Calculator sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepTrackers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).LocationConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &locationservice.ListTrackersInput{}

	err = conn.ListTrackersPagesWithContext(ctx, input, func(page *locationservice.ListTrackersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, entry := range page.Entries {
			r := ResourceTracker()
			d := r.Data(nil)

			id := aws.StringValue(entry.TrackerName)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Location Service Tracker for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Location Service Tracker for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Location Service Tracker sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepTrackerAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).LocationConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &locationservice.ListTrackersInput{}

	err = conn.ListTrackersPagesWithContext(ctx, input, func(page *locationservice.ListTrackersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, entry := range page.Entries {
			input := &locationservice.ListTrackerConsumersInput{
				TrackerName: entry.TrackerName,
			}

			err := conn.ListTrackerConsumersPagesWithContext(ctx, input, func(page *locationservice.ListTrackerConsumersOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, arn := range page.ConsumerArns {
					r := ResourceTrackerAssociation()
					d := r.Data(nil)

					d.SetId(fmt.Sprintf("%s|%s", aws.StringValue(entry.TrackerName), aws.StringValue(arn)))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error listing Location Service Tracker Association for %s: %w", region, err))
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Location Service Tracker for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Location Service Tracker Association for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Location Service Tracker Association sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
