// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_redshift_authentication_profile", sweepAuthenticationProfiles)
	awsv2.Register("aws_redshift_cluster", sweepClusters)
	awsv2.Register("aws_redshift_cluster_snapshot", sweepClusterSnapshots, "aws_redshift_cluster")
	awsv2.Register("aws_redshift_event_subscription", sweepEventSubscriptions)
	awsv2.Register("aws_redshift_hsm_client_certificate", sweepHSMClientCertificates)
	awsv2.Register("aws_redshift_hsm_configuration", sweepHSMConfigurations)
	awsv2.Register("aws_redshift_idc_application", sweepIDCApplications)
	awsv2.Register("aws_redshift_integration", sweepIntegrations)
	awsv2.Register("aws_redshift_scheduled_action", sweepScheduledActions)
	awsv2.Register("aws_redshift_snapshot_schedule", sweepSnapshotSchedules)
	awsv2.Register("aws_redshift_subnet_group", sweepSubnetGroups, "aws_redshift_cluster")
}

func sweepClusterSnapshots(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RedshiftClient(ctx)
	var input redshift.DescribeClusterSnapshotsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := redshift.NewDescribeClusterSnapshotsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Snapshots {
			id := aws.ToString(v.SnapshotIdentifier)

			if typ := aws.ToString(v.SnapshotType); typ != "manual" {
				log.Printf("[INFO] Skipping Redshift Cluster Snapshot %s: SnapshotType=%s", id, typ)
				continue
			}

			r := resourceClusterSnapshot()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RedshiftClient(ctx)
	var input redshift.DescribeClustersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := redshift.NewDescribeClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Clusters {
			r := resourceCluster()
			d := r.Data(nil)
			d.Set("skip_final_snapshot", true)
			d.SetId(aws.ToString(v.ClusterIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepEventSubscriptions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RedshiftClient(ctx)
	var input redshift.DescribeEventSubscriptionsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := redshift.NewDescribeEventSubscriptionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.EventSubscriptionsList {
			r := resourceEventSubscription()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.CustSubscriptionId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepIntegrations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RedshiftClient(ctx)
	var input redshift.DescribeIntegrationsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := redshift.NewDescribeIntegrationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Integrations {
			sweepResources = append(sweepResources, framework.NewSweepResource(newIntegrationResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.IntegrationArn))),
			)
		}
	}

	return sweepResources, nil
}

func sweepScheduledActions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RedshiftClient(ctx)
	var input redshift.DescribeScheduledActionsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := redshift.NewDescribeScheduledActionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ScheduledActions {
			r := resourceScheduledAction()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ScheduledActionName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepSnapshotSchedules(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RedshiftClient(ctx)
	var input redshift.DescribeSnapshotSchedulesInput
	sweepResources := make([]sweep.Sweepable, 0)
	prefixesToSweep := []string{sweep.ResourcePrefix}

	pages := redshift.NewDescribeSnapshotSchedulesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
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

	return sweepResources, nil
}

func sweepSubnetGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RedshiftClient(ctx)
	var input redshift.DescribeClusterSubnetGroupsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := redshift.NewDescribeClusterSubnetGroupsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
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

	return sweepResources, nil
}

func sweepHSMClientCertificates(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RedshiftClient(ctx)
	var input redshift.DescribeHsmClientCertificatesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := redshift.NewDescribeHsmClientCertificatesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.HsmClientCertificates {
			r := resourceHSMClientCertificate()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.HsmClientCertificateIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepHSMConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RedshiftClient(ctx)
	var input redshift.DescribeHsmConfigurationsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := redshift.NewDescribeHsmConfigurationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.HsmConfigurations {
			r := resourceHSMConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.HsmConfigurationIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepAuthenticationProfiles(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RedshiftClient(ctx)
	var input redshift.DescribeAuthenticationProfilesInput
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeAuthenticationProfiles(ctx, &input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.AuthenticationProfiles {
		r := resourceAuthenticationProfile()
		d := r.Data(nil)
		d.SetId(aws.ToString(v.AuthenticationProfileName))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	return sweepResources, nil
}

func sweepIDCApplications(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RedshiftClient(ctx)
	var input redshift.DescribeRedshiftIdcApplicationsInput
	var sweepResources []sweep.Sweepable

	pages := redshift.NewDescribeRedshiftIdcApplicationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.RedshiftIdcApplications {
			sweepResources = append(sweepResources, framework.NewSweepResource(newIDCApplicationResource, client,
				framework.NewAttribute("redshift_idc_application_arn", aws.ToString(v.RedshiftIdcApplicationArn))),
			)
		}
	}

	return sweepResources, nil
}
