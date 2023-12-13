// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"time"
)

const (
	IdentityProviderConfigTypeOIDC = "oidc"
)

const (
	ResourcesSecrets = "secrets"
)

func Resources_Values() []string {
	return []string{
		ResourcesSecrets,
	}
}

const (
	propagationTimeout = 2 * time.Minute
)
