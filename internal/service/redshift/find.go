// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findClusters(ctx context.Context, conn *redshift.Client, input *redshift.DescribeClustersInput) ([]awstypes.Cluster, error) {
	var output []awstypes.Cluster

	pages := redshift.NewDescribeClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ClusterNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
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

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return tfresource.AssertSingleValueResult(output)
}

func findClusterByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.Cluster, error) {
	input := &redshift.DescribeClustersInput{
		ClusterIdentifier: aws.String(id),
	}

	output, err := findCluster(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.ClusterIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findScheduledActionByName(ctx context.Context, conn *redshift.Client, name string) (*awstypes.ScheduledAction, error) {
	input := &redshift.DescribeScheduledActionsInput{
		ScheduledActionName: aws.String(name),
	}

	output, err := conn.DescribeScheduledActions(ctx, input)

	if errs.IsA[*awstypes.ScheduledActionNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ScheduledActions) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.ScheduledActions); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return tfresource.AssertSingleValueResult(output.ScheduledActions)
}

func findHSMClientCertificateByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.HsmClientCertificate, error) {
	input := redshift.DescribeHsmClientCertificatesInput{
		HsmClientCertificateIdentifier: aws.String(id),
	}

	output, err := conn.DescribeHsmClientCertificates(ctx, &input)
	if errs.IsA[*awstypes.HsmClientCertificateNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.HsmClientCertificates) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.HsmClientCertificates); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return tfresource.AssertSingleValueResult(output.HsmClientCertificates)
}

func findHSMConfigurationByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.HsmConfiguration, error) {
	input := redshift.DescribeHsmConfigurationsInput{
		HsmConfigurationIdentifier: aws.String(id),
	}

	output, err := conn.DescribeHsmConfigurations(ctx, &input)
	if errs.IsA[*awstypes.HsmConfigurationNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.HsmConfigurations) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.HsmConfigurations); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return tfresource.AssertSingleValueResult(output.HsmConfigurations)
}

func findUsageLimitByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.UsageLimit, error) {
	input := &redshift.DescribeUsageLimitsInput{
		UsageLimitId: aws.String(id),
	}

	output, err := conn.DescribeUsageLimits(ctx, input)

	if errs.IsA[*awstypes.UsageLimitNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.UsageLimits) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.UsageLimits); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return tfresource.AssertSingleValueResult(output.UsageLimits)
}

func findAuthenticationProfileByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.AuthenticationProfile, error) {
	input := redshift.DescribeAuthenticationProfilesInput{
		AuthenticationProfileName: aws.String(id),
	}

	output, err := conn.DescribeAuthenticationProfiles(ctx, &input)

	if errs.IsA[*awstypes.AuthenticationProfileNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.AuthenticationProfiles) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.AuthenticationProfiles); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return tfresource.AssertSingleValueResult(output.AuthenticationProfiles)
}

func findEventSubscriptionByName(ctx context.Context, conn *redshift.Client, name string) (*awstypes.EventSubscription, error) {
	input := &redshift.DescribeEventSubscriptionsInput{
		SubscriptionName: aws.String(name),
	}

	output, err := conn.DescribeEventSubscriptions(ctx, input)

	if errs.IsA[*awstypes.SubscriptionNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.EventSubscriptionsList) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.EventSubscriptionsList); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return tfresource.AssertSingleValueResult(output.EventSubscriptionsList)
}

func findSubnetGroupByName(ctx context.Context, conn *redshift.Client, name string) (*awstypes.ClusterSubnetGroup, error) {
	input := &redshift.DescribeClusterSubnetGroupsInput{
		ClusterSubnetGroupName: aws.String(name),
	}

	output, err := conn.DescribeClusterSubnetGroups(ctx, input)

	if errs.IsA[*awstypes.ClusterSubnetGroupNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ClusterSubnetGroups) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.ClusterSubnetGroups); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return tfresource.AssertSingleValueResult(output.ClusterSubnetGroups)
}

func findEndpointAccessByName(ctx context.Context, conn *redshift.Client, name string) (*awstypes.EndpointAccess, error) {
	input := &redshift.DescribeEndpointAccessInput{
		EndpointName: aws.String(name),
	}

	output, err := conn.DescribeEndpointAccess(ctx, input)

	if errs.IsA[*awstypes.EndpointNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.EndpointAccessList) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.EndpointAccessList); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return tfresource.AssertSingleValueResult(output.EndpointAccessList)
}

func findEndpointAuthorizationByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.EndpointAuthorization, error) {
	account, clusterId, err := DecodeEndpointAuthorizationID(id)
	if err != nil {
		return nil, err
	}

	input := &redshift.DescribeEndpointAuthorizationInput{
		Account:           aws.String(account),
		ClusterIdentifier: aws.String(clusterId),
	}

	output, err := conn.DescribeEndpointAuthorization(ctx, input)

	if errs.IsA[*awstypes.EndpointAuthorizationNotFoundFault](err) || errs.IsA[*awstypes.ClusterNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.EndpointAuthorizationList) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.EndpointAuthorizationList); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return tfresource.AssertSingleValueResult(output.EndpointAuthorizationList)
}

func findPartnerByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.PartnerIntegrationInfo, error) {
	account, clusterId, dbName, partnerName, err := decodePartnerID(id)
	if err != nil {
		return nil, err
	}

	input := &redshift.DescribePartnersInput{
		AccountId:         aws.String(account),
		ClusterIdentifier: aws.String(clusterId),
		DatabaseName:      aws.String(dbName),
		PartnerName:       aws.String(partnerName),
	}

	output, err := conn.DescribePartners(ctx, input)

	if errs.IsA[*awstypes.ClusterNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.PartnerIntegrationInfoList) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.PartnerIntegrationInfoList); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return tfresource.AssertSingleValueResult(output.PartnerIntegrationInfoList)
}

func findClusterSnapshotByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.Snapshot, error) {
	input := &redshift.DescribeClusterSnapshotsInput{
		SnapshotIdentifier: aws.String(id),
	}

	output, err := conn.DescribeClusterSnapshots(ctx, input)

	if errs.IsA[*awstypes.ClusterNotFoundFault](err) || errs.IsA[*awstypes.ClusterSnapshotNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Snapshots) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Snapshots); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	if status := aws.ToString(output.Snapshots[0].Status); status == clusterSnapshotStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return tfresource.AssertSingleValueResult(output.Snapshots)
}

func findResourcePolicyByARN(ctx context.Context, conn *redshift.Client, arn string) (*awstypes.ResourcePolicy, error) {
	input := &redshift.GetResourcePolicyInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetResourcePolicy(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
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
