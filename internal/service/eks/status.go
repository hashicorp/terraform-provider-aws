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

func statusFargateProfile(ctx context.Context, conn *eks.EKS, clusterName, fargateProfileName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFargateProfileByClusterNameAndFargateProfileName(ctx, conn, clusterName, fargateProfileName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusNodegroup(ctx context.Context, conn *eks.EKS, clusterName, nodeGroupName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNodegroupByClusterNameAndNodegroupName(ctx, conn, clusterName, nodeGroupName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusNodegroupUpdate(ctx context.Context, conn *eks.EKS, clusterName, nodeGroupName, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNodegroupUpdateByClusterNameNodegroupNameAndID(ctx, conn, clusterName, nodeGroupName, id)

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
