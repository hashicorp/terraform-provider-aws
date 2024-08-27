// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_elastic_beanstalk_application", &resource.Sweeper{
		Name:         "aws_elastic_beanstalk_application",
		Dependencies: []string{"aws_elastic_beanstalk_environment"},
		F:            sweepApplications,
	})

	resource.AddTestSweepers("aws_elastic_beanstalk_environment", &resource.Sweeper{
		Name: "aws_elastic_beanstalk_environment",
		F:    sweepEnvironments,
	})
}

func sweepApplications(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ElasticBeanstalkClient(ctx)
	input := &elasticbeanstalk.DescribeApplicationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeApplications(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Elastic Beanstalk Application sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing Elastic Beanstalk Applications (%s): %w", region, err)
	}

	for _, v := range output.Applications {
		r := resourceApplication()
		d := r.Data(nil)
		d.SetId(aws.ToString(v.ApplicationName))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Elastic Beanstalk Applications (%s): %w", region, err)
	}

	return nil
}

func sweepEnvironments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.ElasticBeanstalkClient(ctx)
	input := &elasticbeanstalk.DescribeEnvironmentsInput{
		IncludeDeleted: aws.Bool(false),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeEnvironmentsPages(ctx, conn, input, func(page *elasticbeanstalk.DescribeEnvironmentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Environments {
			r := resourceEnvironment()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.EnvironmentId))
			d.Set("poll_interval", "10s")
			d.Set("wait_for_ready_timeout", "5m")

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Elastic Beanstalk Environment sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing Elastic Beanstalk Environments (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Elastic Beanstalk Environments (%s): %w", region, err)
	}

	return nil
}
