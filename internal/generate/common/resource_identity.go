// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	namesgen "github.com/hashicorp/terraform-provider-aws/names/generate"
)

type ResourceIdentity struct {
	IsARNIdentity                  bool
	IsCustomInherentRegionIdentity bool
	IsSingletonIdentity            bool
	IdentityAttributeName_         string
	IdentityDuplicateAttrNames     []string
}

func (r ResourceIdentity) HasIdentityDuplicateAttrs() bool {
	return len(r.IdentityDuplicateAttrNames) > 0
}

func (r ResourceIdentity) IdentityDuplicateAttrs() []string {
	return tfslices.ApplyToAll(r.IdentityDuplicateAttrNames, func(s string) string {
		return namesgen.ConstOrQuote(s)
	})
}
