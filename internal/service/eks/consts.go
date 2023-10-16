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
	propagationTimeout = 2 * time.Minute
)
