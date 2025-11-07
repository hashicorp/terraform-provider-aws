// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"slices"
	"strings"

	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	namesgen "github.com/hashicorp/terraform-provider-aws/names/generate"
)

type Implementation string

const (
	ImplementationFramework Implementation = "framework"
	ImplementationSDK       Implementation = "sdk"
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

func (r ResourceIdentity) HasResourceIdentity() bool {
	return len(r.IdentityAttributes) > 0 || r.IsARNIdentity || r.IsSingletonIdentity || r.IsCustomInherentRegionIdentity
}

func (d ResourceIdentity) IdentityAttribute() string {
	return namesgen.ConstOrQuote(d.IdentityAttributeName())
}

func (d ResourceIdentity) IdentityAttributeName() string {
	return d.IdentityAttributeName_
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

func ParseResourceIdentity(annotationName string, args Args, implementation Implementation, d *ResourceIdentity) error {
	switch annotationName {
	case "ArnIdentity":
		d.IsARNIdentity = true
		if len(args.Positional) == 0 {
			d.IdentityAttributeName_ = "arn"
		} else {
			d.IdentityAttributeName_ = args.Positional[0]
		}

		var attrs []string
		if attr, ok := args.Keyword["identityDuplicateAttributes"]; ok {
			attrs = strings.Split(attr, ";")
		}
		if implementation == ImplementationSDK {
			attrs = append(attrs, "id")
		}

		// Sort `id` to first position, the rest alphabetically
		slices.SortFunc(attrs, func(a, b string) int {
			if a == "id" {
				return -1
			} else if b == "id" {
				return 1
			} else {
				return strings.Compare(a, b)
			}
		})
		d.IdentityDuplicateAttrNames = slices.Compact(attrs)
	}

	return nil
}
