// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
)

var _ statecheck.StateCheck = expectFullTagsCheck{}

type entity string

const (
	entityResource   entity = "resource"
	entityDataSource entity = "data source"
)

type expectFullTagsCheck struct {
	base           Base
	knownValue     knownvalue.Check
	servicePackage conns.ServicePackage
	tagSpecFinder  tagSpecFinder
	entity         entity
}

func (e expectFullTagsCheck) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	res, ok := e.base.ResourceFromState(req, resp)
	if !ok {
		return
	}

	sp := e.servicePackage

	tagsSpec := e.tagSpecFinder(ctx, sp, res.Type)

	if tagsSpec == nil {
		resp.Error = fmt.Errorf("no tagging specification found for %s type %s", e.entity, res.Type)
		return
	}

	identifierAttr := tagsSpec.IdentifierAttribute
	if identifierAttr == "" {
		resp.Error = fmt.Errorf("no tag identifier attribute defined for %s type %s", e.entity, res.Type)
		return
	}

	identifier, ok := res.AttributeValues[identifierAttr]
	if !ok {
		resp.Error = fmt.Errorf("attribute %q not found in %s %s", identifierAttr, e.entity, e.base.ResourceAddress())
		return
	}

	ctx = tftags.NewContext(ctx, nil, nil)

	var err error
	if v, ok := sp.(tftags.ServiceTagLister); ok {
		err = v.ListTags(ctx, acctest.Provider.Meta(), identifier.(string)) // Sets tags in Context
	} else if v, ok := sp.(tftags.ResourceTypeTagLister); ok {
		if tagsSpec.ResourceType == "" {
			err = fmt.Errorf("ListTags method for service %s requires ResourceType, but none was set", sp.ServicePackageName())
		} else {
			err = v.ListTags(ctx, acctest.Provider.Meta(), identifier.(string), tagsSpec.ResourceType) // Sets tags in Context
		}
	} else {
		err = fmt.Errorf("no ListTags method found for service %s", sp.ServicePackageName())
	}
	if err != nil {
		resp.Error = fmt.Errorf("listing tags for %s: %s", e.base.ResourceAddress(), err)
		return
	}

	tagsInContext, ok := tftags.FromContext(ctx)
	if !ok {
		resp.Error = fmt.Errorf("Unable to retrieve tags from context")
		return
	}

	var tags tftags.KeyValueTags
	if tagsInContext.TagsOut.IsSome() {
		tags, _ = tagsInContext.TagsOut.Unwrap()
	} else {
		resp.Error = fmt.Errorf("No output tags found in context")
		return
	}

	tags = tags.IgnoreSystem(sp.ServicePackageName())

	tagsMap := tfmaps.ApplyToAllValues(tags.Map(), func(s string) any {
		return s
	})

	if err := e.knownValue.CheckValue(tagsMap); err != nil {
		resp.Error = fmt.Errorf("error checking remote tags for %s: %s", e.base.ResourceAddress(), err) // nosemgrep:ci.semgrep.errors.no-fmt.Errorf-leading-error
		return
	}
}

func ExpectFullResourceTags(servicePackage conns.ServicePackage, resourceAddress string, knownValue knownvalue.Check) expectFullTagsCheck {
	return expectFullTagsCheck{
		base:           NewBase(resourceAddress),
		knownValue:     knownValue,
		servicePackage: servicePackage,
		tagSpecFinder:  findResourceTagSpec,
		entity:         entityResource,
	}
}

func ExpectFullResourceTagsSpecTags(servicePackage conns.ServicePackage, resourceAddress string, tagsSpec *types.ServicePackageResourceTags, knownValue knownvalue.Check) expectFullTagsCheck {
	return expectFullTagsCheck{
		base:           NewBase(resourceAddress),
		knownValue:     knownValue,
		servicePackage: servicePackage,
		tagSpecFinder:  identityTagSpec(tagsSpec),
		entity:         entityResource,
	}
}

func ExpectFullDataSourceTags(servicePackage conns.ServicePackage, resourceAddress string, knownValue knownvalue.Check) expectFullTagsCheck {
	return expectFullTagsCheck{
		base:           NewBase(resourceAddress),
		knownValue:     knownValue,
		servicePackage: servicePackage,
		tagSpecFinder:  findDataSourceTagSpec,
		entity:         entityDataSource,
	}
}

func ExpectFullDataSourceTagsSpecTags(servicePackage conns.ServicePackage, resourceAddress string, tagsSpec *types.ServicePackageResourceTags, knownValue knownvalue.Check) expectFullTagsCheck {
	return expectFullTagsCheck{
		base:           NewBase(resourceAddress),
		knownValue:     knownValue,
		servicePackage: servicePackage,
		tagSpecFinder:  identityTagSpec(tagsSpec),
		entity:         entityDataSource,
	}
}

type tagSpecFinder func(context.Context, conns.ServicePackage, string) *types.ServicePackageResourceTags

func findResourceTagSpec(ctx context.Context, sp conns.ServicePackage, typeName string) (tagsSpec *types.ServicePackageResourceTags) {
	for _, r := range sp.FrameworkResources(ctx) {
		if r.TypeName == typeName {
			tagsSpec = r.Tags
			break
		}
	}
	if tagsSpec == nil {
		for _, r := range sp.SDKResources(ctx) {
			if r.TypeName == typeName {
				tagsSpec = r.Tags
				break
			}
		}
	}
	return tagsSpec
}

func findDataSourceTagSpec(ctx context.Context, sp conns.ServicePackage, typeName string) (tagsSpec *types.ServicePackageResourceTags) {
	for _, r := range sp.FrameworkDataSources(ctx) {
		if r.TypeName == typeName {
			tagsSpec = r.Tags
			break
		}
	}
	if tagsSpec == nil {
		for _, r := range sp.SDKDataSources(ctx) {
			if r.TypeName == typeName {
				tagsSpec = r.Tags
				break
			}
		}
	}
	return tagsSpec
}

func identityTagSpec(tagsSpec *types.ServicePackageResourceTags) tagSpecFinder {
	return func(ctx context.Context, sp conns.ServicePackage, typeName string) *types.ServicePackageResourceTags {
		return tagsSpec
	}
}
