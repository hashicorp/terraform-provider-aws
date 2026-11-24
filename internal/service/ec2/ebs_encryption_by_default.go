// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SingletonIdentity
// @Testing(preIdentityVersion="v6.73.0", generator=false)
// @SDKResource("aws_ebs_encryption_by_default", name="EBS Encryption By Default")
func resourceEBSEncryptionByDefault() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEBSEncryptionByDefaultCreate,
		ReadWithoutTimeout:   resourceEBSEncryptionByDefaultRead,
		UpdateWithoutTimeout: resourceEBSEncryptionByDefaultUpdate,
		DeleteWithoutTimeout: resourceEBSEncryptionByDefaultDelete,

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceEBSEncryptionByDefaultV0().CoreConfigSchema().ImpliedType(),
				Upgrade: ebsEncryptionByDefaultStateUpgradeV0,
				Version: 0,
			},
		},

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

func resourceEBSEncryptionByDefaultCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	enabled := d.Get(names.AttrEnabled).(bool)
	if err := setEBSEncryptionByDefault(ctx, conn, enabled); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EBS encryption by default (%t): %s", enabled, err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))

	return append(diags, resourceEBSEncryptionByDefaultRead(ctx, d, meta)...)
}

func resourceEBSEncryptionByDefaultRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := ec2.GetEbsEncryptionByDefaultInput{}
	resp, err := conn.GetEbsEncryptionByDefault(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS encryption by default: %s", err)
	}

	if err := resourceEBSEncryptionByDefaultFlatten(resp, d); err != nil {
		return sdkdiag.AppendErrorf(diags, "flattening EBS encryption by default: %s", err)
	}

	return diags
}

func resourceEBSEncryptionByDefaultUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	enabled := d.Get(names.AttrEnabled).(bool)
	if err := setEBSEncryptionByDefault(ctx, conn, enabled); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EBS encryption by default (%t): %s", enabled, err)
	}

	return append(diags, resourceEBSEncryptionByDefaultRead(ctx, d, meta)...)
}

func resourceEBSEncryptionByDefaultDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	// Removing the resource disables default encryption.
	if err := setEBSEncryptionByDefault(ctx, conn, false); err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling EBS encryption by default: %s", err)
	}

	return diags
}

func setEBSEncryptionByDefault(ctx context.Context, conn *ec2.Client, enabled bool) error {
	var err error

	if enabled {
		input := ec2.EnableEbsEncryptionByDefaultInput{}
		_, err = conn.EnableEbsEncryptionByDefault(ctx, &input)
	} else {
		input := ec2.DisableEbsEncryptionByDefaultInput{}
		_, err = conn.DisableEbsEncryptionByDefault(ctx, &input)
	}

	return err
}

func resourceEBSEncryptionByDefaultFlatten(resp *ec2.GetEbsEncryptionByDefaultOutput, d *schema.ResourceData) error {
	return d.Set(names.AttrEnabled, resp.EbsEncryptionByDefault)
}

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
	// Upgrade the ID from legacy format to regional format.
	if id, ok := rawState["id"].(string); ok && id != "" {
		// If the ID is not a region, replace it with the current region.
		if !inttypes.IsAWSRegion(id) {
			rawState["id"] = meta.(*conns.AWSClient).Region(ctx)
		}
	}

	return rawState, nil
}
