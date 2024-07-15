// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findClusters(ctx context.Context, conn *redshift.Redshift, input *redshift.DescribeClustersInput) ([]*redshift.Cluster, error) {
	var output []*redshift.Cluster

	err := conn.DescribeClustersPagesWithContext(ctx, input, func(page *redshift.DescribeClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Clusters {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findCluster(ctx context.Context, conn *redshift.Redshift, input *redshift.DescribeClustersInput) (*redshift.Cluster, error) {
	output, err := findClusters(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func findClusterByID(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.Cluster, error) {
	input := &redshift.DescribeClustersInput{
		ClusterIdentifier: aws.String(id),
	}

	output, err := findCluster(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.ClusterIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findScheduledActionByName(ctx context.Context, conn *redshift.Redshift, name string) (*redshift.ScheduledAction, error) {
	input := &redshift.DescribeScheduledActionsInput{
		ScheduledActionName: aws.String(name),
	}

	output, err := conn.DescribeScheduledActionsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeScheduledActionNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ScheduledActions) == 0 || output.ScheduledActions[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.ScheduledActions); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.ScheduledActions[0], nil
}

func findHSMClientCertificateByID(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.HsmClientCertificate, error) {
	input := redshift.DescribeHsmClientCertificatesInput{
		HsmClientCertificateIdentifier: aws.String(id),
	}

	out, err := conn.DescribeHsmClientCertificatesWithContext(ctx, &input)
	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeHsmClientCertificateNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.HsmClientCertificates) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(out.HsmClientCertificates); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return out.HsmClientCertificates[0], nil
}

func findHSMConfigurationByID(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.HsmConfiguration, error) {
	input := redshift.DescribeHsmConfigurationsInput{
		HsmConfigurationIdentifier: aws.String(id),
	}

	out, err := conn.DescribeHsmConfigurationsWithContext(ctx, &input)
	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeHsmConfigurationNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.HsmConfigurations) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(out.HsmConfigurations); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return out.HsmConfigurations[0], nil
}

func findUsageLimitByID(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.UsageLimit, error) {
	input := &redshift.DescribeUsageLimitsInput{
		UsageLimitId: aws.String(id),
	}

	output, err := conn.DescribeUsageLimitsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeUsageLimitNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.UsageLimits) == 0 || output.UsageLimits[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.UsageLimits); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.UsageLimits[0], nil
}

func findAuthenticationProfileByID(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.AuthenticationProfile, error) {
	input := redshift.DescribeAuthenticationProfilesInput{
		AuthenticationProfileName: aws.String(id),
	}

	out, err := conn.DescribeAuthenticationProfilesWithContext(ctx, &input)
	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeAuthenticationProfileNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.AuthenticationProfiles) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(out.AuthenticationProfiles); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return out.AuthenticationProfiles[0], nil
}

func findEventSubscriptionByName(ctx context.Context, conn *redshift.Redshift, name string) (*redshift.EventSubscription, error) {
	input := &redshift.DescribeEventSubscriptionsInput{
		SubscriptionName: aws.String(name),
	}

	output, err := conn.DescribeEventSubscriptionsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSubscriptionNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.EventSubscriptionsList) == 0 || output.EventSubscriptionsList[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.EventSubscriptionsList); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.EventSubscriptionsList[0], nil
}

func findSubnetGroupByName(ctx context.Context, conn *redshift.Redshift, name string) (*redshift.ClusterSubnetGroup, error) {
	input := &redshift.DescribeClusterSubnetGroupsInput{
		ClusterSubnetGroupName: aws.String(name),
	}

	output, err := conn.DescribeClusterSubnetGroupsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterSubnetGroupNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ClusterSubnetGroups) == 0 || output.ClusterSubnetGroups[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.ClusterSubnetGroups); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.ClusterSubnetGroups[0], nil
}

func findEndpointAccessByName(ctx context.Context, conn *redshift.Redshift, name string) (*redshift.EndpointAccess, error) {
	input := &redshift.DescribeEndpointAccessInput{
		EndpointName: aws.String(name),
	}

	output, err := conn.DescribeEndpointAccessWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeEndpointNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.EndpointAccessList) == 0 || output.EndpointAccessList[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.EndpointAccessList); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.EndpointAccessList[0], nil
}

func findEndpointAuthorizationByID(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.EndpointAuthorization, error) {
	account, clusterId, err := DecodeEndpointAuthorizationID(id)
	if err != nil {
		return nil, err
	}

	input := &redshift.DescribeEndpointAuthorizationInput{
		Account:           aws.String(account),
		ClusterIdentifier: aws.String(clusterId),
	}

	output, err := conn.DescribeEndpointAuthorizationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeEndpointAuthorizationNotFoundFault) || tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.EndpointAuthorizationList) == 0 || output.EndpointAuthorizationList[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.EndpointAuthorizationList); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.EndpointAuthorizationList[0], nil
}

func findPartnerByID(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.PartnerIntegrationInfo, error) {
	account, clusterId, dbName, partnerName, err := DecodePartnerID(id)
	if err != nil {
		return nil, err
	}

	input := &redshift.DescribePartnersInput{
		AccountId:         aws.String(account),
		ClusterIdentifier: aws.String(clusterId),
		DatabaseName:      aws.String(dbName),
		PartnerName:       aws.String(partnerName),
	}

	output, err := conn.DescribePartnersWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.PartnerIntegrationInfoList) == 0 || output.PartnerIntegrationInfoList[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.PartnerIntegrationInfoList); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.PartnerIntegrationInfoList[0], nil
}

func findClusterSnapshotByID(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.Snapshot, error) {
	input := &redshift.DescribeClusterSnapshotsInput{
		SnapshotIdentifier: aws.String(id),
	}

	output, err := conn.DescribeClusterSnapshotsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterNotFoundFault) || tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterSnapshotNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Snapshots) == 0 || output.Snapshots[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Snapshots); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	if status := aws.StringValue(output.Snapshots[0].Status); status == clusterSnapshotStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output.Snapshots[0], nil
}

func findResourcePolicyByARN(ctx context.Context, conn *redshift.Redshift, arn string) (*redshift.ResourcePolicy, error) {
	input := &redshift.GetResourcePolicyInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetResourcePolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeResourceNotFoundFault) {
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
