// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKResource("aws_backup_global_settings")
func ResourceGlobalSettings() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGlobalSettingsUpdate,
		UpdateWithoutTimeout: resourceGlobalSettingsUpdate,
		ReadWithoutTimeout:   resourceGlobalSettingsRead,
		DeleteWithoutTimeout: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"global_settings": {
				Type:     schema.TypeMap,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceGlobalSettingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	input := &backup.UpdateGlobalSettingsInput{
		GlobalSettings: flex.ExpandStringValueMap(d.Get("global_settings").(map[string]interface{})),
	}

	_, err := conn.UpdateGlobalSettings(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Backup Global Settings (%s): %s", meta.(*conns.AWSClient).AccountID, err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)

	return append(diags, resourceGlobalSettingsRead(ctx, d, meta)...)
}

func resourceGlobalSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	resp, err := conn.DescribeGlobalSettings(ctx, &backup.DescribeGlobalSettingsInput{})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Global Settings (%s): %s", d.Id(), err)
	}

	if err := d.Set("global_settings", resp.GlobalSettings); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting global_settings: %s", err)
	}

	return diags
}
