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

type AWSClient interface {
	AccountID(ctx context.Context) string
	Region(ctx context.Context) string
}

func RegionalSingleton(ctx context.Context, rd *schema.ResourceData, identitySpec inttypes.Identity, client AWSClient) error {
	var region string
	if region = rd.Id(); region != "" {
		if regionAttr, ok := rd.GetOk(names.AttrRegion); ok {
			if region != regionAttr {
				return fmt.Errorf("the region passed for import %q does not match the region %q in the ID", regionAttr, region)
			}
		}
	} else {
		identity, err := rd.Identity()
		if err != nil {
			return err
		}

		if err := validateAccountID(identity, client.AccountID(ctx)); err != nil {
			return err
		}

		regionRaw, ok := identity.GetOk(names.AttrRegion)
		if ok {
			region, ok = regionRaw.(string)
			if !ok {
				return fmt.Errorf("identity attribute %q: expected string, got %T", names.AttrRegion, regionRaw)
			}
		} else {
			region = client.Region(ctx)
		}

		rd.SetId(region)
	}

	rd.Set(names.AttrRegion, region)
	for _, attr := range identitySpec.IdentityDuplicateAttrs {
		setAttribute(rd, attr, region)
	}

	return nil
}

func GlobalSingleton(ctx context.Context, rd *schema.ResourceData, identitySpec inttypes.Identity, client AWSClient) error {
	accountID := client.AccountID(ctx)

	// Historically, we have not validated the Import ID for Global Singletons
	if rd.Id() == "" {
		identity, err := rd.Identity()
		if err != nil {
			return err
		}

		if err := validateAccountID(identity, accountID); err != nil {
			return err
		}
	}

	rd.SetId(accountID)
	for _, attr := range identitySpec.IdentityDuplicateAttrs {
		setAttribute(rd, attr, accountID)
	}

	return nil
}

func validateAccountID(identity *schema.IdentityData, expected string) error {
	accountIDRaw, ok := identity.GetOk(names.AttrAccountID)
	var accountID string
	if ok {
		accountID, ok = accountIDRaw.(string)
		if !ok {
			return fmt.Errorf("identity attribute %q: expected string, got %T", names.AttrAccountID, accountIDRaw)
		}
		if accountID != expected {
			return fmt.Errorf("identity attribute %q: Provider configured with Account ID %q cannot be used to import resources from account %q", names.AttrAccountID, expected, accountID)
		}
	}
	return nil
}
