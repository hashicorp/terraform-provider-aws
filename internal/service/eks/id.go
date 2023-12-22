// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"fmt"
	"strings"
)

const addonResourceIDSeparator = ":"

func AddonCreateResourceID(clusterName, addonName string) string {
	parts := []string{clusterName, addonName}
	id := strings.Join(parts, addonResourceIDSeparator)

	return id
}

func AddonParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, addonResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected cluster-name%[2]saddon-name", id, addonResourceIDSeparator)
}

const accessEntryResourceIDSeparator = ":"

func AccessEntryCreateResourceID(clusterName, principal_arn string) string {
	parts := []string{clusterName, principal_arn}
	id := strings.Join(parts, accessEntryResourceIDSeparator)

	return id
}

func AccessEntryParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, accessEntryResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected cluster-name%[2]sprincipal_arn", id, accessEntryResourceIDSeparator)
}

const associatePolicyResourceIDSeparator = ":"

func AssociatePolicyCreateResourceID(clusterName, principal_arn, policy_arn string) string {
	parts := []string{clusterName, principal_arn, policy_arn}
	id := strings.Join(parts, associatePolicyResourceIDSeparator)

	return id
}

func AssociatePolicyParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, associatePolicyResourceIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected cluster-name%[2]sprincipal_arn", id, associatePolicyResourceIDSeparator)
}

const fargateProfileResourceIDSeparator = ":"

func FargateProfileCreateResourceID(clusterName, fargateProfileName string) string {
	parts := []string{clusterName, fargateProfileName}
	id := strings.Join(parts, fargateProfileResourceIDSeparator)

	return id
}

func FargateProfileParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, fargateProfileResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected cluster-name%[2]sfargate-profile-name", id, fargateProfileResourceIDSeparator)
}

const identityProviderConfigResourceIDSeparator = ":"

func IdentityProviderConfigCreateResourceID(clusterName, configName string) string {
	parts := []string{clusterName, configName}
	id := strings.Join(parts, identityProviderConfigResourceIDSeparator)

	return id
}

func IdentityProviderConfigParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, identityProviderConfigResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected cluster-name%[2]sconfig-name", id, identityProviderConfigResourceIDSeparator)
}

const nodeGroupResourceIDSeparator = ":"

func NodeGroupCreateResourceID(clusterName, nodeGroupName string) string {
	parts := []string{clusterName, nodeGroupName}
	id := strings.Join(parts, nodeGroupResourceIDSeparator)

	return id
}

func NodeGroupParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, nodeGroupResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected cluster-name%[2]snode-group-name", id, nodeGroupResourceIDSeparator)
}
