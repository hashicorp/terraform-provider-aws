package redshift

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindClusterByID(conn *redshift.Redshift, id string) (*redshift.Cluster, error) {
	input := &redshift.DescribeClustersInput{
		ClusterIdentifier: aws.String(id),
	}

	output, err := conn.DescribeClusters(input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Clusters) == 0 || output.Clusters[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Clusters); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.Clusters[0], nil
}

func FindScheduledActionByName(conn *redshift.Redshift, name string) (*redshift.ScheduledAction, error) {
	input := &redshift.DescribeScheduledActionsInput{
		ScheduledActionName: aws.String(name),
	}

	output, err := conn.DescribeScheduledActions(input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeScheduledActionNotFoundFault) {
		return nil, &resource.NotFoundError{
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

func FindScheduleAssociationById(conn *redshift.Redshift, id string) (string, *redshift.ClusterAssociatedToSchedule, error) {
	clusterIdentifier, scheduleIdentifier, err := SnapshotScheduleAssociationParseID(id)
	if err != nil {
		return "", nil, fmt.Errorf("error parsing Redshift Cluster Snapshot Schedule Association ID %s: %s", id, err)
	}

	input := &redshift.DescribeSnapshotSchedulesInput{
		ClusterIdentifier:  aws.String(clusterIdentifier),
		ScheduleIdentifier: aws.String(scheduleIdentifier),
	}
	resp, err := conn.DescribeSnapshotSchedules(input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSnapshotScheduleNotFoundFault) {
		return "", nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return "", nil, err
	}

	if resp.SnapshotSchedules == nil || len(resp.SnapshotSchedules) == 0 {
		return "", nil, tfresource.NewEmptyResultError(input)
	}

	snapshotSchedule := resp.SnapshotSchedules[0]

	if snapshotSchedule == nil {
		return "", nil, tfresource.NewEmptyResultError(input)
	}

	var associatedCluster *redshift.ClusterAssociatedToSchedule
	for _, cluster := range snapshotSchedule.AssociatedClusters {
		if aws.StringValue(cluster.ClusterIdentifier) == clusterIdentifier {
			associatedCluster = cluster
			break
		}
	}

	if associatedCluster == nil {
		return "", nil, tfresource.NewEmptyResultError(input)
	}

	return aws.StringValue(snapshotSchedule.ScheduleIdentifier), associatedCluster, nil
}

func FindHsmClientCertificateByID(conn *redshift.Redshift, id string) (*redshift.HsmClientCertificate, error) {
	input := redshift.DescribeHsmClientCertificatesInput{
		HsmClientCertificateIdentifier: aws.String(id),
	}

	out, err := conn.DescribeHsmClientCertificates(&input)
	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeHsmClientCertificateNotFoundFault) {
		return nil, &resource.NotFoundError{
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

func FindUsageLimitByID(conn *redshift.Redshift, id string) (*redshift.UsageLimit, error) {
	input := &redshift.DescribeUsageLimitsInput{
		UsageLimitId: aws.String(id),
	}

	output, err := conn.DescribeUsageLimits(input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeUsageLimitNotFoundFault) {
		return nil, &resource.NotFoundError{
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

func FindAuthenticationProfileByID(conn *redshift.Redshift, id string) (*redshift.AuthenticationProfile, error) {
	input := redshift.DescribeAuthenticationProfilesInput{
		AuthenticationProfileName: aws.String(id),
	}

	out, err := conn.DescribeAuthenticationProfiles(&input)
	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeAuthenticationProfileNotFoundFault) {
		return nil, &resource.NotFoundError{
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

func FindEventSubscriptionByName(conn *redshift.Redshift, name string) (*redshift.EventSubscription, error) {
	input := &redshift.DescribeEventSubscriptionsInput{
		SubscriptionName: aws.String(name),
	}

	output, err := conn.DescribeEventSubscriptions(input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSubscriptionNotFoundFault) {
		return nil, &resource.NotFoundError{
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
