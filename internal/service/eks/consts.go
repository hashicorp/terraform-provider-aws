// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"time"
)

const (
	identityProviderConfigTypeOIDC = "oidc"
)

const (
	resourcesSecrets = "secrets"
)

func resources_Values() []string {
	return []string{
		resourcesSecrets,
	}
}

const (
	propagationTimeout = 2 * time.Minute
)

const (
	accessEntryTypeEC2          = "EC2"
	accessEntryTypeEC2Linux     = "EC2_LINUX"
	accessEntryTypeEC2Windows   = "EC2_WINDOWS"
	accessEntryTypeFargateLinux = "FARGATE_LINUX"
	accessEntryTypeHybridLinux  = "HYBRID_LINUX"
	accessEntryTypeStandard     = "STANDARD"
)

func accessEntryType_Values() []string {
	return []string{
		accessEntryTypeEC2,
		accessEntryTypeEC2Linux,
		accessEntryTypeEC2Windows,
		accessEntryTypeFargateLinux,
		accessEntryTypeHybridLinux,
		accessEntryTypeStandard,
	}
}

const (
	nodePoolGeneralPurpose = "general-purpose"
	nodePoolSystem         = "system"
)

func nodePoolType_Values() []string {
	return []string{
		nodePoolGeneralPurpose,
		nodePoolSystem,
	}
}
