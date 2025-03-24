// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_appstream_directory_config", sweepDirectoryConfigs)
	awsv2.Register("aws_appstream_fleet", sweepFleets)
	awsv2.Register("aws_appstream_image_builder", sweepImageBuilders)
	awsv2.Register("aws_appstream_stack", sweepStacks)
	awsv2.Register("aws_appstream_user", sweepUsers)
}

func sweepDirectoryConfigs(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	if region := client.Region(ctx); region == endpoints.UsWest1RegionID {
		log.Printf("[WARN] Skipping AppStream Directory Config sweep for region: %s", region)
		return nil, nil
	}
	conn := client.AppStreamClient(ctx)
	var input appstream.DescribeDirectoryConfigsInput
	sweepResources := make([]sweep.Sweepable, 0)

	err := describeDirectoryConfigsPages(ctx, conn, &input, func(page *appstream.DescribeDirectoryConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DirectoryConfigs {
			r := resourceDirectoryConfig()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DirectoryName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepFleets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	if region := client.Region(ctx); region == endpoints.UsWest1RegionID {
		log.Printf("[WARN] Skipping AppStream Fleet sweep for region: %s", region)
		return nil, nil
	}
	conn := client.AppStreamClient(ctx)
	var input appstream.DescribeFleetsInput
	sweepResources := make([]sweep.Sweepable, 0)

	err := describeFleetsPages(ctx, conn, &input, func(page *appstream.DescribeFleetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Fleets {
			r := resourceFleet()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepImageBuilders(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	if region := client.Region(ctx); region == endpoints.UsWest1RegionID {
		log.Printf("[WARN] Skipping AppStream Image Builder sweep for region: %s", region)
		return nil, nil
	}
	conn := client.AppStreamClient(ctx)
	var input appstream.DescribeImageBuildersInput
	sweepResources := make([]sweep.Sweepable, 0)

	err := describeImageBuildersPages(ctx, conn, &input, func(page *appstream.DescribeImageBuildersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ImageBuilders {
			r := resourceImageBuilder()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepStacks(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	if region := client.Region(ctx); region == endpoints.UsWest1RegionID {
		log.Printf("[WARN] Skipping AppStream Stack sweep for region: %s", region)
		return nil, nil
	}
	conn := client.AppStreamClient(ctx)
	var input appstream.DescribeStacksInput
	sweepResources := make([]sweep.Sweepable, 0)

	err := describeStacksPages(ctx, conn, &input, func(page *appstream.DescribeStacksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Stacks {
			r := resourceStack()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepUsers(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	if region := client.Region(ctx); region == endpoints.UsWest1RegionID {
		log.Printf("[WARN] Skipping AppStream User sweep for region: %s", region)
		return nil, nil
	}
	conn := client.AppStreamClient(ctx)
	var input appstream.DescribeUsersInput
	sweepResources := make([]sweep.Sweepable, 0)

	err := describeUsersPages(ctx, conn, &input, func(page *appstream.DescribeUsersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Users {
			r := resourceUser()
			d := r.Data(nil)
			d.SetId(userCreateResourceID(aws.ToString(v.UserName), v.AuthenticationType))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}
