package eks

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusAddon(ctx context.Context, conn *eks.EKS, clusterName, addonName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAddonByClusterNameAndAddonName(ctx, conn, clusterName, addonName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusAddonUpdate(ctx context.Context, conn *eks.EKS, clusterName, addonName, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAddonUpdateByClusterNameAddonNameAndID(ctx, conn, clusterName, addonName, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusCluster(conn *eks.EKS, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClusterByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusClusterUpdate(conn *eks.EKS, name, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClusterUpdateByNameAndID(conn, name, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusFargateProfile(conn *eks.EKS, clusterName, fargateProfileName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFargateProfileByClusterNameAndFargateProfileName(conn, clusterName, fargateProfileName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusNodegroup(conn *eks.EKS, clusterName, nodeGroupName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNodegroupByClusterNameAndNodegroupName(conn, clusterName, nodeGroupName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusNodegroupUpdate(conn *eks.EKS, clusterName, nodeGroupName, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNodegroupUpdateByClusterNameNodegroupNameAndID(conn, clusterName, nodeGroupName, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusOIDCIdentityProviderConfig(ctx context.Context, conn *eks.EKS, clusterName, configName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindOIDCIdentityProviderConfigByClusterNameAndConfigName(ctx, conn, clusterName, configName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
