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
	resource.AddTestSweepers("aws_location_map", &resource.Sweeper{
		Name: "aws_location_map",
		F:    sweepMaps,
	})

	resource.AddTestSweepers("aws_location_place_index", &resource.Sweeper{
		Name: "aws_location_place_index",
		F:    sweepPlaceIndexes,
	})
}

func sweepMaps(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).LocationConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &locationservice.ListMapsInput{}

	err = conn.ListMapsPages(input, func(page *locationservice.ListMapsOutput, lastPage bool) bool {
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

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Location Service Map for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Location Service Map sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepPlaceIndexes(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).LocationConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &locationservice.ListPlaceIndexesInput{}

	err = conn.ListPlaceIndexesPages(input, func(page *locationservice.ListPlaceIndexesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, entry := range page.Entries {
			r := ResourceMap()
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

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Location Service Place Index for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Location Service Place Index sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
