// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/importer"
)

func arnIdentityResourceImporter(attrName string, isGlobal bool) *schema.ResourceImporter {
	if isGlobal {
		return &schema.ResourceImporter{
			StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				if err := importer.GlobalARN(ctx, rd, attrName); err != nil {
					return nil, err
				}

				return []*schema.ResourceData{rd}, nil
			},
		}
	} else {
		return &schema.ResourceImporter{
			StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				if err := importer.RegionalARN(ctx, rd, attrName); err != nil {
					return nil, err
				}

				return []*schema.ResourceData{rd}, nil
			},
		}
	}
}

func singletonIdentityResourceImporter(isGlobal bool) *schema.ResourceImporter {
	if isGlobal {
		// Historically, we haven't validated *any* Import ID value for Global Singletons
		return &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		}
	} else {
		return &schema.ResourceImporter{
			StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				if err := importer.RegionalSingleton(ctx, rd); err != nil {
					return nil, err
				}

				return []*schema.ResourceData{rd}, nil
			},
		}
	}
}
