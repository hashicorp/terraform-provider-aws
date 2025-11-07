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
	IdentityAttributes             []IdentityAttribute
	MutableIdentity                bool
	IdentityVersion                int64
	CustomInherentRegionParser     string
}

func (r ResourceIdentity) HasIdentityDuplicateAttrs() bool {
	return len(r.IdentityDuplicateAttrNames) > 0
}

func (r ResourceIdentity) IdentityDuplicateAttrs() []string {
	return tfslices.ApplyToAll(r.IdentityDuplicateAttrNames, func(s string) string {
		return namesgen.ConstOrQuote(s)
	})
}

type IdentityAttribute struct {
	Name_                  string
	Optional               bool
	ResourceAttributeName_ string
	TestNotNull            bool
}

func (a IdentityAttribute) Name() string {
	return namesgen.ConstOrQuote(a.Name_)
}

func (a IdentityAttribute) ResourceAttributeName() string {
	return namesgen.ConstOrQuote(a.ResourceAttributeName_)
}
