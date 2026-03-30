// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package importer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegionalInherentRegion(_ context.Context, rd *schema.ResourceData, identitySpec inttypes.Identity) error {
	attr := identitySpec.Attributes[0]
	parser := identitySpec.CustomInherentRegionParser()

	if rd.Id() != "" {
		baseIdentity, err := parser(rd.Id())
		if err != nil {
			return fmt.Errorf("parsing import ID %q: %w", rd.Id(), err)
		}

		rd.Set(attr.ResourceAttributeName(), rd.Id())
		for _, attr := range identitySpec.IdentityDuplicateAttrs {
			setAttribute(rd, attr, rd.Id())
		}

		if region, ok := rd.GetOk(names.AttrRegion); ok {
			if region != baseIdentity.Region {
				return fmt.Errorf("the region passed for import %q does not match the region %q in the import ID %q", region, baseIdentity.Region, rd.Id())
			}
		} else {
			rd.Set(names.AttrRegion, baseIdentity.Region)
		}

		return nil
	}

	identity, err := rd.Identity()
	if err != nil {
		return err
	}

	valueRaw, ok := identity.GetOk(attr.Name())
	if !ok {
		return fmt.Errorf("identity attribute %q is required", attr.Name())
	}

	value, ok := valueRaw.(string)
	if !ok {
		return fmt.Errorf("identity attribute %q: expected string, got %T", attr.Name(), valueRaw)
	}

	foo, err := parser(value)
	if err != nil {
		return fmt.Errorf("identity attribute %q: parsing %q: %w", attr.Name(), value, err)
	}

	rd.Set(names.AttrRegion, foo.Region)

	rd.Set(attr.ResourceAttributeName(), value)
	for _, attr := range identitySpec.IdentityDuplicateAttrs {
		setAttribute(rd, attr, value)
	}

	return nil
}
