// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

func GlobalSingleParameterized(ctx context.Context, rd *schema.ResourceData, identitySpec *inttypes.Identity, client AWSClient) error {
	if rd.Id() != "" {
		importID := rd.Id()
		if identitySpec.IdentityAttribute != "id" {
			rd.Set(identitySpec.IdentityAttribute, importID)
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

	attr := identitySpec.Attributes[1]

	valRaw, ok := identity.GetOk(attr.Name)
	if attr.Required && !ok {
		return fmt.Errorf("identity attribute %q is required", attr.Name)
	}
	val, ok := valRaw.(string)
	if !ok {
		return fmt.Errorf("identity attribute %q: expected string, got %T", attr.Name, valRaw)
	}
	setAttribute(rd, attr.Name, val)

	return nil
}
