// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package importer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegionalARN(_ context.Context, rd *schema.ResourceData, identitySpec inttypes.Identity) error {
	attr := identitySpec.Attributes[0]

	if rd.Id() != "" {
		arnARN, err := arn.Parse(rd.Id())
		if err != nil {
			return fmt.Errorf("could not parse import ID %q as ARN: %w", rd.Id(), err)
		}
		rd.Set(attr.ResourceAttributeName(), rd.Id())
		for _, attr := range identitySpec.IdentityDuplicateAttrs {
			setAttribute(rd, attr, rd.Id())
		}

		if region, ok := rd.GetOk(names.AttrRegion); ok {
			if region != arnARN.Region {
				return fmt.Errorf("the region passed for import %q does not match the region %q in the ARN %q", region, arnARN.Region, rd.Id())
			}
		} else {
			rd.Set(names.AttrRegion, arnARN.Region)
		}

		return nil
	}

	identity, err := rd.Identity()
	if err != nil {
		return err
	}

	arnRaw, ok := identity.GetOk(attr.Name())
	if !ok {
		return fmt.Errorf("identity attribute %q is required", attr.Name())
	}

	arnVal, ok := arnRaw.(string)
	if !ok {
		return fmt.Errorf("identity attribute %q: expected string, got %T", attr.Name(), arnRaw)
	}

	arnARN, err := arn.Parse(arnVal)
	if err != nil {
		return fmt.Errorf("identity attribute %q: could not parse %q as ARN: %w", attr.Name(), arnVal, err)
	}

	rd.Set(names.AttrRegion, arnARN.Region)

	rd.Set(attr.ResourceAttributeName(), arnVal)
	for _, attr := range identitySpec.IdentityDuplicateAttrs {
		setAttribute(rd, attr, arnVal)
	}

	return nil
}

func RegionalARNValue(_ context.Context, rd *schema.ResourceData, attrName string, arnValue string) error {
	arnARN, err := arn.Parse(arnValue)
	if err != nil {
		return fmt.Errorf("could not parse %q as ARN: %w", arnValue, err)
	}
	rd.Set(attrName, arnValue)

	if region, ok := rd.GetOk(names.AttrRegion); ok {
		if region != arnARN.Region {
			return fmt.Errorf("the region passed for import %q does not match the region %q in the ARN %q", region, arnARN.Region, rd.Id())
		}
	} else {
		rd.Set(names.AttrRegion, arnARN.Region)
	}

	rd.SetId(arnValue)

	return nil
}

func GlobalARN(_ context.Context, rd *schema.ResourceData, identitySpec inttypes.Identity) error {
	attr := identitySpec.Attributes[0]

	if rd.Id() != "" {
		_, err := arn.Parse(rd.Id())
		if err != nil {
			return fmt.Errorf("could not parse import ID %q as ARN: %w", rd.Id(), err)
		}
		rd.Set(attr.ResourceAttributeName(), rd.Id())
		for _, attr := range identitySpec.IdentityDuplicateAttrs {
			setAttribute(rd, attr, rd.Id())
		}

		return nil
	}

	identity, err := rd.Identity()
	if err != nil {
		return err
	}

	arnRaw, ok := identity.GetOk(attr.Name())
	if !ok {
		return fmt.Errorf("identity attribute %q is required", attr.Name())
	}

	arnVal, ok := arnRaw.(string)
	if !ok {
		return fmt.Errorf("identity attribute %q: expected string, got %T", attr.Name(), arnRaw)
	}

	_, err = arn.Parse(arnVal)
	if err != nil {
		return fmt.Errorf("identity attribute %q: could not parse %q as ARN: %w", attr.Name(), arnVal, err)
	}

	rd.Set(attr.ResourceAttributeName(), arnVal)
	for _, attr := range identitySpec.IdentityDuplicateAttrs {
		setAttribute(rd, attr, arnVal)
	}

	return nil
}

func setAttribute(rd *schema.ResourceData, name string, value any) {
	if name == "id" {
		rd.SetId(value.(string))
		return
	}
	rd.Set(name, value)
}
