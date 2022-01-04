//go:build sweep
// +build sweep

package appstream

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_appstream_directory_config", &resource.Sweeper{
		Name: "aws_appstream_directory_config",
		F:    sweepDirectoryConfigs,
	})

	resource.AddTestSweepers("aws_appstream_fleet", &resource.Sweeper{
		Name: "aws_appstream_fleet",
		F:    sweepFleets,
	})

	resource.AddTestSweepers("aws_appstream_image_builder", &resource.Sweeper{
		Name: "aws_appstream_image_builder",
		F:    sweepImageBuilders,
	})

	resource.AddTestSweepers("aws_appstream_stack", &resource.Sweeper{
		Name: "aws_appstream_stack",
		F:    sweepStacks,
	})
}

func sweepDirectoryConfigs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).AppStreamConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &appstream.DescribeDirectoryConfigsInput{}

	err = describeDirectoryConfigsPagesWithContext(context.TODO(), conn, input, func(page *appstream.DescribeDirectoryConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, directoryConfig := range page.DirectoryConfigs {
			if directoryConfig == nil {
				continue
			}

			id := aws.StringValue(directoryConfig.DirectoryName)

			r := ResourceDirectoryConfig()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Appstream Directory Config sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppStream Directory Configs: %w", err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppStream Directory Configs for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppStream Directory Config sweep for %s: %s", region, err)
		return nil // In case we have completed some pages, but had errors
	}

	return errs.ErrorOrNil()
}

func sweepFleets(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).AppStreamConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &appstream.DescribeFleetsInput{}

	err = describeFleetsPagesWithContext(context.TODO(), conn, input, func(page *appstream.DescribeFleetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, fleet := range page.Fleets {
			if fleet == nil {
				continue
			}

			id := aws.StringValue(fleet.Name)

			r := ResourceImageBuilder()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Appstream Fleets sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppStream Fleets: %w", err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppStream Fleets for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppStream Fleets sweep for %s: %s", region, err)
		return nil // In case we have completed some pages, but had errors
	}

	return errs.ErrorOrNil()
}

func sweepImageBuilders(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).AppStreamConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &appstream.DescribeImageBuildersInput{}

	err = describeImageBuildersPagesWithContext(context.TODO(), conn, input, func(page *appstream.DescribeImageBuildersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, imageBuilder := range page.ImageBuilders {
			if imageBuilder == nil {
				continue
			}

			id := aws.StringValue(imageBuilder.Name)

			r := ResourceImageBuilder()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Appstream Image Builders sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppStream Image Builders: %w", err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppStream Image Builders for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppStream Image Builders sweep for %s: %s", region, err)
		return nil // In case we have completed some pages, but had errors
	}

	return errs.ErrorOrNil()
}

func sweepStacks(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).AppStreamConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &appstream.DescribeStacksInput{}

	err = describeStacksPagesWithContext(context.TODO(), conn, input, func(page *appstream.DescribeStacksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, stack := range page.Stacks {
			if stack == nil {
				continue
			}

			id := aws.StringValue(stack.Name)

			r := ResourceImageBuilder()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Appstream Stacks sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppStream Stacks: %w", err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppStream Stacks for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppStream Stacks sweep for %s: %s", region, err)
		return nil // In case we have completed some pages, but had errors
	}

	return errs.ErrorOrNil()
}
