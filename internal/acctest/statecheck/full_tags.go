// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
)

var _ statecheck.StateCheck = expectFullTagsCheck{}

type expectFullTagsCheck struct {
	base           Base
	knownValue     knownvalue.Check
	servicePackage conns.ServicePackage
}

func (e expectFullTagsCheck) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	res, ok := e.base.ResourceFromState(req, resp)
	if !ok {
		return
	}

	var tagsSpec *types.ServicePackageResourceTags
	sp := e.servicePackage
	for _, r := range sp.FrameworkResources(ctx) {
		foo, _ := r.Factory(ctx)
		var metadata resource.MetadataResponse
		foo.Metadata(ctx, resource.MetadataRequest{}, &metadata)
		if res.Type == metadata.TypeName {
			tagsSpec = r.Tags
			break
		}
	}
	if tagsSpec == nil {
		for _, r := range sp.SDKResources(ctx) {
			if res.Type == r.TypeName {
				tagsSpec = r.Tags
				break
			}
		}
	}

	if tagsSpec == nil {
		resp.Error = fmt.Errorf("no tagging specification found for resource type %s", res.Type)
		return
	}

	identifierAttr := tagsSpec.IdentifierAttribute

	identifier, ok := res.AttributeValues[identifierAttr]
	if !ok {
		resp.Error = fmt.Errorf("attribute %q not found in resource %s", identifierAttr, e.base.ResourceAddress())
		return
	}

	ctx = tftags.NewContext(ctx, nil, nil)

	var err error
	if v, ok := sp.(tftags.ServiceTagLister); ok {
		err = v.ListTags(ctx, acctest.Provider.Meta(), identifier.(string)) // Sets tags in Context
	} else if v, ok := sp.(tftags.ResourceTypeTagLister); ok && tagsSpec.ResourceType != "" {
		err = v.ListTags(ctx, acctest.Provider.Meta(), identifier.(string), tagsSpec.ResourceType) // Sets tags in Context
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

	tagsMap := tfmaps.ApplyToAllValues(tags.Map(), func(s string) any {
		return s
	})

	if err := e.knownValue.CheckValue(tagsMap); err != nil {
		resp.Error = fmt.Errorf("error checking remote tags for %s: %s", e.base.ResourceAddress(), err) // nosemgrep:ci.semgrep.errors.no-fmt.Errorf-leading-error
		return
	}
}

func ExpectFullTags(servicePackage conns.ServicePackage, resourceAddress string, knownValue knownvalue.Check) expectFullTagsCheck {
	return expectFullTagsCheck{
		base:           NewBase(resourceAddress),
		knownValue:     knownValue,
		servicePackage: servicePackage,
	}
}
