// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func FindAddonByClusterNameAndAddonName(ctx context.Context, conn *eks.Client, clusterName, addonName string) (*types.Addon, error) {
	input := &eks.DescribeAddonInput{
		AddonName:   aws.String(addonName),
		ClusterName: aws.String(clusterName),
	}

	output, err := conn.DescribeAddon(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Addon == nil {
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Addon, nil
}

func FindAddonUpdateByClusterNameAddonNameAndID(ctx context.Context, conn *eks.Client, clusterName, addonName, id string) (*types.Update, error) {
	input := &eks.DescribeUpdateInput{
		AddonName: aws.String(addonName),
		Name:      aws.String(clusterName),
		UpdateId:  aws.String(id),
	}

	output, err := conn.DescribeUpdate(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Update == nil {
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Update, nil
}

func FindAddonVersionByAddonNameAndKubernetesVersion(ctx context.Context, conn *eks.Client, addonName, kubernetesVersion string, mostRecent bool) (*types.AddonVersionInfo, error) {
	input := &eks.DescribeAddonVersionsInput{
		AddonName:         aws.String(addonName),
		KubernetesVersion: aws.String(kubernetesVersion),
	}
	var version *types.AddonVersionInfo

	output, err := conn.DescribeAddonVersions(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Addons == nil {
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return version, nil
}

func FindFargateProfileByClusterNameAndFargateProfileName(ctx context.Context, conn *eks.Client, clusterName, fargateProfileName string) (*types.FargateProfile, error) {
	input := &eks.DescribeFargateProfileInput{
		ClusterName:        aws.String(clusterName),
		FargateProfileName: aws.String(fargateProfileName),
	}

	output, err := conn.DescribeFargateProfile(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FargateProfile == nil {
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.FargateProfile, nil
}

func FindNodegroupByClusterNameAndNodegroupName(ctx context.Context, conn *eks.Client, clusterName, nodeGroupName string) (*types.Nodegroup, error) {
	input := &eks.DescribeNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodeGroupName),
	}

	output, err := conn.DescribeNodegroup(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Nodegroup == nil {
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Nodegroup, nil
}

func FindNodegroupUpdateByClusterNameNodegroupNameAndID(ctx context.Context, conn *eks.Client, clusterName, nodeGroupName, id string) (*types.Update, error) {
	input := &eks.DescribeUpdateInput{
		Name:          aws.String(clusterName),
		NodegroupName: aws.String(nodeGroupName),
		UpdateId:      aws.String(id),
	}

	output, err := conn.DescribeUpdate(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Update == nil {
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Update, nil
}

func FindOIDCIdentityProviderConfigByClusterNameAndConfigName(ctx context.Context, conn *eks.Client, clusterName, configName string) (*types.OidcIdentityProviderConfig, error) {
	input := &eks.DescribeIdentityProviderConfigInput{
		ClusterName: aws.String(clusterName),
		IdentityProviderConfig: &types.IdentityProviderConfig{
			Name: aws.String(configName),
			Type: aws.String(identityProviderConfigTypeOIDC),
		},
	}

	output, err := conn.DescribeIdentityProviderConfig(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.IdentityProviderConfig == nil || output.IdentityProviderConfig.Oidc == nil {
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.IdentityProviderConfig.Oidc, nil
}
