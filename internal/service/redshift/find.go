// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findClusters(ctx context.Context, conn *redshift.Client, input *redshift.DescribeClustersInput) ([]awstypes.Cluster, error) {
	var output []awstypes.Cluster

	pages := redshift.NewDescribeClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ClusterNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Clusters...)
	}

	return output, nil
}

func findCluster(ctx context.Context, conn *redshift.Client, input *redshift.DescribeClustersInput) (*awstypes.Cluster, error) {
	output, err := findClusters(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findClusterByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.Cluster, error) {
	input := redshift.DescribeClustersInput{
		ClusterIdentifier: aws.String(id),
	}
	output, err := findCluster(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.ClusterIdentifier) != id {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func findScheduledActions(ctx context.Context, conn *redshift.Client, input *redshift.DescribeScheduledActionsInput) ([]awstypes.ScheduledAction, error) {
	var output []awstypes.ScheduledAction

	pages := redshift.NewDescribeScheduledActionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ScheduledActionNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ScheduledActions...)
	}

	return output, nil
}

func findScheduledAction(ctx context.Context, conn *redshift.Client, input *redshift.DescribeScheduledActionsInput) (*awstypes.ScheduledAction, error) {
	output, err := findScheduledActions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findScheduledActionByName(ctx context.Context, conn *redshift.Client, name string) (*awstypes.ScheduledAction, error) {
	input := redshift.DescribeScheduledActionsInput{
		ScheduledActionName: aws.String(name),
	}

	return findScheduledAction(ctx, conn, &input)
}

func findHSMClientCertificates(ctx context.Context, conn *redshift.Client, input *redshift.DescribeHsmClientCertificatesInput) ([]awstypes.HsmClientCertificate, error) {
	var output []awstypes.HsmClientCertificate

	pages := redshift.NewDescribeHsmClientCertificatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.HsmClientCertificateNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.HsmClientCertificates...)
	}

	return output, nil
}

func findHSMClientCertificate(ctx context.Context, conn *redshift.Client, input *redshift.DescribeHsmClientCertificatesInput) (*awstypes.HsmClientCertificate, error) {
	output, err := findHSMClientCertificates(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findHSMClientCertificateByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.HsmClientCertificate, error) {
	input := redshift.DescribeHsmClientCertificatesInput{
		HsmClientCertificateIdentifier: aws.String(id),
	}

	return findHSMClientCertificate(ctx, conn, &input)
}

func findHSMConfigurations(ctx context.Context, conn *redshift.Client, input *redshift.DescribeHsmConfigurationsInput) ([]awstypes.HsmConfiguration, error) {
	var output []awstypes.HsmConfiguration

	pages := redshift.NewDescribeHsmConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.HsmConfigurationNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.HsmConfigurations...)
	}

	return output, nil
}

func findHSMConfiguration(ctx context.Context, conn *redshift.Client, input *redshift.DescribeHsmConfigurationsInput) (*awstypes.HsmConfiguration, error) {
	output, err := findHSMConfigurations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findHSMConfigurationByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.HsmConfiguration, error) {
	input := redshift.DescribeHsmConfigurationsInput{
		HsmConfigurationIdentifier: aws.String(id),
	}

	return findHSMConfiguration(ctx, conn, &input)
}

func findUsageLimits(ctx context.Context, conn *redshift.Client, input *redshift.DescribeUsageLimitsInput) ([]awstypes.UsageLimit, error) {
	var output []awstypes.UsageLimit

	pages := redshift.NewDescribeUsageLimitsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.UsageLimitNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.UsageLimits...)
	}

	return output, nil
}

func findUsageLimit(ctx context.Context, conn *redshift.Client, input *redshift.DescribeUsageLimitsInput) (*awstypes.UsageLimit, error) {
	output, err := findUsageLimits(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findUsageLimitByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.UsageLimit, error) {
	input := redshift.DescribeUsageLimitsInput{
		UsageLimitId: aws.String(id),
	}

	return findUsageLimit(ctx, conn, &input)
}

func findAuthenticationProfiles(ctx context.Context, conn *redshift.Client, input *redshift.DescribeAuthenticationProfilesInput) ([]awstypes.AuthenticationProfile, error) {
	output, err := conn.DescribeAuthenticationProfiles(ctx, input)

	if errs.IsA[*awstypes.AuthenticationProfileNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AuthenticationProfiles, nil
}

func findAuthenticationProfile(ctx context.Context, conn *redshift.Client, input *redshift.DescribeAuthenticationProfilesInput) (*awstypes.AuthenticationProfile, error) {
	output, err := findAuthenticationProfiles(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAuthenticationProfileByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.AuthenticationProfile, error) {
	input := redshift.DescribeAuthenticationProfilesInput{
		AuthenticationProfileName: aws.String(id),
	}

	return findAuthenticationProfile(ctx, conn, &input)
}

func findEventSubscriptions(ctx context.Context, conn *redshift.Client, input *redshift.DescribeEventSubscriptionsInput) ([]awstypes.EventSubscription, error) {
	var output []awstypes.EventSubscription

	pages := redshift.NewDescribeEventSubscriptionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.SubscriptionNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.EventSubscriptionsList...)
	}

	return output, nil
}

func findEventSubscription(ctx context.Context, conn *redshift.Client, input *redshift.DescribeEventSubscriptionsInput) (*awstypes.EventSubscription, error) {
	output, err := findEventSubscriptions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEventSubscriptionByName(ctx context.Context, conn *redshift.Client, name string) (*awstypes.EventSubscription, error) {
	input := redshift.DescribeEventSubscriptionsInput{
		SubscriptionName: aws.String(name),
	}

	return findEventSubscription(ctx, conn, &input)
}

func findClusterSubnetGroups(ctx context.Context, conn *redshift.Client, input *redshift.DescribeClusterSubnetGroupsInput) ([]awstypes.ClusterSubnetGroup, error) {
	var output []awstypes.ClusterSubnetGroup

	pages := redshift.NewDescribeClusterSubnetGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ClusterSubnetGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ClusterSubnetGroups...)
	}

	return output, nil
}

func findClusterSubnetGroup(ctx context.Context, conn *redshift.Client, input *redshift.DescribeClusterSubnetGroupsInput) (*awstypes.ClusterSubnetGroup, error) {
	output, err := findClusterSubnetGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSubnetGroupByName(ctx context.Context, conn *redshift.Client, name string) (*awstypes.ClusterSubnetGroup, error) {
	input := redshift.DescribeClusterSubnetGroupsInput{
		ClusterSubnetGroupName: aws.String(name),
	}

	return findClusterSubnetGroup(ctx, conn, &input)
}

func findEndpointAccesses(ctx context.Context, conn *redshift.Client, input *redshift.DescribeEndpointAccessInput) ([]awstypes.EndpointAccess, error) {
	var output []awstypes.EndpointAccess

	pages := redshift.NewDescribeEndpointAccessPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.EndpointNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.EndpointAccessList...)
	}

	return output, nil
}

func findEndpointAccess(ctx context.Context, conn *redshift.Client, input *redshift.DescribeEndpointAccessInput) (*awstypes.EndpointAccess, error) {
	output, err := findEndpointAccesses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEndpointAccessByName(ctx context.Context, conn *redshift.Client, name string) (*awstypes.EndpointAccess, error) {
	input := redshift.DescribeEndpointAccessInput{
		EndpointName: aws.String(name),
	}

	return findEndpointAccess(ctx, conn, &input)
}

func findEndpointAuthorizations(ctx context.Context, conn *redshift.Client, input *redshift.DescribeEndpointAuthorizationInput) ([]awstypes.EndpointAuthorization, error) {
	var output []awstypes.EndpointAuthorization

	pages := redshift.NewDescribeEndpointAuthorizationPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ClusterNotFoundFault](err) || errs.IsA[*awstypes.EndpointAuthorizationNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.EndpointAuthorizationList...)
	}

	return output, nil
}

func findEndpointAuthorization(ctx context.Context, conn *redshift.Client, input *redshift.DescribeEndpointAuthorizationInput) (*awstypes.EndpointAuthorization, error) {
	output, err := findEndpointAuthorizations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEndpointAuthorizationByTwoPartKey(ctx context.Context, conn *redshift.Client, accountID, clusterID string) (*awstypes.EndpointAuthorization, error) {
	input := redshift.DescribeEndpointAuthorizationInput{
		Account:           aws.String(accountID),
		ClusterIdentifier: aws.String(clusterID),
	}

	return findEndpointAuthorization(ctx, conn, &input)
}

func findPartners(ctx context.Context, conn *redshift.Client, input *redshift.DescribePartnersInput) ([]awstypes.PartnerIntegrationInfo, error) {
	output, err := conn.DescribePartners(ctx, input)

	if errs.IsA[*awstypes.ClusterNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.PartnerIntegrationInfoList, nil
}

func findPartner(ctx context.Context, conn *redshift.Client, input *redshift.DescribePartnersInput) (*awstypes.PartnerIntegrationInfo, error) {
	output, err := findPartners(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPartnerByFourPartKey(ctx context.Context, conn *redshift.Client, accountID, clusterID, databaseName, partnerName string) (*awstypes.PartnerIntegrationInfo, error) {
	input := redshift.DescribePartnersInput{
		AccountId:         aws.String(accountID),
		ClusterIdentifier: aws.String(clusterID),
		DatabaseName:      aws.String(databaseName),
		PartnerName:       aws.String(partnerName),
	}

	return findPartner(ctx, conn, &input)
}

func findClusterSnapshots(ctx context.Context, conn *redshift.Client, input *redshift.DescribeClusterSnapshotsInput) ([]awstypes.Snapshot, error) {
	var output []awstypes.Snapshot

	pages := redshift.NewDescribeClusterSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ClusterNotFoundFault](err) || errs.IsA[*awstypes.ClusterSnapshotNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Snapshots...)
	}

	return output, nil
}

func findClusterSnapshot(ctx context.Context, conn *redshift.Client, input *redshift.DescribeClusterSnapshotsInput) (*awstypes.Snapshot, error) {
	output, err := findClusterSnapshots(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findClusterSnapshotByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.Snapshot, error) {
	input := redshift.DescribeClusterSnapshotsInput{
		SnapshotIdentifier: aws.String(id),
	}
	output, err := findClusterSnapshot(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if status := aws.ToString(output.Status); status == clusterSnapshotStatusDeleted {
		return nil, &retry.NotFoundError{
			Message: status,
		}
	}

	return output, nil
}

func findResourcePolicy(ctx context.Context, conn *redshift.Client, input *redshift.GetResourcePolicyInput) (*awstypes.ResourcePolicy, error) {
	output, err := conn.GetResourcePolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ResourcePolicy, nil
}

func findResourcePolicyByARN(ctx context.Context, conn *redshift.Client, arn string) (*awstypes.ResourcePolicy, error) {
	input := redshift.GetResourcePolicyInput{
		ResourceArn: aws.String(arn),
	}

	return findResourcePolicy(ctx, conn, &input)
}

func findIntegrations(ctx context.Context, conn *redshift.Client, input *redshift.DescribeIntegrationsInput) ([]awstypes.Integration, error) {
	var output []awstypes.Integration

	pages := redshift.NewDescribeIntegrationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.IntegrationNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Integrations...)
	}

	return output, nil
}

func findIntegration(ctx context.Context, conn *redshift.Client, input *redshift.DescribeIntegrationsInput) (*awstypes.Integration, error) {
	output, err := findIntegrations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIntegrationByARN(ctx context.Context, conn *redshift.Client, arn string) (*awstypes.Integration, error) {
	input := redshift.DescribeIntegrationsInput{
		IntegrationArn: aws.String(arn),
	}

	return findIntegration(ctx, conn, &input)
}

func findLoggingStatus(ctx context.Context, conn *redshift.Client, input *redshift.DescribeLoggingStatusInput) (*redshift.DescribeLoggingStatusOutput, error) {
	output, err := conn.DescribeLoggingStatus(ctx, input)

	if errs.IsA[*awstypes.ClusterNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findLoggingStatusByID(ctx context.Context, conn *redshift.Client, clusterID string) (*redshift.DescribeLoggingStatusOutput, error) {
	input := redshift.DescribeLoggingStatusInput{
		ClusterIdentifier: aws.String(clusterID),
	}

	output, err := findLoggingStatus(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if !aws.ToBool(output.LoggingEnabled) {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func findClusterParameterGroups(ctx context.Context, conn *redshift.Client, input *redshift.DescribeClusterParameterGroupsInput) ([]awstypes.ClusterParameterGroup, error) {
	var output []awstypes.ClusterParameterGroup

	pages := redshift.NewDescribeClusterParameterGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ClusterParameterGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ParameterGroups...)
	}

	return output, nil
}

func findClusterParameterGroup(ctx context.Context, conn *redshift.Client, input *redshift.DescribeClusterParameterGroupsInput) (*awstypes.ClusterParameterGroup, error) {
	output, err := findClusterParameterGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findParameterGroupByName(ctx context.Context, conn *redshift.Client, name string) (*awstypes.ClusterParameterGroup, error) {
	input := redshift.DescribeClusterParameterGroupsInput{
		ParameterGroupName: aws.String(name),
	}
	output, err := findClusterParameterGroup(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.ParameterGroupName) != name {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func findSnapshotCopyByID(ctx context.Context, conn *redshift.Client, clusterID string) (*awstypes.ClusterSnapshotCopyStatus, error) {
	output, err := findClusterByID(ctx, conn, clusterID)

	if err != nil {
		return nil, err
	}

	if output.ClusterSnapshotCopyStatus == nil {
		return nil, tfresource.NewEmptyResultError(clusterID)
	}

	return output.ClusterSnapshotCopyStatus, nil
}

func findSnapshotCopyGrants(ctx context.Context, conn *redshift.Client, input *redshift.DescribeSnapshotCopyGrantsInput) ([]awstypes.SnapshotCopyGrant, error) {
	var output []awstypes.SnapshotCopyGrant

	pages := redshift.NewDescribeSnapshotCopyGrantsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.SnapshotCopyGrantNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.SnapshotCopyGrants...)
	}

	return output, nil
}

func findSnapshotCopyGrant(ctx context.Context, conn *redshift.Client, input *redshift.DescribeSnapshotCopyGrantsInput) (*awstypes.SnapshotCopyGrant, error) {
	output, err := findSnapshotCopyGrants(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSnapshotCopyGrantByName(ctx context.Context, conn *redshift.Client, name string) (*awstypes.SnapshotCopyGrant, error) {
	input := redshift.DescribeSnapshotCopyGrantsInput{
		SnapshotCopyGrantName: aws.String(name),
	}

	return findSnapshotCopyGrant(ctx, conn, &input)
}

func findSnapshotSchedules(ctx context.Context, conn *redshift.Client, input *redshift.DescribeSnapshotSchedulesInput) ([]awstypes.SnapshotSchedule, error) {
	var output []awstypes.SnapshotSchedule

	pages := redshift.NewDescribeSnapshotSchedulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ClusterNotFoundFault](err) || errs.IsA[*awstypes.SnapshotScheduleNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.SnapshotSchedules...)
	}

	return output, nil
}

func findSnapshotSchedule(ctx context.Context, conn *redshift.Client, input *redshift.DescribeSnapshotSchedulesInput) (*awstypes.SnapshotSchedule, error) {
	output, err := findSnapshotSchedules(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSnapshotScheduleByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.SnapshotSchedule, error) {
	input := redshift.DescribeSnapshotSchedulesInput{
		ScheduleIdentifier: aws.String(id),
	}

	return findSnapshotSchedule(ctx, conn, &input)
}

func findSnapshotScheduleAssociationByTwoPartKey(ctx context.Context, conn *redshift.Client, clusterID, scheduleID string) (*awstypes.ClusterAssociatedToSchedule, error) {
	input := redshift.DescribeSnapshotSchedulesInput{
		ClusterIdentifier:  aws.String(clusterID),
		ScheduleIdentifier: aws.String(scheduleID),
	}
	output, err := findSnapshotSchedule(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output.AssociatedClusters, func(v awstypes.ClusterAssociatedToSchedule) bool {
		return aws.ToString(v.ClusterIdentifier) == clusterID
	}))
}

func findDataShares(ctx context.Context, conn *redshift.Client, input *redshift.DescribeDataSharesInput) ([]awstypes.DataShare, error) {
	var output []awstypes.DataShare

	pages := redshift.NewDescribeDataSharesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundFault](err) ||
			errs.IsAErrorMessageContains[*awstypes.InvalidDataShareFault](err, "because the ARN doesn't exist.") ||
			errs.IsAErrorMessageContains[*awstypes.InvalidDataShareFault](err, "either doesn't exist or isn't associated with this data consumer") {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.DataShares...)
	}

	return output, nil
}

func findDataShare(ctx context.Context, conn *redshift.Client, input *redshift.DescribeDataSharesInput) (*awstypes.DataShare, error) {
	output, err := findDataShares(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDataShareAuthorizationByTwoPartKey(ctx context.Context, conn *redshift.Client, dataShareARN, consumerID string) (*awstypes.DataShare, error) {
	input := redshift.DescribeDataSharesInput{
		DataShareArn: aws.String(dataShareARN),
	}
	output, err := findDataShare(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	// Verify a share with the expected consumer identifier is present and the
	// status is one of "AUTHORIZED" or "ACTIVE".
	_, err = tfresource.AssertFirstValueResult(tfslices.Filter(output.DataShareAssociations, func(v awstypes.DataShareAssociation) bool {
		return aws.ToString(v.ConsumerIdentifier) == consumerID && (v.Status == awstypes.DataShareStatusAuthorized || v.Status == awstypes.DataShareStatusActive)
	}))

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findDataShareConsumerAssociationByFourPartKey(ctx context.Context, conn *redshift.Client, dataShareARN, associateEntireAccount, consumerARN, consumerRegion string) (*awstypes.DataShare, error) {
	input := redshift.DescribeDataSharesInput{
		DataShareArn: aws.String(dataShareARN),
	}
	output, err := findDataShare(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	// The data share should include an association in an "ACTIVE" status where
	// one of the following is true:
	// - `associate_entire_account` is `true` and `ConsumerIdentifier` matches the
	//   account number of the data share ARN.
	// - `consumer_arn` is set and `ConsumerIdentifier` matches its value.
	// - `consumer_region` is set and `ConsumerRegion` matches its value.
	_, err = tfresource.AssertSingleValueResult(tfslices.Filter(output.DataShareAssociations, func(v awstypes.DataShareAssociation) bool {
		accountIDFromARN := func(s string) string {
			v, err := arn.Parse(s)
			if err != nil {
				return ""
			}
			return v.AccountID
		}

		return v.Status == awstypes.DataShareStatusActive &&
			((associateEntireAccount == "true" && aws.ToString(v.ConsumerIdentifier) == accountIDFromARN(dataShareARN)) ||
				(consumerARN != "" && aws.ToString(v.ConsumerIdentifier) == consumerARN) ||
				(consumerRegion != "" && aws.ToString(v.ConsumerRegion) == consumerRegion))
	}))

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findClusterParameters(ctx context.Context, conn *redshift.Client, input *redshift.DescribeClusterParametersInput) ([]awstypes.Parameter, error) {
	var output []awstypes.Parameter

	pages := redshift.NewDescribeClusterParametersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ClusterParameterGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Parameters...)
	}

	return output, nil
}

func findRedshiftIDCApplications(ctx context.Context, conn *redshift.Client, input *redshift.DescribeRedshiftIdcApplicationsInput) ([]awstypes.RedshiftIdcApplication, error) { // nosemgrep:ci.redshift-in-func-name
	var output []awstypes.RedshiftIdcApplication

	pages := redshift.NewDescribeRedshiftIdcApplicationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.RedshiftIdcApplicationNotExistsFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.RedshiftIdcApplications...)
	}

	return output, nil
}

func findRedshiftIDCApplication(ctx context.Context, conn *redshift.Client, input *redshift.DescribeRedshiftIdcApplicationsInput) (*awstypes.RedshiftIdcApplication, error) { // nosemgrep:ci.redshift-in-func-name
	output, err := findRedshiftIDCApplications(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findRedshiftIDCApplicationByARN(ctx context.Context, conn *redshift.Client, arn string) (*awstypes.RedshiftIdcApplication, error) { // nosemgrep:ci.redshift-in-func-name
	input := redshift.DescribeRedshiftIdcApplicationsInput{
		RedshiftIdcApplicationArn: aws.String(arn),
	}

	return findRedshiftIDCApplication(ctx, conn, &input)
}
