// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
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
	ctx := sweep.Context(region)
	if region == names.USWest1RegionID {
		log.Printf("[WARN] Skipping AppStream Directory Config sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.AppStreamClient(ctx)
	input := &appstream.DescribeDirectoryConfigsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeDirectoryConfigsPages(ctx, conn, input, func(page *appstream.DescribeDirectoryConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DirectoryConfigs {
			r := ResourceDirectoryConfig()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DirectoryName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppStream Directory Config sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing AppStream Directory Configs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping AppStream Directory Configs (%s): %w", region, err)
	}

	return nil
}

func sweepFleets(region string) error {
	ctx := sweep.Context(region)
	if region == names.USWest1RegionID {
		log.Printf("[WARN] Skipping AppStream Fleet sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.AppStreamClient(ctx)
	input := &appstream.DescribeFleetsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeFleetsPages(ctx, conn, input, func(page *appstream.DescribeFleetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Fleets {
			r := ResourceFleet()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppStream Fleet sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing AppStream Fleets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping AppStream Fleets (%s): %w", region, err)
	}

	return nil
}

func sweepImageBuilders(region string) error {
	ctx := sweep.Context(region)
	if region == names.USWest1RegionID {
		log.Printf("[WARN] Skipping AppStream Image Builder sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.AppStreamClient(ctx)
	input := &appstream.DescribeImageBuildersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeImageBuildersPages(ctx, conn, input, func(page *appstream.DescribeImageBuildersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ImageBuilders {
			r := ResourceImageBuilder()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppStream Image Builder sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing AppStream Image Builders (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping AppStream Image Builders (%s): %w", region, err)
	}

	return nil
}

func sweepStacks(region string) error {
	ctx := sweep.Context(region)
	if region == names.USWest1RegionID {
		log.Printf("[WARN] Skipping AppStream Stack sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.AppStreamClient(ctx)
	input := &appstream.DescribeStacksInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeStacksPages(ctx, conn, input, func(page *appstream.DescribeStacksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Stacks {
			r := ResourceStack()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppStream Stack sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing AppStream Stacks (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping AppStream Stacks (%s): %w", region, err)
	}

	return nil
}
