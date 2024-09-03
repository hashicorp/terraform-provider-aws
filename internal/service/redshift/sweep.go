// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
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
	conn := client.RedshiftClient(ctx)
	input := &redshift.DescribeClusterSnapshotsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := redshift.NewDescribeClusterSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Redshift Snapshot sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Redshift Snapshots (%s): %w", region, err)
		}

		for _, v := range page.Snapshots {
			r := resourceClusterSnapshot()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.SnapshotIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Redshift Snapshots (%s): %w", region, err)
	}

	return nil
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.RedshiftClient(ctx)
	input := &redshift.DescribeClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := redshift.NewDescribeClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Redshift Cluster sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Redshift Clusters (%s): %w", region, err)
		}

		for _, v := range page.Clusters {
			r := resourceCluster()
			d := r.Data(nil)
			d.Set("skip_final_snapshot", true)
			d.SetId(aws.ToString(v.ClusterIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Redshift Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepEventSubscriptions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.RedshiftClient(ctx)
	input := &redshift.DescribeEventSubscriptionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := redshift.NewDescribeEventSubscriptionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Redshift Event Subscription sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Redshift Event Subscriptions (%s): %w", region, err)
		}

		for _, v := range page.EventSubscriptionsList {
			r := resourceEventSubscription()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.CustSubscriptionId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Redshift Event Subscriptions (%s): %w", region, err)
	}

	return nil
}

func sweepScheduledActions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.RedshiftClient(ctx)
	input := &redshift.DescribeScheduledActionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := redshift.NewDescribeScheduledActionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Redshift Scheduled Action sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing Redshift Scheduled Actions (%s): %w", region, err)
		}

		for _, v := range page.ScheduledActions {
			r := resourceScheduledAction()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ScheduledActionName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
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
	conn := client.RedshiftClient(ctx)
	input := &redshift.DescribeSnapshotSchedulesInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	prefixesToSweep := []string{sweep.ResourcePrefix}

	pages := redshift.NewDescribeSnapshotSchedulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Redshift Snapshot Schedule sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing Redshift Snapshot Schedules (%s): %w", region, err)
		}

		for _, v := range page.SnapshotSchedules {
			id := aws.ToString(v.ScheduleIdentifier)

			for _, prefix := range prefixesToSweep {
				if !strings.HasPrefix(id, prefix) {
					r := resourceSnapshotSchedule()
					d := r.Data(nil)
					d.SetId(id)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

					break
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Redshift Snapshot Schedules (%s): %w", region, err)
	}

	return nil
}

func sweepSubnetGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.RedshiftClient(ctx)
	input := &redshift.DescribeClusterSubnetGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := redshift.NewDescribeClusterSubnetGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Redshift Subnet Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing Redshift Subnet Groups (%s): %w", region, err)
		}

		for _, v := range page.ClusterSubnetGroups {
			name := aws.ToString(v.ClusterSubnetGroupName)

			if name == "default" {
				log.Printf("[INFO] Skipping Redshift Subnet Group %s", name)
				continue
			}

			r := resourceSubnetGroup()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Redshift Subnet Groups (%s): %w", region, err)
	}

	return nil
}

func sweepHSMClientCertificates(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	conn := client.RedshiftClient(ctx)
	input := &redshift.DescribeHsmClientCertificatesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := redshift.NewDescribeHsmClientCertificatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Redshift HSM Client Certificate sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing Redshift HSM Client Certificates (%s): %w", region, err)
		}

		for _, v := range page.HsmClientCertificates {
			r := resourceHSMClientCertificate()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.HsmClientCertificateIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Redshift HSM Client Certificates (%s): %w", region, err)
	}

	return nil
}

func sweepHSMConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.RedshiftClient(ctx)
	input := &redshift.DescribeHsmConfigurationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := redshift.NewDescribeHsmConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Redshift HSM Client Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing Redshift HSM Client Configurations (%s): %w", region, err)
		}

		for _, v := range page.HsmConfigurations {
			r := resourceHSMConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.HsmConfigurationIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Redshift HSM Client Configurations (%s): %w", region, err)
	}

	return nil
}

func sweepAuthenticationProfiles(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.RedshiftClient(ctx)
	input := &redshift.DescribeAuthenticationProfilesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeAuthenticationProfiles(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Redshift Authentication Profile sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Redshift Authentication Profiles (%s): %w", region, err)
	}

	for _, v := range output.AuthenticationProfiles {
		r := resourceAuthenticationProfile()
		d := r.Data(nil)
		d.SetId(aws.ToString(v.AuthenticationProfileName))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Redshift Authentication Profiles (%s): %w", region, err)
	}

	return nil
}
