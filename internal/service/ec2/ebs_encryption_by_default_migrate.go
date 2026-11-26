// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// resourceEBSEncryptionByDefaultV0 returns the previous schema (v0) for state upgrading.
func resourceEBSEncryptionByDefaultV0() *schema.Resource {
	return &schema.Resource{
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrEnabled: {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
			}
		},
	}
}

// ebsEncryptionByDefaultStateUpgradeV0 upgrades the state from v0 (pre-identity) to v1 (identity).
// It converts the legacy ID (generated UUID or "default") to the regional identity format.
func ebsEncryptionByDefaultStateUpgradeV0(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if rawState == nil {
		rawState = map[string]any{}
	}

	rawState[names.AttrID] = meta.(*conns.AWSClient).Region(ctx)

	return rawState, nil
}
