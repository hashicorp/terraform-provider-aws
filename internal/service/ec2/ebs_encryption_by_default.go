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
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/importer"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @CustomImport
// @SingletonIdentity
// @Testing(preIdentityVersion="v6.72.0", generator=false)
// @SDKResource("aws_ebs_encryption_by_default", name="EBS Encryption By Default")
func resourceEBSEncryptionByDefault() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEBSEncryptionByDefaultCreate,
		ReadWithoutTimeout:   resourceEBSEncryptionByDefaultRead,
		UpdateWithoutTimeout: resourceEBSEncryptionByDefaultUpdate,
		DeleteWithoutTimeout: resourceEBSEncryptionByDefaultDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				if id := d.Id(); id != "" && !inttypes.IsAWSRegion(id) {
					d.SetId(meta.(*conns.AWSClient).Region(ctx))
				}

				if err := importer.Import(ctx, d, meta); err != nil {
					return nil, err
				}

				return []*schema.ResourceData{d}, nil
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
