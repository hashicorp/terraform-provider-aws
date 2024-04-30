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

// @SDKResource("aws_backup_region_settings")
func ResourceRegionSettings() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegionSettingsUpdate,
		UpdateWithoutTimeout: resourceRegionSettingsUpdate,
		ReadWithoutTimeout:   resourceRegionSettingsRead,
		DeleteWithoutTimeout: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"resource_type_management_preference": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeBool},
			},
			"resource_type_opt_in_preference": {
				Type:     schema.TypeMap,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeBool},
			},
		},
	}
}

func resourceRegionSettingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	input := &backup.UpdateRegionSettingsInput{}

	if v, ok := d.GetOk("resource_type_management_preference"); ok && len(v.(map[string]interface{})) > 0 {
		input.ResourceTypeManagementPreference = flex.ExpandBoolValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("resource_type_opt_in_preference"); ok && len(v.(map[string]interface{})) > 0 {
		input.ResourceTypeOptInPreference = flex.ExpandBoolValueMap(v.(map[string]interface{}))
	}

	_, err := conn.UpdateRegionSettings(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Backup Region Settings (%s): %s", d.Id(), err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	return append(diags, resourceRegionSettingsRead(ctx, d, meta)...)
}

func resourceRegionSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	output, err := conn.DescribeRegionSettings(ctx, &backup.DescribeRegionSettingsInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Region Settings (%s): %s", d.Id(), err)
	}

	d.Set("resource_type_opt_in_preference", output.ResourceTypeOptInPreference)
	d.Set("resource_type_management_preference", output.ResourceTypeManagementPreference)

	return diags
}
