// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func FindAddonByClusterNameAndAddonName(ctx context.Context, conn *eks.EKS, clusterName, addonName string) (*eks.Addon, error) {
	input := &eks.DescribeAddonInput{
		AddonName:   aws.String(addonName),
		ClusterName: aws.String(clusterName),
	}

	output, err := conn.DescribeAddonWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
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

func FindAddonUpdateByClusterNameAddonNameAndID(ctx context.Context, conn *eks.EKS, clusterName, addonName, id string) (*eks.Update, error) {
	input := &eks.DescribeUpdateInput{
		AddonName: aws.String(addonName),
		Name:      aws.String(clusterName),
		UpdateId:  aws.String(id),
	}

	output, err := conn.DescribeUpdateWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
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

func FindAddonVersionByAddonNameAndKubernetesVersion(ctx context.Context, conn *eks.EKS, addonName, kubernetesVersion string, mostRecent bool) (*eks.AddonVersionInfo, error) {
	input := &eks.DescribeAddonVersionsInput{
		AddonName:         aws.String(addonName),
		KubernetesVersion: aws.String(kubernetesVersion),
	}
	var version *eks.AddonVersionInfo

	err := conn.DescribeAddonVersionsPagesWithContext(ctx, input, func(page *eks.DescribeAddonVersionsOutput, lastPage bool) bool {
		if page == nil || len(page.Addons) == 0 {
			return !lastPage
		}

		for _, addon := range page.Addons {
			for i, addonVersion := range addon.AddonVersions {
				if mostRecent && i == 0 {
					version = addonVersion
					return !lastPage
				}
				for _, versionCompatibility := range addonVersion.Compatibilities {
					if aws.BoolValue(versionCompatibility.DefaultVersion) {
						version = addonVersion
						return !lastPage
					}
				}
			}
		}
		return lastPage
	})

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if version == nil || version.AddonVersion == nil {
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return version, nil
}

func FindFargateProfileByClusterNameAndFargateProfileName(ctx context.Context, conn *eks.EKS, clusterName, fargateProfileName string) (*eks.FargateProfile, error) {
	input := &eks.DescribeFargateProfileInput{
		ClusterName:        aws.String(clusterName),
		FargateProfileName: aws.String(fargateProfileName),
	}

	output, err := conn.DescribeFargateProfileWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
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

func FindNodegroupByClusterNameAndNodegroupName(ctx context.Context, conn *eks.EKS, clusterName, nodeGroupName string) (*eks.Nodegroup, error) {
	input := &eks.DescribeNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodeGroupName),
	}

	output, err := conn.DescribeNodegroupWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
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

func FindNodegroupUpdateByClusterNameNodegroupNameAndID(ctx context.Context, conn *eks.EKS, clusterName, nodeGroupName, id string) (*eks.Update, error) {
	input := &eks.DescribeUpdateInput{
		Name:          aws.String(clusterName),
		NodegroupName: aws.String(nodeGroupName),
		UpdateId:      aws.String(id),
	}

	output, err := conn.DescribeUpdateWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
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

func FindOIDCIdentityProviderConfigByClusterNameAndConfigName(ctx context.Context, conn *eks.EKS, clusterName, configName string) (*eks.OidcIdentityProviderConfig, error) {
	input := &eks.DescribeIdentityProviderConfigInput{
		ClusterName: aws.String(clusterName),
		IdentityProviderConfig: &eks.IdentityProviderConfig{
			Name: aws.String(configName),
			Type: aws.String(IdentityProviderConfigTypeOIDC),
		},
	}

	output, err := conn.DescribeIdentityProviderConfigWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
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
