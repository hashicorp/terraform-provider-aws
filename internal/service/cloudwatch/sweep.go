// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_cloudwatch_composite_alarm", &resource.Sweeper{
		Name: "aws_cloudwatch_composite_alarm",
		F:    sweepCompositeAlarms,
	})

	resource.AddTestSweepers("aws_cloudwatch_dashboard", &resource.Sweeper{
		Name: "aws_cloudwatch_dashboard",
		F:    sweepDashboards,
	})

	resource.AddTestSweepers("aws_cloudwatch_metric_alarm", &resource.Sweeper{
		Name: "aws_cloudwatch_metric_alarm",
		F:    sweepMetricAlarms,
	})

	resource.AddTestSweepers("aws_cloudwatch_metric_stream", &resource.Sweeper{
		Name: "aws_cloudwatch_metric_stream",
		F:    sweepMetricStreams,
	})
}

func sweepCompositeAlarms(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CloudWatchClient(ctx)
	input := &cloudwatch.DescribeAlarmsInput{
		AlarmTypes: []types.AlarmType{types.AlarmTypeCompositeAlarm},
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudwatch.NewDescribeAlarmsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] SkippingCloudWatch Composite Alarm sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CloudWatch Composite Alarms (%s): %w", region, err)
		}

		for _, v := range page.CompositeAlarms {
			r := resourceCompositeAlarm()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.AlarmName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudWatch Composite Alarms (%s): %w", region, err)
	}

	return nil
}

func sweepDashboards(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CloudWatchClient(ctx)
	input := &cloudwatch.ListDashboardsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudwatch.NewListDashboardsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] SkippingCloudWatch Dashboard sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CloudWatch Dashboards (%s): %w", region, err)
		}

		for _, v := range page.DashboardEntries {
			r := resourceDashboard()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DashboardName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudWatch Dashboards (%s): %w", region, err)
	}

	return nil
}

func sweepMetricAlarms(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CloudWatchClient(ctx)
	input := &cloudwatch.DescribeAlarmsInput{
		AlarmTypes: []types.AlarmType{types.AlarmTypeMetricAlarm},
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudwatch.NewDescribeAlarmsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] SkippingCloudWatch Metric Alarm sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CloudWatch Metric Alarms (%s): %w", region, err)
		}

		for _, v := range page.MetricAlarms {
			r := resourceMetricAlarm()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.AlarmName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudWatch Metric Alarms (%s): %w", region, err)
	}

	return nil
}

func sweepMetricStreams(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CloudWatchClient(ctx)
	input := &cloudwatch.ListMetricStreamsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudwatch.NewListMetricStreamsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] SkippingCloudWatch Metric Stream sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CloudWatch Metric Streams (%s): %w", region, err)
		}

		for _, v := range page.Entries {
			r := resourceMetricStream()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudWatch Metric Streams (%s): %w", region, err)
	}

	return nil
}
