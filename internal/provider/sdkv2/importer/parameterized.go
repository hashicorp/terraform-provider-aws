// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegionalSingleParameterized(ctx context.Context, rd *schema.ResourceData, attrName string, client AWSClient) error {
	if rd.Id() != "" {
		importID := rd.Id()
		if attrName != names.AttrID {
			rd.Set(attrName, importID)
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

	if regionRaw, ok := identity.GetOk(names.AttrRegion); ok {
		if region, ok := regionRaw.(string); !ok {
			return fmt.Errorf("identity attribute %q: expected string, got %T", names.AttrRegion, regionRaw)
		} else {
			rd.Set(names.AttrRegion, region)
		}
	} else {
		rd.Set(names.AttrRegion, client.Region(ctx))
	}

	valRaw, ok := identity.GetOk(attrName)
	if !ok {
		return fmt.Errorf("identity attribute %q is required", attrName)
	}
	val, ok := valRaw.(string)
	if !ok {
		return fmt.Errorf("identity attribute %q: expected string, got %T", attrName, valRaw)
	}
	setAttribute(rd, attrName, val)

	if attrName != names.AttrID {
		rd.SetId(val)
	}

	return nil
}

func GlobalSingleParameterized(ctx context.Context, rd *schema.ResourceData, attrName string, client AWSClient) error {
	if rd.Id() != "" {
		importID := rd.Id()
		if attrName != names.AttrID {
			rd.Set(attrName, importID)
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

	valRaw, ok := identity.GetOk(attrName)
	if !ok {
		return fmt.Errorf("identity attribute %q is required", attrName)
	}
	val, ok := valRaw.(string)
	if !ok {
		return fmt.Errorf("identity attribute %q: expected string, got %T", attrName, valRaw)
	}
	setAttribute(rd, attrName, val)

	if attrName != names.AttrID {
		rd.SetId(val)
	}

	return nil
}
