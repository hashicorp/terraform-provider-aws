// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusAddon(ctx context.Context, client *eks.Client, clusterName, addonName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAddonByClusterNameAndAddonName(ctx, client, clusterName, addonName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusAddonUpdate(ctx context.Context, client *eks.Client, clusterName, addonName, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAddonUpdateByClusterNameAddonNameAndID(ctx, client, clusterName, addonName, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusFargateProfile(ctx context.Context, client *eks.Client, clusterName, fargateProfileName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFargateProfileByClusterNameAndFargateProfileName(ctx, client, clusterName, fargateProfileName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusNodegroup(ctx context.Context, client *eks.Client, clusterName, nodeGroupName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNodegroupByClusterNameAndNodegroupName(ctx, client, clusterName, nodeGroupName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusNodegroupUpdate(ctx context.Context, client *eks.Client, clusterName, nodeGroupName, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNodegroupUpdateByClusterNameNodegroupNameAndID(ctx, client, clusterName, nodeGroupName, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusOIDCIdentityProviderConfig(ctx context.Context, client *eks.Client, clusterName, configName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindOIDCIdentityProviderConfigByClusterNameAndConfigName(ctx, client, clusterName, configName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}
