// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_backup_global_settings", name="Global Settings")
func resourceGlobalSettings() *schema.Resource {
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
		return sdkdiag.AppendErrorf(diags, "updating Backup Global Settings: %s", err)
	}

	if d.IsNewResource() {
		d.SetId(meta.(*conns.AWSClient).AccountID)
	}

	return append(diags, resourceGlobalSettingsRead(ctx, d, meta)...)
}

func resourceGlobalSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	output, err := findGlobalSettings(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Backup Global Settings (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Global Settings (%s): %s", d.Id(), err)
	}

	if err := d.Set("global_settings", output); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting global_settings: %s", err)
	}

	return diags
}

func findGlobalSettings(ctx context.Context, conn *backup.Client) (map[string]string, error) {
	input := &backup.DescribeGlobalSettingsInput{}
	output, err := conn.DescribeGlobalSettings(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.GlobalSettings == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.GlobalSettings, nil
}
