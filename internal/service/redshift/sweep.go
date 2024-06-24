// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_redshift_cluster_snapshot", &resource.Sweeper{
		Name: "aws_redshift_cluster_snapshot",
		F:    sweepClusterSnapshots,
		Dependencies: []string{
			"aws_redshift_cluster",
		},
	})

	resource.AddTestSweepers("aws_redshift_cluster", &resource.Sweeper{
		Name: "aws_redshift_cluster",
		F:    sweepClusters,
	})

	resource.AddTestSweepers("aws_redshift_hsm_client_certificate", &resource.Sweeper{
		Name: "aws_redshift_hsm_client_certificate",
		F:    sweepHSMClientCertificates,
	})

	resource.AddTestSweepers("aws_redshift_hsm_configuration", &resource.Sweeper{
		Name: "aws_redshift_hsm_configuration",
		F:    sweepHSMConfigurations,
	})

	resource.AddTestSweepers("aws_redshift_authentication_profile", &resource.Sweeper{
		Name: "aws_redshift_authentication_profile",
		F:    sweepAuthenticationProfiles,
	})

	resource.AddTestSweepers("aws_redshift_event_subscription", &resource.Sweeper{
		Name: "aws_redshift_event_subscription",
		F:    sweepEventSubscriptions,
	})

	resource.AddTestSweepers("aws_redshift_scheduled_action", &resource.Sweeper{
		Name: "aws_redshift_scheduled_action",
		F:    sweepScheduledActions,
	})

	resource.AddTestSweepers("aws_redshift_snapshot_schedule", &resource.Sweeper{
		Name: "aws_redshift_snapshot_schedule",
		F:    sweepSnapshotSchedules,
	})

	resource.AddTestSweepers("aws_redshift_subnet_group", &resource.Sweeper{
		Name: "aws_redshift_subnet_group",
		F:    sweepSubnetGroups,
		Dependencies: []string{
			"aws_redshift_cluster",
		},
	})
}

func sweepClusterSnapshots(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.RedshiftConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	err = conn.DescribeClusterSnapshotsPagesWithContext(ctx, &redshift.DescribeClusterSnapshotsInput{}, func(page *redshift.DescribeClusterSnapshotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Snapshots {
			r := resourceClusterSnapshot()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.SnapshotIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("describing Redshift Snapshots: %w", err))
		// in case work can be done, don't jump out yet
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Redshift Snapshots for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Snapshots sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	conn := client.RedshiftConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	err = conn.DescribeClustersPagesWithContext(ctx, &redshift.DescribeClustersInput{}, func(page *redshift.DescribeClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Clusters {
			r := resourceCluster()
			d := r.Data(nil)
			d.Set("skip_final_snapshot", true)
			d.SetId(aws.StringValue(v.ClusterIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("describing Redshift Clusters: %w", err))
		// in case work can be done, don't jump out yet
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Redshift Clusters for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Cluster sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepEventSubscriptions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.RedshiftConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	err = conn.DescribeEventSubscriptionsPagesWithContext(ctx, &redshift.DescribeEventSubscriptionsInput{}, func(page *redshift.DescribeEventSubscriptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.EventSubscriptionsList {
			r := resourceEventSubscription()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.CustSubscriptionId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("describing Redshift Event Subscriptions: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Redshift Event Subscriptions for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Event Subscriptions sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepScheduledActions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.RedshiftConn(ctx)
	input := &redshift.DescribeScheduledActionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeScheduledActionsPagesWithContext(ctx, input, func(page *redshift.DescribeScheduledActionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ScheduledActions {
			r := resourceScheduledAction()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.ScheduledActionName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Redshift Scheduled Action sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing Redshift Scheduled Actions (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Redshift Scheduled Actions (%s): %w", region, err)
	}

	return nil
}

func sweepSnapshotSchedules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	conn := client.RedshiftConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &redshift.DescribeSnapshotSchedulesInput{}
	prefixesToSweep := []string{sweep.ResourcePrefix}

	err = conn.DescribeSnapshotSchedulesPagesWithContext(ctx, input, func(page *redshift.DescribeSnapshotSchedulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.SnapshotSchedules {
			id := aws.StringValue(v.ScheduleIdentifier)

			for _, prefix := range prefixesToSweep {
				if strings.HasPrefix(id, prefix) {
					r := resourceSnapshotSchedule()
					d := r.Data(nil)
					d.SetId(id)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

					break
				}
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("describing Redshift Snapshot Schedules: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Redshift Snapshot Schedules for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Snapshot Schedules sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepSubnetGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.RedshiftConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &redshift.DescribeClusterSubnetGroupsInput{}

	err = conn.DescribeClusterSubnetGroupsPagesWithContext(ctx, input, func(page *redshift.DescribeClusterSubnetGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ClusterSubnetGroups {
			if v == nil {
				continue
			}

			name := aws.StringValue(v.ClusterSubnetGroupName)

			if name == "default" {
				continue
			}

			r := resourceSubnetGroup()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("describing Redshift Subnet Groups: %w", err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Redshift Subnet Groups for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Subnet Group sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepHSMClientCertificates(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	conn := client.RedshiftConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	err = conn.DescribeHsmClientCertificatesPagesWithContext(ctx, &redshift.DescribeHsmClientCertificatesInput{}, func(page *redshift.DescribeHsmClientCertificatesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.HsmClientCertificates {
			r := resourceHSMClientCertificate()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.HsmClientCertificateIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("describing Redshift Hsm Client Certificates: %w", err))
		// in case work can be done, don't jump out yet
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Redshift Hsm Client Certificates for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Hsm Client Certificate sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepHSMConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	conn := client.RedshiftConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	err = conn.DescribeHsmConfigurationsPagesWithContext(ctx, &redshift.DescribeHsmConfigurationsInput{}, func(page *redshift.DescribeHsmConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.HsmConfigurations {
			r := resourceHSMConfiguration()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.HsmConfigurationIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("describing Redshift Hsm Configurations: %w", err))
		// in case work can be done, don't jump out yet
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Redshift Hsm Configurations for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Hsm Configuration sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepAuthenticationProfiles(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	conn := client.RedshiftConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &redshift.DescribeAuthenticationProfilesInput{}
	output, err := conn.DescribeAuthenticationProfilesWithContext(ctx, input)

	if len(output.AuthenticationProfiles) == 0 {
		log.Print("[DEBUG] No Redshift Authentication Profiles to sweep")
	}

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("describing Redshift Authentication Profiles: %w", err))
		// in case work can be done, don't jump out yet
	}

	for _, v := range output.AuthenticationProfiles {
		r := resourceAuthenticationProfile()
		d := r.Data(nil)
		d.SetId(aws.StringValue(v.AuthenticationProfileName))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Redshift Authentication Profiles for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Authentication Profile sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}
