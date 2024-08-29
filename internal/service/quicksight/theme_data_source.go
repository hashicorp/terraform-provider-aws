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
				names.AttrConfiguration: { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ThemeConfiguration.html
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_color_palette": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataColorPalette.html
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"colors": {
											Type:     schema.TypeList,
											Computed: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										"empty_fill_color": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"min_max_gradient": {
											Type:     schema.TypeList,
											Computed: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
									},
								},
							},
							"sheet": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetStyle.html
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"tile": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TileStyle.html
											Type:     schema.TypeList,
											Computed: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"border": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BorderStyle.html
														Type:     schema.TypeList,
														Computed: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"show": {
																	Type:     schema.TypeBool,
																	Computed: true,
																},
															},
														},
													},
												},
											},
										},
										"tile_layout": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TileLayoutStyle.html
											Type:     schema.TypeList,
											Computed: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"gutter": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GutterStyle.html
														Type:     schema.TypeList,
														Computed: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"show": {
																	Type:     schema.TypeBool,
																	Computed: true,
																},
															},
														},
													},
													"margin": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MarginStyle.html
														Type:     schema.TypeList,
														Computed: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"show": {
																	Type:     schema.TypeBool,
																	Computed: true,
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							"typography": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Typography.html
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"font_families": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Font.html
											Type:     schema.TypeList,
											Computed: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"font_family": {
														Type:     schema.TypeString,
														Computed: true,
													},
												},
											},
										},
									},
								},
							},
							"ui_color_palette": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_UIColorPalette.html
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"accent": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"accent_foreground": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"danger": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"danger_foreground": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"dimension": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"dimension_foreground": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"measure": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"measure_foreground": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"primary_background": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"primary_foreground": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"secondary_background": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"secondary_foreground": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"success": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"success_foreground": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"warning": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"warning_foreground": {
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},
						},
					},
				},
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
				names.AttrPermissions: {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrActions: {
								Type:     schema.TypeSet,
								Computed: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							names.AttrPrincipal: {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
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

func dataSourceThemeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID
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
	if err := d.Set(names.AttrConfiguration, flattenThemeConfiguration(theme.Version.Configuration)); err != nil {
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
