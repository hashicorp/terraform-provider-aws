// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_quicksight_theme", name="Theme")
func dataSourceTheme() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceThemeRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrAWSAccountID: {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ValidateFunc: verify.ValidAccountID,
				},
				"base_theme_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrConfiguration: quicksightschema.ThemeConfigurationDataSourceSchema(),
				names.AttrCreatedTime: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"theme_id": {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrLastUpdatedTime: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrPermissions: quicksightschema.PermissionsDataSourceSchema(),
				names.AttrStatus: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags: tftags.TagsSchemaComputed(),
				"version_description": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"version_number": {
					Type:     schema.TypeInt,
					Computed: true,
				},
			}
		},
	}
}

func dataSourceThemeRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	themeID := d.Get("theme_id").(string)
	id := themeCreateResourceID(awsAccountID, themeID)

	theme, err := findThemeByTwoPartKey(ctx, conn, awsAccountID, themeID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Theme (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set(names.AttrARN, theme.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountID)
	d.Set("base_theme_id", theme.Version.BaseThemeId)
	if err := d.Set(names.AttrConfiguration, quicksightschema.FlattenThemeConfiguration(theme.Version.Configuration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting configuration: %s", err)
	}
	d.Set(names.AttrCreatedTime, theme.CreatedTime.Format(time.RFC3339))
	d.Set(names.AttrLastUpdatedTime, theme.LastUpdatedTime.Format(time.RFC3339))
	d.Set(names.AttrName, theme.Name)
	d.Set(names.AttrStatus, theme.Version.Status)
	d.Set("theme_id", theme.ThemeId)
	d.Set("version_description", theme.Version.Description)
	d.Set("version_number", theme.Version.VersionNumber)

	permissions, err := findThemePermissionsByTwoPartKey(ctx, conn, awsAccountID, themeID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Theme (%s) permissions: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPermissions, quicksightschema.FlattenPermissions(permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions: %s", err)
	}

	return diags
}
