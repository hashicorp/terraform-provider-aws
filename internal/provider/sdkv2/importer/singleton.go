// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type Client interface {
	AccountID(ctx context.Context) string
	Region(ctx context.Context) string
}

var _ Client = &conns.AWSClient{}

func RegionalSingleton(ctx context.Context, rd *schema.ResourceData, meta any) error {
	if rd.Id() != "" {
		if region, ok := rd.GetOk(names.AttrRegion); ok {
			if region != rd.Id() {
				return fmt.Errorf("the region passed for import %q does not match the region %q in the ID", region, rd.Id())
			}
		} else {
			rd.Set(names.AttrRegion, rd.Id())
		}

		return nil
	}

	identity, err := rd.Identity()
	if err != nil {
		return err
	}

	client := meta.(Client)

	accountIDRaw, ok := identity.GetOk(names.AttrAccountID)
	var accountID string
	if ok {
		accountID, ok = accountIDRaw.(string)
		if !ok {
			return fmt.Errorf("identity attribute %q: expected string, got %T", names.AttrAccountID, accountIDRaw)
		}
		if accountID != client.AccountID(ctx) {
			return fmt.Errorf("identity attribute %q: Provider configured with Account ID %q cannot be used to import resources from account %q", names.AttrAccountID, client.AccountID(ctx), accountID)
		}
	}

	regionRaw, ok := identity.GetOk(names.AttrRegion)
	var region string
	if ok {
		region, ok = regionRaw.(string)
		if !ok {
			return fmt.Errorf("identity attribute %q: expected string, got %T", names.AttrRegion, regionRaw)
		}
	} else {
		region = client.Region(ctx)
	}

	rd.Set(names.AttrRegion, region)
	rd.SetId(region)

	return nil
}

func GlobalSingleton(ctx context.Context, rd *schema.ResourceData, meta any) error {
	client := meta.(Client)

	if rd.Id() != "" {
		// Historically, we have not validated the Import ID for Global Singletons
		rd.SetId(client.AccountID(ctx))

		return nil
	}

	identity, err := rd.Identity()
	if err != nil {
		return err
	}

	accountIDRaw, ok := identity.GetOk(names.AttrAccountID)
	var accountID string
	if ok {
		accountID, ok = accountIDRaw.(string)
		if !ok {
			return fmt.Errorf("identity attribute %q: expected string, got %T", names.AttrAccountID, accountIDRaw)
		}
		if accountID != client.AccountID(ctx) {
			return fmt.Errorf("Unable to import\n\nidentity attribute %q: Provider configured with Account ID %q, got %q", names.AttrAccountID, client.AccountID(ctx), accountID)
		}
	} else {
		accountID = client.AccountID(ctx)
	}

	rd.SetId(accountID)

	return nil
}
