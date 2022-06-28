//go:build sweep
// +build sweep

package redshift

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).RedshiftConn

	err = conn.DescribeClusterSnapshotsPages(&redshift.DescribeClusterSnapshotsInput{}, func(resp *redshift.DescribeClusterSnapshotsOutput, lastPage bool) bool {
		if len(resp.Snapshots) == 0 {
			log.Print("[DEBUG] No Redshift cluster snapshots to sweep")
			return false
		}

		for _, s := range resp.Snapshots {
			id := aws.StringValue(s.SnapshotIdentifier)

			if !strings.EqualFold(aws.StringValue(s.SnapshotType), "manual") || !strings.EqualFold(aws.StringValue(s.Status), "available") {
				log.Printf("[INFO] Skipping Redshift cluster snapshot: %s", id)
				continue
			}

			log.Printf("[INFO] Deleting Redshift cluster snapshot: %s", id)
			_, err := conn.DeleteClusterSnapshot(&redshift.DeleteClusterSnapshotInput{
				SnapshotIdentifier: s.SnapshotIdentifier,
			})
			if err != nil {
				log.Printf("[ERROR] Failed deleting Redshift cluster snapshot (%s): %s", id, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Redshift Cluster Snapshot sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Redshift cluster snapshots: %w", err)
	}
	return nil
}

func sweepClusters(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).RedshiftConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	err = conn.DescribeClustersPages(&redshift.DescribeClustersInput{}, func(resp *redshift.DescribeClustersOutput, lastPage bool) bool {
		if len(resp.Clusters) == 0 {
			log.Print("[DEBUG] No Redshift clusters to sweep")
			return !lastPage
		}

		for _, c := range resp.Clusters {
			r := ResourceCluster()
			d := r.Data(nil)
			d.Set("skip_final_snapshot", true)
			d.SetId(aws.StringValue(c.ClusterIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Redshift Clusters: %w", err))
		// in case work can be done, don't jump out yet
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Redshift Clusters for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Cluster sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepEventSubscriptions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).RedshiftConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	err = conn.DescribeEventSubscriptionsPages(&redshift.DescribeEventSubscriptionsInput{}, func(page *redshift.DescribeEventSubscriptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, eventSubscription := range page.EventSubscriptionsList {
			r := ResourceEventSubscription()
			d := r.Data(nil)
			d.SetId(aws.StringValue(eventSubscription.CustSubscriptionId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Redshift Event Subscriptions: %w", err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Redshift Event Subscriptions for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Event Subscriptions sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepScheduledActions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).RedshiftConn
	input := &redshift.DescribeScheduledActionsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeScheduledActionsPages(input, func(page *redshift.DescribeScheduledActionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, scheduledAction := range page.ScheduledActions {
			r := ResourceScheduledAction()
			d := r.Data(nil)
			d.SetId(aws.StringValue(scheduledAction.ScheduledActionName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Redshift Scheduled Action sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Redshift Scheduled Actions (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Redshift Scheduled Actions (%s): %w", region, err)
	}

	return nil
}

func sweepSnapshotSchedules(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).RedshiftConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &redshift.DescribeSnapshotSchedulesInput{}
	prefixesToSweep := []string{sweep.ResourcePrefix}

	err = conn.DescribeSnapshotSchedulesPages(input, func(page *redshift.DescribeSnapshotSchedulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, snapshotSchedules := range page.SnapshotSchedules {
			id := aws.StringValue(snapshotSchedules.ScheduleIdentifier)

			for _, prefix := range prefixesToSweep {
				if strings.HasPrefix(id, prefix) {
					r := ResourceSnapshotSchedule()
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
		errs = multierror.Append(errs, fmt.Errorf("error describing Redshift Snapshot Schedules: %w", err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Redshift Snapshot Schedules for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Snapshot Schedules sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepSubnetGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).RedshiftConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &redshift.DescribeClusterSubnetGroupsInput{}

	err = conn.DescribeClusterSubnetGroupsPages(input, func(page *redshift.DescribeClusterSubnetGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, clusterSubnetGroup := range page.ClusterSubnetGroups {
			if clusterSubnetGroup == nil {
				continue
			}

			name := aws.StringValue(clusterSubnetGroup.ClusterSubnetGroupName)

			if name == "default" {
				continue
			}

			r := ResourceSubnetGroup()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Redshift Subnet Groups: %w", err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Redshift Subnet Groups for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Subnet Group sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepHSMClientCertificates(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).RedshiftConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	err = conn.DescribeHsmClientCertificatesPages(&redshift.DescribeHsmClientCertificatesInput{}, func(resp *redshift.DescribeHsmClientCertificatesOutput, lastPage bool) bool {
		if len(resp.HsmClientCertificates) == 0 {
			log.Print("[DEBUG] No Redshift Hsm Client Certificates to sweep")
			return !lastPage
		}

		for _, c := range resp.HsmClientCertificates {
			r := ResourceHSMClientCertificate()
			d := r.Data(nil)
			d.SetId(aws.StringValue(c.HsmClientCertificateIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Redshift Hsm Client Certificates: %w", err))
		// in case work can be done, don't jump out yet
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Redshift Hsm Client Certificates for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Hsm Client Certificate sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepHSMConfigurations(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).RedshiftConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	err = conn.DescribeHsmConfigurationsPages(&redshift.DescribeHsmConfigurationsInput{}, func(resp *redshift.DescribeHsmConfigurationsOutput, lastPage bool) bool {
		if len(resp.HsmConfigurations) == 0 {
			log.Print("[DEBUG] No Redshift Hsm Configurations to sweep")
			return !lastPage
		}

		for _, c := range resp.HsmConfigurations {
			r := ResourceHSMConfiguration()
			d := r.Data(nil)
			d.SetId(aws.StringValue(c.HsmConfigurationIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Redshift Hsm Configurations: %w", err))
		// in case work can be done, don't jump out yet
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Redshift Hsm Configurations for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Hsm Configuration sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepAuthenticationProfiles(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).RedshiftConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &redshift.DescribeAuthenticationProfilesInput{}
	output, err := conn.DescribeAuthenticationProfiles(input)

	if len(output.AuthenticationProfiles) == 0 {
		log.Print("[DEBUG] No Redshift Authentication Profiles to sweep")
	}

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Redshift Authentication Profiles: %w", err))
		// in case work can be done, don't jump out yet
	}

	for _, c := range output.AuthenticationProfiles {
		r := ResourceAuthenticationProfile()
		d := r.Data(nil)
		d.SetId(aws.StringValue(c.AuthenticationProfileName))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Redshift Authentication Profiles for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Authentication Profile sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}
