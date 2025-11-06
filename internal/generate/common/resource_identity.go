// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

type ResourceIdentity struct {
	IsARNIdentity                  bool
	IsCustomInherentRegionIdentity bool
	IsSingletonIdentity            bool
	IdentityAttributeName_         string
}
