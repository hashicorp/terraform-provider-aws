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

func RegionalSingleParameterized(ctx context.Context, rd *schema.ResourceData, identitySpec inttypes.Identity, client AWSClient) error {
	attr := identitySpec.Attributes[len(identitySpec.Attributes)-1]

	if rd.Id() != "" {
		importID := rd.Id()
		if attr.ResourceAttributeName() != names.AttrID {
			rd.Set(attr.ResourceAttributeName(), importID)
		}

		return nil
	}

	identity, err := rd.Identity()
	if err != nil {
		return err
	}

	if err := validateAccountID(identity, client.AccountID(ctx)); err != nil {
		return err
	}

	if err := setRegion(ctx, identity, rd, client); err != nil {
		return err
	}

	valRaw, ok := identity.GetOk(attr.Name())
	if !ok {
		return fmt.Errorf("identity attribute %q is required", attr.Name())
	}

	setAttribute(rd, attr.ResourceAttributeName(), valRaw)

	if attr.ResourceAttributeName() != names.AttrID {
		val, ok := valRaw.(string)
		if !ok {
			return fmt.Errorf("identity attribute %q: expected string, got %T", attr.Name(), valRaw)
		}
		rd.SetId(val)
	}

	return nil
}

func GlobalSingleParameterized(ctx context.Context, rd *schema.ResourceData, identitySpec inttypes.Identity, client AWSClient) error {
	attr := identitySpec.Attributes[len(identitySpec.Attributes)-1]

	if rd.Id() != "" {
		importID := rd.Id()
		if attr.ResourceAttributeName() != names.AttrID {
			rd.Set(attr.ResourceAttributeName(), importID)
		}

		return nil
	}

	identity, err := rd.Identity()
	if err != nil {
		return err
	}

	if err := validateAccountID(identity, client.AccountID(ctx)); err != nil {
		return err
	}

	valRaw, ok := identity.GetOk(attr.Name())
	if !ok {
		return fmt.Errorf("identity attribute %q is required", attr.Name())
	}

	setAttribute(rd, attr.ResourceAttributeName(), valRaw)

	if attr.ResourceAttributeName() != names.AttrID {
		val, ok := valRaw.(string)
		if !ok {
			return fmt.Errorf("identity attribute %q: expected string, got %T", attr.Name(), valRaw)
		}
		rd.SetId(val)
	}

	return nil
}

func RegionalMultipleParameterized(ctx context.Context, rd *schema.ResourceData, identitySpec inttypes.Identity, importSpec inttypes.SDKv2Import, client AWSClient) error {
	if rd.Id() != "" {
		id, parts, err := importSpec.ImportID.Parse(rd.Id())
		if err != nil {
			return err
		}

		rd.SetId(id)
		for attr, val := range parts {
			rd.Set(attr, val)
		}
	} else {
		identity, err := rd.Identity()
		if err != nil {
			return err
		}

		if err := validateAccountID(identity, client.AccountID(ctx)); err != nil {
			return err
		}

		if err := setRegion(ctx, identity, rd, client); err != nil {
			return err
		}

		for _, attr := range identitySpec.Attributes {
			switch attr.Name() {
			case names.AttrAccountID, names.AttrRegion:
				// Do nothing

			default:
				valRaw, ok := identity.GetOk(attr.Name())
				if attr.Required() && !ok {
					return fmt.Errorf("identity attribute %q is required", attr.Name())
				}

				setAttribute(rd, attr.ResourceAttributeName(), valRaw)
			}
		}

		rd.SetId(importSpec.ImportID.Create(rd))
	}

	return nil
}

func GlobalMultipleParameterized(ctx context.Context, rd *schema.ResourceData, identitySpec inttypes.Identity, importSpec inttypes.SDKv2Import, client AWSClient) error {
	if rd.Id() != "" {
		id, parts, err := importSpec.ImportID.Parse(rd.Id())
		if err != nil {
			return err
		}

		rd.SetId(id)
		for attr, val := range parts {
			rd.Set(attr, val)
		}
	} else {
		identity, err := rd.Identity()
		if err != nil {
			return err
		}

		if err := validateAccountID(identity, client.AccountID(ctx)); err != nil {
			return err
		}

		for _, attr := range identitySpec.Attributes {
			switch attr.Name() {
			case names.AttrAccountID:
				// Do nothing

			default:
				valRaw, ok := identity.GetOk(attr.Name())
				if attr.Required() && !ok {
					return fmt.Errorf("identity attribute %q is required", attr.Name())
				}

				setAttribute(rd, attr.ResourceAttributeName(), valRaw)
			}
		}

		rd.SetId(importSpec.ImportID.Create(rd))
	}

	return nil
}

func setRegion(ctx context.Context, identity *schema.IdentityData, rd *schema.ResourceData, client AWSClient) error {
	if regionRaw, ok := identity.GetOk(names.AttrRegion); ok {
		if region, ok := regionRaw.(string); !ok {
			return fmt.Errorf("identity attribute %q: expected string, got %T", names.AttrRegion, regionRaw)
		} else {
			rd.Set(names.AttrRegion, region)
		}
	} else {
		rd.Set(names.AttrRegion, client.Region(ctx))
	}
	return nil
}
