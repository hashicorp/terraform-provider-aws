// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ebs_encryption_by_default", name="EBS Encryption By Default")
func resourceEBSEncryptionByDefault() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEBSEncryptionByDefaultCreate,
		ReadWithoutTimeout:   resourceEBSEncryptionByDefaultRead,
		UpdateWithoutTimeout: resourceEBSEncryptionByDefaultUpdate,
		DeleteWithoutTimeout: resourceEBSEncryptionByDefaultDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceEBSEncryptionByDefaultCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	enabled := d.Get(names.AttrEnabled).(bool)
	if err := setEBSEncryptionByDefault(ctx, conn, enabled); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EBS encryption by default (%t): %s", enabled, err)
	}

	//lintignore:R015 // Allow legacy unstable ID usage in managed resource
	d.SetId(id.UniqueId())

	return append(diags, resourceEBSEncryptionByDefaultRead(ctx, d, meta)...)
}

func resourceEBSEncryptionByDefaultRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	resp, err := conn.GetEbsEncryptionByDefault(ctx, &ec2.GetEbsEncryptionByDefaultInput{})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS encryption by default: %s", err)
	}

	d.Set(names.AttrEnabled, resp.EbsEncryptionByDefault)

	return diags
}

func resourceEBSEncryptionByDefaultUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	enabled := d.Get(names.AttrEnabled).(bool)
	if err := setEBSEncryptionByDefault(ctx, conn, enabled); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EBS encryption by default (%t): %s", enabled, err)
	}

	return append(diags, resourceEBSEncryptionByDefaultRead(ctx, d, meta)...)
}

func resourceEBSEncryptionByDefaultDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		_, err = conn.EnableEbsEncryptionByDefault(ctx, &ec2.EnableEbsEncryptionByDefaultInput{})
	} else {
		_, err = conn.DisableEbsEncryptionByDefault(ctx, &ec2.DisableEbsEncryptionByDefaultInput{})
	}

	return err
}
