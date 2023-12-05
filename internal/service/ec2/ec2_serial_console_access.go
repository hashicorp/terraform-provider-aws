// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_ec2_serial_console_access")
func ResourceSerialConsoleAccess() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSerialConsoleAccessCreate,
		ReadWithoutTimeout:   resourceSerialConsoleAccessRead,
		UpdateWithoutTimeout: resourceSerialConsoleAccessUpdate,
		DeleteWithoutTimeout: resourceSerialConsoleAccessDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceSerialConsoleAccessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	enabled := d.Get("enabled").(bool)
	if err := setSerialConsoleAccess(ctx, conn, enabled); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting EC2 Serial Console Access (%t): %s", enabled, err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	return append(diags, resourceSerialConsoleAccessRead(ctx, d, meta)...)
}

func resourceSerialConsoleAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	output, err := conn.GetSerialConsoleAccessStatusWithContext(ctx, &ec2.GetSerialConsoleAccessStatusInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Serial Console Access: %s", err)
	}

	d.Set("enabled", output.SerialConsoleAccessEnabled)

	return diags
}

func resourceSerialConsoleAccessUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	enabled := d.Get("enabled").(bool)
	if err := setSerialConsoleAccess(ctx, conn, enabled); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EC2 Serial Console Access (%t): %s", enabled, err)
	}

	return append(diags, resourceSerialConsoleAccessRead(ctx, d, meta)...)
}

func resourceSerialConsoleAccessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	// Removing the resource disables serial console access.
	if err := setSerialConsoleAccess(ctx, conn, false); err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling EC2 Serial Console Access: %s", err)
	}

	return diags
}

func setSerialConsoleAccess(ctx context.Context, conn *ec2.EC2, enabled bool) error {
	var err error

	if enabled {
		_, err = conn.EnableSerialConsoleAccessWithContext(ctx, &ec2.EnableSerialConsoleAccessInput{})
	} else {
		_, err = conn.DisableSerialConsoleAccessWithContext(ctx, &ec2.DisableSerialConsoleAccessInput{})
	}

	return err
}
