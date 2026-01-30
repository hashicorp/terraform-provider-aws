// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_serial_console_access", name="Serial Console Access")
// @SingletonIdentity
// @IdentityVersion(1, sdkV2IdentityUpgraders="serialConsoleAccessIdentityUpgradeV0")
// @V60SDKv2Fix
// @Testing(hasExistsFunction=false)
// @Testing(generator=false)
// Generated tests have several issues: (todo: list them)
// @Testing(identityTest=false)
// @Testing(identityVersion="0;v6.0.0")
// @Testing(identityVersion="1;v6.21.0")
func resourceSerialConsoleAccess() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSerialConsoleAccessCreate,
		ReadWithoutTimeout:   resourceSerialConsoleAccessRead,
		UpdateWithoutTimeout: resourceSerialConsoleAccessUpdate,
		DeleteWithoutTimeout: resourceSerialConsoleAccessDelete,

		Schema: map[string]*schema.Schema{
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceSerialConsoleAccessCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	enabled := d.Get(names.AttrEnabled).(bool)
	if err := setSerialConsoleAccess(ctx, conn, enabled); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting EC2 Serial Console Access (%t): %s", enabled, err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))

	return append(diags, resourceSerialConsoleAccessRead(ctx, d, meta)...)
}

func resourceSerialConsoleAccessRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	output, err := findSerialConsoleAccessStatus(ctx, conn)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] EC2 Serial Console Access %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Serial Console Access: %s", err)
	}

	d.Set(names.AttrEnabled, output.SerialConsoleAccessEnabled)

	return diags
}

func resourceSerialConsoleAccessUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	enabled := d.Get(names.AttrEnabled).(bool)
	if err := setSerialConsoleAccess(ctx, conn, enabled); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EC2 Serial Console Access (%t): %s", enabled, err)
	}

	return append(diags, resourceSerialConsoleAccessRead(ctx, d, meta)...)
}

func resourceSerialConsoleAccessDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	// Removing the resource disables serial console access.
	if err := setSerialConsoleAccess(ctx, conn, false); err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling EC2 Serial Console Access: %s", err)
	}

	return diags
}

func setSerialConsoleAccess(ctx context.Context, conn *ec2.Client, enabled bool) error {
	var err error

	if enabled {
		var input ec2.EnableSerialConsoleAccessInput
		_, err = conn.EnableSerialConsoleAccess(ctx, &input)
	} else {
		var input ec2.DisableSerialConsoleAccessInput
		_, err = conn.DisableSerialConsoleAccess(ctx, &input)
	}

	return err
}

var serialConsoleAccessIdentityUpgradeV0 = schema.IdentityUpgrader{
	Version: 0,
	Upgrade: func(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
		rawState[names.AttrRegion] = meta.(*conns.AWSClient).Region(ctx)
		return rawState, nil
	},
}
