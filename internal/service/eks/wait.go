package eks

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfeks "github.com/hashicorp/terraform-provider-aws/aws/internal/service/eks"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	addonCreatedTimeout = 20 * time.Minute
	addonUpdatedTimeout = 20 * time.Minute
	addonDeletedTimeout = 40 * time.Minute
)

func waitAddonCreated(ctx context.Context, conn *eks.EKS, clusterName, addonName string) (*eks.Addon, error) {
	stateConf := resource.StateChangeConf{
		Pending: []string{eks.AddonStatusCreating, eks.AddonStatusDegraded},
		Target:  []string{eks.AddonStatusActive},
		Refresh: statusAddon(ctx, conn, clusterName, addonName),
		Timeout: addonCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.Addon); ok {
		if status, health := aws.StringValue(output.Status), output.Health; status == eks.AddonStatusCreateFailed && health != nil {
			tfresource.SetLastError(err, tfeks.AddonIssuesError(health.Issues))
		}

		return output, err
	}

	return nil, err
}

func waitAddonDeleted(ctx context.Context, conn *eks.EKS, clusterName, addonName string) (*eks.Addon, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.AddonStatusActive, eks.AddonStatusDeleting},
		Target:  []string{},
		Refresh: statusAddon(ctx, conn, clusterName, addonName),
		Timeout: addonDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.Addon); ok {
		if status, health := aws.StringValue(output.Status), output.Health; status == eks.AddonStatusDeleteFailed && health != nil {
			tfresource.SetLastError(err, tfeks.AddonIssuesError(health.Issues))
		}

		return output, err
	}

	return nil, err
}

func waitAddonUpdateSuccessful(ctx context.Context, conn *eks.EKS, clusterName, addonName, id string) (*eks.Update, error) {
	stateConf := resource.StateChangeConf{
		Pending: []string{eks.UpdateStatusInProgress},
		Target:  []string{eks.UpdateStatusSuccessful},
		Refresh: statusAddonUpdate(ctx, conn, clusterName, addonName, id),
		Timeout: addonUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.Update); ok {
		if status := aws.StringValue(output.Status); status == eks.UpdateStatusCancelled || status == eks.UpdateStatusFailed {
			tfresource.SetLastError(err, tfeks.ErrorDetailsError(output.Errors))
		}

		return output, err
	}

	return nil, err
}

func waitClusterCreated(conn *eks.EKS, name string, timeout time.Duration) (*eks.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.ClusterStatusCreating},
		Target:  []string{eks.ClusterStatusActive},
		Refresh: statusCluster(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*eks.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterDeleted(conn *eks.EKS, name string, timeout time.Duration) (*eks.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.ClusterStatusActive, eks.ClusterStatusDeleting},
		Target:  []string{},
		Refresh: statusCluster(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*eks.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterUpdateSuccessful(conn *eks.EKS, name, id string, timeout time.Duration) (*eks.Update, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.UpdateStatusInProgress},
		Target:  []string{eks.UpdateStatusSuccessful},
		Refresh: statusClusterUpdate(conn, name, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*eks.Update); ok {
		if status := aws.StringValue(output.Status); status == eks.UpdateStatusCancelled || status == eks.UpdateStatusFailed {
			tfresource.SetLastError(err, tfeks.ErrorDetailsError(output.Errors))
		}

		return output, err
	}

	return nil, err
}

func waitFargateProfileCreated(conn *eks.EKS, clusterName, fargateProfileName string, timeout time.Duration) (*eks.FargateProfile, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.FargateProfileStatusCreating},
		Target:  []string{eks.FargateProfileStatusActive},
		Refresh: statusFargateProfile(conn, clusterName, fargateProfileName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*eks.FargateProfile); ok {
		return output, err
	}

	return nil, err
}

func waitFargateProfileDeleted(conn *eks.EKS, clusterName, fargateProfileName string, timeout time.Duration) (*eks.FargateProfile, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.FargateProfileStatusActive, eks.FargateProfileStatusDeleting},
		Target:  []string{},
		Refresh: statusFargateProfile(conn, clusterName, fargateProfileName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*eks.FargateProfile); ok {
		return output, err
	}

	return nil, err
}

func waitNodegroupCreated(ctx context.Context, conn *eks.EKS, clusterName, nodeGroupName string, timeout time.Duration) (*eks.Nodegroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.NodegroupStatusCreating},
		Target:  []string{eks.NodegroupStatusActive},
		Refresh: statusNodegroup(conn, clusterName, nodeGroupName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.Nodegroup); ok {
		if status, health := aws.StringValue(output.Status), output.Health; status == eks.NodegroupStatusCreateFailed && health != nil {
			tfresource.SetLastError(err, tfeks.IssuesError(health.Issues))
		}

		return output, err
	}

	return nil, err
}

func waitNodegroupDeleted(ctx context.Context, conn *eks.EKS, clusterName, nodeGroupName string, timeout time.Duration) (*eks.Nodegroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.NodegroupStatusActive, eks.NodegroupStatusDeleting},
		Target:  []string{},
		Refresh: statusNodegroup(conn, clusterName, nodeGroupName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*eks.Nodegroup); ok {
		if status, health := aws.StringValue(output.Status), output.Health; status == eks.NodegroupStatusDeleteFailed && health != nil {
			tfresource.SetLastError(err, tfeks.IssuesError(health.Issues))
		}

		return output, err
	}

	return nil, err
}

func waitNodegroupUpdateSuccessful(ctx context.Context, conn *eks.EKS, clusterName, nodeGroupName, id string, timeout time.Duration) (*eks.Update, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.UpdateStatusInProgress},
		Target:  []string{eks.UpdateStatusSuccessful},
		Refresh: statusNodegroupUpdate(conn, clusterName, nodeGroupName, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*eks.Update); ok {
		if status := aws.StringValue(output.Status); status == eks.UpdateStatusCancelled || status == eks.UpdateStatusFailed {
			tfresource.SetLastError(err, tfeks.ErrorDetailsError(output.Errors))
		}

		return output, err
	}

	return nil, err
}

func waitOIDCIdentityProviderConfigCreated(ctx context.Context, conn *eks.EKS, clusterName, configName string, timeout time.Duration) (*eks.OidcIdentityProviderConfig, error) {
	stateConf := resource.StateChangeConf{
		Pending: []string{eks.ConfigStatusCreating},
		Target:  []string{eks.ConfigStatusActive},
		Refresh: statusOIDCIdentityProviderConfig(ctx, conn, clusterName, configName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.OidcIdentityProviderConfig); ok {
		return output, err
	}

	return nil, err
}

func waitOIDCIdentityProviderConfigDeleted(ctx context.Context, conn *eks.EKS, clusterName, configName string, timeout time.Duration) (*eks.OidcIdentityProviderConfig, error) {
	stateConf := resource.StateChangeConf{
		Pending: []string{eks.ConfigStatusActive, eks.ConfigStatusDeleting},
		Target:  []string{},
		Refresh: statusOIDCIdentityProviderConfig(ctx, conn, clusterName, configName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.OidcIdentityProviderConfig); ok {
		return output, err
	}

	return nil, err
}
